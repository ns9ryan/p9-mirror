package utils

import "time"

// GetTodayDateRange 获取今日时间范围，返回开始时间和结束时间的毫秒时间戳
func GetTodayDateRange() []int64 {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24*time.Hour - time.Millisecond)
	return []int64{startOfDay.UnixMilli(), endOfDay.UnixMilli()}
}
