package vipservicelogic

import (
	"context"

	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/vipicongrouppb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListIconGroupsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListIconGroupsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListIconGroupsLogic {
	return &ListIconGroupsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取VIP图标组列表
func (l *ListIconGroupsLogic) ListIconGroups(in *vipicongrouppb.ListVipIconGroupsRequest) (*vipicongrouppb.ListVipIconGroupsResponse, error) {
	// todo: add your logic here and delete this line

	return &vipicongrouppb.ListVipIconGroupsResponse{}, nil
}
