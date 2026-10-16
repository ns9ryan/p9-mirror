package manage

import (
	"context"

	agentconv "oa.98ent.com/p9/operator-admin-api/internal/logic/agent"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAgentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListAgentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAgentsLogic {
	return &ListAgentsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListAgentsLogic) ListAgents(req *types.ListAgentsReq) (*types.ListAgentsResp, error) {
	if req == nil {
		req = &types.ListAgentsReq{}
	}
	rpcResp, err := l.svcCtx.AgentRpc.ListAgents(l.ctx, &agentpb.ListAgentsReq{
		Page:             req.Page,
		PageSize:         req.PageSize,
		RegisterSource:   req.RegisterSource,
		Level:            req.Level,
		Status:           req.Status,
		PayoutRestricted: req.PayoutRestricted,
		TimeFilterType:   req.TimeFilterType,
		TimeStart:        req.TimeStart,
		TimeEnd:          req.TimeEnd,
		NumberFilterType: req.NumberFilterType,
		NumberMin:        req.NumberMin,
		NumberMax:        req.NumberMax,
		AgentFilterType:  req.AgentFilterType,
		AgentKeyword:     req.AgentKeyword,
	})
	if err != nil {
		return nil, err
	}
	return &types.ListAgentsResp{
		Items: agentconv.ListItemsFromPB(rpcResp.GetItems()),
		Total: rpcResp.GetTotal(),
	}, nil
}
