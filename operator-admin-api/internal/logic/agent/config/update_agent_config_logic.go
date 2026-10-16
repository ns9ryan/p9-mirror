package config

import (
	"context"

	agentconv "oa.98ent.com/p9/operator-admin-api/internal/logic/agent"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateAgentConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateAgentConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAgentConfigLogic {
	return &UpdateAgentConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateAgentConfigLogic) UpdateAgentConfig(req *types.UpdateAgentConfigReq) (*types.AgentConfig, error) {
	if req == nil {
		req = &types.UpdateAgentConfigReq{}
	}
	rpcResp, err := l.svcCtx.AgentRpc.UpdateAgentConfig(l.ctx, &agentpb.UpdateAgentConfigReq{
		MaxShareRate:          req.MaxShareRate,
		ApplyMinShareRate:     req.ApplyMinShareRate,
		ApplyMaxShareRate:     req.ApplyMaxShareRate,
		CashMaxMemberCount:    req.CashMaxMemberCount,
		CashMaxAgentPerIp:     req.CashMaxAgentPerIp,
		CashProductionAgentId: req.CashProductionAgentId,
		CashDemoAgentId:       req.CashDemoAgentId,
		CashTestAgentId:       req.CashTestAgentId,
		CreditDemoAgentId:     req.CreditDemoAgentId,
		CreditTestAgentId:     req.CreditTestAgentId,
	})
	if err != nil {
		return nil, err
	}
	cfg := agentconv.ConfigFromPB(rpcResp)
	return &cfg, nil
}
