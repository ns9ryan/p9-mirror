// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package node

import (
	"context"

	"oa.98ent.com/p9/node-dispatch/api/internal/svc"
	"oa.98ent.com/p9/node-dispatch/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListNodesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListNodesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListNodesLogic {
	return &ListNodesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListNodesLogic) ListNodes(req *types.ListNodesRequest) (resp *types.ListNodesResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
