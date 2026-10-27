package memberfixedlevellogic

import (
	"context"

	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberfixedlevelpb"

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

// 获取固定层级列表
func (l *ListLogic) List(in *memberfixedlevelpb.ListMemberFixedLevelsRequest) (*memberfixedlevelpb.ListMemberFixedLevelsResponse, error) {
	// todo: add your logic here and delete this line

	return &memberfixedlevelpb.ListMemberFixedLevelsResponse{}, nil
}
