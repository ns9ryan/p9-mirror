// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package node

import (
	"context"

	"oa.98ent.com/p9/node-dispatch/api/internal/svc"
	"oa.98ent.com/p9/node-dispatch/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetNodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetNodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNodeLogic {
	return &GetNodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetNodeLogic) GetNode(req *types.GetNodeRequest) (resp *types.GetNodeResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
