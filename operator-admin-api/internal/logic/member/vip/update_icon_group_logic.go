// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package vip

import (
	"context"

	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateIconGroupLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateIconGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateIconGroupLogic {
	return &UpdateIconGroupLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateIconGroupLogic) UpdateIconGroup(req *types.UpdateVipIconGroupRequest) (resp *types.EmptyResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
