package memberwalletlogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/enum"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/svc"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/wallet"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc/gamepb"
)

// GameSettlementLogic 处理会员游戏结算
type GameSettlementLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewGameSettlementLogic 创建游戏结算逻辑
func NewGameSettlementLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GameSettlementLogic {
	return &GameSettlementLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GameSettlement 结算游戏下注并发放派彩
func (l *GameSettlementLogic) GameSettlement(in *gamepb.GameSettlementRequest) (*gamepb.GameSettlementResponse, error) {
	// 校验公共参数
	params, err := parseGameOperationParams(
		in.GetMemberGlobalId(), // 会员全局唯一标识
		in.GetOperationId(),    // 钱包操作唯一标识, UUID
		in.GetProviderBetNo(),  // 游戏厂商原始下注单号
		in.GetCurrencyCode(),   // 钱包结算币种编码
	)
	if err != nil {
		return nil, err
	}

	// 校验下注金额和派彩金额
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

	// 未结算余额扣除原下注金额, 可用余额增加实际派彩
	changes := []wallet.BalanceChange{
		{
			BalanceType:  enum.BalanceTypeUnsettled, // 游戏未结算余额
			ChangeAmount: -betAmount,                // 扣除原始下注金额, 币种最小单位
		},
	}

	// 派彩为零时不产生零金额流水
	if payoutAmount > 0 {
		changes = append(changes, wallet.BalanceChange{
			BalanceType:  enum.BalanceTypeAvailable, // 可用余额
			ChangeAmount: payoutAmount,              // 增加实际派彩金额, 币种最小单位
		})
	}

	// 提交游戏结算操作
	_, err = l.svcCtx.WalletEngine.Submit(
		l.ctx,
		walletGlobalID,
		wallet.RecordParams{
			// 操作与业务来源
			OperationID:    params.operationID,                // 钱包操作唯一标识
			BusinessType:   enum.BusinessTypeGame,             // 一级业务类型: 游戏
			BusinessAction: enum.BusinessActionGameSettlement, // 具体资金操作类型: 游戏结算
			BusinessNo:     params.businessNo,                 // 游戏业务单号

			// 本次操作的余额变动集合
			Changes: changes,
		},
	)
	if err != nil {
		return nil, handleWalletSubmitError(err, "game settlement", params.operationID, l.Logger)
	}

	return &gamepb.GameSettlementResponse{}, nil
}
