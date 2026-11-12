package wallet

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// TestNewRouterRejectsInvalidShardCount 验证 Shard 数量必须大于零
func TestNewRouterRejectsInvalidShardCount(t *testing.T) {
	for _, shardCount := range []int{0, -1} {
		if _, err := newRouter(shardCount); !errors.Is(err, ErrInvalidShardCount) {
			t.Errorf("expected ErrInvalidShardCount for %d, got %v", shardCount, err)
		}
	}
}

// TestRouterRejectsNilReceiver 验证未初始化的 Router 无法执行路由
func TestRouterRejectsNilReceiver(t *testing.T) {
	var r *router

	if _, err := r.route(uuid.New()); !errors.Is(err, errNilRouter) {
		t.Fatalf("expected errNilRouter, got %v", err)
	}
}

// TestRouterRejectsEmptyWalletGlobalID 验证不允许使用空钱包标识
func TestRouterRejectsEmptyWalletGlobalID(t *testing.T) {
	r, err := newRouter(8)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := r.route(uuid.Nil); !errors.Is(err, ErrInvalidWalletGlobalID) {
		t.Fatalf("expected ErrInvalidWalletGlobalID, got %v", err)
	}
}

// TestRouterReturnsValidShardIndex 验证路由结果始终处于有效范围
func TestRouterReturnsValidShardIndex(t *testing.T) {
	const shardCount = 8

	r, err := newRouter(shardCount)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 100; i++ {
		shardIndex, err := r.route(uuid.New())
		if err != nil {
			t.Fatal(err)
		}

		if shardIndex < 0 || shardIndex >= shardCount {
			t.Fatalf("shard index out of range: %d", shardIndex)
		}
	}
}

// TestRouterMappingIsStable 验证不同 Router 实例使用相同的路由映射
func TestRouterMappingIsStable(t *testing.T) {
	const shardCount = 8

	r1, err := newRouter(shardCount)
	if err != nil {
		t.Fatal(err)
	}

	r2, err := newRouter(shardCount)
	if err != nil {
		t.Fatal(err)
	}

	walletGlobalID := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")

	first, err := r1.route(walletGlobalID)
	if err != nil {
		t.Fatal(err)
	}

	second, err := r2.route(walletGlobalID)
	if err != nil {
		t.Fatal(err)
	}

	if first != second {
		t.Fatalf("expected stable shard index %d, got %d", first, second)
	}

	// 固定测试向量, 防止无意改变哈希算法或 UUID 字节顺序
	if first != 5 {
		t.Fatalf("expected shard index 5, got %d", first)
	}
}

// TestRouterWithSingleShard 验证只有一个 Shard 时始终路由到下标零
func TestRouterWithSingleShard(t *testing.T) {
	r, err := newRouter(1)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		shardIndex, err := r.route(uuid.New())
		if err != nil {
			t.Fatal(err)
		}

		if shardIndex != 0 {
			t.Fatalf("expected shard index 0, got %d", shardIndex)
		}
	}
}
