package profile

import (
	"context"

	agentconv "oa.98ent.com/p9/operator-admin-api/internal/logic/agent"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAgentProfileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAgentProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAgentProfileLogic {
	return &GetAgentProfileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAgentProfileLogic) GetAgentProfile(req *types.GetAgentProfileReq) (*types.AgentProfile, error) {
	if req == nil {
		req = &types.GetAgentProfileReq{}
	}
	rpcResp, err := l.svcCtx.AgentRpc.GetAgentProfile(l.ctx, &agentpb.GetAgentProfileReq{
		AgentId: req.AgentId,
	})
	if err != nil {
		return nil, err
	}
	out := agentconv.ProfileFromPB(rpcResp)
	return &out, nil
}
