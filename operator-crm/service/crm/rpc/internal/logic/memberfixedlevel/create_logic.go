package memberfixedlevellogic

import (
	"context"

	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberfixedlevelpb"

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

// 创建固定层级
func (l *CreateLogic) Create(in *memberfixedlevelpb.CreateMemberFixedLevelRequest) (*memberfixedlevelpb.CreateMemberFixedLevelResponse, error) {
	// todo: add your logic here and delete this line

	return &memberfixedlevelpb.CreateMemberFixedLevelResponse{}, nil
}
