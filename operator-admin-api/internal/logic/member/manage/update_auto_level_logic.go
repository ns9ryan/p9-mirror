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

type UpdateAutoLevelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateAutoLevelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAutoLevelLogic {
	return &UpdateAutoLevelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateAutoLevelLogic) UpdateAutoLevel(req *types.UpdateMemberAutoLevelRequest) (resp *types.EmptyResponse, err error) {
	_, err = l.svcCtx.MemberRpc.UpdateAutoLevel(l.ctx, &memberpb.UpdateMemberAutoLevelRequest{MemberId: req.MemberId, AutoLevelId: req.AutoLevelId})
	if err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}
