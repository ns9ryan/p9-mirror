// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package vip

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/vipicongrouppb"
)

// UpdateIconGroupLogic VIP图标修改逻辑
type UpdateIconGroupLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewUpdateIconGroupLogic 创建VIP图标修改逻辑
func NewUpdateIconGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateIconGroupLogic {
	return &UpdateIconGroupLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UpdateIconGroup 修改VIP图标
func (l *UpdateIconGroupLogic) UpdateIconGroup(req *types.UpdateVipIconGroupRequest) (resp *types.EmptyResponse, err error) {
	// 修改当前使用的VIP图标
	_, err = l.svcCtx.MemberVipRpc.UpdateIconGroup(l.ctx, &vipicongrouppb.UpdateVipIconGroupRequest{
		GroupNo: *req.GroupNo, // 图标组编号
		IconUrl: req.IconUrl,  // 自定义图标URL
	})
	if err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}
