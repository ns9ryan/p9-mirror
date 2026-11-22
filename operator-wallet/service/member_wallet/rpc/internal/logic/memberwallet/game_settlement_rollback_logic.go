package memberwalletlogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/enum"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/svc"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/wallet"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc/gamepb"
)

// GameSettlementRollbackLogic 处理会员游戏结算回滚
type GameSettlementRollbackLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewGameSettlementRollbackLogic 创建游戏结算回滚逻辑
func NewGameSettlementRollbackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GameSettlementRollbackLogic {
	return &GameSettlementRollbackLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GameSettlementRollback 撤销原游戏结算, 恢复未结算余额
func (l *GameSettlementRollbackLogic) GameSettlementRollback(in *gamepb.GameSettlementRollbackRequest) (*gamepb.GameSettlementRollbackResponse, error) {
	// 校验公共参数
	params, err := parseGameOperationParams(
		in.GetMemberGlobalId(), // 会员全局唯一标识
		in.GetOperationId(),    // 钱包操作唯一标识, UUID
		in.GetProviderBetNo(),  // 原始游戏厂商下注单号
		in.GetCurrencyCode(),   // 钱包结算币种编码
	)
	if err != nil {
		return nil, err
	}

	// 校验原下注金额和原派彩金额
	betAmount := in.GetBetAmount()
	payoutAmount := in.GetPayoutAmount()

	if err := validateGameAmount(betAmount, false); err != nil {
		return nil, err
	}
	if err := validateGameAmount(payoutAmount, true); err != nil {
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

	// 撤销原结算: 扣回派彩, 恢复原下注金额
	changes := make([]wallet.BalanceChange, 0, 2)

	// 原派彩为零时不产生零金额流水
	if payoutAmount > 0 {
		changes = append(changes, wallet.BalanceChange{
			BalanceType:  enum.BalanceTypeAvailable, // 可用余额
			ChangeAmount: -payoutAmount,             // 扣除
		})
	}

	changes = append(changes, wallet.BalanceChange{
		BalanceType:  enum.BalanceTypeUnsettled, // 游戏未结算余额
		ChangeAmount: betAmount,                 // 增加
	})

	// 提交游戏结算回滚操作
	_, err = l.svcCtx.WalletEngine.Submit(
		l.ctx,
		walletGlobalID,
		wallet.RecordParams{
			// 操作与业务来源
			OperationID:    params.operationID,                        // 钱包操作唯一标识
			BusinessType:   enum.BusinessTypeGame,                     // 一级业务类型: 游戏
			BusinessAction: enum.BusinessActionGameSettlementRollback, // 具体资金操作类型: 结算后回滚
			BusinessNo:     params.businessNo,                         // 游戏业务单号

			// 本次操作的余额变动集合
			Changes: changes,
		},
	)
	if err != nil {
		return nil, handleWalletSubmitError(err, "game settlement rollback", params.operationID, l.Logger)
	}

	return &gamepb.GameSettlementRollbackResponse{}, nil
}
