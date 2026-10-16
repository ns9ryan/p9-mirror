package agentservicelogic

import (
	"context"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchAgentsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSearchAgentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchAgentsLogic {
	return &SearchAgentsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SearchAgents 按 ID/账号搜索代理
func (l *SearchAgentsLogic) SearchAgents(in *agentpb.SearchAgentsReq) (*agentpb.SearchAgentsResp, error) {
	resp, err := l.svcCtx.Agent.SearchAgents(l.ctx, in)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}
	return resp, nil
}
