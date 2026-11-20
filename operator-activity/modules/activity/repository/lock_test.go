package repository_test

import (
	"context"
	"testing"
	"time"

	"oa.98ent.com/p9/operator-activity/modules/activity/perr"
	"oa.98ent.com/p9/operator-activity/modules/activity/repository"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRedisUserPromoLocker_LockUnlock(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	// 短等待：持锁期间第二次抢锁应超时失败
	locker := repository.NewRedisUserPromoLocker(rdb).WithWait(80*time.Millisecond, 20*time.Millisecond)

	unlock, err := locker.Lock(context.Background(), 1, 1002)
	if err != nil {
		t.Fatal(err)
	}

	_, err = locker.Lock(context.Background(), 1, 1002)
	if err == nil || !perr.IsRetryable(err) {
		t.Fatalf("expected retryable busy after wait, got %v", err)
	}

	unlock()
	unlock2, err := locker.Lock(context.Background(), 1, 1002)
	if err != nil {
		t.Fatal(err)
	}
	unlock2()
}

func TestRedisUserPromoLocker_SpinWaitsUntilReleased(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	locker := repository.NewRedisUserPromoLocker(rdb).WithWait(2*time.Second, 20*time.Millisecond)

	unlock, err := locker.Lock(context.Background(), 1, 1002)
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan error, 1)
	go func() {
		_, err := locker.Lock(context.Background(), 1, 1002)
		done <- err
	}()

	time.Sleep(50 * time.Millisecond)
	unlock()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("expected spin acquire after release, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for spin lock")
	}
}

func TestRedisUserPromoLocker_DifferentUser(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	locker := repository.NewRedisUserPromoLockerWithTTL(rdb, time.Second)

	u1, err := locker.Lock(context.Background(), 1, 1001)
	if err != nil {
		t.Fatal(err)
	}
	defer u1()

	u2, err := locker.Lock(context.Background(), 1, 1002)
	if err != nil {
		t.Fatal(err)
	}
	u2()
}
