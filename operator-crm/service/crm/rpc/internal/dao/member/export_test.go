package member

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/ent/migrate"
	_ "oa.98ent.com/p9/operator-crm/ent/runtime"
)

func TestExportPostgres(t *testing.T) {
	dsn := os.Getenv("CRM_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set CRM_TEST_POSTGRES_DSN for disposable-schema PostgreSQL tests")
	}
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	admin := stdlib.OpenDB(*cfg)
	defer admin.Close()
	schema := fmt.Sprintf("test_member_export_%d", time.Now().UnixNano())
	if _, err = admin.Exec("CREATE SCHEMA " + schema); err != nil {
		t.Fatal(err)
	}
	defer admin.Exec("DROP SCHEMA " + schema + " CASCADE")
	cfg.RuntimeParams["search_path"] = schema + ",public"
	db := stdlib.OpenDB(*cfg)
	driver := sql.OpenDB(dialect.Postgres, db)
	client := ent.NewClient(ent.Driver(driver))
	defer client.Close()
	ctx := ctxdata.WithClaims(context.Background(), &ctxdata.Claims{UserID: 1, UserType: ctxdata.UserTypeOperator, OperatorCode: "test_export"})
	if err = client.Schema.Create(ctx, migrate.WithForeignKeys(false)); err != nil {
		t.Fatal(err)
	}
	agent := client.AgentUser.Create().SetAccount("agent").SetCode("agent").SetRegisterSource(3).SetAgentType(1).SetAgentLineCode("CASH_PRODUCTION").SetPromoCode("promo").SetLoginPasswordHash("hash").SetSalt("salt").SaveX(ctx)
	level := client.MemberAutoLevel.Create().SetName("level").SetMinDeposit(0).SetMinWater(0).SaveX(ctx)
	vip := client.MemberVipLevel.Create().SetLevel(0).SetUpgradeDeposit(0).SetUpgradeWater(0).SetRetentionDeposit(0).SetRetentionWater(0).SaveX(ctx)
	instant := time.UnixMilli(1700000000123)
	ids := make([]int64, 205)
	for i := range ids {
		m := client.Member.Create().SetCode(fmt.Sprintf("%012d", 100000000000+i)).SetAgentID(agent.ID).SetInviteCode(fmt.Sprintf("invite%d", i)).SetAutoLevelID(level.ID).SetVipLevelID(vip.ID).SetRegisterMethod("admin").SetRegisterChannel("test").SetCreatedAt(instant).SetRegisterIP("192.0.2.1").SetLastLoginIP("192.0.2.2").SetRegisterDeviceCode("register-device").SetLastLoginDeviceCode("login-device").SaveX(ctx)
		ids[i] = m.ID
		client.MemberAuth.Create().SetMemberID(m.ID).SetUsername(fmt.Sprintf("MiXeD%d", i)).SaveX(ctx)
		client.MemberProfile.Create().SetMemberID(m.ID).SetNickname(fmt.Sprintf("Nick%d", i)).SaveX(ctx)
	}
	slices.Reverse(ids)
	s := New(client, nil, nil, driver)
	assertIDs := func(in ExportFilters, want []int64) {
		t.Helper()
		got, e := s.ExportIDs(ctx, in)
		if e != nil || !slices.Equal(got, want) {
			t.Fatalf("membership got %d want %d, error=%v", len(got), len(want), e)
		}
	}
	assertIDs(ExportFilters{}, ids)
	assertIDs(ExportFilters{IP: "192.0.2.2", Device: "register-device"}, ids)
	assertIDs(ExportFilters{IP: "192.0.2.1", Device: "login-device"}, ids)
	assertIDs(ExportFilters{IP: "192.0.2.3", Device: "login-device"}, nil)
	assertIDs(ExportFilters{IP: "192.0.2.1", Device: "missing"}, nil)
	assertIDs(ExportFilters{ParentUsername: "no-such-inviter"}, nil)
	assertIDs(ExportFilters{List: ListParams{Keyword: "mIxEd204"}}, ids[:1])
	assertIDs(ExportFilters{List: ListParams{Keyword: "nIcK204"}}, ids[:1])
	assertIDs(ExportFilters{AgentType: 2}, nil)
	// Same millisecond across three batches. Each selected ID occurs exactly once.
	var got []int64
	for i := 0; i < len(ids); i += ExportBatchSize {
		batch, e := s.ExportBatch(ctx, ids[i:min(i+ExportBatchSize, len(ids))])
		if e != nil {
			t.Fatal(e)
		}
		for _, m := range batch {
			if !m.CreatedAt.Equal(instant) || m.Edges.Auth == nil {
				t.Fatal("lost exact time/auth")
			}
			got = append(got, m.ID)
		}
	}
	if !slices.Equal(got, ids) {
		t.Fatal("paging lost or duplicated members")
	}
	if _, err = s.ExportBatch(ctxdata.WithClaims(context.Background(), &ctxdata.Claims{OperatorCode: "foreign"}), ids[:1]); err == nil {
		t.Fatal("foreign member exported")
	}
	client.AgentUser.UpdateOneID(agent.ID).SetAgentLineCode("CASH_DEMO").SaveX(ctx)
	assertIDs(ExportFilters{}, nil)
	assertIDs(ExportFilters{IncludeDemo: true}, ids)
	assertIDs(ExportFilters{IncludeDemo: true, AgentType: 2}, nil)
	client.AgentUser.UpdateOneID(agent.ID).SetAgentLineCode("CREDIT_DEMO").SaveX(ctx)
	assertIDs(ExportFilters{}, nil)
	assertIDs(ExportFilters{IncludeDemo: true}, ids)
	client.AgentUser.UpdateOneID(agent.ID).SetAgentLineCode("CASH_PRODUCTION").SaveX(ctx)
	// A deleted selected member fails the batch instead of delivering a short file.
	client.Member.DeleteOneID(ids[0]).ExecX(ctx)
	if _, err = s.ExportBatch(ctx, ids[:1]); err == nil {
		t.Fatal("membership deletion ignored")
	}
	// A failed aggregate query must fail the batch, never fabricate zero statistics.
	if _, err = db.Exec("DROP TABLE member_stats"); err != nil {
		t.Fatal(err)
	}
	if _, err = s.ExportBatch(ctx, ids[1:2]); err == nil {
		t.Fatal("statistics query failure ignored")
	}
}

