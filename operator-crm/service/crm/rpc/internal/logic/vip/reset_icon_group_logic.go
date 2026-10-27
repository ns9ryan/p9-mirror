package viplogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/vipicongrouppb"
)

// ResetIconGroupLogic VIP图标恢复默认逻辑
type ResetIconGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewResetIconGroupLogic 创建VIP图标恢复默认逻辑
func NewResetIconGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetIconGroupLogic {
	return &ResetIconGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ResetIconGroup 恢复默认VIP图标
func (l *ResetIconGroupLogic) ResetIconGroup(in *vipicongrouppb.ResetVipIconGroupRequest) (*vipicongrouppb.ResetVipIconGroupResponse, error) {
	if in == nil || in.GetGroupNo() < 0 {
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
	}

	// 恢复系统默认图标
	if err := l.svcCtx.DAO.MemberVipIconGroup.Reset(l.ctx, in.GetGroupNo()); err != nil {
		switch {
		case ent.IsNotFound(err):
			// 图标组不存在或不属于当前厅
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

	return &vipicongrouppb.ResetVipIconGroupResponse{}, nil
}
