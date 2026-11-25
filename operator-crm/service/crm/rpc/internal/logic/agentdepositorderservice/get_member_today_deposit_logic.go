package agentdepositorderservicelogic

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentdepositorderpb"
)

type GetMemberTodayDepositLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMemberTodayDepositLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMemberTodayDepositLogic {
	return &GetMemberTodayDepositLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// group: agent_deposit_order 获取会员今日累计有效充值金额
// 默认统计今日（UTC）范围内创建时间、支付时间或审核时间满足的 status=4/6 订单。
func (l *GetMemberTodayDepositLogic) GetMemberTodayDeposit(in *agentdepositorderpb.GetMemberTodayDepositReq) (*agentdepositorderpb.GetMemberTodayDepositResp, error) {
	if in == nil || in.MemberId <= 0 {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}

	// 默认取今日 UTC 零点至当前时刻
	now := time.Now().UTC()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	timeStart := todayStart.Unix()
	timeEnd := now.Unix()

	amount, count, err := l.svcCtx.Agent.GetMemberDepositSummary(l.ctx, in.MemberId, timeStart, timeEnd)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}

	return &agentdepositorderpb.GetMemberTodayDepositResp{
		TodayDepositAmount: amount,
		TodayDepositCount:  count,
	}, nil
}

