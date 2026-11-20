package manage

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	contract "oa.98ent.com/p9/operator-crm/pkg/memberexport"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"
	pb "oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/memberexportpb"
)

type ExportLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewExportLogic(ctx context.Context, s *svc.ServiceContext) *ExportLogic {
	return &ExportLogic{logx.WithContext(ctx), ctx, s}
}

var apiExportSlots = make(chan struct{}, 2)

func (l *ExportLogic) Export(req *types.ExportMembersRequest) (resp *types.ExportMembersResponse, err error) {
	ctx, cancel := context.WithTimeout(l.ctx, 40*time.Second)
	defer cancel()
	started := time.Now()
	task := uuid.NewString()
	var rows int64
	defer func() {
		if err != nil && ctx.Err() != nil {
			err = exportHTTPError(ctx, ctx.Err())
			resp = nil
		}
		l.Infow("member export finished", logx.Field("task_id", task), logx.Field("rows", rows), logx.Field("elapsed_ms", time.Since(started).Milliseconds()), logx.Field("error_class", exportErrorClass(err)))
	}()
	claims, ok := ctxdata.OperatorClaimsFromCtx(ctx)
	if !ok || claims.IsPlatform || claims.TokenType != ctxdata.TokenTypeAccess || ctxdata.RawTokenFromCtx(ctx) == "" {
		return nil, xerr.Forbidden("common.forbidden")
	}
	if req == nil {
		return nil, xerr.BadRequest("missing export request")
	}
	if req.RangeTodayDeposit != nil || req.RangeTodayWithdraw != nil {
		return nil, xerr.BadRequest("today amount filters are unavailable: no daily statistics source")
	}
	if len(req.Keyword) > 100 {
		return nil, xerr.BadRequest("keyword too long")
	}
	select {
	case apiExportSlots <- struct{}{}:
		defer func() { <-apiExportSlots }()
	default:
		return nil, xerr.Err(429, "member export service busy")
	}
	indices, e := contract.ColumnIndices(req.Columns)
	if e != nil {
		return nil, xerr.BadRequest(e.Error())
	}
	if l.svcCtx.ExportRedis == nil || l.svcCtx.MemberExportRpc == nil || l.svcCtx.Config.OssApiBaseURL == "" {
		return nil, xerr.Err(503, "private export dependencies unavailable")
	}
	lock := exportLock{redis: l.svcCtx.ExportRedis, key: "member:export:v2:" + claims.OperatorCode, token: task}
	if e = lock.acquire(ctx); e != nil {
		return nil, e
	}
	defer func() {
		if e := lock.release(); e != nil {
			l.Errorw("member export lock release failed", logx.Field("task_id", task))
		}
	}()
	file, e := os.CreateTemp("", "member-export-*.csv")
	if e != nil {
		return nil, xerr.InternalServerError("export temporary file failed")
	}
	defer func(file *os.File) {
		file.Close()
		if e := os.Remove(file.Name()); e != nil {
			l.Errorw("member export temporary cleanup failed", logx.Field("task_id", task))
		}
	}(file)
	// A short RPC child deadline leaves time for verification and upload.
	queryCtx, queryCancel := context.WithTimeout(ctx, 22*time.Second)
	stream, e := l.svcCtx.MemberExportRpc.Export(queryCtx, exportRPCRequest(req))
	if e != nil {
		queryCancel()
		return nil, exportHTTPError(ctx, e)
	}
	var complete bool
	var size int64
	var sourceStarted int64
	for {
		chunk, e := stream.Recv()
		if e == io.EOF {
			break
		}
		if e != nil {
			queryCancel()
			return nil, exportHTTPError(ctx, e)
		}
		if complete || chunk.Rows < rows || chunk.Rows > contract.MaxRows || chunk.StartedAt <= 0 || (sourceStarted != 0 && sourceStarted != chunk.StartedAt) {
			queryCancel()
			return nil, xerr.InternalServerError("invalid export stream")
		}
		sourceStarted = chunk.StartedAt
		rows = chunk.Rows
		size += int64(len(chunk.Csv))
		if size > contract.MaxBytes {
			queryCancel()
			return nil, xerr.Err(413, "CSV exceeds size limit")
		}
		if chunk.Complete {
			if len(chunk.Csv) != 0 {
				queryCancel()
				return nil, xerr.InternalServerError("invalid export trailer")
			}
			complete = true
		} else if _, e = file.Write(chunk.Csv); e != nil {
			queryCancel()
			return nil, xerr.InternalServerError("export file write failed")
		}
	}
	queryCancel()
	if !complete {
		return nil, xerr.InternalServerError("incomplete export stream")
	}
	if e = verifyCSV(ctx, file, rows); e != nil {
		return nil, e
	}
	if e = lock.check(ctx); e != nil {
		return nil, e
	}

	if len(req.Columns) > 0 {
		projected, err := projectCSV(ctx, file, indices, rows)
		if err != nil {
			return nil, err
		}
		defer func() {
			projected.Close()
			if e := os.Remove(projected.Name()); e != nil {
				l.Errorw("member export projection cleanup failed", logx.Field("task_id", task))
			}
		}()
		file = projected
	}
	stat, e := file.Stat()
	if e != nil {
		return nil, xerr.InternalServerError("export file stat failed")
	}
	size = stat.Size()
	uploadCtx, uploadCancel := context.WithTimeout(ctx, 12*time.Second)
	result, e := uploadExportToOSS(uploadCtx, l.svcCtx.Config.OssApiBaseURL, claims.OperatorCode, task+".csv", file, size)
	uploadCancel()
	if e != nil {
		return nil, exportHTTPError(ctx, e)
	}
	if e = lock.check(ctx); e != nil {
		return nil, e
	}
	if ctx.Err() != nil {
		return nil, exportHTTPError(ctx, ctx.Err())
	}
	return &types.ExportMembersResponse{TaskId: task, Url: result.URL, Key: result.Key, Rows: rows, StartedAt: sourceStarted, ExpiresAt: time.Now().Add(time.Duration(result.URLExpiresIn) * time.Second).UnixMilli(), Visibility: result.Visibility}, nil

}

