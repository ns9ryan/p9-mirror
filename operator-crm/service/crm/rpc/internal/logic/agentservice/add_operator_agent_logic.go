package agentservicelogic

import (
	"context"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddOperatorAgentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddOperatorAgentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddOperatorAgentLogic {
	return &AddOperatorAgentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AddOperatorAgent 分站添加代理
func (l *AddOperatorAgentLogic) AddOperatorAgent(in *agentpb.AddOperatorAgentReq) (*agentpb.AddAgentResp, error) {
	resp, err := l.svcCtx.Agent.AddOperatorAgent(l.ctx, in)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}
	return resp, nil
}
