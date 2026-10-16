package authservicelogic

import (
	"context"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/authpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type LogoutAllLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLogoutAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogoutAllLogic {
	return &LogoutAllLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// LogoutAll 退出全部登录
func (l *LogoutAllLogic) LogoutAll(_ *authpb.Empty) (*authpb.Empty, error) {
	if err := l.svcCtx.Auth.LogoutAll(l.ctx); err != nil {
		return nil, xerr.RpcErr(err)
	}
	return &authpb.Empty{}, nil
}
