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

// GameBetCancelLogic 处理会员游戏下注取消
type GameBetCancelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewGameBetCancelLogic 创建游戏下注取消逻辑
func NewGameBetCancelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GameBetCancelLogic {
	return &GameBetCancelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// gameBetCancelParams 游戏下注取消参数
type gameBetCancelParams struct {
	memberGlobalID uuid.UUID // 会员全局唯一标识
	operationID    uuid.UUID // 钱包操作唯一标识
	providerBetNo  string    // 原始游戏下注单号
	currencyCode   string    // 币种编码
}

// GameBetCancel 游戏下注取消
func (l *GameBetCancelLogic) GameBetCancel(in *gamepb.GameBetCancelRequest) (*gamepb.GameBetCancelResponse, error) {
	// 校验请求参数
	params, err := parseGameBetCancelRequest(in)
	if err != nil {
		return nil, err
	}

	// 校验厅身份, 锁定会员钱包并执行事务
	err = withLockedMemberWalletTx(
		l.ctx,
		l.svcCtx,
		params.memberGlobalID,
		params.currencyCode,
		"取消币种与钱包币种不一致",
		func(tx *ent.Tx, memberWallet *ent.MemberWallet) error {
			// 查询原始下注流水
			betLedgers, err := l.svcCtx.DAO.MemberWalletLedger.ListByBusinessTx(
				l.ctx,
				tx,
				int64(enum.BusinessTypeGame),
				int64(enum.BusinessActionGameBet),
				params.providerBetNo,
			)
			if err != nil {
				return l.internalError("查询原始下注流水失败", err)
			}

			if len(betLedgers) == 0 {
				return status.Error(codes.NotFound, "原始下注记录不存在")
			}

			// 原始下注必须属于当前会员钱包
			for _, ledger := range betLedgers {
				if ledger == nil {
					return l.internalError("原始下注流水异常", errors.New("nil ledger"))
				}
				if ledger.WalletID != memberWallet.ID {
					return status.Error(codes.NotFound, "原始下注记录不存在")
				}
			}

			// 从原始下注流水中读取真实下注金额
			betAmount, ok := gameBetAmountFromRecords(
				betLedgers,
				memberWallet.ID,
				params.providerBetNo,
			)
			if !ok {
				return l.internalError("原始下注流水异常", errors.New("invalid game bet ledgers"))
			}

			// 已经产生结算流水的下注不允许取消
			settlementLedgers, err := l.svcCtx.DAO.MemberWalletLedger.ListByBusinessTx(
				l.ctx,
				tx,
				int64(enum.BusinessTypeGame),
				int64(enum.BusinessActionGameSettlement),
				params.providerBetNo,
			)
			if err != nil {
				return l.internalError("查询游戏结算流水失败", err)
			}

			if len(settlementLedgers) > 0 {
				return status.Error(codes.FailedPrecondition, "游戏下注已经结算，不能取消")
			}

			// 按原始下注单号检查取消业务幂等
			cancelLedgers, err := l.svcCtx.DAO.MemberWalletLedger.ListByBusinessTx(
				l.ctx,
				tx,
				int64(enum.BusinessTypeGame),
				int64(enum.BusinessActionGameBetCancel),
				params.providerBetNo,
			)
			if err != nil {
				return l.internalError("查询下注取消流水失败", err)
			}

			if len(cancelLedgers) > 0 {
				if !matchGameBetCancelRecords(
					cancelLedgers,
					memberWallet.ID,
					params.providerBetNo,
					betAmount,
				) {
					return status.Error(codes.AlreadyExists, "游戏下注已经取消")
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
				BusinessAction: enum.BusinessActionGameBetCancel,
				BusinessNo:     params.providerBetNo,
				Changes: []wallet.BalanceChange{
					{
						BalanceType:  enum.BalanceTypeAvailable,
						ChangeAmount: betAmount,
					},
					{
						BalanceType:  enum.BalanceTypeUnsettled,
						ChangeAmount: -betAmount,
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
		return nil, l.internalError("执行下注取消事务失败", err)
	}

	return &gamepb.GameBetCancelResponse{}, nil
}

// parseGameBetCancelRequest 校验并解析下注取消请求
func parseGameBetCancelRequest(in *gamepb.GameBetCancelRequest) (gameBetCancelParams, error) {
	// 校验会员和操作标识
	memberGlobalID, err := uuid.Parse(in.GetMemberGlobalId())
	if err != nil || memberGlobalID == uuid.Nil {
		return gameBetCancelParams{}, status.Error(codes.InvalidArgument, "会员全局标识无效")
	}

	operationID, err := uuid.Parse(in.GetOperationId())
	if err != nil || operationID == uuid.Nil {
		return gameBetCancelParams{}, status.Error(codes.InvalidArgument, "钱包操作标识无效")
	}

	// 校验原始下注单号
	providerBetNo := in.GetProviderBetNo()
	if providerBetNo == "" ||
		strings.TrimSpace(providerBetNo) != providerBetNo ||
		utf8.RuneCountInString(providerBetNo) > 128 {
		return gameBetCancelParams{}, status.Error(codes.InvalidArgument, "游戏下注单号无效")
	}

	// 校验币种编码
	currencyCode := in.GetCurrencyCode()
	if currencyCode == "" ||
		strings.TrimSpace(currencyCode) != currencyCode ||
		utf8.RuneCountInString(currencyCode) > 16 {
		return gameBetCancelParams{}, status.Error(codes.InvalidArgument, "币种编码无效")
	}

	return gameBetCancelParams{
		memberGlobalID: memberGlobalID,
		operationID:    operationID,
		providerBetNo:  providerBetNo,
		currencyCode:   currencyCode,
	}, nil
}

// gameBetAmountFromRecords 从原始下注流水中读取下注金额
func gameBetAmountFromRecords(existing []*ent.MemberWalletLedger, walletID int64, providerBetNo string) (int64, bool) {
	if len(existing) != 2 {
		return 0, false
	}

	var betAmount int64

	for _, ledger := range existing {
		if ledger == nil {
			return 0, false
		}

		if enum.BalanceType(ledger.BalanceType) != enum.BalanceTypeUnsettled {
			continue
		}

		if betAmount != 0 || ledger.ChangeAmount <= 0 {
			return 0, false
		}

		betAmount = ledger.ChangeAmount
	}

	if betAmount == 0 {
		return 0, false
	}

	// 校验原始下注的两条余额流水及业务归属
	if !matchBusinessLedgerChanges(
		existing,
		walletID,
		enum.BusinessTypeGame,
		enum.BusinessActionGameBet,
		providerBetNo,
		map[enum.BalanceType]int64{
			enum.BalanceTypeAvailable: -betAmount,
			enum.BalanceTypeUnsettled: betAmount,
		},
	) {
		return 0, false
	}

	return betAmount, true
}

// matchGameBetCancelRecords 判断已有取消流水是否与原下注完全一致
func matchGameBetCancelRecords(existing []*ent.MemberWalletLedger, walletID int64, providerBetNo string, betAmount int64) bool {
	if betAmount <= 0 {
		return false
	}

	return matchBusinessLedgerChanges(
		existing,
		walletID,
		enum.BusinessTypeGame,
		enum.BusinessActionGameBetCancel,
		providerBetNo,
		map[enum.BalanceType]int64{
			enum.BalanceTypeAvailable: betAmount,
			enum.BalanceTypeUnsettled: -betAmount,
		},
	)
}

// internalError 记录内部错误并返回统一RPC错误
func (l *GameBetCancelLogic) internalError(action string, err error) error {
	l.Errorf("%s: %v", action, err)
	return status.Error(codes.Internal, "下注取消失败")
}

// mapRecordError 将统一记账错误转换为RPC业务错误
func (l *GameBetCancelLogic) mapRecordError(err error) error {
	switch {
	case errors.Is(err, wallet.ErrInvalidRecordParams):
		return status.Error(codes.InvalidArgument, "下注取消记账参数无效")

	case errors.Is(err, wallet.ErrOperationConflict):
		return status.Error(codes.AlreadyExists, "钱包操作标识已被使用")

	case errors.Is(err, wallet.ErrBalanceOverflow):
		return status.Error(codes.OutOfRange, "下注取消后余额超出允许范围")

	case errors.Is(err, wallet.ErrInsufficientBalance):
		return status.Error(codes.FailedPrecondition, "游戏未结算余额不足")

	case dao.IsGameBetCancelUniqueConflict(err):
		return status.Error(codes.AlreadyExists, "游戏下注已经取消")

	case dao.IsWalletOperationUniqueConflict(err):
		return status.Error(codes.AlreadyExists, "钱包操作标识已被使用")

	default:
		return l.internalError("统一资金记账失败", err)
	}
}
