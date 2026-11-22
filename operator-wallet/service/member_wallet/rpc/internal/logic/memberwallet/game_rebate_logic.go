package memberwalletlogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/enum"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/svc"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/wallet"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc/gamepb"
)

// GameRebateLogic 处理会员游戏返水
type GameRebateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewGameRebateLogic 创建游戏返水逻辑
func NewGameRebateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GameRebateLogic {
	return &GameRebateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GameRebate 发放会员游戏返水
func (l *GameRebateLogic) GameRebate(in *gamepb.GameRebateRequest) (*gamepb.GameRebateResponse, error) {
	// 校验公共参数
	params, err := parseGameOperationParams(
		in.GetMemberGlobalId(), // 会员全局唯一标识
		in.GetOperationId(),    // 钱包操作唯一标识, UUID
		in.GetRebateOrderNo(),  // 返水业务订单号
		in.GetCurrencyCode(),   // 钱包结算币种编码
	)
	if err != nil {
		return nil, err
	}

	// 校验返水金额
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

	// 提交游戏返水操作
	_, err = l.svcCtx.WalletEngine.Submit(
		l.ctx,
		walletGlobalID,
		wallet.RecordParams{
			// 操作与业务来源
			OperationID:    params.operationID,            // 钱包操作唯一标识
			BusinessType:   enum.BusinessTypeGame,         // 一级业务类型: 游戏
			BusinessAction: enum.BusinessActionGameRebate, // 具体资金操作类型: 游戏返水/返利
			BusinessNo:     params.businessNo,             // 游戏业务单号

			// 增加可用余额
			Changes: []wallet.BalanceChange{
				{
					BalanceType:  enum.BalanceTypeAvailable, // 可用余额
					ChangeAmount: amount,                    // 增加
				},
			},
		},
	)
	if err != nil {
		return nil, handleWalletSubmitError(err, "game rebate", params.operationID, l.Logger)
	}

	return &gamepb.GameRebateResponse{}, nil
}
