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
	// 通过代理RPC获取代理个人信息
	rpcResp, err := l.svcCtx.AgentRpc.GetAgentProfile(l.ctx, &agentpb.GetAgentProfileReq{
		AgentId: req.AgentId,
	})
	if err != nil {
		logx.Errorw("get agent profile failed", logx.Field("error", err))
		return nil, err
	}
	// 转换为API响应格式
	out := agentconv.ProfileFromPB(rpcResp)

	return &out, nil
}
