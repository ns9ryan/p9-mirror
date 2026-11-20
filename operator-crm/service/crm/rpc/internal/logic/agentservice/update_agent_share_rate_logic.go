package agentservicelogic

import (
	"context"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateAgentShareRateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateAgentShareRateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAgentShareRateLogic {
	return &UpdateAgentShareRateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateAgentShareRateLogic) UpdateAgentShareRate(in *agentpb.UpdateAgentShareRateReq) (*agentpb.Empty, error) {
	resp, err := l.svcCtx.Agent.UpdateShareRate(l.ctx, in)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}
	return resp, nil
}
