package apply

import (
	"context"

	agentconv "oa.98ent.com/p9/operator-admin-api/internal/logic/agent"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAgentAppliesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 代理申请列表
func NewListAgentAppliesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAgentAppliesLogic {
	return &ListAgentAppliesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListAgentAppliesLogic) ListAgentApplies(req *types.ListAgentAppliesReq) (*types.ListAgentAppliesResp, error) {
	if req == nil {
		req = &types.ListAgentAppliesReq{}
	}
	rpcResp, err := l.svcCtx.AgentRpc.ListAgentApplies(l.ctx, &agentpb.ListAgentAppliesReq{
		Page:        req.Page,
		PageSize:    req.PageSize,
		AuditStatus: req.AuditStatus,
		AccountType: req.AccountType,
		Account:     req.Account,
		TimeStart:   req.TimeStart,
		TimeEnd:     req.TimeEnd,
	})
	if err != nil {
		return nil, err
	}
	return &types.ListAgentAppliesResp{
		List:  agentconv.ApplyItemsFromPB(rpcResp.GetList()),
		Total: rpcResp.GetTotal(),
	}, nil
}
