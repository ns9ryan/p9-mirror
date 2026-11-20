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

// GetLevelLogic VIP等级详情逻辑
type GetLevelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetLevelLogic 创建VIP等级详情逻辑
func NewGetLevelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLevelLogic {
	return &GetLevelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetLevel 获取VIP等级
func (l *GetLevelLogic) GetLevel(req *types.GetVipLevelRequest) (resp *types.GetVipLevelResponse, err error) {
	// 查询VIP等级详情
	rpcResp, err := l.svcCtx.MemberVipRpc.GetLevel(l.ctx, &viplevelpb.GetVipLevelRequest{
		Id: req.Id, // VIP等级ID
	})
	if err != nil {
		return nil, err
	}

	level := rpcResp.GetLevel()

	// 转换VIP等级详情
	return &types.GetVipLevelResponse{
		VipLevelInfo: types.VipLevelInfo{
			Id:               level.GetId(),               // VIP等级ID
			Level:            level.GetLevel(),            // VIP等级编号
			UpgradeDeposit:   level.GetUpgradeDeposit(),   // 晋级所需存款
			UpgradeWater:     level.GetUpgradeWater(),     // 晋级所需洗码量
			RetentionDeposit: level.GetRetentionDeposit(), // 保级所需存款
			RetentionWater:   level.GetRetentionWater(),   // 保级所需洗码量
			UpgradeReward:    level.GetUpgradeReward(),    // 晋级奖励金额
			MemberCount:      level.GetMemberCount(),      // 当前会员数
			UpdatedAt:        level.GetUpdatedAt(),        // 更新时间
		},
	}, nil
}
