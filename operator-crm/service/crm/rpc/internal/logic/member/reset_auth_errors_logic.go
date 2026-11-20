package memberlogic

import (
	"context"
	"oa.98ent.com/p9/common/xerr"

	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ResetAuthErrorsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewResetAuthErrorsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetAuthErrorsLogic {
	return &ResetAuthErrorsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ResetAuthErrorsLogic) ResetAuthErrors(in *memberpb.ResetMemberAuthErrorsRequest) (*memberpb.MemberOperationResponse, error) {
	if err := l.svcCtx.Member.ResetAuthErrors(l.ctx, in.GetMemberId(), in.GetField()); err != nil {
		return nil, xerr.RpcErr(err)
	}

	return &memberpb.MemberOperationResponse{}, nil
}
