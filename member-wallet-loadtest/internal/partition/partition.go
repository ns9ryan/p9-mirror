package partition

import (
	"errors"
	"fmt"
	"hash/fnv"

	"github.com/google/uuid"
)

// ForWallet 根据钱包全局标识计算目标 Partition, 路由规则必须与 operator-wallet 保持一致
func ForWallet(walletGlobalID string, partitionCount int) (int, error) {
	if partitionCount <= 0 {
		return 0, errors.New("invalid partition count")
	}

	id, err := uuid.Parse(walletGlobalID)
	if err != nil || id == uuid.Nil {
		return 0, errors.New("invalid wallet global ID")
	}

	// 使用 UUID 原始字节计算 FNV-1a 哈希
	h := fnv.New64a()
	_, _ = h.Write(id[:])

	return int(h.Sum64() % uint64(partitionCount)), nil
}

// Distribution 统计测试钱包在各 Partition 上的分布数量
func Distribution(walletGlobalIDs []string, partitionCount int) ([]int, error) {
	if partitionCount <= 0 {
		return nil, errors.New("invalid partition count")
	}

	counts := make([]int, partitionCount)

	for i, walletGlobalID := range walletGlobalIDs {
		index, err := ForWallet(walletGlobalID, partitionCount)
		if err != nil {
			return nil, fmt.Errorf("wallet index %d: %w", i, err)
		}

		counts[index]++
	}

	return counts, nil
}
