package memberlogic

import (
	"context"
	"oa.98ent.com/p9/common/xerr"

	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateBettingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateBettingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateBettingLogic {
	return &UpdateBettingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateBettingLogic) UpdateBetting(in *memberpb.UpdateMemberSwitchRequest) (*memberpb.MemberOperationResponse, error) {
	if in.Disabled == nil {
		return nil, xerr.RpcErr(xerr.BadRequest("common.invalidParam"))
	}
	if err := l.svcCtx.Member.SetBettingDisabled(l.ctx, in.GetMemberId(), in.GetDisabled()); err != nil {
		return nil, xerr.RpcErr(err)
	}

	return &memberpb.MemberOperationResponse{}, nil
}
