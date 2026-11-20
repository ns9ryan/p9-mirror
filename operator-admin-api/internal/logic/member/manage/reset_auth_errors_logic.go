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

type ResetAuthErrorsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewResetAuthErrorsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetAuthErrorsLogic {
	return &ResetAuthErrorsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ResetAuthErrorsLogic) ResetAuthErrors(req *types.ResetMemberAuthErrorsRequest) (resp *types.EmptyResponse, err error) {
	_, err = l.svcCtx.MemberRpc.ResetAuthErrors(l.ctx, &memberpb.ResetMemberAuthErrorsRequest{MemberId: req.MemberId, Field: req.Field})
	if err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}
