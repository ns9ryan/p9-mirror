package agentservicelogic

import (
	"context"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateAgentLevelsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateAgentLevelsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAgentLevelsLogic {
	return &UpdateAgentLevelsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateAgentLevels 批量更新代理等级
func (l *UpdateAgentLevelsLogic) UpdateAgentLevels(in *agentpb.UpdateAgentLevelsReq) (*agentpb.UpdateAgentLevelsResp, error) {
	resp, err := l.svcCtx.Agent.UpdateLevels(l.ctx, in)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}
	return resp, nil
}
