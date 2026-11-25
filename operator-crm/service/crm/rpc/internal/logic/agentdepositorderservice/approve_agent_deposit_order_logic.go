package agentdepositorderservicelogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/constants"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/kafka"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentdepositorderpb"
)

type ApproveAgentDepositOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewApproveAgentDepositOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApproveAgentDepositOrderLogic {
	return &ApproveAgentDepositOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// group: agent_deposit_order 审核通过
func (l *ApproveAgentDepositOrderLogic) ApproveAgentDepositOrder(in *agentdepositorderpb.ApproveAgentDepositOrderReq) (*agentdepositorderpb.ApproveAgentDepositOrderResp, error) {
	if in == nil || in.OrderId <= 0 {
		return nil, xerr.RpcErr(xerr.BadRequest("common.invalidParam"))
	}
	resp, err := l.svcCtx.Agent.ApproveAgentDepositOrder(l.ctx, in)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}

	// 仅在最终通过（status=4）时投递活动事件；二审流转（status=12）不触发
	if resp.Status != int32(constants.DepositOrderStatusApproved) {
		return resp, nil
	}

	// 查询会员 global_id，供调用方触发钱包入账，失败降级为空串（调用方自行处理）
	claims := ctxdata.ClaimsFromCtx(l.ctx)
	if claims != nil {
		globalID, gErr := l.svcCtx.Member.GetMemberGlobalID(l.ctx, claims.OperatorCode, resp.MemberId)
		if gErr != nil {
			l.Errorf("[审核通过] 查询会员 global_id 失败，member_id=%d err=%v", resp.MemberId, gErr)
		} else {
			resp.MemberGlobalId = globalID
		}
	}

	// 异步投递，失败只记日志不影响主流程
	go func() {
		asyncCtx := context.WithoutCancel(l.ctx)
		claims := ctxdata.ClaimsFromCtx(asyncCtx)
		if claims == nil {
			l.Errorf("[活动事件] claims 为空，跳过事件投递: order_id=%d", in.OrderId)
			return
		}
		// 查询该会员所有有效存款订单总数（全时间范围，status=4/6），用于活动侧首充/N充条件判断
		_, depositCount, cntErr := l.svcCtx.Agent.GetMemberDepositSummary(asyncCtx, resp.MemberId, 0, 0)
		if cntErr != nil {
			l.Errorf("[活动事件] 查询存款次数失败，deposit_count 置为 0: member_id=%d err=%v", resp.MemberId, cntErr)
		}
		l.svcCtx.KafkaEventHandler.PublishDepositOrderApprovedEvent(
			asyncCtx,
			kafka.DepositApprovedEventInput{
				OperatorCode:   claims.OperatorCode,
				MemberID:       resp.MemberId,
				OrderID:        in.OrderId,
				Amount:         resp.DepositAmount,
				Currency:       constants.DepositCurrency,
				DepositCount:   int(depositCount),
				ClientType:     constants.ClientTypePCWeb,
				MemberGlobalID: resp.MemberGlobalId,
			},
		)
	}()

	return resp, nil
}
