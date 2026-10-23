package viplogic

import (
	"context"

	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/viplevelpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListLevelsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListLevelsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLevelsLogic {
	return &ListLevelsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取VIP等级列表
func (l *ListLevelsLogic) ListLevels(in *viplevelpb.ListVipLevelsRequest) (*viplevelpb.ListVipLevelsResponse, error) {
	// todo: add your logic here and delete this line

	return &viplevelpb.ListVipLevelsResponse{}, nil
}
