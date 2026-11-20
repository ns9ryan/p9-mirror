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

type UpdateWithdrawalLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateWithdrawalLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateWithdrawalLogic {
	return &UpdateWithdrawalLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateWithdrawalLogic) UpdateWithdrawal(req *types.UpdateMemberSwitchRequest) (resp *types.EmptyResponse, err error) {
	_, err = l.svcCtx.MemberRpc.UpdateWithdrawal(l.ctx, &memberpb.UpdateMemberSwitchRequest{MemberId: req.MemberId, Disabled: req.Disabled})
	if err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}
