package member

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/operator-crm/ent"
	assignment "oa.98ent.com/p9/operator-crm/ent/memberfixedlevelassignment"
	"oa.98ent.com/p9/operator-crm/ent/migrate"
	_ "oa.98ent.com/p9/operator-crm/ent/runtime"
)

// Uses a disposable schema; never migrates the existing application's tables.
func TestMemberLevelReplacement(t *testing.T) {
	dsn := os.Getenv("CRM_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set CRM_TEST_POSTGRES_DSN to run PostgreSQL integration test")
	}
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	admin := stdlib.OpenDB(*cfg)
	defer admin.Close()
	schemaName := fmt.Sprintf("test_member_levels_%d", time.Now().UnixNano())
	if _, err = admin.Exec("CREATE SCHEMA " + schemaName); err != nil {
		t.Fatal(err)
	}
	defer admin.Exec("DROP SCHEMA " + schemaName + " CASCADE")
	cfg.RuntimeParams["search_path"] = schemaName + ",public"
	var db *sql.DB = stdlib.OpenDB(*cfg)
	client := ent.NewClient(ent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	defer client.Close()
	ctx := ctxdata.WithClaims(context.Background(), &ctxdata.Claims{UserID: 1, UserType: ctxdata.UserTypeOperator, OperatorCode: "test_a"})
	if err = client.Schema.Create(ctx, migrate.WithForeignKeys(false)); err != nil {
		t.Fatal(err)
	}
	a := client.MemberAutoLevel.Create().SetName("A").SetMinDeposit(0).SetMinWater(0).SaveX(ctx)
	a2 := client.MemberAutoLevel.Create().SetName("A2").SetMinDeposit(1).SetMinWater(1).SaveX(ctx)
	f1 := client.MemberFixedLevel.Create().SetName("F1").SaveX(ctx)
	f2 := client.MemberFixedLevel.Create().SetName("F2").SaveX(ctx)
	other := ctxdata.WithClaims(context.Background(), &ctxdata.Claims{UserID: 1, UserType: ctxdata.UserTypeOperator, OperatorCode: "test_b"})
	foreign := client.MemberFixedLevel.Create().SetName("foreign").SaveX(other)
	foreignAuto := client.MemberAutoLevel.Create().SetName("foreign").SetMinDeposit(0).SetMinWater(0).SaveX(other)
	m := client.Member.Create().SetCode("123456789012").SetAgentID(1).SetInviteCode("test_invite").SetAutoLevelID(a.ID).SetVipLevelID(1).SetRegisterMethod("admin").SetRegisterChannel("test").SaveX(ctx)
	s := New(client, nil, nil)
	if err = s.UpdateAutoLevel(ctx, m.ID, a2.ID); err != nil {
		t.Fatal(err)
	}
	if got := client.Member.GetX(ctx, m.ID).AutoLevelID; got != a2.ID {
		t.Fatalf("auto level = %d", got)
	}
	if err = s.UpdateAutoLevel(ctx, m.ID, foreignAuto.ID); err == nil {
		t.Fatal("accepted foreign auto level")
	}
	if err = s.UpdateFixedLevels(ctx, m.ID, []int64{f1.ID, f1.ID}); err != nil {
		t.Fatal(err)
	}
	check := func(want int64) {
		t.Helper()
		rows := client.MemberFixedLevelAssignment.Query().Where(assignment.MemberIDEQ(m.ID)).AllX(ctx)
		if len(rows) != 1 || rows[0].FixedLevelID != want {
			t.Fatalf("assignments = %+v", rows)
		}
	}
	check(f1.ID)
	if err = s.UpdateFixedLevels(ctx, m.ID, []int64{foreign.ID}); err == nil {
		t.Fatal("accepted foreign fixed level")
	}
	check(f1.ID)
	if err = s.UpdateFixedLevels(other, m.ID, []int64{foreign.ID}); err == nil {
		t.Fatal("accepted foreign member")
	}
	check(f1.ID)
	if err = s.UpdateFixedLevels(ctx, m.ID, []int64{f2.ID}); err != nil {
		t.Fatal(err)
	}
	check(f2.ID)
	// Concurrent full replacements must leave one complete request's set.
	errs := make(chan error, 2)
	go func() { errs <- s.UpdateFixedLevels(ctx, m.ID, []int64{f1.ID}) }()
	go func() { errs <- s.UpdateFixedLevels(ctx, m.ID, []int64{f2.ID}) }()
	for i := 0; i < 2; i++ {
		if err := <-errs; err != nil {
			t.Fatal(err)
		}
	}
	if n := client.MemberFixedLevelAssignment.Query().Where(assignment.MemberIDEQ(m.ID)).CountX(ctx); n != 1 {
		t.Fatalf("concurrent updates left %d rows", n)
	}
	if err = s.UpdateFixedLevels(ctx, m.ID, []int64{f2.ID}); err != nil {
		t.Fatal(err)
	}
	// Fail insertion after deleting old assignments; the transaction must restore them.
	_, err = db.Exec("ALTER TABLE member_fixed_level_assignment ADD CONSTRAINT reject_test_level CHECK (fixed_level_id <> " + fmt.Sprint(f1.ID) + ")")
	if err != nil {
		t.Fatal(err)
	}
	if err = s.UpdateFixedLevels(ctx, m.ID, []int64{f1.ID}); err == nil {
		t.Fatal("expected insertion failure")
	}
	check(f2.ID)
	if err = s.UpdateFixedLevels(ctx, m.ID, nil); err != nil {
		t.Fatal(err)
	}
	if n := client.MemberFixedLevelAssignment.Query().Where(assignment.MemberIDEQ(m.ID)).CountX(ctx); n != 0 {
		t.Fatalf("clear left %d rows", n)
	}
	if err = s.UpdateFixedLevels(context.Background(), m.ID, nil); err == nil {
		t.Fatal("accepted missing operator")
	}
}
