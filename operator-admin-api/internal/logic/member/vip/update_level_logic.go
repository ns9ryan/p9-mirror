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

// UpdateLevelLogic VIP等级修改逻辑
type UpdateLevelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewUpdateLevelLogic 创建VIP等级修改逻辑
func NewUpdateLevelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateLevelLogic {
	return &UpdateLevelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UpdateLevel 修改VIP等级
func (l *UpdateLevelLogic) UpdateLevel(req *types.UpdateVipLevelRequest) (resp *types.EmptyResponse, err error) {
	// 修改VIP等级配置
	_, err = l.svcCtx.MemberVipRpc.UpdateLevel(l.ctx, &viplevelpb.UpdateVipLevelRequest{
		Id:               req.Id,               // VIP等级ID
		UpgradeDeposit:   req.UpgradeDeposit,   // 晋级所需存款
		UpgradeWater:     req.UpgradeWater,     // 晋级所需洗码量
		RetentionDeposit: req.RetentionDeposit, // 保级所需存款
		RetentionWater:   req.RetentionWater,   // 保级所需洗码量
		UpgradeReward:    req.UpgradeReward,    // 晋级奖励金额
	})
	if err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}
