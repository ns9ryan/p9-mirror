// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package vip

import (
	"context"

	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteLevelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteLevelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteLevelLogic {
	return &DeleteLevelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteLevelLogic) DeleteLevel(req *types.DeleteVipLevelRequest) error {
	// todo: add your logic here and delete this line

	return nil
}
