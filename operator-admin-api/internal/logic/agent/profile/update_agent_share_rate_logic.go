package profile

import (
	"context"

	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateAgentShareRateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateAgentShareRateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAgentShareRateLogic {
	return &UpdateAgentShareRateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateAgentShareRateLogic) UpdateAgentShareRate(req *types.UpdateAgentShareRateReq) (*types.AgentEmpty, error) {
	if req == nil {
		req = &types.UpdateAgentShareRateReq{}
	}
	_, err := l.svcCtx.AgentRpc.UpdateAgentShareRate(l.ctx, &agentpb.UpdateAgentShareRateReq{
		AgentId:   req.AgentId,
		ShareRate: req.ShareRate,
	})
	if err != nil {
		return nil, err
	}
	return &types.AgentEmpty{}, nil
}
