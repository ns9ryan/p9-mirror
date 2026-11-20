package manage

import (
	"context"
	"strings"

	agentconv "oa.98ent.com/p9/agent-admin-api/internal/logic/agent"
	"oa.98ent.com/p9/agent-admin-api/internal/svc"
	"oa.98ent.com/p9/agent-admin-api/internal/types"
	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddSubAgentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddSubAgentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddSubAgentLogic {
	return &AddSubAgentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddSubAgentLogic) AddSubAgent(req *types.AddSubAgentReq) (*types.AddAgentResp, error) {
	if req == nil {
		req = &types.AddSubAgentReq{}
	}

	if strings.TrimSpace(req.RegisterIp) == "" {
		req.RegisterIp = ctxdata.ClientIPFromCtx(l.ctx)
	}

	rpcResp, err := l.svcCtx.AgentRpc.AddSubAgent(l.ctx, &agentpb.AddSubAgentReq{
		Agent: agentconv.CreateFieldsToPB(req),
	})
	if err != nil {
		return nil, err
	}
	return agentconv.AddAgentFromPB(rpcResp), nil
}
