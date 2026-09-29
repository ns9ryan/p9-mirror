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

type CreateNodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateNodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateNodeLogic {
	return &CreateNodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// CreateNode 创建节点
func (l *CreateNodeLogic) CreateNode(req *types.CreateNodeRequest) (resp *types.CreateNodeResponse, err error) {
	// 创建节点
	result, err := l.svcCtx.NodeRpc.Create(
		l.ctx,
		&nodepb.CreateNodeRequest{
			Name:   req.Name,   // 节点名称
			Status: req.Status, // 节点状态: 1启用, 2停用
			Remark: req.Remark, // 运维备注
		},
	)
	if err != nil {
		return nil, err
	}

	// 返回创建结果
	return &types.CreateNodeResponse{
		Id:         result.Id,         // 节点ID
		Code:       result.Code,       // 节点业务编码
		AuthSecret: result.AuthSecret, // 节点认证密钥, 仅本次返回
	}, nil
}
