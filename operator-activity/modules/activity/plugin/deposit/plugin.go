package deposit

import (
	"context"
	"time"

	"oa.98ent.com/p9/operator-activity/modules/activity/plugin"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"

	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xkafka/topic"
)

// Plugin 充值活动插件
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

func (p *Plugin) Name() string { return plugin.NameDeposit }

func (p *Plugin) ActivityType() types.ActivityType { return types.ActivityTypeDeposit }

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
	// 设置分站编码到上下文
	ctx = ctxdata.WithOperatorCode(ctx, evt.OperatorCode)
	// 处理事件
	return p.Handle(ctx, evt)
}

func buildEvent(payload types.Payload) (*Event, error) {
	if payload.Name != topic.GameUserDepositOrderApproved {
		return nil, nil
	}
	var depositPayload types.DepositPayload
	if err := payload.UnmarshalData(&depositPayload); err != nil {
		return nil, err
	}
	eventTime := time.Now()
	if payload.Timestamp > 0 {
		eventTime = time.UnixMilli(payload.Timestamp)
	}
	return &Event{
		OperatorCode: depositPayload.OperatorCode,
		MemberID:     depositPayload.MemberID,
		OrderID:      depositPayload.OrderID,
		Amount:       depositPayload.Amount,
		DepositCount: depositPayload.DepositCount,
		ClientType:   depositPayload.ClientType,
		EventTime:    eventTime,
	}, nil
}

// Handle 校验补全 → 驱动引擎
func (p *Plugin) Handle(ctx context.Context, evt *Event) error {
	processed, memberInfo, err := p.svc.Process(ctx, evt)
	if err != nil || processed == nil {
		return err
	}
	return p.engine.ProcessEvent(ctx, processed, types.ActivityTypeDeposit, memberInfo, p.Service())
}
