package memberwalletlogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/enum"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/svc"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/wallet"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc/withdrawpb"
)

// WithdrawalFreezeLogic 处理会员提现冻结
type WithdrawalFreezeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewWithdrawalFreezeLogic 创建提现冻结逻辑
func NewWithdrawalFreezeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WithdrawalFreezeLogic {
	return &WithdrawalFreezeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// WithdrawalFreeze 冻结会员提现金额
func (l *WithdrawalFreezeLogic) WithdrawalFreeze(in *withdrawpb.WithdrawalFreezeRequest) (*withdrawpb.WithdrawalFreezeResponse, error) {
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

	// 提交提现冻结操作
	_, err = l.svcCtx.WalletEngine.Submit(
		l.ctx,
		walletGlobalID,
		wallet.RecordParams{
			// 操作与业务来源
			OperationID:    params.operationID,                  // 钱包操作唯一标识
			BusinessType:   enum.BusinessTypeWithdraw,           // 一级业务类型: 提现
			BusinessAction: enum.BusinessActionWithdrawalFreeze, // 具体资金操作类型: 提现冻结
			BusinessNo:     params.withdrawalNo,                 // 提现订单号

			// 可用余额转入冻结余额
			Changes: []wallet.BalanceChange{
				{
					BalanceType:  enum.BalanceTypeAvailable, // 可用余额
					ChangeAmount: -params.amount,            // 扣除
				},
				{
					BalanceType:  enum.BalanceTypeFrozen, // 冻结余额
					ChangeAmount: params.amount,          // 增加
				},
			},
		},
	)
	if err != nil {
		return nil, handleWalletSubmitError(err, "withdrawal freeze", params.operationID, l.Logger)
	}

	return &withdrawpb.WithdrawalFreezeResponse{}, nil
}
