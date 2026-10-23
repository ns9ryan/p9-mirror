package memberautolevellogic

import (
	"context"

	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberautolevelpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateLogic {
	return &UpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 修改自动层级
func (l *UpdateLogic) Update(in *memberautolevelpb.UpdateMemberAutoLevelRequest) (*memberautolevelpb.UpdateMemberAutoLevelResponse, error) {
	// todo: add your logic here and delete this line

	return &memberautolevelpb.UpdateMemberAutoLevelResponse{}, nil
}
