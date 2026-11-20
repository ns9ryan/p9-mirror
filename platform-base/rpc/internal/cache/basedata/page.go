package basedata

// Page 先过滤有序列表，再计算总数和分页边界，保留并发调用方共享的缓存切片。
func Page[T any](rows []*T, page, size int64, matches func(*T) bool) ([]*T, int64) {
	if matches != nil {
		filtered := make([]*T, 0, len(rows))
		for _, row := range rows {
			if matches(row) {
				filtered = append(filtered, row)
			}
		}
		rows = filtered
	}
	total := int64(len(rows))
	// 先检查页码边界，再计算偏移量，避免大页码导致乘法溢出。
	if total == 0 || page-1 > (total-1)/size {
		return rows[:0], total
	}
	start := (page - 1) * size
	return rows[start:min(start+size, total)], total
}
