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

// UpdateLogic 会员固定层级修改逻辑
type UpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewUpdateLogic 创建固定层级修改逻辑
func NewUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateLogic {
	return &UpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Update 修改固定层级
func (l *UpdateLogic) Update(req *types.UpdateMemberFixedLevelRequest) (resp *types.EmptyResponse, err error) {
	// 修改固定层级配置
	_, err = l.svcCtx.MemberFixedLevelRpc.Update(l.ctx, &memberfixedlevelpb.UpdateMemberFixedLevelRequest{
		Id:              req.Id,              // 层级ID
		Name:            req.Name,            // 层级名称
		Description:     req.Description,     // 层级说明
		WithdrawalLimit: req.WithdrawalLimit, // 单笔提款金额上限
	})
	if err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}