// exportRPCRequest 映射已校验的筛选条件，保留 CRM 的独立租户范围校验。
func exportRPCRequest(request *types.ExportMembersRequest) *pb.ExportRequest {
	return &pb.ExportRequest{
		FilterIp:     request.FilterIp,
		FilterDevice: request.FilterDevice,
		PUsername:    request.PUsername,
		AgentType:    request.AgentType,
		IncludeDemo:  request.IncludeDemo,
		Filters: &memberpb.ListMembersRequest{
			Keyword:         request.Keyword,
			TimeType:        memberpb.MemberTimeType(request.TimeType),
			StartAt:         request.StartAt,
			EndAt:           request.EndAt,
			AutoLevelId:     request.AutoLevelId,
			FixedLevelId:    request.FixedLevelId,
			RegisterMethod:  request.RegisterMethod,
			RegisterChannel: request.RegisterChannel,
			Status:          request.Status,
			AgentId:         request.AgentId,
			BettingDisabled: request.BettingDisabled,
			NoPhone:         request.NoPhone,
			NoEmail:         request.NoEmail,
			Deposited:       request.Deposited,
			NotDeposited:    request.NotDeposited,
		},
	}
}

// exportErrorClass 将错误归类为稳定日志字段，不记录请求内容或签名链接。
func exportErrorClass(err error) string {
	if err == nil {
		return "success"
	}
	if errors.Is(err, context.Canceled) {
		return "canceled"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "timeout"
	}
	var businessError *xerr.Error
	if errors.As(err, &businessError) {
		return strconv.Itoa(businessError.Status)
	}
	return "dependency"
}

// 标准库没有定义 499；用于标识调用方取消请求。
const memberExportStatusClientClosedRequest = 499

// exportHTTPError 优先处理请求取消与超时，再将 RPC 错误映射为 HTTP 状态。
// 已明确分类的业务错误保持原语义。
func exportHTTPError(ctx context.Context, err error) error {
	if ctx.Err() != nil {
		err = ctx.Err()
	}
	var businessError *xerr.Error
	if errors.As(err, &businessError) {
		return businessError
	}
	if errors.Is(err, context.Canceled) || status.Code(err) == codes.Canceled {
		return xerr.Err(memberExportStatusClientClosedRequest, "member export canceled")
	}
	if errors.Is(err, context.DeadlineExceeded) || status.Code(err) == codes.DeadlineExceeded {
		return xerr.Err(http.StatusGatewayTimeout, "member export timed out")
	}
	switch status.Code(err) {
	case codes.InvalidArgument:
		return xerr.BadRequest(status.Convert(err).Message())
	case codes.Unauthenticated:
		return xerr.Err(http.StatusUnauthorized, "export authentication failed")
	case codes.PermissionDenied:
		return xerr.Forbidden("common.forbidden")
	case codes.OutOfRange:
		return xerr.Err(http.StatusRequestEntityTooLarge, "member export exceeds 5000 rows")
	case codes.ResourceExhausted:
		return xerr.Err(http.StatusTooManyRequests, "member export service busy")
	case codes.Unavailable:
		return xerr.Err(http.StatusServiceUnavailable, "member export dependency failed")
	case codes.DataLoss:
		return xerr.InternalServerError("member export data validation failed")
	default:
		return xerr.Err(http.StatusBadGateway, "member export dependency failed")
	}
}
