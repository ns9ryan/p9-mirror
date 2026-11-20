package agentservicelogic

import (
	"context"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddAgentApplyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddAgentApplyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddAgentApplyLogic {
	return &AddAgentApplyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AddAgentApply 申请代理添加
func (l *AddAgentApplyLogic) AddAgentApply(in *agentpb.AddAgentApplyReq) (*agentpb.AgentApplyItem, error) {
	resp, err := l.svcCtx.Agent.AddAgentApply(l.ctx, in)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}
	return resp, nil
}
