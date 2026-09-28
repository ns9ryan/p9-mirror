// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package operator

import (
	"context"

	"oa.98ent.com/p9/platform-operator/api/internal/svc"
	"oa.98ent.com/p9/platform-operator/api/internal/types"
	"oa.98ent.com/p9/platform-operator/rpc/pb/platformoperatorrpc/operatorpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type PublishOperatorLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPublishOperatorLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishOperatorLogic {
	return &PublishOperatorLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// PublishOperator 发布分站
func (l *PublishOperatorLogic) PublishOperator(req *types.PublishOperatorRequest) (resp *types.PublishOperatorResponse, err error) {
	// 发布分站
	_, err = l.svcCtx.OperatorRpc.Publish(
		l.ctx,
		&operatorpb.PublishOperatorRequest{
			Id: req.Id, // 分站ID
		},
	)
	if err != nil {
		return nil, err
	}

	// 返回发布结果
	return &types.PublishOperatorResponse{}, nil
}
