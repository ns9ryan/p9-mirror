package memberlogic

import (
	"context"

	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDetailLogic {
	return &GetDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取会员详情
func (l *GetDetailLogic) GetDetail(in *memberpb.GetMemberDetailRequest) (*memberpb.GetMemberDetailResponse, error) {
	// todo: add your logic here and delete this line

	return &memberpb.GetMemberDetailResponse{}, nil
}
