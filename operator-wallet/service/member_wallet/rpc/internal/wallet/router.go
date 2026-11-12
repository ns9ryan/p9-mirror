// ============================================================================
// 钱包 Shard 路由
// ============================================================================

package wallet

import (
	"errors"
	"fmt"
	"hash/fnv"

	"github.com/google/uuid"
)

var (
	// ErrInvalidShardCount 表示钱包 Shard 数量无效
	ErrInvalidShardCount = errors.New("invalid wallet shard count")

	// ErrInvalidWalletGlobalID 表示钱包全局标识无效
	ErrInvalidWalletGlobalID = errors.New("invalid wallet global ID")

	errNilRouter = errors.New("wallet router is nil")
)

// router 根据钱包全局标识选择固定 Shard
//
// FNV-1a 算法和 Shard 数量属于钱包分片映射契约
// 接入 Kafka 后, Producer 必须使用相同的路由下标选择 Partition
type router struct {
	shardCount int // 固定 Shard 数量
}

// newRouter 创建钱包路由器
func newRouter(shardCount int) (*router, error) {
	if shardCount <= 0 {
		return nil, fmt.Errorf("create wallet router: %w: %d", ErrInvalidShardCount, shardCount)
	}

	return &router{
		shardCount: shardCount,
	}, nil
}

// route 根据钱包全局标识计算目标 Shard 下标
func (r *router) route(walletGlobalID uuid.UUID) (int, error) {
	if r == nil {
		return 0, errNilRouter
	}

	if r.shardCount <= 0 {
		return 0, ErrInvalidShardCount
	}

	if walletGlobalID == uuid.Nil {
		return 0, ErrInvalidWalletGlobalID
	}

	// 使用 UUID 原始字节计算稳定哈希, 不依赖字符串格式
	h := fnv.New64a()
	_, _ = h.Write(walletGlobalID[:])

	return int(h.Sum64() % uint64(r.shardCount)), nil
}
