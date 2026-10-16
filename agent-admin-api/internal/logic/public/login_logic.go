// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package public

import (
	"context"

	"oa.98ent.com/p9/agent-admin-api/internal/svc"
	"oa.98ent.com/p9/agent-admin-api/internal/types"
	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/authpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 登录
func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Login 登录
func (l *LoginLogic) Login(req *types.LoginReq) (resp *types.LoginResp, err error) {
	if req == nil {
		req = &types.LoginReq{}
	}
	out, err := l.svcCtx.AgentAuthRpc.Login(l.ctx, &authpb.LoginReq{
		Account:      req.Account,
		Password:     req.Password,
		OperatorCode: req.OperatorCode,
		ClientIp:     ctxdata.ClientIPFromCtx(l.ctx),
	})
	if err != nil {
		return nil, err
	}
	return toLoginResp(out), nil
}

// 转换登录响应结构体
func toLoginResp(out *authpb.LoginResp) *types.LoginResp {
	if out == nil {
		return &types.LoginResp{}
	}
	resp := &types.LoginResp{}
	if token := out.GetToken(); token != nil {
		resp.Token = types.TokenInfo{
			AccessToken:   token.GetAccessToken(),
			RefreshToken:  token.GetRefreshToken(),
			Expire:        token.GetExpire(),
			RefreshExpire: token.GetRefreshExpire(),
		}
	}
	if user := out.GetUser(); user != nil {
		resp.User = types.UserPublic{
			Id:          user.GetId(),
			Realname:    user.GetRealname(),
			DisplayName: user.GetNickname(),
			Status:      user.GetStatus(),
			CreatedAt:   user.GetCreatedAt(),
			LastLoginAt: user.GetLastLoginAt(),
			Mobile:      user.GetMobile(),
			Email:       user.GetEmail(),
		}
	}
	return resp
}
