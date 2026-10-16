package config

import (
	"context"

	agentconv "oa.98ent.com/p9/operator-admin-api/internal/logic/agent"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAgentConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAgentConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAgentConfigLogic {
	return &GetAgentConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAgentConfigLogic) GetAgentConfig() (*types.AgentConfig, error) {
	rpcResp, err := l.svcCtx.AgentRpc.GetAgentConfig(l.ctx, &agentpb.GetAgentConfigReq{})
	if err != nil {
		return nil, err
	}
	cfg := agentconv.ConfigFromPB(rpcResp)
	return &cfg, nil
}
