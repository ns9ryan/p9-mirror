// ============================================================================
// Package paging 提供通用分页处理, 统一默认值、数量限制和偏移量计算.
// ============================================================================

package paging

import "math"

const (
	defaultPageSize int64 = 20  // 默认每页数量
	maxPageSize     int64 = 100 // 最大每页数量
)

// OffsetLimit 计算数据库分页参数, 偏移量溢出时返回 false
func OffsetLimit(page, pageSize int64) (offset, limit int, ok bool) {
	// 规范分页参数
	page = max(page, 1)
	switch {
	case pageSize <= 0:
		pageSize = defaultPageSize
	case pageSize > maxPageSize:
		pageSize = maxPageSize
	}

	// 防止偏移量溢出
	if page-1 > int64(math.MaxInt)/pageSize {
		return 0, 0, false
	}

	return int((page - 1) * pageSize), int(pageSize), true
}
