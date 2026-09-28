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

type SyncPublishStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSyncPublishStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncPublishStatusLogic {
	return &SyncPublishStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// SyncPublishStatus 同步分站发布状态
func (l *SyncPublishStatusLogic) SyncPublishStatus(req *types.SyncPublishStatusRequest) (resp *types.SyncPublishStatusResponse, err error) {
	// 同步分站发布状态
	result, err := l.svcCtx.OperatorRpc.SyncPublishStatus(
		l.ctx,
		&operatorpb.SyncPublishStatusRequest{
			Id: req.Id, // 分站ID
		},
	)
	if err != nil {
		return nil, err
	}

	// 返回同步结果
	return &types.SyncPublishStatusResponse{
		OperatorId:    result.OperatorId,    // 分站ID
		OperatorCode:  result.OperatorCode,  // 分站业务编码
		PublishStatus: result.PublishStatus, // 发布状态: 2发布中, 3已发布, 4发布失败
	}, nil
}
