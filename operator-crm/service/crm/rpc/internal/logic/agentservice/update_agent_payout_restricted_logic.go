package agentservicelogic

import (
	"context"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateAgentPayoutRestrictedLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateAgentPayoutRestrictedLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAgentPayoutRestrictedLogic {
	return &UpdateAgentPayoutRestrictedLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateAgentPayoutRestrictedLogic) UpdateAgentPayoutRestricted(in *agentpb.UpdateAgentPayoutRestrictedReq) (*agentpb.Empty, error) {
	resp, err := l.svcCtx.Agent.UpdatePayoutRestricted(l.ctx, in)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}
	return resp, nil
}
