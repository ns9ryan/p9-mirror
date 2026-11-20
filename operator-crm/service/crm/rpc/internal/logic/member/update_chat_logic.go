package memberlogic

import (
	"context"
	"oa.98ent.com/p9/common/xerr"

	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateChatLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateChatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateChatLogic {
	return &UpdateChatLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateChatLogic) UpdateChat(in *memberpb.UpdateMemberSwitchRequest) (*memberpb.MemberOperationResponse, error) {
	if in.Disabled == nil {
		return nil, xerr.RpcErr(xerr.BadRequest("common.invalidParam"))
	}
	if err := l.svcCtx.Member.SetChatDisabled(l.ctx, in.GetMemberId(), in.GetDisabled()); err != nil {
		return nil, xerr.RpcErr(err)
	}

	return &memberpb.MemberOperationResponse{}, nil
}
