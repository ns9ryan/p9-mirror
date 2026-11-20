package repository_test

import (
	"context"
	"testing"
	"time"

	"oa.98ent.com/p9/operator-activity/modules/activity/repository"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRedisDepositCounter_IncrOrLoad(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	counter := repository.NewRedisDepositCounter(rdb)
	ctx := context.Background()
	key := "promo:dep:win:1:1001:1:2"
	ttl := time.Hour

	loadCalls := 0
	load := func(context.Context) (int, int64, error) {
		loadCalls++
		return 1, 10000, nil // DB 已含当前单
	}

	count, sum, err := counter.IncrOrLoad(ctx, key, 9001, 10000, ttl, load)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 || sum != 10000 || loadCalls != 1 {
		t.Fatalf("init: count=%d sum=%d loadCalls=%d", count, sum, loadCalls)
	}

	// 同单重试：不重复累加、不回源
	count, sum, err = counter.IncrOrLoad(ctx, key, 9001, 10000, ttl, load)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 || sum != 10000 || loadCalls != 1 {
		t.Fatalf("dup: count=%d sum=%d loadCalls=%d", count, sum, loadCalls)
	}

	// 新单：Redis 增量，不再 load
	count, sum, err = counter.IncrOrLoad(ctx, key, 9002, 5000, ttl, load)
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 || sum != 15000 || loadCalls != 1 {
		t.Fatalf("incr: count=%d sum=%d loadCalls=%d", count, sum, loadCalls)
	}
}
