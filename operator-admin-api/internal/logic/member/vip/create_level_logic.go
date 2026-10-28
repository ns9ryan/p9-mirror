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

// CreateLevelLogic VIP等级创建逻辑
type CreateLevelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewCreateLevelLogic 创建VIP等级创建逻辑
func NewCreateLevelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateLevelLogic {
	return &CreateLevelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// CreateLevel 创建VIP等级
func (l *CreateLevelLogic) CreateLevel(req *types.CreateVipLevelRequest) (resp *types.CreateVipLevelResponse, err error) {
	// 创建VIP等级, 等级编号由CRM分配
	rpcResp, err := l.svcCtx.MemberVipRpc.CreateLevel(l.ctx, &viplevelpb.CreateVipLevelRequest{
		UpgradeDeposit:   *req.UpgradeDeposit,   // 晋级所需存款
		UpgradeWater:     *req.UpgradeWater,     // 晋级所需洗码量
		RetentionDeposit: *req.RetentionDeposit, // 保级所需存款
		RetentionWater:   *req.RetentionWater,   // 保级所需洗码量
		UpgradeReward:    *req.UpgradeReward,    // 晋级奖励金额
	})
	if err != nil {
		return nil, err
	}

	// 返回新建等级信息
	return &types.CreateVipLevelResponse{
		Id:    rpcResp.GetId(),    // VIP等级ID
		Level: rpcResp.GetLevel(), // 服务端分配的等级编号
	}, nil
}
