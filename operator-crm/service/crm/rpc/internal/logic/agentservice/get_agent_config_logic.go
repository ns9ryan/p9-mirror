package agentservicelogic

import (
	"context"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAgentConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAgentConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAgentConfigLogic {
	return &GetAgentConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetAgentConfig 获取代理配置
func (l *GetAgentConfigLogic) GetAgentConfig(in *agentpb.GetAgentConfigReq) (*agentpb.AgentConfigResp, error) {
	resp, err := l.svcCtx.Agent.GetConfig(l.ctx, in)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}
	return resp, nil
}
