package viplogic

import (
	"context"

	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/viplevelpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteLevelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteLevelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteLevelLogic {
	return &DeleteLevelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除VIP等级
func (l *DeleteLevelLogic) DeleteLevel(in *viplevelpb.DeleteVipLevelRequest) (*viplevelpb.DeleteVipLevelResponse, error) {
	// todo: add your logic here and delete this line

	return &viplevelpb.DeleteVipLevelResponse{}, nil
}
