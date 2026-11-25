// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package deposit_order

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
	agentconst "oa.98ent.com/p9/operator-admin-api/internal/constants"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentdepositorderpb"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc/depositpb"
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

	// 仅最终通过（status=4）时触发推荐人存款事件及钱包入账；二审流转（status=12）不触发
	if rpcResp.GetStatus() == agentconst.DepositOrderStatusApproved {
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

		// 钱包入账：operation_id 由订单号确定性生成，确保重试幂等
		if err = l.creditDepositWallet(rpcResp); err != nil {
			return nil, err
		}
	}

	return &types.AgentDepositOrderActionResp{
		OrderId: rpcResp.GetOrderId(),
		Status:  rpcResp.GetStatus(),
	}, nil
}

// creditDepositWallet 审核终态通过后向钱包服务发起存款本金入账。
// operation_id 由订单号通过 SHA-1 命名空间 UUID 确定性生成，保证接口幂等。
func (l *ApproveAgentDepositOrderLogic) creditDepositWallet(resp *agentdepositorderpb.ApproveAgentDepositOrderResp) error {
	if resp.GetMemberGlobalId() == "" {
		return xerr.Err(http.StatusInternalServerError, "member global id missing for wallet credit")
	}
	if resp.GetOrderNo() == "" || resp.GetSiteCurrency() == "" {
		return xerr.Err(http.StatusInternalServerError, "order info missing for wallet credit")
	}
	if resp.GetActualReceivedAmount() <= 0 {
		return xerr.Err(http.StatusInternalServerError, "actual received amount invalid for wallet credit")
	}

	// 确定性 operation_id：同一订单重试始终产生相同 UUID，满足钱包幂等要求
	operationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("agent_deposit_credit:"+resp.GetOrderNo()))

	_, err := l.svcCtx.MemberWalletRpc.DepositCredit(l.ctx, &depositpb.DepositCreditRequest{
		MemberGlobalId: resp.GetMemberGlobalId(),
		OperationId:    operationID.String(),
		DepositOrderNo: resp.GetOrderNo(),
		CurrencyCode:   resp.GetSiteCurrency(),
		Amount:         resp.GetActualReceivedAmount(),
	})
	if err != nil {
		l.Errorw("agent deposit wallet credit failed",
			logx.Field("order_no", resp.GetOrderNo()),
			logx.Field("member_global_id", resp.GetMemberGlobalId()),
			logx.Field("amount", resp.GetActualReceivedAmount()),
			logx.Field("err", err),
		)
		return xerr.Err(http.StatusBadGateway, "wallet credit failed")
	}
	return nil
}
