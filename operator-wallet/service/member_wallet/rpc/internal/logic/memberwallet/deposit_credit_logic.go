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

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/operator-wallet/ent"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/dao"
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

	// 校验当前厅身份, 钱包数据范围由 Ent Interceptor 根据上下文隔离
	operatorCode := ctxdata.OperatorCodeFromCtx(l.ctx)
	if operatorCode == "" ||
		strings.TrimSpace(operatorCode) != operatorCode ||
		utf8.RuneCountInString(operatorCode) > 64 {
		return nil, status.Error(codes.Unauthenticated, "缺少有效的厅身份")
	}

	// 余额更新和流水写入必须在同一个事务中
	err = dao.WithTx(l.ctx, l.svcCtx.DB, func(tx *ent.Tx) error {
		// 查询并锁定会员钱包
		memberWallet, err := l.svcCtx.DAO.MemberWallet.GetForUpdateTx(l.ctx, tx, params.memberGlobalID)
		if err != nil {
			if ent.IsNotFound(err) {
				return status.Error(codes.NotFound, "会员钱包不存在")
			}
			return l.internalError("查询会员钱包失败", err)
		}

		// 校验钱包币种
		if memberWallet.CurrencyCode != params.currencyCode {
			return status.Error(codes.FailedPrecondition, "存款币种与钱包币种不一致")
		}

		// 创建绑定当前事务和钱包的记账器
		recorder, err := wallet.NewRecorder(tx, l.svcCtx.DAO, memberWallet)
		if err != nil {
			return l.internalError("初始化钱包记账器失败", err)
		}

		// 执行统一资金记账
		if err := recorder.Record(l.ctx, wallet.RecordParams{
			OperationID:    params.operationID,
			BusinessType:   enum.BusinessTypeDeposit,
			BusinessAction: enum.BusinessActionDepositCredit,
			BusinessNo:     params.depositOrderNo,
			Changes: []wallet.BalanceChange{
				{
					BalanceType:  enum.BalanceTypeAvailable,
					ChangeAmount: params.amount,
				},
			},
		}); err != nil {
			return l.mapRecordError(err)
		}

		return nil
	})
	if err != nil {
		// 保留事务回调中已经转换好的业务错误
		if _, ok := status.FromError(err); ok {
			return nil, err
		}
		return nil, l.internalError("执行存款入账事务失败", err)
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

	operationID, err := uuid.Parse(in.GetOperationId())
	if err != nil || operationID == uuid.Nil {
		return depositCreditParams{}, status.Error(codes.InvalidArgument, "钱包操作标识无效")
	}

	// 校验存款订单信息
	orderNo := in.GetDepositOrderNo()
	if orderNo == "" ||
		strings.TrimSpace(orderNo) != orderNo ||
		utf8.RuneCountInString(orderNo) > 128 {
		return depositCreditParams{}, status.Error(codes.InvalidArgument, "存款订单号无效")
	}

	currencyCode := in.GetCurrencyCode()
	if currencyCode == "" ||
		strings.TrimSpace(currencyCode) != currencyCode ||
		utf8.RuneCountInString(currencyCode) > 16 {
		return depositCreditParams{}, status.Error(codes.InvalidArgument, "币种编码无效")
	}

	// 校验入账金额
	amount := in.GetAmount()
	if amount <= 0 {
		return depositCreditParams{}, status.Error(codes.InvalidArgument, "入账金额必须大于零")
	}

	return depositCreditParams{
		memberGlobalID: memberGlobalID, // 会员全局唯一标识
		operationID:    operationID,    // 钱包操作唯一标识
		depositOrderNo: orderNo,        // 存款订单号
		currencyCode:   currencyCode,   // 币种编码
		amount:         amount,         // 入账金额
	}, nil
}

// internalError 记录内部错误并返回统一RPC错误
func (l *DepositCreditLogic) internalError(action string, err error) error {
	l.Errorf("%s: %v", action, err)
	return status.Error(codes.Internal, "存款本金入账失败")
}

// mapRecordError 将统一记账错误转换为RPC业务错误
func (l *DepositCreditLogic) mapRecordError(err error) error {
	switch {
	case errors.Is(err, wallet.ErrInvalidRecordParams):
		return status.Error(codes.InvalidArgument, "存款记账参数无效")

	case errors.Is(err, wallet.ErrOperationConflict):
		return status.Error(codes.AlreadyExists, "钱包操作标识已被使用")

	case errors.Is(err, wallet.ErrBalanceOverflow):
		return status.Error(codes.OutOfRange, "入账后余额超出允许范围")

	case errors.Is(err, wallet.ErrInsufficientBalance):
		return status.Error(codes.FailedPrecondition, "钱包余额不足")

	case dao.IsDepositOrderUniqueConflict(err):
		return status.Error(codes.AlreadyExists, "存款订单已存在")

	case dao.IsWalletOperationUniqueConflict(err):
		return status.Error(codes.AlreadyExists, "钱包操作标识已被使用")

	default:
		return l.internalError("统一资金记账失败", err)
	}
}
