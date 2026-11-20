// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package vip

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/viplevelpb"
)

// DeleteLevelLogic VIP等级删除逻辑
type DeleteLevelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewDeleteLevelLogic 创建VIP等级删除逻辑
func NewDeleteLevelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteLevelLogic {
	return &DeleteLevelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// DeleteLevel 删除VIP等级
func (l *DeleteLevelLogic) DeleteLevel(req *types.DeleteVipLevelRequest) (resp *types.EmptyResponse, err error) {
	// 删除VIP等级
	_, err = l.svcCtx.MemberVipRpc.DeleteLevel(l.ctx, &viplevelpb.DeleteVipLevelRequest{
		Id: req.Id, // VIP等级ID
	})
	if err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}
