// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package deposit_order

import (
	"context"

	"oa.98ent.com/p9/common/ctxdata"
	agentconv "oa.98ent.com/p9/operator-admin-api/internal/logic/agent"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentdepositorderpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAgentDepositOrdersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 人工存款订单列表
func NewListAgentDepositOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAgentDepositOrdersLogic {
	return &ListAgentDepositOrdersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListAgentDepositOrdersLogic) ListAgentDepositOrders(req *types.ListAgentDepositOrdersReq) (*types.ListAgentDepositOrdersResp, error) {
	if req == nil {
		req = &types.ListAgentDepositOrdersReq{}
	}
	pbReq := &agentdepositorderpb.ListAgentDepositOrdersReq{
		Page:        req.Page,
		PageSize:    req.PageSize,
		Status:      req.Status,
		MemberId:    req.MemberId,
		DepositType: req.DepositType,
		TimeStart:   req.TimeStart,
		TimeEnd:     req.TimeEnd,
	}
	depositCurrency := "USDT"
	pbReq.DepositCurrency = &depositCurrency
	if req.OrderNo != "" {
		pbReq.OrderNo = &req.OrderNo
	}
	if req.MemberAccount != "" {
		pbReq.MemberAccount = &req.MemberAccount
	}
	if req.ThirdPartyOrderNo != "" {
		pbReq.ThirdPartyOrderNo = &req.ThirdPartyOrderNo
	}
	if req.ReviewerId > 0 {
		pbReq.ReviewerId = &req.ReviewerId
	}
	if req.ClaimStatus > 0 {
		pbReq.ClaimStatus = &req.ClaimStatus
	}
	rpcResp, err := l.svcCtx.AgentDepositOrderRpc.ListAgentDepositOrders(l.ctx, pbReq)
	if err != nil {
		return nil, err
	}
	// 从 JWT claims 获取当前用户 ID，用于计算 ClaimedBySelf
	var selfUID int64
	if claims := ctxdata.ClaimsFromCtx(l.ctx); claims != nil {
		selfUID = claims.UserID
	}
	return &types.ListAgentDepositOrdersResp{
		List:  agentconv.DepositOrderListItemsFromPB(rpcResp.GetList(), selfUID),
		Total: rpcResp.GetTotal(),
	}, nil
}
