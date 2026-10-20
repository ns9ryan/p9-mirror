package manage

import (
	"context"

	agentconv "oa.98ent.com/p9/operator-admin-api/internal/logic/agent"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchAgentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSearchAgentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchAgentsLogic {
	return &SearchAgentsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SearchAgentsLogic) SearchAgents(req *types.SearchAgentsReq) (*types.SearchAgentsResp, error) {
	if req == nil {
		req = &types.SearchAgentsReq{}
	}
	rpcResp, err := l.svcCtx.AgentRpc.SearchAgents(l.ctx, &agentpb.SearchAgentsReq{
		LineCode: &req.LineCode,
		Keyword:  req.Keyword,
		Level:    req.Level,
		Limit:    req.Limit,
	})
	if err != nil {
		return nil, err
	}
	return &types.SearchAgentsResp{
		Items: agentconv.SearchItemsFromPB(rpcResp.GetItems()),
	}, nil
}
