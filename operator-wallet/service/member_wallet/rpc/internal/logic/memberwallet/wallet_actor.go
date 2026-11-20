package memberwalletlogic

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/operator-wallet/ent"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/svc"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/wallet"
)

// walletActorParams 钱包 Actor 加载参数
type walletActorParams struct {
	memberGlobalID uuid.UUID // 会员全局唯一标识
	currencyCode   string    // 请求币种编码
}

// validateOperatorCode 校验当前请求的可信厅身份
func validateOperatorCode(ctx context.Context) error {
	operatorCode := ctxdata.OperatorCodeFromCtx(ctx)
	if operatorCode == "" || strings.TrimSpace(operatorCode) != operatorCode || utf8.RuneCountInString(operatorCode) > 64 {
		return status.Error(codes.Unauthenticated, "缺少有效的厅身份")
	}

	return nil
}

// ensureWalletActor 确保会员钱包已加载并注册到 Engine
func ensureWalletActor(ctx context.Context, svcCtx *svc.ServiceContext, params walletActorParams) (uuid.UUID, error) {
	if ctx == nil || svcCtx == nil {
		return uuid.Nil, errors.New("ensure wallet actor: invalid context")
	}

	if err := validateOperatorCode(ctx); err != nil {
		return uuid.Nil, err
	}

	if svcCtx.WalletEngine == nil {
		return uuid.Nil, errors.New("wallet engine is not initialized")
	}

	operatorCode := ctxdata.OperatorCodeFromCtx(ctx)

	// 已缓存时直接复用已有 Actor
	if walletGlobalID, cachedCurrency, ok := svcCtx.CachedWallet(operatorCode, params.memberGlobalID); ok {
		if cachedCurrency != params.currencyCode {
			return uuid.Nil, status.Error(codes.FailedPrecondition, "请求币种与钱包币种不一致")
		}

		exists, err := svcCtx.WalletEngine.HasWallet(walletGlobalID)
		if err != nil {
			if errors.Is(err, wallet.ErrEngineClosed) {
				return uuid.Nil, status.Error(codes.Unavailable, "钱包服务正在关闭")
			}
			return uuid.Nil, fmt.Errorf("check wallet actor: %w", err)
		}

		if !exists {
			return uuid.Nil, errors.New("wallet cache references an unregistered actor")
		}

		return walletGlobalID, nil
	}

	// 首次请求, 从当前厅数据库加载钱包
	memberWallet, err := svcCtx.DAO.MemberWallet.GetByMemberGlobalID(ctx, params.memberGlobalID)
	if err != nil {
		if ent.IsNotFound(err) {
			return uuid.Nil, status.Error(codes.NotFound, "会员钱包不存在")
		}
		return uuid.Nil, fmt.Errorf("load member wallet: %w", err)
	}

	if memberWallet == nil {
		return uuid.Nil, errors.New("loaded member wallet is nil")
	}

	// 再次确认钱包归属, 不依赖请求方提交的厅标识
	if memberWallet.OperatorCode != operatorCode || memberWallet.MemberGlobalID != params.memberGlobalID {
		return uuid.Nil, status.Error(codes.NotFound, "会员钱包不存在")
	}

	// 校验请求币种
	if memberWallet.CurrencyCode != params.currencyCode {
		return uuid.Nil, status.Error(codes.FailedPrecondition, "请求币种与钱包币种不一致")
	}

	// 将数据库快照转换为 Actor 初始状态
	state := wallet.State{
		// 钱包标识
		WalletID:       memberWallet.ID,             // 数据库钱包ID
		WalletGlobalID: memberWallet.GlobalID,       // 钱包全局唯一标识
		MemberGlobalID: memberWallet.MemberGlobalID, // 会员全局唯一标识

		// 归属信息
		OperatorCode: memberWallet.OperatorCode, // 所属厅业务编码
		CurrencyCode: memberWallet.CurrencyCode, // 钱包结算币种编码

		// 余额
		Balance:          memberWallet.Balance,          // 可用余额
		FrozenBalance:    memberWallet.FrozenBalance,    // 冻结余额
		UnsettledBalance: memberWallet.UnsettledBalance, // 游戏未结算余额

		Version: memberWallet.Version, // 钱包状态版本, 随资金操作递增
	}

	// 校验钱包运行时状态
	if err := state.Validate(); err != nil {
		return uuid.Nil, fmt.Errorf("validate loaded wallet state: %w", err)
	}

	// Engine 内部保证同一个钱包只创建一个 Actor, 已存在时复用原状态, 不会覆盖正在运行的余额和版本
	if err := svcCtx.WalletEngine.RegisterWallet(state); err != nil {
		if errors.Is(err, wallet.ErrEngineClosed) {
			return uuid.Nil, status.Error(codes.Unavailable, "钱包服务正在关闭")
		}
		return uuid.Nil, fmt.Errorf("register wallet actor: %w", err)
	}

	// Actor 注册成功后才缓存映射
	if err := svcCtx.CacheWallet(operatorCode, params.memberGlobalID, state.WalletGlobalID, state.CurrencyCode); err != nil {
		return uuid.Nil, fmt.Errorf("cache registered wallet: %w", err)
	}

	return state.WalletGlobalID, nil
}

// ensureWalletActorForRPC 确保钱包 Actor 可用并统一处理加载错误
func ensureWalletActorForRPC(ctx context.Context, svcCtx *svc.ServiceContext, params walletActorParams, operationID uuid.UUID, logger logx.Logger) (uuid.UUID, error) {
	walletGlobalID, err := ensureWalletActor(ctx, svcCtx, params)
	if err == nil {
		return walletGlobalID, nil
	}

	// 保留已经转换好的业务错误
	if _, ok := status.FromError(err); ok {
		return uuid.Nil, err
	}

	// 此时尚未提交资金操作, 不存在操作已被接收的情况
	switch {
	case errors.Is(err, context.Canceled):
		return uuid.Nil, status.Error(codes.Canceled, "加载会员钱包请求已取消")

	case errors.Is(err, context.DeadlineExceeded):
		return uuid.Nil, status.Error(codes.DeadlineExceeded, "加载会员钱包超时")
	}

	// 其他钱包错误统一映射
	mappedErr := mapWalletError(err)

	// 保留原始错误和操作标识, 方便排查
	if status.Code(mappedErr) == codes.Internal {
		logger.Errorf("ensure wallet actor failed, operation_id=%s: %v", operationID, err)
	}

	return uuid.Nil, mappedErr
}
