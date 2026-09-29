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

type GetOperatorNodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetOperatorNodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOperatorNodeLogic {
	return &GetOperatorNodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetOperatorNode 获取分站部署节点
func (l *GetOperatorNodeLogic) GetOperatorNode(req *types.GetOperatorNodeRequest) (resp *types.GetOperatorNodeResponse, err error) {
	// 获取分站部署节点
	result, err := l.svcCtx.OperatorNodeRpc.Get(
		l.ctx,
		&operatornodepb.GetOperatorNodeRequest{
			OperatorId: req.OperatorId, // 分站ID
		},
	)
	if err != nil {
		return nil, err
	}

	// 返回分站部署节点
	return &types.GetOperatorNodeResponse{
		OperatorNode: toOperatorNodeInfo(result.OperatorNode), // 当前部署节点, 尚未选择时为空
	}, nil
}
