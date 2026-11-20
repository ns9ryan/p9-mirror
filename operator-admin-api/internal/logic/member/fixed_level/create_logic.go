// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package fixed_level

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberfixedlevelpb"
)

// CreateLogic 会员固定层级创建逻辑
type CreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewCreateLogic 创建固定层级创建逻辑
func NewCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateLogic {
	return &CreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Create 创建固定层级
func (l *CreateLogic) Create(req *types.CreateMemberFixedLevelRequest) (resp *types.CreateMemberFixedLevelResponse, err error) {
	rpcResp, err := l.svcCtx.MemberFixedLevelRpc.Create(l.ctx, &memberfixedlevelpb.CreateMemberFixedLevelRequest{
		Name:            req.Name,            // 层级名称
		Description:     req.Description,     // 层级说明
		WithdrawalLimit: req.WithdrawalLimit, // 单笔提款金额上限
	})
	if err != nil {
		return nil, err
	}

	return &types.CreateMemberFixedLevelResponse{
		Id: rpcResp.GetId(), // 新建层级ID
	}, nil
}
