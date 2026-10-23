// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package auto_level

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberautolevelpb"

	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
)

// CreateLogic 自动层级创建逻辑
type CreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewCreateLogic 创建自动层级创建逻辑
func NewCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateLogic {
	return &CreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Create 创建自动层级
func (l *CreateLogic) Create(req *types.CreateMemberAutoLevelRequest) (resp *types.CreateMemberAutoLevelResponse, err error) {
	// 创建自动层级
	rpcResp, err := l.svcCtx.MemberAutoLevelRpc.Create(l.ctx, &memberautolevelpb.CreateMemberAutoLevelRequest{
		Name:       req.Name,       // 层级名称
		MinDeposit: req.MinDeposit, // 晋级要求的累计存款金额
		MinWater:   req.MinWater,   // 晋级要求的累计洗码量
	})
	if err != nil {
		return nil, err
	}

	return &types.CreateMemberAutoLevelResponse{
		Id: rpcResp.GetId(), // 新建层级ID
	}, nil
}
