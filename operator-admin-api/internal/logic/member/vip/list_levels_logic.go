// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package vip

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/viplevelpb"
)

// ListLevelsLogic VIP等级列表逻辑
type ListLevelsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewListLevelsLogic 创建VIP等级列表逻辑
func NewListLevelsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLevelsLogic {
	return &ListLevelsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ListLevels 获取VIP等级列表
func (l *ListLevelsLogic) ListLevels(req *types.ListVipLevelsRequest) (resp *types.ListVipLevelsResponse, err error) {
	// 获取VIP等级分页列表
	rpcResp, err := l.svcCtx.MemberVipRpc.ListLevels(l.ctx, &viplevelpb.ListVipLevelsRequest{
		Page:     req.Page,     // 当前页码
		PageSize: req.PageSize, // 每页数量
	})
	if err != nil {
		return nil, err
	}

	// 转换VIP等级列表
	levels := rpcResp.GetList()
	list := make([]types.VipLevelInfo, 0, len(levels))
	for _, level := range levels {
		list = append(list, types.VipLevelInfo{
			Id:               level.GetId(),               // VIP等级ID
			Level:            level.GetLevel(),            // VIP等级编号
			UpgradeDeposit:   level.GetUpgradeDeposit(),   // 晋级所需存款
			UpgradeWater:     level.GetUpgradeWater(),     // 晋级所需洗码量
			RetentionDeposit: level.GetRetentionDeposit(), // 保级所需存款
			RetentionWater:   level.GetRetentionWater(),   // 保级所需洗码量
			UpgradeReward:    level.GetUpgradeReward(),    // 晋级奖励金额
			MemberCount:      level.GetMemberCount(),      // 当前会员数
			UpdatedAt:        level.GetUpdatedAt(),        // 更新时间
		})
	}

	return &types.ListVipLevelsResponse{
		Total: rpcResp.GetTotal(), // 数据总数
		List:  list,               // VIP等级列表
	}, nil
}
