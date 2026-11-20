package agentservicelogic

import (
	"context"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateAgentConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateAgentConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAgentConfigLogic {
	return &UpdateAgentConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateAgentConfig 更新代理配置
func (l *UpdateAgentConfigLogic) UpdateAgentConfig(in *agentpb.UpdateAgentConfigReq) (*agentpb.AgentConfigResp, error) {
	resp, err := l.svcCtx.Agent.UpdateConfig(l.ctx, in)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}
	return resp, nil
}
