package memberwalletlogic

import (
	"context"
	"errors"
	"math"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/operator-wallet/ent"
	"oa.98ent.com/p9/operator-wallet/ent/memberwallet"
	"oa.98ent.com/p9/operator-wallet/ent/memberwalletledger"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/enum"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/svc"
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

	// 获取当前厅编码
	operatorCode := ctxdata.OperatorCodeFromCtx(l.ctx)
	if operatorCode == "" ||
		strings.TrimSpace(operatorCode) != operatorCode ||
		utf8.RuneCountInString(operatorCode) > 64 {
		return nil, status.Error(codes.Unauthenticated, "缺少有效的厅身份")
	}

	// 开启事务并锁定钱包
	tx, err := l.svcCtx.DB.Tx(l.ctx)
	if err != nil {
		return nil, l.internalError("开启存款入账事务失败", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	wallet, err := tx.MemberWallet.Query().
		Where(memberwallet.MemberGlobalIDEQ(params.memberGlobalID)).
		ForUpdate().
		Only(l.ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, "会员钱包不存在")
		}
		return nil, l.internalError("查询会员钱包失败", err)
	}

	// 校验钱包币种
	if wallet.CurrencyCode != params.currencyCode {
		return nil, status.Error(codes.FailedPrecondition, "存款币种与钱包币种不一致")
	}

	// 根据存款订单检查幂等
	existing, err := tx.MemberWalletLedger.Query().
		Where(
			memberwalletledger.BusinessTypeEQ(int64(enum.BusinessTypeDeposit)),
			memberwalletledger.BusinessActionEQ(int64(enum.BusinessActionDepositCredit)),
			memberwalletledger.BusinessNoEQ(params.depositOrderNo),
		).
		Only(l.ctx)
	if err == nil {
		if existing.WalletID == wallet.ID &&
			existing.BalanceType == int64(enum.BalanceTypeAvailable) &&
			existing.OperationID == params.operationID &&
			existing.ChangeAmount == params.amount {
			return &depositpb.DepositCreditResponse{}, nil
		}

		return nil, status.Error(codes.AlreadyExists, "存款订单已存在, 请求信息不一致")
	}
	if !ent.IsNotFound(err) {
		return nil, l.internalError("查询存款入账流水失败", err)
	}

	// 检查钱包操作标识是否已使用
	existing, err = tx.MemberWalletLedger.Query().
		Where(
			memberwalletledger.WalletIDEQ(wallet.ID),
			memberwalletledger.OperationIDEQ(params.operationID),
			memberwalletledger.BalanceTypeEQ(int64(enum.BalanceTypeAvailable)),
		).
		Only(l.ctx)
	if err == nil {
		return nil, status.Error(codes.AlreadyExists, "钱包操作标识已被使用")
	}
	if !ent.IsNotFound(err) {
		return nil, l.internalError("查询钱包操作标识失败", err)
	}

	// 检查余额是否溢出
	if wallet.Balance > math.MaxInt64-params.amount {
		return nil, status.Error(codes.OutOfRange, "入账后余额超出允许范围")
	}

	// 计算余额变动
	balanceBefore := wallet.Balance
	balanceAfter := balanceBefore + params.amount

	// 更新会员可用余额
	if _, err := tx.MemberWallet.UpdateOneID(wallet.ID).
		SetBalance(balanceAfter).
		Save(l.ctx); err != nil {
		return nil, l.internalError("更新会员钱包余额失败", err)
	}

	// 写入会员钱包资金流水
	_, err = tx.MemberWalletLedger.Create().
		SetWalletID(wallet.ID).
		SetBalanceType(int64(enum.BalanceTypeAvailable)).
		SetChangeAmount(params.amount).
		SetBalanceBefore(balanceBefore).
		SetBalanceAfter(balanceAfter).
		SetOperationID(params.operationID).
		SetBusinessType(int64(enum.BusinessTypeDeposit)).
		SetBusinessAction(int64(enum.BusinessActionDepositCredit)).
		SetBusinessNo(params.depositOrderNo).
		Save(l.ctx)
	if err != nil {
		if isDepositCreditUniqueConflict(err) {
			return nil, status.Error(codes.AlreadyExists, "存款订单或钱包操作标识已被使用")
		}
		return nil, l.internalError("创建存款资金流水失败", err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return nil, l.internalError("提交存款入账事务失败", err)
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

// isDepositCreditUniqueConflict 判断是否触发存款入账唯一约束
func isDepositCreditUniqueConflict(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		return false
	}

	switch pgErr.ConstraintName {
	case "uk_member_wallet_ledger_deposit_credit",
		"uk_member_wallet_ledger_operation":
		return true
	default:
		return false
	}
}

// internalError 记录数据库错误并返回统一RPC错误
func (l *DepositCreditLogic) internalError(action string, err error) error {
	l.Errorf("%s: %v", action, err)
	return status.Error(codes.Internal, "存款本金入账失败")
}
