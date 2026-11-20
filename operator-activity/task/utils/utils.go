package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"oa.98ent.com/p9/operator-activity/ent"
	"oa.98ent.com/p9/operator-activity/task/constants"
	"oa.98ent.com/p9/operator-activity/task/types"
)

func CalculatePeriodKey(config *ent.TaskConfig, occurredAt time.Time) string {
	configTimezone := config.Timezone
	loc, err := time.LoadLocation(configTimezone)
	if err == nil {
		occurredAt = occurredAt.In(loc)
	}
	switch config.Category {
	case constants.TaskCategoryOnce:
		return constants.PeriodKeyOnce
	case constants.TaskCategoryDaily:
		return occurredAt.Format(constants.DateFormatDaily)
	case constants.TaskCategoryWeekly:
		year, week := occurredAt.ISOWeek()
		return fmt.Sprintf(constants.DateFormatWeekly, year, week)
	case constants.TaskCategoryMonthly:
		return occurredAt.Format(constants.DateFormatMonthly)
	default:
		return constants.PeriodKeyUnknown
	}
}

func RandomStr() string {
	const digits = "0123456789"

	randomPart := make([]byte, 10)
	for i := range randomPart {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return ""
		}
		randomPart[i] = digits[n.Int64()]
	}

	return fmt.Sprintf("%s%s", time.Now().Format("20060102150405"), string(randomPart))
}

func FieldKey(event *types.Event, periodKey string) string {
	return fmt.Sprintf("%s:%s:%d:%s", event.OperatorCode, event.EventType, event.MemberID, periodKey)
}

// ========== 批次处理相关工具函数 ==========

// GenerateBatchID 生成批次 ID
// 格式: timeWindow:sequence (如 20261116-101033:1)
// GenerateBatchID 生成批次 ID
// 简化设计：只使用时间窗口，不使用序号
// 返回: 批次 ID（只包含时间窗口）
func GenerateBatchID() *types.BatchID {
	return &types.BatchID{
		TimeWindow: time.Now().Format(constants.BatchTimeWindowFormat),
		Sequence:   0,  // 不使用序号
	}
}

// BatchIDToString 将 BatchID 转为字符串（只使用时间窗口）
func BatchIDToString(bid *types.BatchID) string {
	return bid.TimeWindow
}

// ParseBatchIDFromString 从字符串解析 BatchID（只解析时间窗口）
func ParseBatchIDFromString(s string) (*types.BatchID, error) {
	// 简化设计：直接使用字符串作为时间窗口
	if len(s) != len(constants.BatchTimeWindowFormat) {
		return nil, fmt.Errorf("invalid batch id format: %s", s)
	}

	return &types.BatchID{
		TimeWindow: s,
		Sequence:   0,  // 不使用序号
	}, nil
}

// FormatBatchDataKey 格式化批次数据存储的 Redis Key
// 简化设计：格式为 prefix:serviceName:timeWindow（不需要序号）
func FormatBatchDataKey(prefix, serviceName string, batchID *types.BatchID) string {
	return fmt.Sprintf("%s:%s:%s", prefix, serviceName, batchID.TimeWindow)
}

// FormatBatchDataKeyFromString 从时间窗口字符串格式化数据 Key
func FormatBatchDataKeyFromString(prefix, serviceName, timeWindow string) string {
	return fmt.Sprintf("%s:%s:%s", prefix, serviceName, timeWindow)
}

// FormatBatchLockKey 格式化分布式锁的 Redis Key
// 简化设计：格式为 task:lock:prefix:serviceName:timeWindow
func FormatBatchLockKey(prefix, serviceName string, batchID *types.BatchID) string {
	return fmt.Sprintf("task:lock:%s:%s:%s", prefix, serviceName, batchID.TimeWindow)
}

// FormatBatchLockKeyFromString 从 batchID 字符串格式化锁 Key
func FormatBatchLockKeyFromString(prefix, serviceName, batchIDStr string) string {
	return fmt.Sprintf("task:lock:%s:%s:%s", prefix, serviceName, batchIDStr)
}

// GenerateInstanceID 生成实例唯一标识符（用于分布式锁 Owner）
// 格式: hostname-uuid 或 hostname-pid
func GenerateInstanceID() string {
	return uuid.New().String()
}

// IsBatchKey 检查 Redis Key 是否是批次数据 Key
// 简化设计：批次 Key 格式为 prefix:serviceName:timeWindow（3部分）
func IsBatchKey(key string, expectedPrefix string) bool {
	if !strings.HasPrefix(key, expectedPrefix) {
		return false
	}
	parts := strings.Split(key, ":")
	// prefix:serviceName:timeWindow 至少3部分
	// 例：task:task:operator-activity-task:20261116101033
	return len(parts) >= 4
}

// ExtractBatchIDFromKey 从 Redis Key 提取 batchID 字符串
// 简化设计：格式为 prefix:serviceName:timeWindow -> 返回 timeWindow
func ExtractBatchIDFromKey(key string) string {
	parts := strings.Split(key, ":")
	if len(parts) < 4 {
		return ""
	}
	// 取最后一部分 timeWindow
	return parts[len(parts)-1]
}
