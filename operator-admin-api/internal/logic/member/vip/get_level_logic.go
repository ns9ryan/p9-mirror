// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package vip

import (
	"context"

	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLevelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetLevelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLevelLogic {
	return &GetLevelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetLevelLogic) GetLevel(req *types.GetVipLevelRequest) (resp *types.GetVipLevelResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
