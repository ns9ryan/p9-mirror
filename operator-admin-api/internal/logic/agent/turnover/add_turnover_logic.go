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
	// consume_amount / balance_before 前端以主币单位（如 USDT）输入，转换为分后传给 RPC
	consumeAmountCents := int64(math.Round(req.ConsumeAmount * agentconv.AmountCentsFactor))
	balanceBeforeCents := int64(math.Round(req.BalanceBefore * agentconv.AmountCentsFactor))

	pbReq := &turnoverpb.AddTurnoverReq{
		MemberId:      req.MemberId,
		ConsumeAmount: consumeAmountCents,
		BalanceBefore: balanceBeforeCents,
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
