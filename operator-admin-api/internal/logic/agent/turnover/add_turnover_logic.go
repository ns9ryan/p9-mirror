// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package turnover

import (
	"context"
	"math"

	agentconv "oa.98ent.com/p9/operator-admin-api/internal/logic/agent"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/turnoverpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddTurnoverLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 添加流水
func NewAddTurnoverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddTurnoverLogic {
	return &AddTurnoverLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddTurnoverLogic) AddTurnover(req *types.AddTurnoverReq) (*types.AddTurnoverResp, error) {
	// 从分站配置读取结算货币编码及换算倍率，向下传递给 RPC。货币未配置时直接报错。
	currency, amountFactor, err := agentconv.GetSiteCurrency(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}
	f := float64(amountFactor)
	consumeAmountCents := int64(math.Round(req.ConsumeAmount * f))
	balanceBeforeCents := int64(math.Round(req.BalanceBefore * f))

	pbReq := &turnoverpb.AddTurnoverReq{
		MemberId:      req.MemberId,
		ConsumeAmount: consumeAmountCents,
		BalanceBefore: balanceBeforeCents,
		Currency:      currency,
	}
	if req.Notes != "" {
		pbReq.Notes = &req.Notes
	}
	if req.GameId != 0 {
		pbReq.GameId = &req.GameId
	}

	rpcResp, err := l.svcCtx.TurnoverRpc.AddTurnover(l.ctx, pbReq)
	if err != nil {
		return nil, err
	}
	return &types.AddTurnoverResp{
		Id: rpcResp.GetId(),
	}, nil
}
