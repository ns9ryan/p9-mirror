package memberfixedlevellogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberfixedlevelpb"
)

// DeleteLogic 固定层级删除逻辑
type DeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewDeleteLogic 创建固定层级删除逻辑
func NewDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteLogic {
	return &DeleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Delete 删除固定层级
func (l *DeleteLogic) Delete(in *memberfixedlevelpb.DeleteMemberFixedLevelRequest) (*memberfixedlevelpb.DeleteMemberFixedLevelResponse, error) {
	if in == nil || in.GetId() <= 0 {
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
	}

	// 删除固定层级及其会员关联
	err := l.svcCtx.DAO.MemberFixedLevel.Delete(l.ctx, in.GetId())
	if err != nil {
		switch {
		case ent.IsNotFound(err):
			// 层级不存在或不属于当前厅
			return nil, xerr.RpcErr(xerr.EntNotFound(i18n.InvalidParam, err))
		case ent.IsConstraintError(err):
			// 数据约束冲突
			return nil, xerr.RpcErr(xerr.EntConstraintError(i18n.InvalidParam, err))
		default:
			// 其他错误
			return nil, xerr.RpcErr(err)
		}
	}

	return &memberfixedlevelpb.DeleteMemberFixedLevelResponse{}, nil
}
