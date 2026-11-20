// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package manage

import (
	"context"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"

	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateBettingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateBettingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateBettingLogic {
	return &UpdateBettingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateBettingLogic) UpdateBetting(req *types.UpdateMemberSwitchRequest) (resp *types.EmptyResponse, err error) {
	_, err = l.svcCtx.MemberRpc.UpdateBetting(l.ctx, &memberpb.UpdateMemberSwitchRequest{MemberId: req.MemberId, Disabled: req.Disabled})
	if err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}
