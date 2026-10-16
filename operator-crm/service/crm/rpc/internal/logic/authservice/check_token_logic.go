package authservicelogic

import (
	"context"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/authpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckTokenLogic {
	return &CheckTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CheckToken 校验访问令牌
func (l *CheckTokenLogic) CheckToken(in *authpb.CheckTokenReq) (*authpb.CheckTokenResp, error) {
	raw := ""
	if in != nil {
		raw = in.AccessToken
	}
	resp, err := l.svcCtx.Auth.CheckToken(l.ctx, raw)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}
	return resp, nil
}
