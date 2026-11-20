// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package operator_game_allocation

import (
	"context"
	"strings"

	"oa.98ent.com/p9/platform-game/rpc/pb/platform_game"
	"oa.98ent.com/p9/platform-operator/api/internal/svc"
	"oa.98ent.com/p9/platform-operator/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PublishOperatorGameAllocationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewPublishOperatorGameAllocationLogic 发布游戏资源分配到指定分站
func NewPublishOperatorGameAllocationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishOperatorGameAllocationLogic {
	return &PublishOperatorGameAllocationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PublishOperatorGameAllocationLogic) PublishOperatorGameAllocation(req *types.PublishOperatorGameAllocationRequest) (resp *types.PublishOperatorGameAllocationResponse, err error) {
	// 校验请求参数
	opCode := strings.TrimSpace(req.OpCode)
	if opCode == "" {
		return nil, err
	}

	// 调用 platform-game RPC 服务发布游戏资源分配
	_, err = l.svcCtx.GameGrpcClient.GetOperatorGameAllocationServiceClient().PublishOperatorGameAllocation(l.ctx, &platform_game.PublishOperatorGameAllocationRequest{
		OpCode: opCode,
	})
	if err != nil {
		return nil, err
	}

	return &types.PublishOperatorGameAllocationResponse{}, nil
}
