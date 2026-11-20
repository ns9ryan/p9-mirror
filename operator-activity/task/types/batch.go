package types

import "time"

// BatchID 表示批次的时间窗口和序列号
// 格式: timeWindow:sequenceNum (如 20261116-101033:1)
type BatchID struct {
	TimeWindow string // 时间窗口 (YYYYMMDDhhmmss)
	Sequence   int    // 序列号 (同一秒内的批次序号)
}

// BatchInfo 批次的基本信息
type BatchInfo struct {
	ID         BatchID            // 批次 ID
	Type       string             // 批次类型 (task/payout/taskretry/payoutretry)
	ServiceName string             // 服务名称
	CreatedAt  time.Time          // 批次创建时间
	MessageCount int                // 消息数量
	OwnerID    string             // 当前所有者（多实例认领时的所有者标识）
}

// BatchLockInfo 分布式锁的信息
type BatchLockInfo struct {
	BatchID    string    // 批次 ID 字符串 (timeWindow:sequence)
	OwnerID    string    // 所有者标识（实例 ID）
	AcquiredAt time.Time // 获取锁的时间
	ExpiresAt  time.Time // 锁过期时间
}

// BatchResult 批次处理的结果
type BatchResult struct {
	BatchID string
	Type string
	Total int           // 总消息数
	Success int         // 成功处理数
	Failed int          // 失败数
	FailedEvents []*Event // 失败的事件
	Duration time.Duration
}
