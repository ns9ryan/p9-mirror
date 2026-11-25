package turnoverservicelogic

import (
	"context"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/constants"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/kafka"

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
	if in == nil {
		return nil, xerr.RpcErr(xerr.BadRequest("invalid param"))
	}
	if in.Currency == "" {
		return nil, xerr.RpcErr(xerr.BadRequest("invalid param"))
	}
	resp, err := l.svcCtx.Agent.AddTurnover(l.ctx, in)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}

	// 流水类型固定为游戏，始终投递投注结算事件
	operatorCode := ctxdata.ClaimsFromCtx(l.ctx).OperatorCode
	memberGlobalID, gErr := l.svcCtx.Member.GetMemberGlobalID(l.ctx, operatorCode, in.MemberId)
	if gErr != nil {
		l.Errorf("[活动事件] 查询会员 global_id 失败，member_global_id 置为空串: member_id=%d err=%v", in.MemberId, gErr)
	}

	var gameID int64
	if in.GameId != nil {
		gameID = *in.GameId
	}
	l.svcCtx.KafkaEventHandler.PublishGameOrderSettledEvent(
		l.ctx,
		kafka.GameOrderSettledEventInput{
			OperatorCode:   operatorCode,
			MemberID:       in.MemberId,
			OrderID:        resp.Id,
			GameID:         gameID,
			Amount:         in.ConsumeAmount,
			Currency:       in.Currency,
			ClientType:     constants.ClientTypePCWeb,
			MemberGlobalID: memberGlobalID,
		},
	)

	return resp, nil
}
