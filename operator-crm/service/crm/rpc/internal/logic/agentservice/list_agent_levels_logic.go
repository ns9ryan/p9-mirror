package agentservicelogic

import (
	"context"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAgentLevelsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListAgentLevelsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAgentLevelsLogic {
	return &ListAgentLevelsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListAgentLevels 获取代理等级列表
func (l *ListAgentLevelsLogic) ListAgentLevels(in *agentpb.ListAgentLevelsReq) (*agentpb.ListAgentLevelsResp, error) {
	resp, err := l.svcCtx.Agent.ListLevels(l.ctx, in)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}
	return resp, nil
}
