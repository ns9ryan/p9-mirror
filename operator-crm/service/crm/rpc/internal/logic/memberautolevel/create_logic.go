package memberautolevellogic

import (
	"context"

	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberautolevelpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateLogic {
	return &CreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建自动层级
func (l *CreateLogic) Create(in *memberautolevelpb.CreateMemberAutoLevelRequest) (*memberautolevelpb.CreateMemberAutoLevelResponse, error) {
	// todo: add your logic here and delete this line

	return &memberautolevelpb.CreateMemberAutoLevelResponse{}, nil
}
