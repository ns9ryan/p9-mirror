// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package basic_resource_allocation

import (
	"context"

	"oa.98ent.com/p9/platform-operator/api/internal/svc"
	"oa.98ent.com/p9/platform-operator/api/internal/types"
	"oa.98ent.com/p9/platform-operator/rpc/pb/platformoperatorrpc/basicresourceallocationpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type PublishBasicResourceAllocationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPublishBasicResourceAllocationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishBasicResourceAllocationLogic {
	return &PublishBasicResourceAllocationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// PublishBasicResourceAllocation 发布基础资源分配到指定分站
func (l *PublishBasicResourceAllocationLogic) PublishBasicResourceAllocation(req *types.PublishBasicResourceAllocationRequest) (resp *types.PublishBasicResourceAllocationResponse, err error) {
	// 调用 platform-operator RPC 服务发布基础资源分配
	_, err = l.svcCtx.BasicResourceAllocationRpc.Publish(
		l.ctx,
		&basicresourceallocationpb.PublishBasicResourceAllocationRequest{
			OperatorCode: req.OpCode,
		},
	)
	if err != nil {
		return nil, err
	}

	return &types.PublishBasicResourceAllocationResponse{}, nil
}
