package viplogic

import (
	"context"

	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/vipicongrouppb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateIconGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateIconGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateIconGroupLogic {
	return &UpdateIconGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 修改VIP图标
func (l *UpdateIconGroupLogic) UpdateIconGroup(in *vipicongrouppb.UpdateVipIconGroupRequest) (*vipicongrouppb.UpdateVipIconGroupResponse, error) {
	// todo: add your logic here and delete this line

	return &vipicongrouppb.UpdateVipIconGroupResponse{}, nil
}
