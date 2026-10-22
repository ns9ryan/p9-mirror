package agentservicelogic

import (
	"context"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateAgentStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateAgentStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAgentStatusLogic {
	return &UpdateAgentStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateAgentStatus 更新代理账号状态：1正常 2冻结 3禁用。
func (l *UpdateAgentStatusLogic) UpdateAgentStatus(in *agentpb.UpdateAgentStatusReq) (*agentpb.Empty, error) {
	resp, err := l.svcCtx.Agent.UpdateStatus(l.ctx, in)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}
	return resp, nil
}
