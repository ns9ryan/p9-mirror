package viplogic

import (
	"context"

	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/vipicongrouppb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ResetIconGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewResetIconGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetIconGroupLogic {
	return &ResetIconGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 恢复默认VIP图标
func (l *ResetIconGroupLogic) ResetIconGroup(in *vipicongrouppb.ResetVipIconGroupRequest) (*vipicongrouppb.ResetVipIconGroupResponse, error) {
	// todo: add your logic here and delete this line

	return &vipicongrouppb.ResetVipIconGroupResponse{}, nil
}
