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

// GetNode 获取节点
func (l *GetNodeLogic) GetNode(req *types.GetNodeRequest) (resp *types.GetNodeResponse, err error) {
	// 获取节点
	result, err := l.svcCtx.NodeRpc.Get(
		l.ctx,
		&nodepb.GetNodeRequest{
			Id: req.Id, // 节点ID
		},
	)
	if err != nil {
		return nil, err
	}

	// 返回节点信息
	return &types.GetNodeResponse{
		NodeInfo: toNodeInfo(result.Node), // 节点信息
	}, nil
}
