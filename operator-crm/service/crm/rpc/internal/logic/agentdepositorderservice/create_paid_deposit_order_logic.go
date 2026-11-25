package agentdepositorderservicelogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/constants"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/kafka"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentdepositorderpb"
)

type CreatePaidDepositOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreatePaidDepositOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreatePaidDepositOrderLogic {
	return &CreatePaidDepositOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// group: agent_deposit_order 正式入款（直接写入支付成功终态，幂等）
func (l *CreatePaidDepositOrderLogic) CreatePaidDepositOrder(in *agentdepositorderpb.CreatePaidDepositOrderReq) (*agentdepositorderpb.CreatePaidDepositOrderResp, error) {
	if in == nil {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	if in.ThirdPartyOrderNo == "" {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	if in.MemberId <= 0 {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	// deposit_type 仅允许 1:加密货币 或 3:第三方代收
	if in.DepositType != 1 && in.DepositType != 3 {
		return nil, xerr.BadRequest(i18n.DepositTypeInvalid)
	}
	if in.DepositAmount <= 0 {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	if in.DepositCurrency == "" {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}

	// order_no 内部生成，调用方无需提供
	orderNo := orderNoGen.GenerateWithPrefix(depositOrderNoPrefix)

	resp, err := l.svcCtx.Agent.CreatePaidDepositOrder(l.ctx, in, orderNo)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}

	// 仅新建成功时推送 Kafka（幂等命中已有记录不重复投递）
	if resp.Created {
		go func() {
			asyncCtx := context.WithoutCancel(l.ctx)
			claims := ctxdata.ClaimsFromCtx(asyncCtx)
			if claims == nil {
				l.Errorf("[活动事件] claims 为空，跳过支付成功事件投递: order_id=%d", resp.Id)
				return
			}
			// 查询该会员所有有效存款订单总数（全时间范围，status=4/6），用于活动侧首充/N充条件判断
			_, depositCount, cntErr := l.svcCtx.Agent.GetMemberDepositSummary(asyncCtx, in.MemberId, 0, 0)
			if cntErr != nil {
				l.Errorf("[活动事件] 查询存款次数失败，deposit_count 置为 0: member_id=%d err=%v", in.MemberId, cntErr)
			}
			memberGlobalID, gErr := l.svcCtx.Member.GetMemberGlobalID(asyncCtx, claims.OperatorCode, in.MemberId)
			if gErr != nil {
				l.Errorf("[活动事件] 查询会员 global_id 失败，member_global_id 置为空串: member_id=%d err=%v", in.MemberId, gErr)
			}
			l.svcCtx.KafkaEventHandler.PublishDepositOrderApprovedEvent(
				asyncCtx,
				kafka.DepositApprovedEventInput{
					OperatorCode:   claims.OperatorCode,
					MemberID:       in.MemberId,
					OrderID:        resp.Id,
					Amount:         in.DepositAmount,
					Currency:       constants.DepositCurrency,
					DepositCount:   int(depositCount),
					ClientType:     constants.ClientTypePCWeb,
					MemberGlobalID: memberGlobalID,
				},
			)
		}()
	}

	return resp, nil
}

