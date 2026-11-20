package agentservicelogic

import (
	"context"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAgentLoginAssociationsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListAgentLoginAssociationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAgentLoginAssociationsLogic {
	return &ListAgentLoginAssociationsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListAgentLoginAssociations 关联检索登录日志
func (l *ListAgentLoginAssociationsLogic) ListAgentLoginAssociations(in *agentpb.ListAgentLoginAssociationsReq) (*agentpb.ListAgentLoginAssociationsResp, error) {
	resp, err := l.svcCtx.Agent.ListAgentLoginAssociations(l.ctx, in)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}
	return resp, nil
}
