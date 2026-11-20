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

type UpdateNodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateNodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateNodeLogic {
	return &UpdateNodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UpdateNode 修改节点
func (l *UpdateNodeLogic) UpdateNode(req *types.UpdateNodeRequest) (resp *types.UpdateNodeResponse, err error) {
	// 修改节点
	_, err = l.svcCtx.NodeRpc.Update(
		l.ctx,
		&nodepb.UpdateNodeRequest{
			Id:     req.Id,     // 节点ID
			Name:   req.Name,   // 节点名称
			Status: req.Status, // 节点状态: 1启用, 2停用
			Remark: req.Remark, // 运维备注
		},
	)
	if err != nil {
		return nil, err
	}

	// 返回修改结果
	return &types.UpdateNodeResponse{}, nil
}
