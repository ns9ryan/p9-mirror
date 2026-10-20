package config

import (
	"context"

	agentlogic "oa.98ent.com/p9/operator-admin-api/internal/logic/agent"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/pkg/agentline"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLineDefaultAgentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetLineDefaultAgentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLineDefaultAgentsLogic {
	return &GetLineDefaultAgentsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetLineDefaultAgentsLogic) GetLineDefaultAgents() (*types.LineDefaultAgentsResp, error) {
	codes, err := agentlogic.OpenedLineCodes(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}
	cfg, err := l.svcCtx.AgentRpc.GetAgentConfig(l.ctx, &agentpb.GetAgentConfigReq{})
	if err != nil {
		return nil, err
	}
	items := make([]types.LineDefaultAgent, 0, len(codes))
	for _, code := range codes {
		agentID := defaultAgentID(cfg, code)
		item := types.LineDefaultAgent{LineCode: code, AgentId: agentID}
		if agentID > 0 {
			profile, perr := l.svcCtx.AgentRpc.GetAgentProfile(l.ctx, &agentpb.GetAgentProfileReq{AgentId: agentID})
			if perr != nil {
				return nil, perr
			}
			item.ShareRate = profile.GetShareRate()
		}
		items = append(items, item)
	}
	return &types.LineDefaultAgentsResp{Items: items}, nil
}

func defaultAgentID(cfg *agentpb.AgentConfigResp, code string) int64 {
	if cfg == nil {
		return 0
	}
	switch code {
	case agentline.CashProduction:
		return cfg.GetCashProductionAgentId()
	case agentline.CashDemo:
		return cfg.GetCashDemoAgentId()
	case agentline.CashTest:
		return cfg.GetCashTestAgentId()
	case agentline.CreditDemo:
		return cfg.GetCreditDemoAgentId()
	case agentline.CreditTest:
		return cfg.GetCreditTestAgentId()
	default:
		return 0
	}
}
