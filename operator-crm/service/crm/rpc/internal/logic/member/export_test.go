package memberlogic

import (
	"context"
	"errors"
	"io"
	"strconv"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"oa.98ent.com/p9/common/modules/fieldcrypto"
	"oa.98ent.com/p9/operator-crm/ent"
	contract "oa.98ent.com/p9/operator-crm/pkg/memberexport"
	dao "oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/dao/member"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"
	pb "oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/memberexportpb"
)

func TestParseExportIncludesDemoSelection(t *testing.T) {
	for _, includeDemo := range []bool{false, true} {
		filters, err := parseExport(&pb.ExportRequest{IncludeDemo: includeDemo, AgentType: 1})
		if err != nil || filters.IncludeDemo != includeDemo || filters.AgentType != 1 {
			t.Fatalf("demo or agent type selection lost: %v", err)
		}
	}
}

func fixtureExportMember(id int64) *ent.Member {
	username := "MiXeD"
	return &ent.Member{ID: id, Code: "123456789012", CreatedAt: time.UnixMilli(1700000000123), AgentID: 1, Edges: ent.MemberEdges{Auth: &ent.MemberAuth{Username: &username}, Agent: &ent.AgentUser{Account: "agent"}, AutoLevel: &ent.MemberAutoLevel{Name: "L1"}, VipLevel: &ent.MemberVipLevel{Level: 0}}}
}

type fakeExportSource struct {
	n                  int
	probeErr, batchErr error
	calls              int
	mutate             func(*ent.Member)
}

func (f *fakeExportSource) ExportIDs(context.Context, dao.ExportFilters) ([]int64, error) {
	f.calls++
	ids := make([]int64, f.n)
	for i := range ids {
		ids[i] = int64(f.n - i)
	}
	return ids, f.probeErr
}
func (f *fakeExportSource) ExportBatch(_ context.Context, ids []int64) ([]*ent.Member, error) {
	if f.batchErr != nil {
		return nil, f.batchErr
	}
	rows := make([]*ent.Member, len(ids))
	for i, id := range ids {
		rows[i] = fixtureExportMember(id)
		if f.mutate != nil {
			f.mutate(rows[i])
		}
	}
	return rows, nil
}
func TestExportBoundariesAndIdenticalTimes(t *testing.T) {
	for _, n := range []int{0, 101, contract.MaxRows, contract.MaxRows + 1} {
		t.Run(strconv.Itoa(n), func(t *testing.T) {
			source := &fakeExportSource{n: n}
			var complete bool
			var rows int64
			err := runExport(context.Background(), source, dao.ExportFilters{}, contract.Policy{}, nil, func(c *pb.ExportChunk) error { complete = c.Complete; rows = c.Rows; return nil })
			if n > contract.MaxRows {
				if status.Code(err) != codes.OutOfRange || complete {
					t.Fatal("over limit delivered")
				}
			} else if err != nil || !complete || rows != int64(n) {
				t.Fatalf("rows=%d complete=%v err=%v", rows, complete, err)
			}
			if source.calls != 1 {
				t.Fatal("repeated membership probe")
			}
		})
	}
}
func TestExportFailuresNeverComplete(t *testing.T) {
	for _, source := range []*fakeExportSource{{n: 1, probeErr: errors.New("SQL sensitive filter")}, {n: 1, batchErr: errors.New("wallet/stat query failed")}, {n: 1, mutate: func(m *ent.Member) { m.Edges.Auth = nil }}, {n: 1, mutate: func(m *ent.Member) { m.ID = 999 }}} {
		completed := false
		err := runExport(context.Background(), source, dao.ExportFilters{}, contract.Policy{}, nil, func(c *pb.ExportChunk) error { completed = c.Complete; return nil })
		if err == nil || completed || strings.Contains(err.Error(), "sensitive filter") {
			t.Fatal("failure yielded file or leaked predicate")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := runExport(ctx, &fakeExportSource{n: 1}, dao.ExportFilters{}, contract.Policy{}, nil, func(*pb.ExportChunk) error { return nil }); status.Code(err) != codes.Canceled {
		t.Fatal("cancel ignored")
	}
	ctx, cancel = context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()
	if err := runExport(ctx, &fakeExportSource{n: 1}, dao.ExportFilters{}, contract.Policy{}, nil, func(*pb.ExportChunk) error { return nil }); status.Code(err) != codes.DeadlineExceeded {
		t.Fatal("timeout ignored")
	}
	if err := runExport(context.Background(), &fakeExportSource{}, dao.ExportFilters{}, contract.Policy{}, nil, func(*pb.ExportChunk) error { return io.ErrClosedPipe }); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatal("send error ignored")
	}
}
func TestExportContactsAndExactAmounts(t *testing.T) {
	cryptor, err := fieldcrypto.New([]byte("0123456789abcdef"), []byte("key"))
	if err != nil {
		t.Fatal(err)
	}
	phone, _ := cryptor.Seal("0013800000000")
	email, _ := cryptor.Seal("name@example.com")
	m := fixtureExportMember(9223372036854775807)
	m.Edges.Auth.PhoneCiphertext = &phone
	m.Edges.Auth.EmailCiphertext = &email
	m.Edges.Stats = &ent.MemberStats{TotalDeposit: 9223372036854775807, TotalWithdrawal: 9007199254740993}
	row, err := exportRow(m, contract.Policy{Phone: true, Email: true}, cryptor)
	if err != nil || row[4] != "'0013800000000" || row[5] != "name@example.com" || row[16] != "9223372036854775807" || row[17] != "9007199254740993" {
		t.Fatalf("precision/contact error: %v", err)
	}
	row, err = exportRow(m, contract.Policy{}, nil)
	if err != nil || row[4] != "[REDACTED]" || row[5] != "[REDACTED]" {
		t.Fatal("mask policy failed")
	}
	corrupt := "corrupt-ciphertext"
	m.Edges.Auth.PhoneCiphertext = &corrupt
	if _, err = exportRow(m, contract.Policy{Phone: true}, cryptor); err == nil {
		t.Fatal("decryption failure ignored")
	}
	m.Edges.Auth.PhoneCiphertext = nil
	m.Edges.Stats = nil
	row, err = exportRow(m, contract.Policy{}, nil)
	if err != nil || row[16] != contract.Missing || row[18] != contract.Unavailable {
		t.Fatal("missing aggregate converted to zero")
	}
	m.Edges.Stats = &ent.MemberStats{TotalDeposit: -1}
	if _, err = exportRow(m, contract.Policy{}, nil); err == nil {
		t.Fatal("negative funding accepted")
	}
}
func TestExportRejectsListPagingAndSorting(t *testing.T) {
	for _, f := range []*memberpb.ListMembersRequest{{Page: 1}, {PageSize: 100}, {SortBy: 1}, {SortOrder: 1}, {TimeType: 99}, {Deposited: true, NotDeposited: true}} {
		if _, err := parseExport(&pb.ExportRequest{Filters: f}); status.Code(err) != codes.InvalidArgument {
			t.Fatal("invalid list option silently ignored")
		}
	}
}
