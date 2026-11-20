package member

import (
	"context"
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
	"oa.98ent.com/p9/operator-crm/ent/migrate"
	_ "oa.98ent.com/p9/operator-crm/ent/runtime"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/dao"
)

func TestBatchCreateTx(t *testing.T) {
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
	schema := fmt.Sprintf("test_member_batch_%d", time.Now().UnixNano())
	if _, err := admin.Exec("CREATE SCHEMA " + schema); err != nil {
		t.Fatal(err)
	}
	defer admin.Exec("DROP SCHEMA " + schema + " CASCADE")
	cfg.RuntimeParams["search_path"] = schema + ",public"
	client := ent.NewClient(ent.Driver(entsql.OpenDB(dialect.Postgres, stdlib.OpenDB(*cfg))))
	defer client.Close()
	ctx := ctxdata.WithClaims(context.Background(), &ctxdata.Claims{
		UserID: 1, UserType: ctxdata.UserTypeOperator, OperatorCode: "test_batch",
	})
	if err := client.Schema.Create(ctx, migrate.WithForeignKeys(false)); err != nil {
		t.Fatal(err)
	}
	client.MemberAutoLevel.Create().SetName("普通会员").SetMinDeposit(0).SetMinWater(0).SaveX(ctx)
	client.MemberVipLevel.Create().SetLevel(0).SetUpgradeDeposit(0).SetUpgradeWater(0).
		SetRetentionDeposit(0).SetRetentionWater(0).SaveX(ctx)

	group := dao.New(client)
	service := New(client, &group.MemberAutoLevel, &group.MemberVipLevel, nil)
	params := make([]CreateParams, 2000)
	for i := range params {
		account := fmt.Sprintf("batch_%04d", i)
		params[i] = CreateParams{
			Code: fmt.Sprintf("%012d", 100000000000+i), InviteCode: fmt.Sprintf("INVITE%d", i),
			AgentID: 1, Account: &account, Nickname: account, RegisterMethod: "account", RegisterChannel: "admin",
		}
	}
	checkCounts := func(want int) {
		t.Helper()
		if members, auths, profiles, progress := client.Member.Query().CountX(ctx), client.MemberAuth.Query().CountX(ctx),
			client.MemberProfile.Query().CountX(ctx), client.MemberVipProgress.Query().CountX(ctx); members != want || auths != want || profiles != want || progress != want {
			t.Fatalf("counts = %d, %d, %d, %d, want %d each", members, auths, profiles, progress, want)
		}
	}

	// Force a duplicate in the second SQL batch, after the first batch was written.
	validCode := params[memberInsertBatchSize].Code
	params[memberInsertBatchSize].Code = params[0].Code
	err = dao.WithTx(ctx, client, func(tx *ent.Tx) error {
		_, err := service.BatchCreateTx(ctx, tx, params[:memberInsertBatchSize+1])
		return err
	})
	if !ent.IsConstraintError(err) {
		t.Fatalf("expected duplicate constraint error, got %v", err)
	}
	checkCounts(0)

	params[memberInsertBatchSize].Code = validCode
	err = dao.WithTx(ctx, client, func(tx *ent.Tx) error {
		members, err := service.BatchCreateTx(ctx, tx, params)
		if err != nil {
			return err
		}
		if len(members) != len(params) {
			return fmt.Errorf("created %d members, want %d", len(members), len(params))
		}
		for i, member := range members {
			if member.ID == 0 || member.Code != params[i].Code {
				return fmt.Errorf("result %d does not match its input", i)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	checkCounts(len(params))
}
