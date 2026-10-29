// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package manage

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"
)

// CreateLogic 会员创建逻辑
type CreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewCreateLogic 创建会员逻辑
func NewCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateLogic {
	return &CreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Create 创建会员
func (l *CreateLogic) Create(req *types.CreateMemberRequest) (resp *types.CreateMemberResponse, err error) {
	// 调用CRM创建会员RPC
	rpcResp, err := l.svcCtx.MemberRpc.Create(l.ctx, &memberpb.CreateMemberRequest{
		Account:  req.Account,  // 会员登录账号
		Password: req.Password, // 初始密码
		Nickname: req.Nickname, // 会员昵称
		AgentId:  req.AgentId,  // 所属分层代理ID
		Remark:   req.Remark,   // 后台会员备注
	})
	if err != nil {
		return nil, err
	}

	// 返回新建会员信息
	return &types.CreateMemberResponse{
		Id:   rpcResp.GetId(),   // 会员数据库主键
		Code: rpcResp.GetCode(), // 会员对外编号
	}, nil
}
