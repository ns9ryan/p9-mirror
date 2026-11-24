package turnoverservicelogic

import (
	"context"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/turnoverpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddTurnoverLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddTurnoverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddTurnoverLogic {
	return &AddTurnoverLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// group: turnover 添加流水
func (l *AddTurnoverLogic) AddTurnover(in *turnoverpb.AddTurnoverReq) (*turnoverpb.AddTurnoverResp, error) {
	resp, err := l.svcCtx.Agent.AddTurnover(l.ctx, in)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}

	// 流水类型固定为游戏，始终投递投注结算事件
	operatorCode := ctxdata.ClaimsFromCtx(l.ctx).OperatorCode
	l.svcCtx.KafkaEventHandler.PublishGameOrderSettledEvent(
		l.ctx,
		operatorCode,
		in.MemberId,
		resp.Id,
		in.ConsumeAmount,
	)

	return resp, nil
}
