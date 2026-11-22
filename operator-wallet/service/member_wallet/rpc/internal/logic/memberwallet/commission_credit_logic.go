package memberwalletlogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/enum"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/svc"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/wallet"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc/rewardpb"
)

// CommissionCreditLogic 处理全民代理佣金入账
type CommissionCreditLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewCommissionCreditLogic 创建全民代理佣金入账逻辑
func NewCommissionCreditLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CommissionCreditLogic {
	return &CommissionCreditLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CommissionCredit 发放全民代理佣金
func (l *CommissionCreditLogic) CommissionCredit(in *rewardpb.CommissionCreditRequest) (*rewardpb.CommissionCreditResponse, error) {
	// 校验请求参数
	params, err := parseRewardOperationParams(
		in.GetMemberGlobalId(),    // 会员全局唯一标识
		in.GetOperationId(),       // 钱包操作唯一标识, UUID
		in.GetCommissionOrderNo(), // 佣金结算订单号
		in.GetCurrencyCode(),      // 钱包结算币种编码
		in.GetAmount(),            // 佣金金额, 币种最小单位
	)
	if err != nil {
		return nil, err
	}

	// 确保会员钱包已加载并注册到 Engine
	walletGlobalID, err := ensureWalletActorForRPC(
		l.ctx,
		l.svcCtx,
		walletActorParams{
			memberGlobalID: params.memberGlobalID, // 会员全局唯一标识
			currencyCode:   params.currencyCode,   // 请求币种编码
		},
		params.operationID, // 钱包操作唯一标识
		l.Logger,
	)
	if err != nil {
		return nil, err
	}

	// 提交全民代理佣金入账操作
	_, err = l.svcCtx.WalletEngine.Submit(
		l.ctx,
		walletGlobalID,
		wallet.RecordParams{
			// 操作与业务来源
			OperationID:    params.operationID,                  // 钱包操作唯一标识
			BusinessType:   enum.BusinessTypeCommission,         // 一级业务类型: 佣金
			BusinessAction: enum.BusinessActionCommissionCredit, // 具体资金操作类型: 全民代理佣金入账
			BusinessNo:     params.businessNo,                   // 奖励业务订单号

			// 增加可用余额
			Changes: []wallet.BalanceChange{
				{
					BalanceType:  enum.BalanceTypeAvailable, // 可用余额
					ChangeAmount: params.amount,             // 增加
				},
			},
		},
	)
	if err != nil {
		return nil, handleWalletSubmitError(err, "commission credit", params.operationID, l.Logger)
	}

	return &rewardpb.CommissionCreditResponse{}, nil
}
