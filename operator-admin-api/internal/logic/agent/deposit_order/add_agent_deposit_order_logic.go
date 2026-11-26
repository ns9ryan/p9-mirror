// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package deposit_order

import (
	"context"
	"math"

	agentconst "oa.98ent.com/p9/operator-admin-api/internal/constants"
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
	// 读取结算货币编码及换算倍率，传递给 RPC（倍率由 RPC 层用于内部计算，API 层不再换算金额）。
	depositCurrency, amountFactor, err := agentconv.GetSiteCurrency(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}
	// 前端直接发送已放大后的整数金额（分），API 层不再乘倍率，直接取整传入。
	depositAmountCents := int64(math.Round(req.DepositAmount))

	rpcResp, err := l.svcCtx.AgentDepositOrderRpc.AddAgentDepositOrder(l.ctx, &agentdepositorderpb.AddAgentDepositOrderReq{
		MemberId:           req.MemberId,
		DepositAmount:      depositAmountCents,
		RolloverMultiplier: int64(math.Round(req.RolloverMultiplier * 100)),
		BonusRate:          int64(math.Round(req.BonusRate * 100)),
		DepositType:        agentconst.DepositTypeManual,
		DepositCurrency:    depositCurrency,
		CurrencyFactor:     amountFactor,
	})
	if err != nil {
		return nil, err
	}
	return &types.AddAgentDepositOrderResp{
		Id: rpcResp.GetId(),
	}, nil
}
