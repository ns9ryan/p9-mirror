package viplogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/vipicongrouppb"
)

// ListIconGroupsLogic VIP图标组列表逻辑
type ListIconGroupsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewListIconGroupsLogic 创建VIP图标组列表逻辑
func NewListIconGroupsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListIconGroupsLogic {
	return &ListIconGroupsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListIconGroups 获取VIP图标组列表
func (l *ListIconGroupsLogic) ListIconGroups(in *vipicongrouppb.ListVipIconGroupsRequest) (*vipicongrouppb.ListVipIconGroupsResponse, error) {
	// 获取VIP图标组列表
	groups, err := l.svcCtx.DAO.MemberVipIconGroup.List(l.ctx)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}

	// 组装图标组列表
	list := make([]*vipicongrouppb.VipIconGroupInfo, 0, len(groups))
	for _, group := range groups {
		startLevel := group.GroupNo * 10 // 起始VIP等级
		endLevel := startLevel + 9       // 结束VIP等级

		list = append(list, &vipicongrouppb.VipIconGroupInfo{
			Id:             group.ID,                    // 图标组ID
			GroupNo:        group.GroupNo,               // 图标组编号
			StartLevel:     startLevel,                  // 起始VIP等级
			EndLevel:       endLevel,                    // 结束VIP等级
			DefaultIconUrl: group.DefaultIconURL,        // 系统默认图标URL
			IconUrl:        group.IconURL,               // 当前使用的图标URL
			UpdatedAt:      group.UpdatedAt.UnixMilli(), // 更新时间
		})
	}

	return &vipicongrouppb.ListVipIconGroupsResponse{
		List: list, // VIP图标组列表
	}, nil
}
