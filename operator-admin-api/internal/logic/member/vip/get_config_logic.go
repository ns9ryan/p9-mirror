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

// GetConfigLogic VIP基础配置查询逻辑
type GetConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetConfigLogic 创建VIP基础配置查询逻辑
func NewGetConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetConfigLogic {
	return &GetConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetConfig 获取VIP基础配置
func (l *GetConfigLogic) GetConfig(req *types.GetVipConfigRequest) (resp *types.GetVipConfigResponse, err error) {
	// 查询VIP基础配置
	rpcResp, err := l.svcCtx.MemberVipRpc.GetConfig(l.ctx, &vipconfigpb.GetVipConfigRequest{})
	if err != nil {
		return nil, err
	}

	config := rpcResp.GetConfig()

	// 转换VIP基础配置
	return &types.GetVipConfigResponse{
		VipConfigInfo: types.VipConfigInfo{
			VipEnabled:            config.GetVipEnabled(),            // VIP总开关
			UpgradeRewardEnabled:  config.GetUpgradeRewardEnabled(),  // 晋级奖励开关
			RetentionCheckEnabled: config.GetRetentionCheckEnabled(), // 保级考核开关
			RewardGrantMode:       config.GetRewardGrantMode(),       // 奖励派发模式
			RewardWaterMultiplier: config.GetRewardWaterMultiplier(), // 晋级奖励洗码倍率
			LastEnabledAt:         config.LastEnabledAt,              // 最近一次开启时间
			UpdatedAt:             config.GetUpdatedAt(),             // 更新时间
		},
	}, nil
}
