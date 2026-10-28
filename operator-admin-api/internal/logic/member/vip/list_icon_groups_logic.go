// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package vip

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/vipicongrouppb"
)

// ListIconGroupsLogic VIP图标组列表逻辑
type ListIconGroupsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewListIconGroupsLogic 创建VIP图标组列表逻辑
func NewListIconGroupsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListIconGroupsLogic {
	return &ListIconGroupsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ListIconGroups 获取VIP图标组列表
func (l *ListIconGroupsLogic) ListIconGroups(req *types.ListVipIconGroupsRequest) (resp *types.ListVipIconGroupsResponse, err error) {
	// 获取VIP图标组列表
	rpcResp, err := l.svcCtx.MemberVipRpc.ListIconGroups(l.ctx, &vipicongrouppb.ListVipIconGroupsRequest{})
	if err != nil {
		return nil, err
	}

	// 转换图标组列表
	list := make([]types.VipIconGroupInfo, 0, len(rpcResp.GetList()))
	for _, group := range rpcResp.GetList() {
		list = append(list, types.VipIconGroupInfo{
			Id:             group.GetId(),             // 图标组ID
			GroupNo:        group.GetGroupNo(),        // 图标组编号
			StartLevel:     group.GetStartLevel(),     // 起始VIP等级
			EndLevel:       group.GetEndLevel(),       // 结束VIP等级
			DefaultIconUrl: group.GetDefaultIconUrl(), // 系统默认图标URL
			IconUrl:        group.GetIconUrl(),        // 当前使用的图标URL
			UpdatedAt:      group.GetUpdatedAt(),      // 更新时间
		})
	}

	return &types.ListVipIconGroupsResponse{
		List: list, // VIP图标组列表
	}, nil
}
