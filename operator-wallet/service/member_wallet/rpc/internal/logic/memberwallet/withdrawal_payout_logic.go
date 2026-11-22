package memberwalletlogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/enum"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/svc"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/wallet"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc/withdrawpb"
)

// WithdrawalPayoutLogic 处理会员提现出款
type WithdrawalPayoutLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewWithdrawalPayoutLogic 创建提现出款逻辑
func NewWithdrawalPayoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WithdrawalPayoutLogic {
	return &WithdrawalPayoutLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// WithdrawalPayout 扣除已冻结的提现金额
func (l *WithdrawalPayoutLogic) WithdrawalPayout(in *withdrawpb.WithdrawalPayoutRequest) (*withdrawpb.WithdrawalPayoutResponse, error) {
	// 校验请求参数
	params, err := parseWithdrawalParams(
		in.GetMemberGlobalId(),    // 会员全局唯一标识
		in.GetOperationId(),       // 钱包操作唯一标识, UUID
		in.GetWithdrawalOrderNo(), // 提现订单号
		in.GetCurrencyCode(),      // 钱包结算币种编码
		in.GetAmount(),            // 冻结金额, 币种最小单位
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

	// 提交提现出款操作
	_, err = l.svcCtx.WalletEngine.Submit(
		l.ctx,
		walletGlobalID,
		wallet.RecordParams{
			// 操作与业务来源
			OperationID:    params.operationID,                  // 钱包操作唯一标识
			BusinessType:   enum.BusinessTypeWithdraw,           // 一级业务类型: 提现
			BusinessAction: enum.BusinessActionWithdrawalPayout, // 具体资金操作类型: 提现出款
			BusinessNo:     params.withdrawalNo,                 // 提现订单号

			// 扣除冻结余额
			Changes: []wallet.BalanceChange{
				{
					BalanceType:  enum.BalanceTypeFrozen, // 冻结余额
					ChangeAmount: -params.amount,         // 扣除
				},
			},
		},
	)
	if err != nil {
		return nil, handleWalletSubmitError(err, "withdrawal payout", params.operationID, l.Logger)
	}

	return &withdrawpb.WithdrawalPayoutResponse{}, nil
}
