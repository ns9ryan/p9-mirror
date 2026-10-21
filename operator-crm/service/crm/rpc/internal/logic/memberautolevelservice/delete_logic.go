package memberautolevelservicelogic

import (
	"context"

	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberautolevelpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteLogic {
	return &DeleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除自动层级
func (l *DeleteLogic) Delete(in *memberautolevelpb.DeleteMemberAutoLevelRequest) (*memberautolevelpb.DeleteMemberAutoLevelResponse, error) {
	// todo: add your logic here and delete this line

	return &memberautolevelpb.DeleteMemberAutoLevelResponse{}, nil
}
