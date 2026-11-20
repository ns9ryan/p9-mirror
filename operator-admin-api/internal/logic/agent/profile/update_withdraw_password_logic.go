package profile

import (
	"context"

	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateWithdrawPasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateWithdrawPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateWithdrawPasswordLogic {
	return &UpdateWithdrawPasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateWithdrawPasswordLogic) UpdateWithdrawPassword(req *types.UpdateWithdrawPasswordReq) (*types.AgentEmpty, error) {
	if req == nil {
		req = &types.UpdateWithdrawPasswordReq{}
	}
	_, err := l.svcCtx.AgentRpc.UpdateWithdrawPassword(l.ctx, &agentpb.UpdateWithdrawPasswordReq{
		AgentId:  req.AgentId,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}
	return &types.AgentEmpty{}, nil
}
