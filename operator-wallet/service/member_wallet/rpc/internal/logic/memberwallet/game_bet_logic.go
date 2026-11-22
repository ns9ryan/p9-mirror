package memberwalletlogic

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/enum"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/svc"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/wallet"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc/gamepb"
)

// GameBetLogic 处理会员游戏下注
type GameBetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewGameBetLogic 创建游戏下注逻辑
func NewGameBetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GameBetLogic {
	return &GameBetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// gameBetParams 游戏下注参数
type gameBetParams struct {
	// 基础信息
	memberGlobalID uuid.UUID // 会员全局唯一标识
	operationID    uuid.UUID // 钱包操作唯一标识

	// 业务信息
	providerBetNo string // 游戏厂商原始下注单号
	currencyCode  string // 币种编码
	betAmount     int64  // 下注金额
}

// GameBet 游戏下注
func (l *GameBetLogic) GameBet(in *gamepb.GameBetRequest) (*gamepb.GameBetResponse, error) {
	// 校验请求参数
	params, err := parseGameBetRequest(in)
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

	// 提交游戏下注操作
	// 可用余额扣除下注金额, 未结算余额增加相同金额
	_, err = l.svcCtx.WalletEngine.Submit(
		l.ctx,
		walletGlobalID,
		wallet.RecordParams{
			// 操作与业务来源
			OperationID:    params.operationID,         // 钱包操作唯一标识
			BusinessType:   enum.BusinessTypeGame,      // 一级业务类型: 游戏
			BusinessAction: enum.BusinessActionGameBet, // 具体资金操作类型: 游戏下注
			BusinessNo:     params.providerBetNo,       // 游戏厂商原始下注单号

			// 本次操作的余额变动集合
			Changes: []wallet.BalanceChange{
				{
					BalanceType:  enum.BalanceTypeAvailable, // 可用余额
					ChangeAmount: -params.betAmount,         // 扣除下注金额
				},
				{
					BalanceType:  enum.BalanceTypeUnsettled, // 游戏未结算余额
					ChangeAmount: params.betAmount,          // 增加下注金额
				},
			},
		},
	)
	if err != nil {
		return nil, handleWalletSubmitError(err, "game bet", params.operationID, l.Logger)
	}

	return &gamepb.GameBetResponse{}, nil
}

// parseGameBetRequest 校验并解析游戏下注请求
func parseGameBetRequest(in *gamepb.GameBetRequest) (gameBetParams, error) {
	// 校验会员和操作标识
	memberGlobalID, err := uuid.Parse(in.GetMemberGlobalId())
	if err != nil || memberGlobalID == uuid.Nil {
		return gameBetParams{}, status.Error(codes.InvalidArgument, "会员全局标识无效")
	}

	// 校验钱包操作标识
	operationID, err := uuid.Parse(in.GetOperationId())
	if err != nil || operationID == uuid.Nil {
		return gameBetParams{}, status.Error(codes.InvalidArgument, "钱包操作标识无效")
	}

	// 校验游戏厂商原始下注单号
	providerBetNo := in.GetProviderBetNo()
	if providerBetNo == "" || strings.TrimSpace(providerBetNo) != providerBetNo || utf8.RuneCountInString(providerBetNo) > 128 {
		return gameBetParams{}, status.Error(codes.InvalidArgument, "游戏下注单号无效")
	}

	// 校验币种编码
	currencyCode := in.GetCurrencyCode()
	if currencyCode == "" || strings.TrimSpace(currencyCode) != currencyCode || utf8.RuneCountInString(currencyCode) > 16 {
		return gameBetParams{}, status.Error(codes.InvalidArgument, "币种编码无效")
	}

	// 校验下注金额
	betAmount := in.GetBetAmount()
	if betAmount <= 0 {
		return gameBetParams{}, status.Error(codes.InvalidArgument, "下注金额必须大于零")
	}

	// 游戏下注参数
	return gameBetParams{
		// 基础信息
		memberGlobalID: memberGlobalID, // 会员全局唯一标识
		operationID:    operationID,    // 钱包操作唯一标识

		// 业务信息
		providerBetNo: providerBetNo, // 游戏厂商原始下注单号
		currencyCode:  currencyCode,  // 币种编码
		betAmount:     betAmount,     // 下注金额
	}, nil
}
