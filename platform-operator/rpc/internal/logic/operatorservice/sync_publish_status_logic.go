package operatorservicelogic

import (
	"context"

	"oa.98ent.com/p9/platform-operator/rpc/internal/svc"
	"oa.98ent.com/p9/platform-operator/rpc/pb/platformoperatorrpc/operatorpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type SyncPublishStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSyncPublishStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncPublishStatusLogic {
	return &SyncPublishStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 同步分站发布状态
func (l *SyncPublishStatusLogic) SyncPublishStatus(in *operatorpb.SyncPublishStatusRequest) (*operatorpb.SyncPublishStatusResponse, error) {
	// todo: add your logic here and delete this line

	return &operatorpb.SyncPublishStatusResponse{}, nil
}
