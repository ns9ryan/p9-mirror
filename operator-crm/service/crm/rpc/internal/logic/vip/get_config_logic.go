package viplogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/vipconfigpb"
)

// GetConfigLogic VIP基础配置查询逻辑
type GetConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewGetConfigLogic 创建VIP基础配置查询逻辑
func NewGetConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetConfigLogic {
	return &GetConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetConfig 获取VIP基础配置
func (l *GetConfigLogic) GetConfig(in *vipconfigpb.GetVipConfigRequest) (*vipconfigpb.GetVipConfigResponse, error) {
	// 查询当前厅VIP基础配置
	config, err := l.svcCtx.DAO.MemberVipConfig.Get(l.ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			// 配置不存在, 请检查分站初始化状态
			return nil, xerr.RpcErr(xerr.EntNotFound(i18n.InvalidParam, err))
		}
		return nil, xerr.RpcErr(err)
	}

	// 转换最近一次开启时间
	var lastEnabledAt *int64
	if config.LastEnabledAt != nil {
		timestamp := config.LastEnabledAt.UnixMilli()
		lastEnabledAt = &timestamp
	}

	// 组装VIP基础配置
	return &vipconfigpb.GetVipConfigResponse{
		Config: &vipconfigpb.VipConfigInfo{
			VipEnabled:            config.VipEnabled,            // VIP总开关
			UpgradeRewardEnabled:  config.UpgradeRewardEnabled,  // 晋级奖励开关
			RetentionCheckEnabled: config.RetentionCheckEnabled, // 保级考核开关
			RewardGrantMode:       config.RewardGrantMode,       // 奖励派发模式
			RewardWaterMultiplier: config.RewardWaterMultiplier, // 晋级奖励洗码倍率
			LastEnabledAt:         lastEnabledAt,                // 最近一次开启时间
			UpdatedAt:             config.UpdatedAt.UnixMilli(), // 更新时间
		},
	}, nil
}
