package metric

import (
	"context"
	"time"
)

// DepositWindowCounter 窗口期充值累计计数器（Redis）
// key 不存在时由 load 从 DB 回源并写入；已存在则对当前订单做幂等累加。
type DepositWindowCounter interface {
	// IncrOrLoad 返回窗口内累计次数与金额
	IncrOrLoad(
		ctx context.Context,
		key string,
		orderID int64,
		amount int64,
		ttl time.Duration,
		load func(context.Context) (count int, sum int64, err error),
	) (count int, sum int64, err error)
}
