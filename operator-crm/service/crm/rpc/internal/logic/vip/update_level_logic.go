package viplogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/dao"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/viplevelpb"
)

// UpdateLevelLogic VIP等级修改逻辑
type UpdateLevelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewUpdateLevelLogic 创建VIP等级修改逻辑
func NewUpdateLevelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateLevelLogic {
	return &UpdateLevelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateLevel 修改VIP等级
func (l *UpdateLevelLogic) UpdateLevel(in *viplevelpb.UpdateVipLevelRequest) (*viplevelpb.UpdateVipLevelResponse, error) {
	if in == nil || in.GetId() <= 0 {
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
	}

	// 至少提供一个待修改字段
	hasUpdate := in.UpgradeDeposit != nil || in.UpgradeWater != nil ||
		in.RetentionDeposit != nil || in.RetentionWater != nil ||
		in.UpgradeReward != nil
	if !hasUpdate {
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
	}

	// 查询当前等级
	level, err := l.svcCtx.DAO.MemberVipLevel.Get(l.ctx, in.GetId())
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, xerr.RpcErr(xerr.EntNotFound(i18n.InvalidParam, err))
		}
		return nil, xerr.RpcErr(err)
	}

	// VIP0的晋级、保级及奖励条件不可修改
	if level.Level == 0 {
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
	}

	// 检查晋级和保级条件是否实际发生变化
	upgradeDepositChanged := in.UpgradeDeposit != nil && *in.UpgradeDeposit != level.UpgradeDeposit         // 晋级存款条件是否变化
	upgradeWaterChanged := in.UpgradeWater != nil && *in.UpgradeWater != level.UpgradeWater                 // 晋级洗码条件是否变化
	retentionDepositChanged := in.RetentionDeposit != nil && *in.RetentionDeposit != level.RetentionDeposit // 保级存款条件是否变化
	retentionWaterChanged := in.RetentionWater != nil && *in.RetentionWater != level.RetentionWater         // 保级洗码条件是否变化

	conditionChanged := upgradeDepositChanged || upgradeWaterChanged || retentionDepositChanged || retentionWaterChanged

	if conditionChanged {
		// 查询当前等级关联的会员数
		memberCounts, err := l.svcCtx.DAO.MemberVipLevel.MemberCounts(l.ctx, []int64{level.ID})
		if err != nil {
			return nil, xerr.RpcErr(err)
		}

		// TODO(VIP): 会员等级重新计算实现后, 支持修改有关联会员的晋级、保级条件
		if memberCounts[level.ID] > 0 {
			return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
		}
	}

	// 更新VIP等级配置
	if err := l.svcCtx.DAO.MemberVipLevel.Update(l.ctx, level.ID, dao.MemberVipLevelUpdate{
		UpgradeDeposit:   in.UpgradeDeposit,   // 晋级所需存款金额
		UpgradeWater:     in.UpgradeWater,     // 晋级所需洗码量
		RetentionDeposit: in.RetentionDeposit, // 保级所需存款金额
		RetentionWater:   in.RetentionWater,   // 保级所需洗码量
		UpgradeReward:    in.UpgradeReward,    // 晋级奖励金额
	}); err != nil {
		switch {
		case ent.IsNotFound(err):
			// 等级不存在或不属于当前厅
			return nil, xerr.RpcErr(xerr.EntNotFound(i18n.InvalidParam, err))
		case ent.IsConstraintError(err):
			// 数据库约束冲突, 如唯一约束冲突
			return nil, xerr.RpcErr(xerr.EntConstraintError(i18n.InvalidParam, err))
		case ent.IsValidationError(err):
			// Ent字段校验失败
			return nil, xerr.RpcErr(xerr.EntValidationError(i18n.InvalidParam, err))
		default:
			// 其他未分类错误
			return nil, xerr.RpcErr(err)
		}
	}

	return &viplevelpb.UpdateVipLevelResponse{}, nil
}
