package profile

import (
	"context"

	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateAgentPayoutRestrictedLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateAgentPayoutRestrictedLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAgentPayoutRestrictedLogic {
	return &UpdateAgentPayoutRestrictedLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateAgentPayoutRestrictedLogic) UpdateAgentPayoutRestricted(req *types.UpdateAgentPayoutRestrictedReq) (*types.AgentEmpty, error) {
	if req == nil {
		req = &types.UpdateAgentPayoutRestrictedReq{}
	}
	_, err := l.svcCtx.AgentRpc.UpdateAgentPayoutRestricted(l.ctx, &agentpb.UpdateAgentPayoutRestrictedReq{
		AgentId:          req.AgentId,
		PayoutRestricted: req.PayoutRestricted,
	})
	if err != nil {
		return nil, err
	}
	return &types.AgentEmpty{}, nil
}
