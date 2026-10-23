package viplogic

import (
	"context"

	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/vipconfigpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateConfigLogic {
	return &UpdateConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 修改VIP基础配置
func (l *UpdateConfigLogic) UpdateConfig(in *vipconfigpb.UpdateVipConfigRequest) (*vipconfigpb.UpdateVipConfigResponse, error) {
	// todo: add your logic here and delete this line

	return &vipconfigpb.UpdateVipConfigResponse{}, nil
}
