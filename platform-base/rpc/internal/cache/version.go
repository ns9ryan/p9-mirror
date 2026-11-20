package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

const readVersioned = `
local revision = redis.call('GET', KEYS[1]) or '0'
return {revision, redis.call('GET', ARGV[1] .. revision)}
`

// ReadVersioned 在单机 Redis 上用 Lua 原子读取版本号和对应的缓存数据。
func ReadVersioned(ctx context.Context, client redis.UniversalClient, revisionKey, dataPrefix string) (key, payload string, err error) {
	result, err := client.Eval(ctx, readVersioned, []string{revisionKey}, dataPrefix).Slice()
	if err != nil {
		return "", "", err
	}
	if len(result) != 2 {
		return "", "", fmt.Errorf("unexpected versioned cache response length: %d", len(result))
	}
	revision, ok := result[0].(string)
	if !ok {
		return "", "", fmt.Errorf("unexpected cache revision type: %T", result[0])
	}
	key = dataPrefix + revision
	// 缓存未命中时仍返回版本化键，供回源后写入。
	if result[1] == nil {
		return key, "", nil
	}
	payload, ok = result[1].(string)
	if !ok {
		return "", "", fmt.Errorf("unexpected cache payload type: %T", result[1])
	}
	return key, payload, nil
}
