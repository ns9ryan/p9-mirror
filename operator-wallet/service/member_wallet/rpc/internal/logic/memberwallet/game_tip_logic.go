package memberwalletlogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/enum"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/svc"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/wallet"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc/gamepb"
)

// GameTipLogic 处理会员游戏消费打赏
type GameTipLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewGameTipLogic 创建游戏消费打赏逻辑
func NewGameTipLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GameTipLogic {
	return &GameTipLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GameTip 扣除会员游戏消费打赏金额
func (l *GameTipLogic) GameTip(in *gamepb.GameTipRequest) (*gamepb.GameTipResponse, error) {
	// 校验公共参数
	params, err := parseGameOperationParams(
		in.GetMemberGlobalId(),  // 会员全局唯一标识
		in.GetOperationId(),     // 钱包操作唯一标识, UUID
		in.GetProviderOrderNo(), // 游戏厂商打赏订单号
		in.GetCurrencyCode(),    // 钱包结算币种编码
	)
	if err != nil {
		return nil, err
	}

	// 校验打赏金额
	amount := in.GetAmount()
	if err := validateGameAmount(amount, false); err != nil {
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

	// 提交游戏打赏操作
	_, err = l.svcCtx.WalletEngine.Submit(
		l.ctx,
		walletGlobalID,
		wallet.RecordParams{
			// 操作与业务来源
			OperationID:    params.operationID,         // 钱包操作唯一标识
			BusinessType:   enum.BusinessTypeGame,      // 一级业务类型: 游戏
			BusinessAction: enum.BusinessActionGameTip, // 具体资金操作类型: 游戏消费打赏
			BusinessNo:     params.businessNo,          // 游戏业务单号

			// 扣除可用余额
			Changes: []wallet.BalanceChange{
				{
					BalanceType:  enum.BalanceTypeAvailable, // 可用余额
					ChangeAmount: -amount,                   // 扣除
				},
			},
		},
	)
	if err != nil {
		return nil, handleWalletSubmitError(err, "game tip", params.operationID, l.Logger)
	}

	return &gamepb.GameTipResponse{}, nil
}
