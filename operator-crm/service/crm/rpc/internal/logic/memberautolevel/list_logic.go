package memberautolevellogic

import (
	"context"

	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberautolevelpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLogic {
	return &ListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取自动层级列表
func (l *ListLogic) List(in *memberautolevelpb.ListMemberAutoLevelsRequest) (*memberautolevelpb.ListMemberAutoLevelsResponse, error) {
	// todo: add your logic here and delete this line

	return &memberautolevelpb.ListMemberAutoLevelsResponse{}, nil
}
