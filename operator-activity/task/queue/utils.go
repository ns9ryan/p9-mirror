package queue

import (
	"fmt"
	"strings"
)

// parseInt64 将字符串解析为 int64
func parseInt64(value string) int64 {
	var result int64
	_, _ = fmt.Sscan(value, &result)
	return result
}

// IsMissingQueueError 判断 Redis RENAME 相关的 Key 不存在错误（保留给调用方扩展使用）。
func IsMissingQueueError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "no such key")
}
