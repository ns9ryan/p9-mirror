// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package operator_node

import (
	"context"

	"oa.98ent.com/p9/platform-operator/api/internal/svc"
	"oa.98ent.com/p9/platform-operator/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveOperatorNodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSaveOperatorNodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveOperatorNodeLogic {
	return &SaveOperatorNodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SaveOperatorNodeLogic) SaveOperatorNode(req *types.SaveOperatorNodeRequest) (resp *types.SaveOperatorNodeResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
