package vipservicelogic

import (
	"context"

	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/vipconfigpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetConfigLogic {
	return &GetConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取VIP基础配置
func (l *GetConfigLogic) GetConfig(in *vipconfigpb.GetVipConfigRequest) (*vipconfigpb.GetVipConfigResponse, error) {
	// todo: add your logic here and delete this line

	return &vipconfigpb.GetVipConfigResponse{}, nil
}
