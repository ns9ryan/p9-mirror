package level

import (
	"context"

	agentconv "oa.98ent.com/p9/operator-admin-api/internal/logic/agent"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateAgentLevelsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateAgentLevelsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAgentLevelsLogic {
	return &UpdateAgentLevelsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateAgentLevelsLogic) UpdateAgentLevels(req *types.UpdateAgentLevelsReq) (*types.UpdateAgentLevelsResp, error) {
	if req == nil {
		req = &types.UpdateAgentLevelsReq{}
	}
	rpcResp, err := l.svcCtx.AgentRpc.UpdateAgentLevels(l.ctx, &agentpb.UpdateAgentLevelsReq{
		List: agentconv.LevelItemsToPB(req.List),
	})
	if err != nil {
		return nil, err
	}
	return &types.UpdateAgentLevelsResp{
		List: agentconv.LevelItemsFromPB(rpcResp.GetList()),
	}, nil
}
