package memberlogic

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"oa.98ent.com/p9/common/modules/fieldcrypto"
	"oa.98ent.com/p9/operator-crm/ent"
	contract "oa.98ent.com/p9/operator-crm/pkg/memberexport"
	dao "oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/dao/member"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"
	pb "oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/memberexportpb"
)

var exportSlots = make(chan struct{}, 2)

func parseExport(in *pb.ExportRequest) (dao.ExportFilters, error) {
	var out dao.ExportFilters
	if in == nil {
		return out, status.Error(codes.InvalidArgument, "missing export request")
	}
	filters := in.Filters
	if filters == nil {
		filters = &memberpb.ListMembersRequest{}
	}
	// Export order is explicitly ID DESC. List sort/paging cannot be silently ignored.
	if filters.Page != 0 || filters.PageSize != 0 || filters.SortBy != 0 || filters.SortOrder != 0 {
		return out, status.Error(codes.InvalidArgument, "export does not accept paging or list sorting")
	}
	filters = proto.Clone(filters).(*memberpb.ListMembersRequest)
	filters.Page = 1
	filters.PageSize = 100
	params, err := parseListParams(filters)
	if err != nil {
		return out, status.Error(codes.InvalidArgument, "invalid member filters")
	}
	return dao.ExportFilters{List: params, IP: in.FilterIp, Device: in.FilterDevice, ParentUsername: in.PUsername, AgentType: in.AgentType, IncludeDemo: in.IncludeDemo}, nil
}

// StreamMembers is synchronous and bounded even for direct internal RPC callers.
func StreamMembers(ctx context.Context, s *svc.ServiceContext, in *pb.ExportRequest, send func(context.Context, *pb.ExportChunk) error) error {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	ctx, policy, err := contract.Authorize(ctx, s.ExportAuth)
	if err != nil {
		if ctx.Err() != nil {
			return status.FromContextError(ctx.Err()).Err()
		}
		return err
	}
	filters, err := parseExport(in)
	if err != nil {
		return err
	}
	select {
	case exportSlots <- struct{}{}:
		defer func() { <-exportSlots }()
	default:
		return status.Error(codes.ResourceExhausted, "export service busy")
	}
	// Validate before accessing DB, including all extra filters.
	if err := dao.ValidateExportFilters(filters); err != nil {
		return status.Error(codes.InvalidArgument, "invalid export filters")
	}
	return runExport(ctx, s.Member, filters, policy, s.Cryptor, func(chunk *pb.ExportChunk) error { return send(ctx, chunk) })
}

type exportSource interface {
	ExportIDs(context.Context, dao.ExportFilters) ([]int64, error)
	ExportBatch(context.Context, []int64) ([]*ent.Member, error)
}

func runExport(ctx context.Context, source exportSource, filters dao.ExportFilters, policy contract.Policy, cryptor *fieldcrypto.Cryptor, send func(*pb.ExportChunk) error) error {
	started := time.Now().UnixMilli()
	queryCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	ids, err := source.ExportIDs(queryCtx, filters)
	cancel()
	if err != nil {
		return exportQueryError(ctx, err)
	}
	if len(ids) > contract.MaxRows {
		return status.Error(codes.OutOfRange, "export exceeds 5000 members")
	}
	// Check the probe too: duplicates or broken ordering must never yield a file.
	for i, id := range ids {
		if id <= 0 || (i > 0 && id >= ids[i-1]) {
			return status.Error(codes.DataLoss, "invalid export membership")
		}
	}
	var rows int64
	var size int
	emit := func(records [][]string) error {
		if err := ctx.Err(); err != nil {
			return status.FromContextError(err).Err()
		}
		var b bytes.Buffer
		w := csv.NewWriter(&b)
		w.UseCRLF = true
		if err := w.WriteAll(records); err != nil {
			return status.Error(codes.Internal, "CSV write failed")
		}
		size += b.Len()
		if size > contract.MaxBytes {
			return status.Error(codes.ResourceExhausted, "CSV exceeds size limit")
		}
		return send(&pb.ExportChunk{Csv: b.Bytes(), Rows: rows, StartedAt: started})
	}
	if err = emit([][]string{contract.Header}); err != nil {
		return err
	}
	for offset := 0; offset < len(ids); offset += dao.ExportBatchSize {
		end := min(offset+dao.ExportBatchSize, len(ids))
		batch := ids[offset:end]
		queryCtx, cancel = context.WithTimeout(ctx, 5*time.Second)
		members, e := source.ExportBatch(queryCtx, batch)
		cancel()
		if e != nil {
			return exportQueryError(ctx, e)
		}
		if len(members) != len(batch) {
			return status.Error(codes.DataLoss, "export membership changed")
		}
		records := make([][]string, 0, len(members))
		for i, m := range members {
			if m == nil || m.ID != batch[i] {
				return status.Error(codes.DataLoss, "export membership mismatch")
			}
			row, e := exportRow(m, policy, cryptor)
			if e != nil {
				return status.Error(codes.DataLoss, e.Error())
			}
			records = append(records, row)
		}
		rows += int64(len(records))
		if err = emit(records); err != nil {
			return err
		}
	}
	if rows != int64(len(ids)) {
		return status.Error(codes.DataLoss, "export row count mismatch")
	}
	if err = ctx.Err(); err != nil {
		return status.FromContextError(err).Err()
	}
	return send(&pb.ExportChunk{Rows: rows, Complete: true, StartedAt: started})
}

func exportQueryError(ctx context.Context, err error) error {
	if ctx.Err() != nil {
		return status.FromContextError(ctx.Err()).Err()
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return status.FromContextError(err).Err()
	}
	// Never include database predicates or encrypted values in an RPC error.
	return status.Error(codes.Unavailable, "member export query failed")
}
