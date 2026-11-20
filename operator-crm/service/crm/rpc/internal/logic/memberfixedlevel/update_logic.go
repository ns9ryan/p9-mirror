package memberfixedlevellogic

import (
	"context"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberfixedlevelpb"
)

// UpdateLogic 固定层级修改逻辑
type UpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewUpdateLogic 创建固定层级修改逻辑
func NewUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateLogic {
	return &UpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Update 修改固定层级
func (l *UpdateLogic) Update(in *memberfixedlevelpb.UpdateMemberFixedLevelRequest) (*memberfixedlevelpb.UpdateMemberFixedLevelResponse, error) {
	if in == nil || in.GetId() <= 0 {
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
	}

	// 规范化输入
	name := strings.TrimSpace(in.GetName())
	description := strings.TrimSpace(in.GetDescription())

	// 修改固定层级, 所属厅由Ent Hook限定
	err := l.svcCtx.DAO.MemberFixedLevel.Update(l.ctx, in.GetId(), name, description, in.WithdrawalLimit)
	if err != nil {
		switch {
		case ent.IsNotFound(err):
			// 层级不存在或不属于当前厅
			return nil, xerr.RpcErr(xerr.EntNotFound(i18n.InvalidParam, err))
		case ent.IsConstraintError(err):
			// 数据约束冲突, 如同一厅内层级名称重复
			return nil, xerr.RpcErr(xerr.EntConstraintError(i18n.InvalidParam, err))
		case ent.IsValidationError(err):
			// Ent字段校验失败
			return nil, xerr.RpcErr(xerr.EntValidationError(i18n.InvalidParam, err))
		default:
			// 其他错误
			return nil, xerr.RpcErr(err)
		}
	}

	return &memberfixedlevelpb.UpdateMemberFixedLevelResponse{}, nil
}
