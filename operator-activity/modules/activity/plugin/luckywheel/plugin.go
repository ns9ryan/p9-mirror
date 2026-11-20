package luckywheel

import (
	"context"
	"time"

	"oa.98ent.com/p9/operator-activity/modules/activity/plugin"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"

	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xkafka/topic"
)

// Plugin 幸运转盘：监听投注/充值达标送抽奖机会
type Plugin struct {
	svc *Service
}

func New(crmServ CrmLookup, store ChanceStore, deps *plugin.StrategyDeps, locker Locker) *Plugin {
	return &Plugin{
		svc: NewService(crmServ, store, deps, locker),
	}
}

func (p *Plugin) Name() string { return plugin.NameLuckyWheel }

func (p *Plugin) ActivityType() types.ActivityType { return types.ActivityTypeLuckyWheel }

func (p *Plugin) Service() *Service { return p.svc }

func (p *Plugin) Dispatch(ctx context.Context, payload types.Payload) error {
	evt, err := buildEvent(payload)
	if err != nil || evt == nil {
		return err
	}
	if evt.OperatorCode == "" {
		logx.Errorf("plugin %s: operator code is empty", p.Name())
		return nil
	}
	ctx = ctxdata.WithOperatorCode(ctx, evt.OperatorCode)
	return p.Handle(ctx, evt)
}

func buildEvent(payload types.Payload) (*Event, error) {
	switch payload.Name {
	case topic.GameUserGameOrderSettled:
		var betPayload types.BetPayload
		if err := payload.UnmarshalData(&betPayload); err != nil {
			return nil, err
		}
		eventTime := time.Now()
		if betPayload.Timestamp > 0 {
			eventTime = time.UnixMilli(betPayload.Timestamp)
		}
		return &Event{
			EventType:    EventBet,
			OperatorCode: betPayload.OperatorCode,
			MemberID:     betPayload.MemberID,
			OrderID:      betPayload.OrderID,
			Amount:       betPayload.Amount,
			ClientType:   betPayload.ClientType,
			GameID:       betPayload.GameID,
			EventTime:    eventTime,
		}, nil
	case topic.GameUserDepositOrderApproved:
		var depositPayload types.DepositPayload
		if err := payload.UnmarshalData(&depositPayload); err != nil {
			return nil, err
		}
		eventTime := time.Now()
		if depositPayload.Timestamp > 0 {
			eventTime = time.UnixMilli(depositPayload.Timestamp)
		}
		return &Event{
			EventType:    EventDeposit,
			OperatorCode: depositPayload.OperatorCode,
			MemberID:     depositPayload.MemberID,
			OrderID:      depositPayload.OrderID,
			Amount:       depositPayload.Amount,
			ClientType:   depositPayload.ClientType,
			EventTime:    eventTime,
		}, nil
	}
	return nil, nil
}

func (p *Plugin) Handle(ctx context.Context, evt *Event) error {
	processed, memberInfo, err := p.svc.Process(ctx, evt)
	if err != nil || processed == nil {
		return err
	}
	return p.svc.GrantChances(ctx, processed, memberInfo)
}
