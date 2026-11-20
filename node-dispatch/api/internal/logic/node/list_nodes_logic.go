// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package node

import (
	"context"

	"oa.98ent.com/p9/node-dispatch/api/internal/svc"
	"oa.98ent.com/p9/node-dispatch/api/internal/types"
	"oa.98ent.com/p9/node-dispatch/rpc/pb/nodedispatchrpc/nodepb"

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

// ListNodes 获取节点列表
func (l *ListNodesLogic) ListNodes(req *types.ListNodesRequest) (resp *types.ListNodesResponse, err error) {
	// 获取节点列表
	result, err := l.svcCtx.NodeRpc.List(
		l.ctx,
		&nodepb.ListNodesRequest{
			Page:     req.Page,     // 页码, 从1开始
			PageSize: req.PageSize, // 每页数量
			Keyword:  req.Keyword,  // 搜索关键字, 匹配节点编码或名称
			Status:   req.Status,   // 节点状态: 1启用, 2停用
			Online:   req.Online,   // 节点是否在线
		},
	)
	if err != nil {
		return nil, err
	}

	// 转换节点列表
	list := make([]types.NodeInfo, 0, len(result.List))
	for _, item := range result.List {
		list = append(list, toNodeInfo(item))
	}

	// 返回节点列表
	return &types.ListNodesResponse{
		Total: result.Total, // 数据总数
		List:  list,         // 节点列表
	}, nil
}
