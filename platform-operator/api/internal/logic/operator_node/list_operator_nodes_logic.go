// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package operator_node

import (
	"context"

	"oa.98ent.com/p9/node-dispatch/rpc/pb/nodedispatchrpc/nodepb"
	"oa.98ent.com/p9/platform-operator/api/internal/svc"
	"oa.98ent.com/p9/platform-operator/api/internal/types"
	"oa.98ent.com/p9/platform-operator/rpc/pb/platformoperatorrpc/operatornodepb"

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

// ListOperatorNodes 获取分站部署节点列表
func (l *ListOperatorNodesLogic) ListOperatorNodes(req *types.ListOperatorNodesRequest) (resp *types.ListOperatorNodesResponse, err error) {
	// 获取分站当前部署节点
	operatorNodeResult, err := l.svcCtx.OperatorNodeRpc.Get(
		l.ctx,
		&operatornodepb.GetOperatorNodeRequest{
			OperatorId: req.OperatorId, // 分站ID
		},
	)
	if err != nil {
		return nil, err
	}

	// 获取当前已选择的节点编码
	selectedNodeCode := ""
	if operatorNodeResult.OperatorNode != nil {
		selectedNodeCode = operatorNodeResult.OperatorNode.NodeCode
	}

	// 获取节点列表
	nodeResult, err := l.svcCtx.NodeDispatchNodeRpc.List(
		l.ctx,
		&nodepb.ListNodesRequest{
			Page:     req.Page,     // 页码, 从1开始
			PageSize: req.PageSize, // 每页数量
			Keyword:  req.Keyword,  // 搜索关键字, 匹配节点编码或名称
		},
	)
	if err != nil {
		return nil, err
	}

	// 组合节点列表和当前部署关系
	list := make([]types.OperatorNodeInfo, 0, len(nodeResult.List))
	for _, item := range nodeResult.List {
		list = append(list, toOperatorNodeInfo(item, selectedNodeCode))
	}

	// 返回分站部署节点列表
	return &types.ListOperatorNodesResponse{
		Total: nodeResult.Total, // 数据总数
		List:  list,             // 部署节点列表
	}, nil
}
