package level

import (
	"context"

	agentconv "oa.98ent.com/p9/operator-admin-api/internal/logic/agent"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAgentLevelsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListAgentLevelsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAgentLevelsLogic {
	return &ListAgentLevelsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListAgentLevelsLogic) ListAgentLevels() (*types.ListAgentLevelsResp, error) {
	rpcResp, err := l.svcCtx.AgentRpc.ListAgentLevels(l.ctx, &agentpb.ListAgentLevelsReq{})
	if err != nil {
		return nil, err
	}
	return &types.ListAgentLevelsResp{
		Items: agentconv.LevelItemsFromPB(rpcResp.GetItems()),
	}, nil
}
