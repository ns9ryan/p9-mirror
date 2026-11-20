package luckywheel_test

import (
	"context"
	"testing"
	"time"

	"oa.98ent.com/p9/operator-activity/modules/activity/plugin"
	"oa.98ent.com/p9/operator-activity/modules/activity/plugin/luckywheel"
	"oa.98ent.com/p9/operator-activity/modules/activity/repository"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"
)

type mockCrmServ struct {
	member     *types.MemberInfo
	opBaseInfo *types.OpBaseInfo
}

func (m *mockCrmServ) LoadMember(ctx context.Context, id int64) (*types.MemberInfo, error) {
	return m.member, nil
}

func (m *mockCrmServ) LoadOpBaseInfo(ctx context.Context) (*types.OpBaseInfo, error) {
	return m.opBaseInfo, nil
}

func TestService_Process(t *testing.T) {
	svc := luckywheel.NewService(&mockCrmServ{
		member: &types.MemberInfo{ID: 1001, OperatorCode: "demo"},
	}, nil, plugin.NewStrategyDeps(nil, nil), nil)

	evt := &luckywheel.Event{
		EventType: luckywheel.EventBet,
		MemberID:  1001,
		Amount:    20000,
		EventTime: time.Now(),
	}
	out, user, err := svc.Process(context.Background(), evt)
	if err != nil {
		t.Fatal(err)
	}
	if out.OperatorCode != "demo" || out.ClientType != 2 {
		t.Fatalf("expected operator_code/client filled, got operator_code=%s client=%d", out.OperatorCode, out.ClientType)
	}
	if user == nil || user.ID != 1001 {
		t.Fatalf("expected user 1001")
	}
}

func TestService_Process_IgnoreOtherEvents(t *testing.T) {
	svc := luckywheel.NewService(&mockCrmServ{member: &types.MemberInfo{ID: 1}}, nil, plugin.NewStrategyDeps(nil, nil), nil)
	out, _, err := svc.Process(context.Background(), &luckywheel.Event{
		EventType: luckywheel.EventSpin,
		MemberID:  1,
		Amount:    100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if out != nil {
		t.Fatal("expected nil for non bet/deposit events")
	}
}

func TestWheelTriggerMode_Has(t *testing.T) {
	cases := []struct {
		mode   types.WheelTriggerMode
		flag   types.WheelTriggerMode
		expect bool
	}{
		{0, types.WheelTriggerBet, true},
		{0, types.WheelTriggerDeposit, false},
		{types.WheelTriggerBet, types.WheelTriggerBet, true},
		{types.WheelTriggerBet, types.WheelTriggerDeposit, false},
		{types.WheelTriggerDeposit, types.WheelTriggerDeposit, true},
		{types.WheelTriggerDeposit, types.WheelTriggerBet, false},
		{types.WheelTriggerBet | types.WheelTriggerDeposit, types.WheelTriggerBet, true},
		{types.WheelTriggerBet | types.WheelTriggerDeposit, types.WheelTriggerDeposit, true},
	}
	for _, c := range cases {
		if got := c.mode.Has(c.flag); got != c.expect {
			t.Fatalf("mode=%d Has(%d)=%v want %v", c.mode, c.flag, got, c.expect)
		}
	}
}

func TestDrawPrize_Weighted(t *testing.T) {
	// 通过 Spin 间接验证：空奖池应失败；有奖池应返回合法 schema
	// 这里直接测 BuildPipeline + 配置
	promo := &types.Activity{
		ID:     4,
		Type:   types.ActivityTypeLuckyWheel,
		Status: types.ActivityStatusOngoing,
		Settings: types.ActivitySettings{
			RewardSchemas: []types.RewardSchema{
				{ID: 1, Amount: 100, Rate: 100},
			},
		},
	}
	store := &memStore{promo: promo, balance: 1, member: &types.MemberInfo{ID: 1, OperatorCode: "demo"}}
	svc := luckywheel.NewService(&mockCrmServ{member: &types.MemberInfo{ID: 1}}, store, plugin.NewStrategyDeps(nil, nil), nil)

	res, err := svc.Spin(context.Background(), 4, 1)
	if err != nil {
		t.Fatal(err)
	}
	if res.Prize.ID != 1 || res.Prize.Amount != 100 {
		t.Fatalf("unexpected prize %+v", res.Prize)
	}
	if res.Order == nil || res.Order.RewardsTotal != 100 {
		t.Fatalf("unexpected order %+v", res.Order)
	}
}

// memStore 最小内存实现，仅支撑 Spin 单测
type memStore struct {
	promo   *types.Activity
	member  *types.MemberInfo
	balance int
	orders  []*types.ActivityOrder
}

func (m *memStore) GetActivity(context.Context, int64) (*types.Activity, error) {
	return m.promo, nil
}
func (m *memStore) ExistsWheelChanceGrant(context.Context, string) (bool, error) {
	return false, nil
}
func (m *memStore) GrantWheelChances(context.Context, *repository.WheelChanceGrantModel) error {
	return nil
}
func (m *memStore) GetWheelChanceBalance(context.Context, int64, int64) (int, error) {
	return m.balance, nil
}
func (m *memStore) ConsumeWheelChance(context.Context, int64, int64) error {
	if m.balance < 1 {
		return context.Canceled
	}
	m.balance--
	return nil
}
func (m *memStore) LoadOngoingActivities(context.Context, types.ActivityType) ([]*types.Activity, error) {
	return []*types.Activity{m.promo}, nil
}
func (m *memStore) IdempKeyExists(context.Context, string) (bool, error) { return false, nil }
func (m *memStore) SaveOrders(_ context.Context, orders []*types.ActivityOrder, _ *types.Activity, _ *types.MemberInfo) error {
	m.orders = append(m.orders, orders...)
	return nil
}
func (m *memStore) PublishOrderCreated(context.Context, []*types.ActivityOrder) error { return nil }
