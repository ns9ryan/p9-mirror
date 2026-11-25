package agentdepositorderservicelogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/common/utils/orderno"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentdepositorderpb"
)

// orderNoGen 存款订单号生成器，nodeID 自动从本机内网 IP 推导。
// 订单号前缀 DEP（Deposit）。
var orderNoGen = orderno.NewSnowflake(0)

const depositOrderNoPrefix = "DEP"

type AddAgentDepositOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddAgentDepositOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddAgentDepositOrderLogic {
	return &AddAgentDepositOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// group: agent_deposit_order 添加人工存款订单
func (l *AddAgentDepositOrderLogic) AddAgentDepositOrder(in *agentdepositorderpb.AddAgentDepositOrderReq) (*agentdepositorderpb.AddAgentDepositOrderResp, error) {
	if in == nil {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	// deposit_type 必填，仅允许 1 加密货币 / 2 人工入账
	if in.DepositType < 1 || in.DepositType > 2 {
		return nil, xerr.BadRequest(i18n.DepositTypeInvalid)
	}
	if in.MemberId <= 0 {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	if in.DepositAmount <= 0 {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	if in.DepositCurrency == "" {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	if in.CurrencyFactor <= 0 {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	if in.RolloverMultiplier < 0 {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	// bonus_rate 单位为百分比×100，0~10000 对应 0%~100%
	if in.BonusRate < 0 || in.BonusRate > 10000 {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}

	orderNo := orderNoGen.GenerateWithPrefix(depositOrderNoPrefix)
	resp, err := l.svcCtx.Agent.AddAgentDepositOrder(l.ctx, in, orderNo)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}
	// TODO 在有统一保存用户洗码后需要把记录的洗码数据保存过去
	return resp, nil
}
