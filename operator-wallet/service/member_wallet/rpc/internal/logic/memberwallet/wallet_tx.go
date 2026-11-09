// ============================================================================
// 厅身份、钱包事务
// ============================================================================

package memberwalletlogic

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/operator-wallet/ent"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/dao"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/svc"
)

// validateOperatorCode 校验当前请求的可信厅身份
func validateOperatorCode(ctx context.Context) error {
	operatorCode := ctxdata.OperatorCodeFromCtx(ctx)
	if operatorCode == "" ||
		strings.TrimSpace(operatorCode) != operatorCode ||
		utf8.RuneCountInString(operatorCode) > 64 {
		return status.Error(codes.Unauthenticated, "缺少有效的厅身份")
	}

	return nil
}

// withLockedMemberWalletTx 校验厅身份, 锁定会员钱包并执行事务
func withLockedMemberWalletTx(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	memberGlobalID uuid.UUID,
	currencyCode string,
	currencyMismatchMessage string,
	run func(tx *ent.Tx, memberWallet *ent.MemberWallet) error,
) error {
	// 确保钱包操作来自有效的厅身份
	if err := validateOperatorCode(ctx); err != nil {
		return err
	}

	// 钱包余额更新和资金流水写入必须处于同一事务
	return dao.WithTx(ctx, svcCtx.DB, func(tx *ent.Tx) error {
		// 查询并锁定会员钱包
		memberWallet, err := svcCtx.DAO.MemberWallet.GetForUpdateTx(
			ctx, tx, memberGlobalID,
		)
		if err != nil {
			if ent.IsNotFound(err) {
				return status.Error(codes.NotFound, "会员钱包不存在")
			}
			return fmt.Errorf("查询会员钱包失败: %w", err)
		}

		// 校验钱包结算币种
		if memberWallet.CurrencyCode != currencyCode {
			return status.Error(codes.FailedPrecondition, currencyMismatchMessage)
		}

		return run(tx, memberWallet)
	})
}
