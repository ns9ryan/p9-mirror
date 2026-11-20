package betting

import (
	"context"
	"time"

	"oa.98ent.com/p9/operator-activity/modules/activity/plugin"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"

	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xkafka/topic"
)

// Plugin 投注活动插件
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

func (p *Plugin) Name() string { return plugin.NameBetting }

func (p *Plugin) ActivityType() types.ActivityType { return types.ActivityTypeWagering }

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
	if payload.Name != topic.GameUserGameOrderSettled {
		return nil, nil
	}
	var betPayload types.BetPayload
	if err := payload.UnmarshalData(&betPayload); err != nil {
		return nil, err
	}
	eventTime := time.Now()
	if payload.Timestamp > 0 {
		eventTime = time.UnixMilli(payload.Timestamp)
	}
	return &Event{
		OperatorCode: betPayload.OperatorCode,
		MemberID:     betPayload.MemberID,
		OrderID:      betPayload.OrderID,
		Amount:       betPayload.Amount,
		ClientType:   betPayload.ClientType,
		GameID:       betPayload.GameID,
		EventTime:    eventTime,
	}, nil
}

// Handle 校验补全 → 驱动引擎
func (p *Plugin) Handle(ctx context.Context, evt *Event) error {
	processed, memberInfo, err := p.svc.Process(ctx, evt)
	if err != nil || processed == nil {
		return err
	}
	return p.engine.ProcessEvent(ctx, processed, types.ActivityTypeWagering, memberInfo, p.Service())
}
