package memberlogic

import (
	"context"

	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSensitiveInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSensitiveInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSensitiveInfoLogic {
	return &GetSensitiveInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查看会员敏感信息
func (l *GetSensitiveInfoLogic) GetSensitiveInfo(in *memberpb.GetMemberSensitiveInfoRequest) (*memberpb.GetMemberSensitiveInfoResponse, error) {
	// todo: add your logic here and delete this line

	return &memberpb.GetMemberSensitiveInfoResponse{}, nil
}
