package member

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"
)

type GetMemberCreateOptionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMemberCreateOptionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMemberCreateOptionsLogic {
	return &GetMemberCreateOptionsLogic{ctx:ctx, svcCtx:svcCtx, Logger:logx.WithContext(ctx)}
}

// GetMemberCreateOptions 返回会员创建表单的临时图片选项。
func (l *GetMemberCreateOptionsLogic) GetMemberCreateOptions() (*types.GetMemberCreateOptionsResp, error) {
	resp, err := l.svcCtx.MemberRpc.GetCreateOptions(l.ctx, &memberpb.GetMemberCreateOptionsRequest{})
	if err != nil {
		return nil, err
	}
	return CreateOptionsFromPB(resp), nil
}
