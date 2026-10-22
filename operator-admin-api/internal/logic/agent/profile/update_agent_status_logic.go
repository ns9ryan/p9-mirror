package profile

import (
	"context"

	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateAgentStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateAgentStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAgentStatusLogic {
	return &UpdateAgentStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 修改代理状态
func (l *UpdateAgentStatusLogic) UpdateAgentStatus(req *types.UpdateAgentStatusReq) (*types.AgentEmpty, error) {
	if req == nil {
		req = &types.UpdateAgentStatusReq{}
	}
	_, err := l.svcCtx.AgentRpc.UpdateAgentStatus(l.ctx, &agentpb.UpdateAgentStatusReq{
		AgentId: req.AgentId,
		Status:  req.Status,
	})
	if err != nil {
		return nil, err
	}
	return &types.AgentEmpty{}, nil
}
