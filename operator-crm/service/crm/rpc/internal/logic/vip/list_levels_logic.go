package viplogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/common/paging"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/viplevelpb"
)

// ListLevelsLogic VIP等级列表逻辑
type ListLevelsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewListLevelsLogic 创建VIP等级列表逻辑
func NewListLevelsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLevelsLogic {
	return &ListLevelsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListLevels 获取VIP等级列表
func (l *ListLevelsLogic) ListLevels(in *viplevelpb.ListVipLevelsRequest) (*viplevelpb.ListVipLevelsResponse, error) {
	// 分页参数
	offset, limit, ok := paging.OffsetLimit(in.GetPage(), in.GetPageSize())
	if !ok {
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
	}

	// 获取VIP等级分页列表
	levels, total, err := l.svcCtx.DAO.MemberVipLevel.List(l.ctx, offset, limit)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}

	// 获取当前页等级ID
	levelIDs := make([]int64, 0, len(levels))
	for _, level := range levels {
		levelIDs = append(levelIDs, level.ID)
	}

	// 统计当前页各等级的会员数
	memberCounts, err := l.svcCtx.DAO.MemberVipLevel.MemberCounts(l.ctx, levelIDs)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}

	// 组装VIP等级列表
	list := make([]*viplevelpb.VipLevelInfo, 0, len(levels))
	for _, level := range levels {
		list = append(list, &viplevelpb.VipLevelInfo{
			Id:               level.ID,                    // VIP等级ID
			Level:            level.Level,                 // VIP等级编号
			UpgradeDeposit:   level.UpgradeDeposit,        // 晋级所需存款金额
			UpgradeWater:     level.UpgradeWater,          // 晋级所需洗码量
			RetentionDeposit: level.RetentionDeposit,      // 保级所需存款金额
			RetentionWater:   level.RetentionWater,        // 保级所需洗码量
			UpgradeReward:    level.UpgradeReward,         // 晋级奖励金额
			MemberCount:      memberCounts[level.ID],      // 当前会员数
			UpdatedAt:        level.UpdatedAt.UnixMilli(), // 更新时间
		})
	}

	return &viplevelpb.ListVipLevelsResponse{
		Total: total, // 数据总数
		List:  list,  // VIP等级列表
	}, nil
}
