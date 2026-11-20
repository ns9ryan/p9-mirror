package memberwalletlogic

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/enum"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/svc"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/wallet"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc/depositpb"

	"github.com/zeromicro/go-zero/core/logx"
)

// DepositBonusLogic 处理会员存款赠送入账
type DepositBonusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewDepositBonusLogic 创建存款赠送入账逻辑
func NewDepositBonusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DepositBonusLogic {
	return &DepositBonusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// depositBonusParams 存款赠送入账参数
type depositBonusParams struct {
	memberGlobalID uuid.UUID // 会员全局唯一标识
	operationID    uuid.UUID // 钱包操作唯一标识
	depositOrderNo string    // 存款订单号
	currencyCode   string    // 币种编码
	amount         int64     // 赠送金额
}

// DepositBonus 存款赠送入账
func (l *DepositBonusLogic) DepositBonus(in *depositpb.DepositBonusRequest) (*depositpb.DepositBonusResponse, error) {
	// 校验请求参数
	params, err := parseDepositBonusRequest(in)
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
		params.operationID,
		l.Logger,
	)
	if err != nil {
		return nil, err
	}

	// 提交存款赠送入账操作
	_, err = l.svcCtx.WalletEngine.Submit(
		l.ctx,
		walletGlobalID,
		wallet.RecordParams{
			// 操作与业务来源
			OperationID:    params.operationID,              // 钱包操作唯一标识
			BusinessType:   enum.BusinessTypeDeposit,        // 一级业务类型
			BusinessAction: enum.BusinessActionDepositBonus, // 具体资金操作类型
			BusinessNo:     params.depositOrderNo,           // 原始业务单号

			// 本次操作的余额变动集合
			Changes: []wallet.BalanceChange{
				{
					BalanceType:  enum.BalanceTypeAvailable, // 可用余额
					ChangeAmount: params.amount,             // 增加赠送金额
				},
			},
		},
	)
	if err != nil {
		return nil, handleWalletSubmitError(err, "deposit bonus", params.operationID, l.Logger)
	}

	return &depositpb.DepositBonusResponse{}, nil
}

// parseDepositBonusRequest 校验并解析存款赠送请求
func parseDepositBonusRequest(in *depositpb.DepositBonusRequest) (depositBonusParams, error) {
	// 校验会员和操作标识
	memberGlobalID, err := uuid.Parse(in.GetMemberGlobalId())
	if err != nil || memberGlobalID == uuid.Nil {
		return depositBonusParams{}, status.Error(codes.InvalidArgument, "会员全局标识无效")
	}

	// 校验钱包操作标识
	operationID, err := uuid.Parse(in.GetOperationId())
	if err != nil || operationID == uuid.Nil {
		return depositBonusParams{}, status.Error(codes.InvalidArgument, "钱包操作标识无效")
	}

	// 校验存款订单信息
	orderNo := in.GetDepositOrderNo()
	if orderNo == "" || strings.TrimSpace(orderNo) != orderNo || utf8.RuneCountInString(orderNo) > 128 {
		return depositBonusParams{}, status.Error(codes.InvalidArgument, "存款订单号无效")
	}

	// 校验币种编码
	currencyCode := in.GetCurrencyCode()
	if currencyCode == "" || strings.TrimSpace(currencyCode) != currencyCode || utf8.RuneCountInString(currencyCode) > 16 {
		return depositBonusParams{}, status.Error(codes.InvalidArgument, "币种编码无效")
	}

	// 校验赠送金额
	amount := in.GetAmount()
	if amount <= 0 {
		return depositBonusParams{}, status.Error(codes.InvalidArgument, "赠送金额必须大于零")
	}

	// 存款赠送入账参数
	return depositBonusParams{
		memberGlobalID: memberGlobalID, // 会员全局唯一标识
		operationID:    operationID,    // 钱包操作唯一标识
		depositOrderNo: orderNo,        // 存款订单号
		currencyCode:   currencyCode,   // 币种编码
		amount:         amount,         // 赠送金额
	}, nil
}
