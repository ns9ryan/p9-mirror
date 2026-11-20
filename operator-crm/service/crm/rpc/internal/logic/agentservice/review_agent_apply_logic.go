package agentservicelogic

import (
	"context"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/constants"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReviewAgentApplyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReviewAgentApplyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReviewAgentApplyLogic {
	return &ReviewAgentApplyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ReviewAgentApply 申请代理审批
func (l *ReviewAgentApplyLogic) ReviewAgentApply(in *agentpb.ReviewAgentApplyReq) (*agentpb.ReviewAgentApplyResp, error) {
	resp, updatedAt, err := l.svcCtx.Agent.ReviewAgentApply(l.ctx, in)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}
	if in.Action == constants.ReviewActionApprove {
		operatorCode := ctxdata.ClaimsFromCtx(l.ctx).OperatorCode
		l.svcCtx.KafkaEventHandler.PublishAgentApplicationApprovedEvent(l.ctx, resp.Apply.MemberId, operatorCode, updatedAt, constants.AuditStatusApproved)
	}
	return resp, nil
}
