// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package deposit_order

import (
	"context"
	"time"

	"oa.98ent.com/p9/common/ctxdata"
	agentconv "oa.98ent.com/p9/operator-admin-api/internal/logic/agent"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentdepositorderpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ApproveAgentDepositOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 审核通过
func NewApproveAgentDepositOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApproveAgentDepositOrderLogic {
	return &ApproveAgentDepositOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ApproveAgentDepositOrderLogic) ApproveAgentDepositOrder(req *types.AgentDepositOrderAuditReq) (*types.AgentDepositOrderActionResp, error) {
	pbReq := &agentdepositorderpb.ApproveAgentDepositOrderReq{
		OrderId: req.OrderId,
	}
	if req.Notes != "" {
		pbReq.AuditNotes = &req.Notes
	}
	rpcResp, err := l.svcCtx.AgentDepositOrderRpc.ApproveAgentDepositOrder(l.ctx, pbReq)
	if err != nil {
		return nil, err
	}

	// 仅最终通过（status=4）时触发推荐人存款事件；二审流转（status=12）不触发
	if rpcResp.GetStatus() == agentconv.DepositOrderStatusApproved {
		// operator_code 从登录态获取，用于任务配置查询与上级代理查询
		operatorCode := ""
		if claims := ctxdata.ClaimsFromCtx(l.ctx); claims != nil {
			operatorCode = claims.OperatorCode
		}
		// occurred_at 使用 RPC 返回的真实审核时间
		l.svcCtx.KafkaEventHandler.PublishReferralRegisteredDepositEvent(
			l.ctx,
			rpcResp.GetMemberId(),
			operatorCode,
			time.Unix(rpcResp.GetApprovedAt(), 0),
			rpcResp.GetDepositAmount(), // 金额单位为"分"，与任务配置 CompletionAmount 同单位
		)
	}

	return &types.AgentDepositOrderActionResp{
		OrderId: rpcResp.GetOrderId(),
		Status:  rpcResp.GetStatus(),
	}, nil
}
