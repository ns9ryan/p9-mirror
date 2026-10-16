package agentservicelogic

import (
	"context"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddSubAgentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddSubAgentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddSubAgentLogic {
	return &AddSubAgentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AddSubAgent 代理添加下级
func (l *AddSubAgentLogic) AddSubAgent(in *agentpb.AddSubAgentReq) (*agentpb.AddAgentResp, error) {
	resp, err := l.svcCtx.Agent.AddSubAgent(l.ctx, in)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}
	return resp, nil
}
