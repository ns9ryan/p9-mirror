// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package operator_node

import (
	"context"

	"oa.98ent.com/p9/platform-operator/api/internal/svc"
	"oa.98ent.com/p9/platform-operator/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListOperatorNodesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListOperatorNodesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListOperatorNodesLogic {
	return &ListOperatorNodesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListOperatorNodesLogic) ListOperatorNodes(req *types.ListOperatorNodesRequest) (resp *types.ListOperatorNodesResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
