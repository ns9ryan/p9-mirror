package viplogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/viplevelpb"
)

// GetLevelLogic VIP等级详情逻辑
type GetLevelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewGetLevelLogic 创建VIP等级详情逻辑
func NewGetLevelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLevelLogic {
	return &GetLevelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetLevel 获取VIP等级
func (l *GetLevelLogic) GetLevel(in *viplevelpb.GetVipLevelRequest) (*viplevelpb.GetVipLevelResponse, error) {
	if in == nil || in.GetId() <= 0 {
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
	}

	// 查询VIP等级
	level, err := l.svcCtx.DAO.MemberVipLevel.Get(l.ctx, in.GetId())
	if err != nil {
		if ent.IsNotFound(err) {
			// 等级不存在或不属于当前厅
			return nil, xerr.RpcErr(xerr.EntNotFound(i18n.InvalidParam, err))
		}
		return nil, xerr.RpcErr(err)
	}

	// 统计当前等级的会员数
	memberCounts, err := l.svcCtx.DAO.MemberVipLevel.MemberCounts(l.ctx, []int64{level.ID})
	if err != nil {
		return nil, xerr.RpcErr(err)
	}

	// 组装VIP等级详情
	return &viplevelpb.GetVipLevelResponse{
		Level: &viplevelpb.VipLevelInfo{
			Id:               level.ID,                    // VIP等级ID
			Level:            level.Level,                 // VIP等级编号
			UpgradeDeposit:   level.UpgradeDeposit,        // 晋级所需存款金额
			UpgradeWater:     level.UpgradeWater,          // 晋级所需洗码量
			RetentionDeposit: level.RetentionDeposit,      // 保级所需存款金额
			RetentionWater:   level.RetentionWater,        // 保级所需洗码量
			UpgradeReward:    level.UpgradeReward,         // 晋级奖励金额
			MemberCount:      memberCounts[level.ID],      // 当前会员数
			UpdatedAt:        level.UpdatedAt.UnixMilli(), // 更新时间
		},
	}, nil
}
