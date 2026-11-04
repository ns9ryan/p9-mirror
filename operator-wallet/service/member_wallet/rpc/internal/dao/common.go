package dao

import (
	"context"
	"fmt"

	"oa.98ent.com/p9/operator-wallet/ent"
)

// WithTx 在一个事务中执行操作, 成功时提交, 失败时回滚
func WithTx(ctx context.Context, client *ent.Client, run func(tx *ent.Tx) error) error {
	tx, err := client.Tx(ctx)
	if err != nil {
		return err
	}

	// 发生 panic 时先回滚事务, 再继续向上抛出
	defer func() {
		if recovered := recover(); recovered != nil {
			_ = tx.Rollback()
			panic(recovered)
		}
	}()

	// 执行业务操作, 失败时回滚事务
	if err := run(tx); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("%w: rollback: %v", err, rollbackErr)
		}
		return err
	}

	// 业务操作成功后提交事务
	return tx.Commit()
}
