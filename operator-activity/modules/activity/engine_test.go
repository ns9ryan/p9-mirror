package promotion_test

import (
	"context"
	"testing"
	"time"

	"oa.98ent.com/p9/operator-activity/ent"
	"oa.98ent.com/p9/operator-activity/internal/config"
	"oa.98ent.com/p9/operator-activity/internal/svc"
	"oa.98ent.com/p9/operator-activity/modules/activity/plugin"

	"entgo.io/ent/dialect"
	"oa.98ent.com/p9/operator-activity/ent/enttest"
	promotion "oa.98ent.com/p9/operator-activity/modules/activity"
	"oa.98ent.com/p9/operator-activity/modules/activity/plugin/referral"
	"oa.98ent.com/p9/operator-activity/modules/activity/repository"
	"oa.98ent.com/p9/operator-activity/modules/activity/repository/crm"
	"oa.98ent.com/p9/operator-activity/modules/activity/repository/seed"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"
)

func TestEngine_ReferralReward(t *testing.T) {
	client := openTestDB(t)
	if err := seed.Run(client); err != nil {
		t.Fatal(err)
	}

	repo := repository.New(client)
	memberServ := crm.NewCrmServ(svc.NewServiceContext(config.Config{}))
	engine := promotion.NewEngine(repo, nil)
	deps := plugin.NewStrategyDeps(repo, nil)
	deps.Workflows = nil
	pipelines := referral.NewService(memberServ, deps)

	evt := &referral.Event{
		EventType:    referral.EventRegister,
		OperatorCode: seed.OperatorCode,
		MemberID:     1001,
		InviterID:    1001,
		EventTime:    time.Now(),
	}

	if err := engine.ProcessEvent(context.Background(), evt, types.ActivityTypeReferralMember, nil, pipelines); err != nil {
		t.Fatal(err)
	}

	count, err := repo.CountOrdersByPromo(context.Background(), 3)
	if err != nil {
		t.Fatal(err)
	}
	if count < 1 {
		t.Fatalf("expected at least 1 referral order, got %d", count)
	}
}

func openTestDB(t *testing.T) *ent.Client {
	t.Helper()
	client := enttest.Open(t, dialect.SQLite, "file:"+t.Name()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	return client
}
