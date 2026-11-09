package memberwalletlogic

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"oa.98ent.com/p9/operator-wallet/ent"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/dao"
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
	memberGlobalID uuid.UUID // 会员全局唯一标识
	operationID    uuid.UUID // 钱包操作唯一标识
	providerBetNo  string    // 游戏厂商原始下注单号
	currencyCode   string    // 币种编码
	betAmount      int64     // 下注金额
}

// GameBet 游戏下注
func (l *GameBetLogic) GameBet(in *gamepb.GameBetRequest) (*gamepb.GameBetResponse, error) {
	// 校验请求参数
	params, err := parseGameBetRequest(in)
	if err != nil {
		return nil, err
	}

	// 校验厅身份, 锁定会员钱包并执行事务
	err = withLockedMemberWalletTx(
		l.ctx,
		l.svcCtx,
		params.memberGlobalID,
		params.currencyCode,
		"下注币种与钱包币种不一致",
		func(tx *ent.Tx, memberWallet *ent.MemberWallet) error {
			// 按原始下注单号检查业务幂等
			existing, err := l.svcCtx.DAO.MemberWalletLedger.ListByBusinessTx(
				l.ctx,
				tx,
				int64(enum.BusinessTypeGame),
				int64(enum.BusinessActionGameBet),
				params.providerBetNo,
			)
			if err != nil {
				return l.internalError("查询游戏下注流水失败", err)
			}

			if len(existing) > 0 {
				if !matchGameBetRecords(
					existing,
					memberWallet.ID,
					params.providerBetNo,
					params.betAmount,
				) {
					return status.Error(codes.AlreadyExists, "游戏下注订单已存在")
				}

				return nil
			}

			// 创建绑定当前事务和钱包的记账器
			recorder, err := wallet.NewRecorder(tx, l.svcCtx.DAO, memberWallet)
			if err != nil {
				return l.internalError("初始化钱包记账器失败", err)
			}

			// 执行统一资金记账
			if err := recorder.Record(l.ctx, wallet.RecordParams{
				OperationID:    params.operationID,
				BusinessType:   enum.BusinessTypeGame,
				BusinessAction: enum.BusinessActionGameBet,
				BusinessNo:     params.providerBetNo,
				Changes: []wallet.BalanceChange{
					{
						BalanceType:  enum.BalanceTypeAvailable,
						ChangeAmount: -params.betAmount,
					},
					{
						BalanceType:  enum.BalanceTypeUnsettled,
						ChangeAmount: params.betAmount,
					},
				},
			}); err != nil {
				return l.mapRecordError(err)
			}

			return nil
		},
	)
	if err != nil {
		// 保留事务回调中已经转换好的业务错误
		if _, ok := status.FromError(err); ok {
			return nil, err
		}
		return nil, l.internalError("执行游戏下注事务失败", err)
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

	operationID, err := uuid.Parse(in.GetOperationId())
	if err != nil || operationID == uuid.Nil {
		return gameBetParams{}, status.Error(codes.InvalidArgument, "钱包操作标识无效")
	}

	// 校验游戏厂商原始下注单号
	providerBetNo := in.GetProviderBetNo()
	if providerBetNo == "" ||
		strings.TrimSpace(providerBetNo) != providerBetNo ||
		utf8.RuneCountInString(providerBetNo) > 128 {
		return gameBetParams{}, status.Error(codes.InvalidArgument, "游戏下注单号无效")
	}

	// 校验币种编码
	currencyCode := in.GetCurrencyCode()
	if currencyCode == "" ||
		strings.TrimSpace(currencyCode) != currencyCode ||
		utf8.RuneCountInString(currencyCode) > 16 {
		return gameBetParams{}, status.Error(codes.InvalidArgument, "币种编码无效")
	}

	// 校验下注金额
	betAmount := in.GetBetAmount()
	if betAmount <= 0 {
		return gameBetParams{}, status.Error(codes.InvalidArgument, "下注金额必须大于零")
	}

	return gameBetParams{
		memberGlobalID: memberGlobalID,
		operationID:    operationID,
		providerBetNo:  providerBetNo,
		currencyCode:   currencyCode,
		betAmount:      betAmount,
	}, nil
}

// matchGameBetRecords 判断已有流水是否与本次下注完全一致
func matchGameBetRecords(existing []*ent.MemberWalletLedger, walletID int64, providerBetNo string, betAmount int64) bool {
	if betAmount <= 0 {
		return false
	}

	return matchBusinessLedgerChanges(
		existing,
		walletID,
		enum.BusinessTypeGame,
		enum.BusinessActionGameBet,
		providerBetNo,
		map[enum.BalanceType]int64{
			enum.BalanceTypeAvailable: -betAmount,
			enum.BalanceTypeUnsettled: betAmount,
		},
	)
}

// internalError 记录内部错误并返回统一RPC错误
func (l *GameBetLogic) internalError(action string, err error) error {
	l.Errorf("%s: %v", action, err)
	return status.Error(codes.Internal, "游戏下注失败")
}

// mapRecordError 将统一记账错误转换为RPC业务错误
func (l *GameBetLogic) mapRecordError(err error) error {
	switch {
	case errors.Is(err, wallet.ErrInvalidRecordParams):
		return status.Error(codes.InvalidArgument, "游戏下注记账参数无效")

	case errors.Is(err, wallet.ErrOperationConflict):
		return status.Error(codes.AlreadyExists, "钱包操作标识已被使用")

	case errors.Is(err, wallet.ErrBalanceOverflow):
		return status.Error(codes.OutOfRange, "游戏下注后余额超出允许范围")

	case errors.Is(err, wallet.ErrInsufficientBalance):
		return status.Error(codes.FailedPrecondition, "可用余额不足")

	case dao.IsGameBetUniqueConflict(err):
		return status.Error(codes.AlreadyExists, "游戏下注订单已存在")

	case dao.IsWalletOperationUniqueConflict(err):
		return status.Error(codes.AlreadyExists, "钱包操作标识已被使用")

	default:
		return l.internalError("统一资金记账失败", err)
	}
}
