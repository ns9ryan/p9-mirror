package referral

import (
	"context"
	"time"

	"oa.98ent.com/p9/operator-activity/modules/activity/plugin"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"

	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xkafka/topic"
)

// Plugin 推荐活动插件
type Plugin struct {
	svc    *Service
	engine plugin.EventProcessor
}

func New(engine plugin.EventProcessor, crmServ CrmLookup, deps *plugin.StrategyDeps) *Plugin {
	return &Plugin{
		svc:    NewService(crmServ, deps),
		engine: engine,
	}
}

func (p *Plugin) Name() string { return plugin.NameReferral }

func (p *Plugin) ActivityType() types.ActivityType { return types.ActivityTypeReferralMember }

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
	case topic.GameUserRegister:
		var registerPayload types.RegisterPayload
		if err := payload.UnmarshalData(&registerPayload); err != nil {
			return nil, err
		}
		eventTime := time.Now()
		if registerPayload.Timestamp > 0 {
			eventTime = time.UnixMilli(registerPayload.Timestamp)
		}
		return &Event{
			EventType:    EventRegister,
			OperatorCode: registerPayload.OperatorCode,
			MemberID:     registerPayload.MemberID,
			InviterID:    registerPayload.InviterID,
			PromoRefID:   registerPayload.PromoRefID,
			EventTime:    eventTime,
		}, nil

	case topic.GameUserDepositOrderApproved:
		var depositPayload types.DepositPayload
		if err := payload.UnmarshalData(&depositPayload); err != nil {
			return nil, err
		}
		eventTime := time.Now()
		if depositPayload.Timestamp > 0 {
			eventTime = time.UnixMilli(payload.Timestamp)
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
	case topic.GameUserGameOrderSettled:
		var betPayload types.BetPayload
		if err := payload.UnmarshalData(&betPayload); err != nil {
			return nil, err
		}
		eventTime := time.Now()
		if betPayload.Timestamp > 0 {
			eventTime = time.UnixMilli(payload.Timestamp)
		}
		return &Event{
			EventType:    EventBet,
			OperatorCode: betPayload.OperatorCode,
			MemberID:     betPayload.MemberID,
			OrderID:      betPayload.OrderID,
			Amount:       betPayload.Amount,
			GameID:       betPayload.GameID,
			ClientType:   betPayload.ClientType,
			EventTime:    eventTime,
		}, nil
	}
	return nil, nil
}

// Handle 校验补全邀请关系 → 驱动引擎
func (p *Plugin) Handle(ctx context.Context, evt *Event) error {
	processed, memberInfo, err := p.svc.Process(ctx, evt)
	if err != nil || processed == nil {
		return err
	}
	return p.engine.ProcessEvent(ctx, processed, types.ActivityTypeReferralMember, memberInfo, p.Service())
}
