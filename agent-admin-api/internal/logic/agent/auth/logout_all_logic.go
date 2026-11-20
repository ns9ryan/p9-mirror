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

type LogoutAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 退出全部登录
func NewLogoutAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogoutAllLogic {
	return &LogoutAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// LogoutAll 退出全部登录
func (l *LogoutAllLogic) LogoutAll() (resp *types.BaseMsgResp, err error) {
	_, err = l.svcCtx.AgentAuthRpc.LogoutAll(l.ctx, &authpb.Empty{})
	if err != nil {
		return nil, err
	}
	return &types.BaseMsgResp{Result: "success"}, nil
}
