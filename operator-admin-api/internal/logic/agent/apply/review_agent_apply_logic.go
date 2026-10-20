package apply

import (
	"context"

	agentconv "oa.98ent.com/p9/operator-admin-api/internal/logic/agent"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReviewAgentApplyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 代理申请审批
func NewReviewAgentApplyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReviewAgentApplyLogic {
	return &ReviewAgentApplyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReviewAgentApplyLogic) ReviewAgentApply(req *types.ReviewAgentApplyReq) (*types.ReviewAgentApplyResp, error) {
	if req == nil {
		req = &types.ReviewAgentApplyReq{}
	}
	rpcResp, err := l.svcCtx.AgentRpc.ReviewAgentApply(l.ctx, &agentpb.ReviewAgentApplyReq{
		Id:           req.Id,
		Action:       req.Action,
		RejectReason: req.RejectReason,
	})
	if err != nil {
		return nil, err
	}
	return &types.ReviewAgentApplyResp{
		Apply:    agentconv.ApplyItemFromPB(rpcResp.GetApply()),
		Password: rpcResp.GetPassword(),
	}, nil
}
