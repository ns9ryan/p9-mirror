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

// CreateLogic 固定层级创建逻辑
type CreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewCreateLogic 创建固定层级逻辑
func NewCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateLogic {
	return &CreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Create 创建固定层级
func (l *CreateLogic) Create(in *memberfixedlevelpb.CreateMemberFixedLevelRequest) (*memberfixedlevelpb.CreateMemberFixedLevelResponse, error) {
	if in == nil {
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
	}

	// 规范化输入
	name := strings.TrimSpace(in.GetName())
	description := strings.TrimSpace(in.GetDescription())

	// 创建固定层级, 所属厅由Ent Hook从可信上下文绑定
	level, err := l.svcCtx.DAO.MemberFixedLevel.Create(l.ctx, name, description, in.WithdrawalLimit)
	if err != nil {
		switch {
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

	// 返回新建层级ID
	return &memberfixedlevelpb.CreateMemberFixedLevelResponse{
		Id: level.ID, // 固定层级ID
	}, nil
}
