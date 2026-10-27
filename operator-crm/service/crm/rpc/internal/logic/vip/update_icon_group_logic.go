package viplogic

import (
	"context"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/vipicongrouppb"
)

// UpdateIconGroupLogic VIP图标修改逻辑
type UpdateIconGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewUpdateIconGroupLogic 创建VIP图标修改逻辑
func NewUpdateIconGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateIconGroupLogic {
	return &UpdateIconGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateIconGroup 修改VIP图标
func (l *UpdateIconGroupLogic) UpdateIconGroup(in *vipicongrouppb.UpdateVipIconGroupRequest) (*vipicongrouppb.UpdateVipIconGroupResponse, error) {
	// 校验图标组编号及图标URL
	iconURL := strings.TrimSpace(in.GetIconUrl())
	if in.GetGroupNo() < 0 || iconURL == "" {
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
	}

	// 修改当前使用的VIP图标
	if err := l.svcCtx.DAO.MemberVipIconGroup.Update(l.ctx, in.GetGroupNo(), iconURL); err != nil {
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

	return &vipicongrouppb.UpdateVipIconGroupResponse{}, nil
}
