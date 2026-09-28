// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package operator_node

import (
	"context"

	"oa.98ent.com/p9/platform-operator/api/internal/svc"
	"oa.98ent.com/p9/platform-operator/api/internal/types"
	"oa.98ent.com/p9/platform-operator/rpc/pb/platformoperatorrpc/operatornodepb"

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

// SaveOperatorNode 保存分站部署节点
func (l *SaveOperatorNodeLogic) SaveOperatorNode(req *types.SaveOperatorNodeRequest) (resp *types.SaveOperatorNodeResponse, err error) {
	// 保存分站部署节点
	_, err = l.svcCtx.OperatorNodeRpc.Save(
		l.ctx,
		&operatornodepb.SaveOperatorNodeRequest{
			OperatorId: req.OperatorId, // 分站ID
			NodeCode:   req.NodeCode,   // 节点业务编码
		},
	)
	if err != nil {
		return nil, err
	}

	// 返回保存结果
	return &types.SaveOperatorNodeResponse{}, nil
}
