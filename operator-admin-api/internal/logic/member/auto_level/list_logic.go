// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package auto_level

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
)

// ListLogic 自动层级列表逻辑
type ListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLogic {
	return &ListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// List 获取自动层级列表
func (l *ListLogic) List(req *types.ListMemberAutoLevelsRequest) (resp *types.ListMemberAutoLevelsResponse, err error) {

	return
}
