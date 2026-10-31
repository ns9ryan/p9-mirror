package viplogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent"
	memberdao "oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/dao/member"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/viplevelpb"
)

// CreateLevelLogic VIP等级创建逻辑
type CreateLevelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewCreateLevelLogic 创建VIP等级创建逻辑
func NewCreateLevelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateLevelLogic {
	return &CreateLevelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreateLevel 创建VIP等级
func (l *CreateLevelLogic) CreateLevel(in *viplevelpb.CreateVipLevelRequest) (*viplevelpb.CreateVipLevelResponse, error) {
	if in == nil {
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
	}

	// 创建VIP等级, 等级编号由DAO根据当前最高等级分配
	level, err := l.svcCtx.DAO.MemberVipLevel.Create(l.ctx, memberdao.VIPLevelCreate{
		UpgradeDeposit:   in.UpgradeDeposit,   // 晋级所需存款金额
		UpgradeWater:     in.UpgradeWater,     // 晋级所需洗码量
		RetentionDeposit: in.RetentionDeposit, // 保级所需存款金额
		RetentionWater:   in.RetentionWater,   // 保级所需洗码量
		UpgradeReward:    in.UpgradeReward,    // 晋级奖励金额
	})
	if err != nil {
		switch {
		case ent.IsConstraintError(err):
			// 数据库约束冲突, 如等级编号重复
			return nil, xerr.RpcErr(xerr.EntConstraintError(i18n.InvalidParam, err))
		case ent.IsValidationError(err):
			// Ent字段校验失败
			return nil, xerr.RpcErr(xerr.EntValidationError(i18n.InvalidParam, err))
		default:
			// 其他错误, 包括分站VIP初始数据缺失
			return nil, xerr.RpcErr(err)
		}
	}

	// 返回新建等级信息
	return &viplevelpb.CreateVipLevelResponse{
		Id:    level.ID,    // VIP等级ID
		Level: level.Level, // 服务端分配的等级编号
	}, nil
}
