package profile

import (
	"context"

	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ResetLoginPasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewResetLoginPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetLoginPasswordLogic {
	return &ResetLoginPasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ResetLoginPasswordLogic) ResetLoginPassword(req *types.ResetLoginPasswordReq) (*types.ResetLoginPasswordResp, error) {
	if req == nil {
		req = &types.ResetLoginPasswordReq{}
	}
	rpcResp, err := l.svcCtx.AgentRpc.ResetLoginPassword(l.ctx, &agentpb.ResetLoginPasswordReq{
		AgentId: req.AgentId,
	})
	if err != nil {
		return nil, err
	}
	return &types.ResetLoginPasswordResp{Password: rpcResp.GetPassword()}, nil
}
