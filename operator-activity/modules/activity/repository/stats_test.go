package repository_test

import (
	"context"
	"testing"

	"oa.98ent.com/p9/operator-activity/ent"
	"oa.98ent.com/p9/operator-activity/ent/enttest"
	"oa.98ent.com/p9/operator-activity/modules/activity/repository"

	"entgo.io/ent/dialect"
	_ "github.com/mattn/go-sqlite3"
)

func setupStatDB(t *testing.T) *ent.Client {
	t.Helper()
	client := enttest.Open(t, dialect.SQLite, "file:stat_"+t.Name()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func TestIncrPromoClaimed_AmountAndUniqueUsers(t *testing.T) {
	client := setupStatDB(t)
	ctx := context.Background()

	if err := repository.IncrPromoClaimed(ctx, client, "demo", 10, 1001, 9001, 5000); err != nil {
		t.Fatal(err)
	}
	if err := repository.IncrPromoClaimed(ctx, client, "demo", 10, 1001, 9002, 3000); err != nil {
		t.Fatal(err)
	}
	if err := repository.IncrPromoClaimed(ctx, client, "demo", 10, 1002, 9003, 2000); err != nil {
		t.Fatal(err)
	}
	if err := repository.IncrPromoClaimed(ctx, client, "demo", 10, 1002, 9003, 2000); err != nil {
		t.Fatal(err)
	}

	st, err := repository.GetActivityStats(client, 10)
	if err != nil {
		t.Fatal(err)
	}
	if st.ClaimMemberCount != 2 {
		t.Fatalf("claimed_member_count=%d want 2", st.ClaimMemberCount)
	}
	if st.ClaimedAmount != 10000 {
		t.Fatalf("claimed_amount=%d want 10000", st.ClaimedAmount)
	}
	if st.ClaimedOrderCount != 3 {
		t.Fatalf("claimed_order_count=%d want 3", st.ClaimedOrderCount)
	}
}
