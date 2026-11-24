// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package deposit_order

import (
	"context"
	"math"

	agentconv "oa.98ent.com/p9/operator-admin-api/internal/logic/agent"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentdepositorderpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddAgentDepositOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 添加人工存款订单
func NewAddAgentDepositOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddAgentDepositOrderLogic {
	return &AddAgentDepositOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddAgentDepositOrderLogic) AddAgentDepositOrder(req *types.AddAgentDepositOrderReq) (*types.AddAgentDepositOrderResp, error) {
	// deposit_amount 前端以主币单位（如 USDT）输入，转换为分后传给 RPC
	depositAmountCents := int64(math.Round(req.DepositAmount * agentconv.AmountCentsFactor))

	rpcResp, err := l.svcCtx.AgentDepositOrderRpc.AddAgentDepositOrder(l.ctx, &agentdepositorderpb.AddAgentDepositOrderReq{
		MemberId:           req.MemberId,
		DepositAmount:      depositAmountCents,
		RolloverMultiplier: req.RolloverMultiplier,
		BonusRate:          req.BonusRate,
		DepositType:        agentconv.DepositTypeManual, // 人工入账，API 层固定值
	})
	if err != nil {
		return nil, err
	}
	return &types.AddAgentDepositOrderResp{
		Id: rpcResp.GetId(),
	}, nil
}
