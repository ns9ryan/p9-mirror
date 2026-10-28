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

// DeleteLevelLogic VIP等级删除逻辑
type DeleteLevelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewDeleteLevelLogic 创建VIP等级删除逻辑
func NewDeleteLevelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteLevelLogic {
	return &DeleteLevelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeleteLevel 删除VIP等级
func (l *DeleteLevelLogic) DeleteLevel(in *viplevelpb.DeleteVipLevelRequest) (*viplevelpb.DeleteVipLevelResponse, error) {
	if in == nil || in.GetId() <= 0 {
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

	// VIP0是系统基础等级, 禁止删除
	if level.Level == 0 {
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
	}

	// 只允许删除当前最高等级, 保持等级连续
	highest, err := l.svcCtx.DAO.MemberVipLevel.GetHighest(l.ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, xerr.RpcErr(xerr.EntNotFound(i18n.InvalidParam, err))
		}
		return nil, xerr.RpcErr(err)
	}
	if highest.ID != level.ID {
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
	}

	// 查询当前等级关联的会员数
	memberCounts, err := l.svcCtx.DAO.MemberVipLevel.MemberCounts(l.ctx, []int64{level.ID})
	if err != nil {
		return nil, xerr.RpcErr(err)
	}

	// TODO(VIP): 会员等级重新计算实现后, 支持删除有关联会员的等级
	if memberCounts[level.ID] > 0 {
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
	}

	// 删除VIP等级
	if err := l.svcCtx.DAO.MemberVipLevel.Delete(l.ctx, level.ID); err != nil {
		switch {
		case ent.IsNotFound(err):
			// 等级不存在或不属于当前厅
			return nil, xerr.RpcErr(xerr.EntNotFound(i18n.InvalidParam, err))
		case ent.IsConstraintError(err):
			// 数据库约束冲突, 删除操作无法满足数据约束
			return nil, xerr.RpcErr(xerr.EntConstraintError(i18n.InvalidParam, err))
		default:
			// 其他未分类错误
			return nil, xerr.RpcErr(err)
		}
	}

	return &viplevelpb.DeleteVipLevelResponse{}, nil
}
