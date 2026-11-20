package memberlogic

import (
	"context"
	"oa.98ent.com/p9/common/xerr"

	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateStatusLogic {
	return &UpdateStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateStatusLogic) UpdateStatus(in *memberpb.UpdateMemberStatusRequest) (*memberpb.MemberOperationResponse, error) {
	if err := l.svcCtx.Member.UpdateStatus(l.ctx, in.GetMemberId(), in.GetStatus()); err != nil {
		return nil, xerr.RpcErr(err)
	}

	return &memberpb.MemberOperationResponse{}, nil
}
