package profile

import (
	"context"

	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateAgentRemarkLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateAgentRemarkLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAgentRemarkLogic {
	return &UpdateAgentRemarkLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateAgentRemarkLogic) UpdateAgentRemark(req *types.UpdateAgentRemarkReq) (*types.AgentEmpty, error) {
	if req == nil {
		req = &types.UpdateAgentRemarkReq{}
	}
	_, err := l.svcCtx.AgentRpc.UpdateAgentRemark(l.ctx, &agentpb.UpdateAgentRemarkReq{
		AgentId: req.AgentId,
		Remark:  req.Remark,
	})
	if err != nil {
		return nil, err
	}
	return &types.AgentEmpty{}, nil
}
