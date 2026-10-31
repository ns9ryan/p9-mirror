// ============================================================================
// Package dbtx 提供 Ent 本地事务的执行与回滚处理
// ============================================================================

package dbtx

import (
	"context"
	"fmt"

	"oa.98ent.com/p9/operator-crm/ent"
)

// WithTx 执行本地事务, 成功提交, 失败回滚
func WithTx(ctx context.Context, client *ent.Client, fn func(tx *ent.Tx) error) error {
	tx, err := client.Tx(ctx)
	if err != nil {
		return err
	}

	// 发生 panic 时回滚事务, 并交由上层处理
	defer func() {
		if recovered := recover(); recovered != nil {
			_ = tx.Rollback()
			panic(recovered)
		}
	}()

	if err := fn(tx); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("%w: rollback: %v", err, rollbackErr)
		}
		return err
	}

	return tx.Commit()
}
