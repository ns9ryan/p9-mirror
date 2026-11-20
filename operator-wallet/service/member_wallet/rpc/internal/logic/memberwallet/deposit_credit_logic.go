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
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc/depositpb"
)

// DepositCreditLogic 处理会员存款本金入账
type DepositCreditLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewDepositCreditLogic 创建存款本金入账逻辑
func NewDepositCreditLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DepositCreditLogic {
	return &DepositCreditLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// depositCreditParams 存款本金入账参数
type depositCreditParams struct {
	memberGlobalID uuid.UUID // 会员全局唯一标识
	operationID    uuid.UUID // 钱包操作唯一标识
	depositOrderNo string    // 存款订单号
	currencyCode   string    // 币种编码
	amount         int64     // 入账金额
}

// DepositCredit 存款本金入账
func (l *DepositCreditLogic) DepositCredit(in *depositpb.DepositCreditRequest) (*depositpb.DepositCreditResponse, error) {
	// 校验请求参数
	params, err := parseDepositCreditRequest(in)
	if err != nil {
		return nil, err
	}

	// 确保会员钱包已加载并注册到 Engine
	walletGlobalID, err := ensureWalletActorForRPC(l.ctx, l.svcCtx,
		walletActorParams{
			memberGlobalID: params.memberGlobalID, // 会员全局唯一标识
			currencyCode:   params.currencyCode,   // 请求币种编码
		},
		params.operationID, l.Logger,
	)
	if err != nil {
		return nil, err
	}

	// 提交存款本金入账操作, Actor 串行处理, Kafka 确认后才更新正式钱包状态
	_, err = l.svcCtx.WalletEngine.Submit(
		l.ctx,
		walletGlobalID, // 钱包全局唯一标识
		wallet.RecordParams{
			// 操作与业务来源
			OperationID:    params.operationID,               // 钱包操作唯一标识
			BusinessType:   enum.BusinessTypeDeposit,         // 一级业务类型
			BusinessAction: enum.BusinessActionDepositCredit, // 具体资金操作类型
			BusinessNo:     params.depositOrderNo,            // 原始业务单号

			// 本次操作的余额变动集合
			Changes: []wallet.BalanceChange{
				{
					BalanceType:  enum.BalanceTypeAvailable, // 余额类型: 可用余额
					ChangeAmount: params.amount,             // 变动金额, 正数增加, 负数减少
				},
			},
		},
	)
	if err != nil {
		return nil, handleWalletSubmitError(err, "deposit credit", params.operationID, l.Logger)
	}

	return &depositpb.DepositCreditResponse{}, nil
}

// parseDepositCreditRequest 校验并解析存款入账请求
func parseDepositCreditRequest(in *depositpb.DepositCreditRequest) (depositCreditParams, error) {
	// 校验会员和操作标识
	memberGlobalID, err := uuid.Parse(in.GetMemberGlobalId())
	if err != nil || memberGlobalID == uuid.Nil {
		return depositCreditParams{}, status.Error(codes.InvalidArgument, "会员全局标识无效")
	}

	// 校验钱包操作标识
	operationID, err := uuid.Parse(in.GetOperationId())
	if err != nil || operationID == uuid.Nil {
		return depositCreditParams{}, status.Error(codes.InvalidArgument, "钱包操作标识无效")
	}

	// 校验存款订单信息
	orderNo := in.GetDepositOrderNo()
	if orderNo == "" || strings.TrimSpace(orderNo) != orderNo || utf8.RuneCountInString(orderNo) > 128 {
		return depositCreditParams{}, status.Error(codes.InvalidArgument, "存款订单号无效")
	}

	// 校验币种编码
	currencyCode := in.GetCurrencyCode()
	if currencyCode == "" || strings.TrimSpace(currencyCode) != currencyCode || utf8.RuneCountInString(currencyCode) > 16 {
		return depositCreditParams{}, status.Error(codes.InvalidArgument, "币种编码无效")
	}

	// 校验入账金额
	amount := in.GetAmount()
	if amount <= 0 {
		return depositCreditParams{}, status.Error(codes.InvalidArgument, "入账金额必须大于零")
	}

	// 存款本金入账参数
	return depositCreditParams{
		memberGlobalID: memberGlobalID, // 会员全局唯一标识
		operationID:    operationID,    // 钱包操作唯一标识
		depositOrderNo: orderNo,        // 存款订单号
		currencyCode:   currencyCode,   // 币种编码
		amount:         amount,         // 入账金额
	}, nil
}
