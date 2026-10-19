package agentservicelogic

import (
	"context"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ResetLoginPasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewResetLoginPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetLoginPasswordLogic {
	return &ResetLoginPasswordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ResetLoginPasswordLogic) ResetLoginPassword(in *agentpb.ResetLoginPasswordReq) (*agentpb.ResetLoginPasswordResp, error) {
	resp, err := l.svcCtx.Agent.ResetLoginPassword(l.ctx, in)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}
	return resp, nil
}
