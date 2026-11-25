// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package turnover

import (
	"context"

	agentconv "oa.98ent.com/p9/operator-admin-api/internal/logic/agent"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/turnoverpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTurnoversLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 流水列表
func NewListTurnoversLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTurnoversLogic {
	return &ListTurnoversLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListTurnoversLogic) ListTurnovers(req *types.ListTurnoversReq) (*types.ListTurnoversResp, error) {
	if req == nil {
		req = &types.ListTurnoversReq{}
	}
	rpcResp, err := l.svcCtx.TurnoverRpc.ListTurnovers(l.ctx, &turnoverpb.ListTurnoversReq{
		Page:      req.Page,
		PageSize:  req.PageSize,
		MemberId:  req.MemberId,
		TimeStart: req.TimeStart,
		TimeEnd:   req.TimeEnd,
	})
	if err != nil {
		return nil, err
	}
	_, amountFactor, err := agentconv.GetSiteCurrency(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}
	return &types.ListTurnoversResp{
		List:  agentconv.TurnoverListItemsFromPB(rpcResp.GetList(), amountFactor),
		Total: rpcResp.GetTotal(),
	}, nil
}
