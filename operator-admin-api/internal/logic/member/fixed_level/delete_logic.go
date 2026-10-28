// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package fixed_level

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberfixedlevelpb"
)

// DeleteLogic 会员固定层级删除逻辑
type DeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewDeleteLogic 创建固定层级删除逻辑
func NewDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteLogic {
	return &DeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Delete 删除固定层级
func (l *DeleteLogic) Delete(req *types.DeleteMemberFixedLevelRequest) (resp *types.EmptyResponse, err error) {
	// 删除固定层级
	_, err = l.svcCtx.MemberFixedLevelRpc.Delete(l.ctx, &memberfixedlevelpb.DeleteMemberFixedLevelRequest{
		Id: req.Id, // 层级ID
	})
	if err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}