func TestExportIncludeDemoPreservesTenantAndAgentTypeFilters(t *testing.T) {
	for _, includeDemo := range []bool{false, true} {
		filters, err := exportWhere(ExportFilters{IncludeDemo: includeDemo, AgentType: 1}, "OP_scope")
		if err != nil {
			t.Fatal(err)
		}
		selector := sql.Dialect(dialect.Postgres).Select("id").From(sql.Table("member"))
		for _, filter := range filters {
			filter(selector)
		}
		query, args := selector.Query()
		if !strings.Contains(query, "agent_type") || !strings.Contains(query, "operator_code") {
			t.Fatal("including demo removed mandatory scope or agent type filter")
		}
		containsDemoFilter := false
		scopeCount := 0
		for _, argument := range args {
			if argument == "CASH_DEMO" || argument == "CREDIT_DEMO" {
				containsDemoFilter = true
			}
			if argument == "OP_scope" {
				scopeCount++
			}
		}
		if containsDemoFilter == includeDemo || scopeCount < 2 {
			t.Fatalf("include_demo=%v: incorrect demo or tenant conditions", includeDemo)
		}
	}
}

func TestInvalidExportFilters(t *testing.T) {
	for _, f := range []ExportFilters{{IP: "not-an-ip"}, {Device: string(make([]byte, 129))}, {ParentUsername: "   "}, {AgentType: 3}, {List: ListParams{TimeField: 99}}} {
		if err := ValidateExportFilters(f); err == nil {
			t.Fatal("invalid filter accepted")
		}
	}
	if _, err := exportWhere(ExportFilters{}, ""); err == nil {
		t.Fatal("missing tenant allowed")
	}
}

func TestExportSQLScopeAndCombination(t *testing.T) {
	filters, err := exportWhere(ExportFilters{List: ListParams{Keyword: "MiXeD"}, IP: "192.0.2.1", Device: "device", ParentUsername: "inviter", AgentType: 1}, "OP_scope")
	if err != nil {
		t.Fatal(err)
	}
	selector := sql.Dialect(dialect.Postgres).Select("id").From(sql.Table("member"))
	for _, filter := range filters {
		filter(selector)
	}
	query, args := selector.Query()
	for _, column := range []string{"operator_code", "register_ip", "last_login_ip", "register_device_code", "last_login_device_code", "username", "nickname", "inviter_member_id", "agent_line_code", "agent_type"} {
		if !strings.Contains(query, column) {
			t.Fatalf("missing SQL condition %s", column)
		}
	}
	scopeCount := 0
	for _, arg := range args {
		if arg == "OP_scope" {
			scopeCount++
		}
	}
	if scopeCount < 5 {
		t.Fatal("association predicates omitted explicit tenant scopes")
	}
	if strings.Count(query, " OR ") < 3 || !strings.Contains(query, " AND ") {
		t.Fatal("IP/device and keyword combination lost")
	}
}
