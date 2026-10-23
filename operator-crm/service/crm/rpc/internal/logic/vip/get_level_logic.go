package viplogic

import (
	"context"

	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/viplevelpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLevelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetLevelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLevelLogic {
	return &GetLevelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取VIP等级
func (l *GetLevelLogic) GetLevel(in *viplevelpb.GetVipLevelRequest) (*viplevelpb.GetVipLevelResponse, error) {
	// todo: add your logic here and delete this line

	return &viplevelpb.GetVipLevelResponse{}, nil
}
