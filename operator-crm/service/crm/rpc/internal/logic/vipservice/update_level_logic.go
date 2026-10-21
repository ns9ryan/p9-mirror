package vipservicelogic

import (
	"context"

	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/viplevelpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateLevelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateLevelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateLevelLogic {
	return &UpdateLevelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 修改VIP等级
func (l *UpdateLevelLogic) UpdateLevel(in *viplevelpb.UpdateVipLevelRequest) (*viplevelpb.UpdateVipLevelResponse, error) {
	// todo: add your logic here and delete this line

	return &viplevelpb.UpdateVipLevelResponse{}, nil
}
