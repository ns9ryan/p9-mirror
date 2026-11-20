// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package vip

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/vipconfigpb"
)

// UpdateConfigLogic VIP基础配置修改逻辑
type UpdateConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewUpdateConfigLogic 创建VIP基础配置修改逻辑
func NewUpdateConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateConfigLogic {
	return &UpdateConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UpdateConfig 修改VIP基础配置
func (l *UpdateConfigLogic) UpdateConfig(req *types.UpdateVipConfigRequest) (resp *types.EmptyResponse, err error) {
	// 修改VIP基础配置
	_, err = l.svcCtx.MemberVipRpc.UpdateConfig(l.ctx, &vipconfigpb.UpdateVipConfigRequest{
		VipEnabled:            req.VipEnabled,            // VIP总开关
		UpgradeRewardEnabled:  req.UpgradeRewardEnabled,  // 晋级奖励开关
		RetentionCheckEnabled: req.RetentionCheckEnabled, // 保级考核开关
		RewardGrantMode:       req.RewardGrantMode,       // 奖励派发模式
		RewardWaterMultiplier: req.RewardWaterMultiplier, // 晋级奖励洗码倍率
	})
	if err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}
