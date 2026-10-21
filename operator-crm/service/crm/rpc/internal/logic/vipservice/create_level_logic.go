package vipservicelogic

import (
	"context"

	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/viplevelpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateLevelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateLevelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateLevelLogic {
	return &CreateLevelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 新增VIP等级
func (l *CreateLevelLogic) CreateLevel(in *viplevelpb.CreateVipLevelRequest) (*viplevelpb.CreateVipLevelResponse, error) {
	// todo: add your logic here and delete this line

	return &viplevelpb.CreateVipLevelResponse{}, nil
}
