package manage

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-admin-api/internal/config"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	contract "oa.98ent.com/p9/operator-crm/pkg/memberexport"
	pb "oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/memberexportpb"
)

func TestExportFileVerificationAndProjection(t *testing.T) {
	f, e := os.CreateTemp(t.TempDir(), "source.csv")
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	w.Write(contract.Header)
	row := make([]string, len(contract.Header))
	for i := range row {
		row[i] = contract.Missing
	}
	row[0] = "'9223372036854775807"
	row[1] = "'123456789012"
	row[2] = "Alice"
	row[3] = "comma, quote\"\nnewline"
	w.Write(row)
	w.Flush()
	if e = verifyCSV(context.Background(), f, 1); e != nil {
		t.Fatal(e)
	}
	if e = verifyCSV(context.Background(), f, 0); e == nil {
		t.Fatal("count mismatch accepted")
	}
	indices, e := contract.ColumnIndices([]string{"nickname", "username"})
	if e != nil {
		t.Fatal(e)
	}
	projected, e := projectCSV(context.Background(), f, indices, 1)
	if e != nil {
		t.Fatal(e)
	}
	defer func() { projected.Close(); os.Remove(projected.Name()) }()
	records, e := csv.NewReader(projected).ReadAll()
	if e != nil || len(records) != 2 || records[0][0] != "nickname" || records[1][0] != row[3] || records[1][1] != "Alice" {
		t.Fatal("projection changed column order or content")
	}
	f.Seek(0, io.SeekEnd)
	w = csv.NewWriter(f)
	row[0] = "9223372036854775807"
	w.Write(row)
	w.Flush()
	if e = verifyCSV(context.Background(), f, 2); e == nil {
		t.Fatal("duplicate numeric member ID accepted")
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if e = verifyCSV(canceled, f, 2); e == nil {
		t.Fatal("cancel ignored")
	}
}

func redisForExportTest(t *testing.T) redis.UniversalClient {
	t.Helper()
	addr := os.Getenv("EXPORT_TEST_REDIS_ADDR")
	if addr == "" {
		addr = miniredis.RunT(t).Addr()
	}
	client := redis.NewClient(&redis.Options{Addr: addr, MaxRetries: -1, DialTimeout: time.Second, ReadTimeout: time.Second, WriteTimeout: time.Second})
	t.Cleanup(func() { client.Close() })
	return client
}
func TestExportLockOwnership(t *testing.T) {
	rdb := redisForExportTest(t)
	ctx := context.Background()
	lock := exportLock{redis: rdb, key: "test:member-export:" + uuid.NewString(), token: "first"}
	defer rdb.Del(ctx, lock.key)
	if err := lock.acquire(ctx); err != nil {
		t.Fatal(err)
	}
	ttl := rdb.PTTL(ctx, lock.key).Val()
	if ttl <= 40*time.Second || ttl > exportLockTTL {
		t.Fatal("unsafe lock TTL")
	}
	if err := lock.acquire(ctx); httpStatus(err) != 409 {
		t.Fatal("conflict not distinguished")
	}
	rdb.Set(ctx, lock.key, "second", exportLockTTL)
	if err := lock.check(ctx); httpStatus(err) != 409 {
		t.Fatal("lost ownership accepted")
	}
	if err := lock.release(); err == nil {
		t.Fatal("stale release ignored")
	}
	if rdb.Get(ctx, lock.key).Val() != "second" {
		t.Fatal("stale owner deleted new lock")
	}
	lock.token = "second"
	if err := lock.release(); err != nil {
		t.Fatal(err)
	}
	if rdb.Exists(ctx, lock.key).Val() != 0 {
		t.Fatal("lock leaked")
	}
}
func TestExportRedisFailure(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1", MaxRetries: -1, DialTimeout: 50 * time.Millisecond})
	defer rdb.Close()
	lock := exportLock{redis: rdb, key: "unreachable", token: "x"}
	if e := lock.acquire(context.Background()); httpStatus(e) != 503 {
		t.Fatal("redis failure disguised as conflict")
	}
}
func httpStatus(err error) int {
	var e *xerr.Error
	if errors.As(err, &e) {
		return e.Status
	}
	return 0
}

type testExportClient struct {
	err    error
	chunks []*pb.ExportChunk
}

func (c testExportClient) Export(context.Context, *pb.ExportRequest, ...grpc.CallOption) (grpc.ServerStreamingClient[pb.ExportChunk], error) {
	return &testExportStream{err: c.err, chunks: c.chunks}, nil
}

type testExportStream struct {
	grpc.ClientStream
	err    error
	chunks []*pb.ExportChunk
}

func (s *testExportStream) Recv() (*pb.ExportChunk, error) {
	if s.err != nil {
		return nil, s.err
	}
	if len(s.chunks) == 0 {
		return nil, io.EOF
	}
	c := s.chunks[0]
	s.chunks = s.chunks[1:]
	return c, nil
}

func TestExportWorkflowCleanup(t *testing.T) {
	rdb := redisForExportTest(t)
	for _, scenario := range []string{"empty", "columns", "upload-failure", "stream-failure", "canceled", "timeout", "upload-canceled", "upload-timeout", "incomplete", "public-result", "wrong-tenant"} {
		t.Run(scenario, func(t *testing.T) {
			dir := t.TempDir()
			t.Setenv("TMP", dir)
			t.Setenv("TEMP", dir)
			t.Setenv("TMPDIR", dir)
			op := "test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
			ctx := ctxdata.WithRawToken(ctxdata.WithClaims(context.Background(), &ctxdata.Claims{UserID: 1, UserType: ctxdata.UserTypeOperator, OperatorCode: op, TokenType: ctxdata.TokenTypeAccess}), "test")
			var cancelUpload context.CancelFunc
			if scenario == "upload-canceled" {
				ctx, cancelUpload = context.WithCancel(ctx)
				defer cancelUpload()
			}
			if scenario == "upload-timeout" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 100*time.Millisecond)
				defer cancel()
			}
			if scenario == "canceled" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}
			if scenario == "timeout" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithDeadline(ctx, time.Now().Add(-time.Second))
				defer cancel()
			}
			var uploaded []byte
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/p9/oss/file/upload" || r.Method != "POST" || r.Header.Get("Authorization") != "Bearer test" {
					t.Error("wrong OSS public API or missing token")
					w.WriteHeader(400)
					return
				}
				if r.Header.Get(memberExportOSSAuthScopeHeader) != memberExportOSSAuthScope {
					t.Error("missing operator OSS authentication scope")
				}
				if scenario == "upload-failure" {
					w.WriteHeader(503)
					return
				}
				if e := r.ParseMultipartForm(1 << 20); e != nil {
					t.Error(e)
					w.WriteHeader(400)
					return
				}
				defer r.MultipartForm.RemoveAll()
				if r.FormValue("scene") != "member_export" {
					t.Error("wrong OSS scene")
				}
				file, _, e := r.FormFile("file")
				if e != nil {
					t.Error(e)
					return
				}
				uploaded, _ = io.ReadAll(file)
				file.Close()
				if scenario == "upload-canceled" || scenario == "upload-timeout" {
					if cancelUpload != nil {
						cancelUpload()
					}
					<-r.Context().Done()
					return
				}
				owner := op
				if scenario == "wrong-tenant" {
					owner = "foreign"
				}
				visibility := "private"
				if scenario == "public-result" {
					visibility = "public"
				}
				key := owner + "/member_export/20261119/" + strings.Repeat("a", 32) + ".csv"
				json.NewEncoder(w).Encode(map[string]any{"code": 0, "data": map[string]any{"key": key, "url": "https://storage.invalid/bucket/" + key + "?signature=test", "scene": "member_export", "visibility": visibility, "url_expires_in": 600}})
			}))
			defer server.Close()
			var source bytes.Buffer
			writer := csv.NewWriter(&source)
			writer.Write(contract.Header)
			count := int64(0)
			req := &types.ExportMembersRequest{}
			if scenario == "columns" {
				row := make([]string, len(contract.Header))
				for i := range row {
					row[i] = contract.Missing
				}
				row[0] = "'1"
				row[1] = "'123456789012"
				row[2] = "Alice"
				row[6] = "1"
				writer.Write(row)
				count = 1
				req.Columns = []string{"username", "status"}
			}
			writer.Flush()
			started := time.Now().UnixMilli()
			rpc := testExportClient{chunks: []*pb.ExportChunk{{Csv: source.Bytes(), Rows: count, StartedAt: started}, {Complete: true, Rows: count, StartedAt: started}}}
			if scenario == "stream-failure" {
				rpc.err = errors.New("RPC failed")
			}
			if scenario == "incomplete" {
				rpc.chunks = rpc.chunks[:1]
			}
			service := &svc.ServiceContext{ExportRedis: rdb, MemberExportRpc: rpc, Config: config.Config{OssApiBaseURL: server.URL}}
			result, err := NewExportLogic(ctx, service).Export(req)
			if scenario == "empty" || scenario == "columns" {
				if err != nil || result.Rows != count || result.Visibility != "private" {
					t.Fatalf("export failed: %v", err)
				}
				records, e := csv.NewReader(bytes.NewReader(uploaded)).ReadAll()
				if e != nil || len(records) != int(count)+1 {
					t.Fatal("wrong uploaded CSV row count")
				}
				if scenario == "columns" && (len(records[0]) != 2 || records[0][0] != "username" || records[1][0] != "Alice" || records[1][1] != "1") {
					t.Fatal("requested columns not applied")
				}
			} else {
				if err == nil {
					t.Fatal("failed export delivered link")
				}
				if (scenario == "canceled" || scenario == "upload-canceled") && httpStatus(err) != 499 {
					t.Fatal("cancellation classification lost")
				}
				if (scenario == "timeout" || scenario == "upload-timeout") && httpStatus(err) != 504 {
					t.Fatal("timeout classification lost")
				}
			}
			files, _ := filepath.Glob(filepath.Join(dir, "member-export-*.csv"))
			if len(files) != 0 {
				t.Fatal("temporary file leaked")
			}
			if rdb.Exists(context.Background(), "member:export:v2:"+op).Val() != 0 {
				t.Fatal("lock leaked")
			}
		})
	}
}

func TestExportUnavailableFiltersAndIdentity(t *testing.T) {
	service := &svc.ServiceContext{}
	if _, err := NewExportLogic(context.Background(), service).Export(&types.ExportMembersRequest{}); httpStatus(err) != 403 {
		t.Fatal("missing identity allowed")
	}
	ctx := ctxdata.WithRawToken(ctxdata.WithClaims(context.Background(), &ctxdata.Claims{UserID: 1, UserType: ctxdata.UserTypeOperator, OperatorCode: "OP_test", TokenType: ctxdata.TokenTypeAccess}), "test")
	for _, req := range []*types.ExportMembersRequest{{RangeTodayDeposit: []string{"0", "1"}}, {RangeTodayWithdraw: []string{}}, {Columns: []string{"unknown"}}, {Columns: []string{"username", "username"}}} {
		if _, err := NewExportLogic(ctx, service).Export(req); httpStatus(err) != 400 {
			t.Fatal("today filter ignored")
		}
	}
	if exportErrorClass(xerr.Err(504, "timeout")) != "504" {
		t.Fatal("wrong error class")
	}
}
