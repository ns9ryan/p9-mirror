package manage

import (
	"context"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"

	"oa.98ent.com/p9/common/xerr"
)

// 锁有效期覆盖整个导出请求；不启动续租协程。
const (
	exportLockTTL                     = 60 * time.Second
	memberExportRedisOperationTimeout = 2 * time.Second
	memberExportReleasedLockCount     = 1
	// 只删除仍属于当前令牌的锁，避免误删其他请求新取得的锁。
	memberExportReleaseLockScript = `if redis.call('get',KEYS[1]) == ARGV[1] then return redis.call('del',KEYS[1]) else return 0 end`
)

type exportLock struct {
	redis      redis.UniversalClient
	key, token string
}

// acquire 使用原子 SET NX 获取租户锁，竞争失败与依赖故障分别处理。
func (exportLockState exportLock) acquire(ctx context.Context) error {
	operationCtx, cancel := context.WithTimeout(ctx, memberExportRedisOperationTimeout)
	defer cancel()
	lockAcquired, err := exportLockState.redis.SetNX(operationCtx, exportLockState.key, exportLockState.token, exportLockTTL).Result()
	if err != nil {
		if operationCtx.Err() != nil {
			return exportHTTPError(operationCtx, operationCtx.Err())
		}
		return xerr.Err(http.StatusServiceUnavailable, "export lock unavailable")
	}
	if !lockAcquired {
		return xerr.Err(http.StatusConflict, "another member export is running")
	}
	return nil
}

// check 在上传前及交付结果前核对锁归属，失锁后不交付链接。
func (exportLockState exportLock) check(ctx context.Context) error {
	operationCtx, cancel := context.WithTimeout(ctx, memberExportRedisOperationTimeout)
	defer cancel()
	token, err := exportLockState.redis.Get(operationCtx, exportLockState.key).Result()
	if err != nil && err != redis.Nil {
		if operationCtx.Err() != nil {
			return exportHTTPError(operationCtx, operationCtx.Err())
		}
		return xerr.Err(http.StatusServiceUnavailable, "export lock unavailable")
	}
	if token != exportLockState.token {
		return xerr.Err(http.StatusConflict, "member export lock lost")
	}
	return nil
}

// release 使用独立超时上下文清理锁，确保原请求取消后仍能尝试释放。
func (exportLockState exportLock) release() error {
	ctx, cancel := context.WithTimeout(context.Background(), memberExportRedisOperationTimeout)
	defer cancel()
	releasedLockCount, err := exportLockState.redis.Eval(ctx, memberExportReleaseLockScript, []string{exportLockState.key}, exportLockState.token).Int()
	if err != nil {
		return err
	}
	if releasedLockCount != memberExportReleasedLockCount {
		return xerr.Err(http.StatusConflict, "export lock no longer owned")
	}
	return nil
}
