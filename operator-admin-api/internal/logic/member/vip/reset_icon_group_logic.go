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

// ResetIconGroupLogic VIP图标恢复默认逻辑
type ResetIconGroupLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewResetIconGroupLogic 创建VIP图标恢复默认逻辑
func NewResetIconGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetIconGroupLogic {
	return &ResetIconGroupLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ResetIconGroup 恢复默认VIP图标
func (l *ResetIconGroupLogic) ResetIconGroup(req *types.ResetVipIconGroupRequest) (resp *types.EmptyResponse, err error) {
	// 恢复系统默认图标
	_, err = l.svcCtx.MemberVipRpc.ResetIconGroup(l.ctx, &vipicongrouppb.ResetVipIconGroupRequest{
		GroupNo: *req.GroupNo, // 图标组编号
	})
	if err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}
