package viplogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/dao"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/vipconfigpb"
)

// UpdateConfigLogic VIP基础配置修改逻辑
type UpdateConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewUpdateConfigLogic 创建VIP基础配置修改逻辑
func NewUpdateConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateConfigLogic {
	return &UpdateConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateConfig 修改VIP基础配置
func (l *UpdateConfigLogic) UpdateConfig(in *vipconfigpb.UpdateVipConfigRequest) (*vipconfigpb.UpdateVipConfigResponse, error) {
	// 没有待修改字段时直接返回成功
	if in.VipEnabled == nil && in.UpgradeRewardEnabled == nil &&
		in.RetentionCheckEnabled == nil && in.RewardGrantMode == nil &&
		in.RewardWaterMultiplier == nil {
		return &vipconfigpb.UpdateVipConfigResponse{}, nil
	}

	// 修改VIP基础配置, 最近开启时间由DAO维护
	if err := l.svcCtx.DAO.MemberVipConfig.UpdateConfig(l.ctx, dao.MemberVipConfigUpdate{
		VipEnabled:            in.VipEnabled,            // VIP总开关
		UpgradeRewardEnabled:  in.UpgradeRewardEnabled,  // 晋级奖励开关
		RetentionCheckEnabled: in.RetentionCheckEnabled, // 保级考核开关
		RewardGrantMode:       in.RewardGrantMode,       // 奖励派发模式
		RewardWaterMultiplier: in.RewardWaterMultiplier, // 晋级奖励洗码倍率
	}); err != nil {
		switch {
		case ent.IsNotFound(err):
			// VIP基础配置不存在, 请检查分站初始化状态
			return nil, xerr.RpcErr(xerr.EntNotFound(i18n.InvalidParam, err))
		case ent.IsConstraintError(err):
			// 数据库约束冲突
			return nil, xerr.RpcErr(xerr.EntConstraintError(i18n.InvalidParam, err))
		case ent.IsValidationError(err):
			// Ent字段校验失败
			return nil, xerr.RpcErr(xerr.EntValidationError(i18n.InvalidParam, err))
		default:
			// 其他未分类错误
			return nil, xerr.RpcErr(err)
		}
	}

	return &vipconfigpb.UpdateVipConfigResponse{}, nil
}
