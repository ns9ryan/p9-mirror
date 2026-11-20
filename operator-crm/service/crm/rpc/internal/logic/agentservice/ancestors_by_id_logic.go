package agentservicelogic

import (
	"context"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type AncestorsByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAncestorsByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AncestorsByIdLogic {
	return &AncestorsByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// group: agent
func (l *AncestorsByIdLogic) AncestorsById(in *agentpb.AncestorsByIDReq) (*agentpb.AncestorsByIDResp, error) {
	resp, err := l.svcCtx.Agent.AncestorsByIDRPC(l.ctx, in)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}
	return resp, nil
}
