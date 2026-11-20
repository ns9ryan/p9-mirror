package memberlogic

import (
	"context"
	"oa.98ent.com/p9/common/xerr"

	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateRemarkLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateRemarkLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateRemarkLogic {
	return &UpdateRemarkLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateRemarkLogic) UpdateRemark(in *memberpb.UpdateMemberRemarkRequest) (*memberpb.MemberOperationResponse, error) {
	if err := l.svcCtx.Member.UpdateRemark(l.ctx, in.GetMemberId(), in.GetRemark()); err != nil {
		return nil, xerr.RpcErr(err)
	}

	return &memberpb.MemberOperationResponse{}, nil
}
