// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package auth

import (
	"context"

	"oa.98ent.com/p9/agent-admin-api/internal/svc"
	"oa.98ent.com/p9/agent-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/authpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type LogoutLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 退出当前登录
func NewLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogoutLogic {
	return &LogoutLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Logout 退出当前登录
func (l *LogoutLogic) Logout(req *types.LogoutReq) (resp *types.BaseMsgResp, err error) {
	if req == nil {
		req = &types.LogoutReq{}
	}
	_, err = l.svcCtx.AgentAuthRpc.Logout(l.ctx, &authpb.LogoutReq{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		return nil, err
	}
	return &types.BaseMsgResp{Result: "success"}, nil
}
