package queue

import (
	"fmt"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	dto "github.com/prometheus/client_model/go"
	"github.com/zeromicro/go-zero/core/logx"
)

// 性能指标定义

// BatchProcessingDuration 批次处理耗时分布（直方图）
var BatchProcessingDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "batch_processing_duration_seconds",
		Help:    "批次处理耗时，单位秒",
		Buckets: prometheus.DefBuckets, // 默认：0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10
	},
	[]string{"batch_type", "result"}, // 标签：批次类型（task/payout/task_retry/payout_retry），结果（success/failed）
)

// EventProcessingDuration 单个事件处理耗时分布（直方图）
var EventProcessingDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "event_processing_duration_seconds",
		Help:    "单个事件处理耗时，单位秒",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"event_type", "result"}, // 标签：事件类型（task/payout），结果（success/failed）
)

// BatchProcessingCount 批次处理总数（计数器）
var BatchProcessingCount = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "batch_processing_total",
		Help: "处理的批次总数",
	},
	[]string{"batch_type", "result"}, // 标签：批次类型，结果
)

// EventProcessingCount 事件处理总数（计数器）
var EventProcessingCount = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "event_processing_total",
		Help: "处理的事件总数",
	},
	[]string{"event_type", "result"}, // 标签：事件类型，结果
)

// ActiveBatches 当前处理中的批次数（仪表）
var ActiveBatches = promauto.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "active_batches",
		Help: "当前处理中的批次数",
	},
	[]string{"batch_type"},
)

// BatchSize 批次大小（直方图）
var BatchSize = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "batch_size",
		Help:    "批次中的事件数量",
		Buckets: []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000, 2500},
	},
	[]string{"batch_type"},
)

// RedisOperationDuration Redis 操作延迟（直方图）
var RedisOperationDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "redis_operation_duration_seconds",
		Help:    "Redis 操作延迟，单位秒",
		Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
	},
	[]string{"operation", "result"}, // 标签：操作类型（read/write/delete/lock），结果
)

// DBOperationDuration 数据库操作延迟（直方图）
var DBOperationDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "db_operation_duration_seconds",
		Help:    "数据库操作延迟，单位秒",
		Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
	},
	[]string{"operation", "result"}, // 标签：操作类型（create/update），结果
)

// PayoutAttempts 派发尝试次数分布（直方图）
var PayoutAttempts = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "payout_attempts",
		Help:    "派发尝试次数",
		Buckets: []float64{1, 2, 3, 4, 5, 10},
	},
	[]string{"result"}, // 标签：结果（success/failed）
)

// StoredSuccessfullyCount 实际入库成功的数据（计数器）
var StoredSuccessfullyCount = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "stored_successfully_total",
		Help: "实际入库成功的数据总数",
	},
	[]string{"event_type"}, // 标签：事件类型（task/payout）
)

// DuplicatePeriodKeyCount 周期键重复的数据（计数器）
var DuplicatePeriodKeyCount = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "duplicate_period_key_total",
		Help: "周期键重复的数据总数",
	},
	[]string{"event_type"}, // 标签：事件类型（task/payout）
)

// RecordBatchProcessing 记录批次处理指标
func RecordBatchProcessing(batchType string, duration time.Duration, result string, eventCount int) {
	BatchProcessingDuration.WithLabelValues(batchType, result).Observe(duration.Seconds())
	BatchProcessingCount.WithLabelValues(batchType, result).Inc()
	BatchSize.WithLabelValues(batchType).Observe(float64(eventCount))
}

// RecordEventProcessing 记录事件处理指标
func RecordEventProcessing(eventType string, duration time.Duration, result string) {
	EventProcessingDuration.WithLabelValues(eventType, result).Observe(duration.Seconds())
	EventProcessingCount.WithLabelValues(eventType, result).Inc()
}

// RecordRedisOperation 记录 Redis 操作指标
func RecordRedisOperation(operation string, duration time.Duration, result string) {
	RedisOperationDuration.WithLabelValues(operation, result).Observe(duration.Seconds())
}

// RecordDBOperation 记录数据库操作指标
func RecordDBOperation(operation string, duration time.Duration, result string) {
	DBOperationDuration.WithLabelValues(operation, result).Observe(duration.Seconds())
}

// RecordPayoutAttempt 记录派发尝试指标
func RecordPayoutAttempt(result string, attempts int) {
	PayoutAttempts.WithLabelValues(result).Observe(float64(attempts))
}

// RecordStoredSuccessfully 记录实际入库成功的数据
func RecordStoredSuccessfully(eventType string) {
	StoredSuccessfullyCount.WithLabelValues(eventType).Inc()
}

// RecordDuplicatePeriodKey 记录周期键重复的数据
func RecordDuplicatePeriodKey(eventType string) {
	DuplicatePeriodKeyCount.WithLabelValues(eventType).Inc()
}

// metricsNameMapping 指标名称到中文的映射
var metricsNameMapping = map[string]string{
	"batch_processing_duration_seconds": "批次处理耗时(秒)",
	"batch_processing_total":            "批次处理总数",
	"batch_size":                        "批次大小",
	"active_batches":                    "活跃批次数",
	"event_processing_duration_seconds": "事件处理耗时(秒)",
	"event_processing_total":            "事件处理总数",
	"redis_operation_duration_seconds":  "Redis操作耗时(秒)",
	"db_operation_duration_seconds":     "数据库操作耗时(秒)",
	"payout_attempts":                   "派发尝试次数",
	"stored_successfully_total":         "实际入库总数",
	"duplicate_period_key_total":        "周期键重复总数",
}

// labelsNameMapping 标签名称到中文的映射
var labelsNameMapping = map[string]string{
	"batch_type": "批次类型",
	"event_type": "事件类型",
	"result":     "结果",
	"operation":  "操作",
}

// labelsValueMapping 标签值到中文的映射
var labelsValueMapping = map[string]string{
	"task":        "任务",
	"payout":      "派发",
	"success":     "成功",
	"failed":      "失败",
	"read":        "读取",
	"write":       "写入",
	"delete":      "删除",
	"lock":        "锁定",
	"create":      "创建",
	"update":      "更新",
	"retry":       "重试",
	"payoutretry": "派发重试",
	"taskretry":   "任务重试",
}

// translateMetricName 将指标名称翻译为中文
func translateMetricName(metricName string) string {
	if cn, ok := metricsNameMapping[metricName]; ok {
		return cn
	}
	return metricName
}

// translateLabels 将标签翻译为中文
func translateLabels(labels string) string {
	// 解析标签字符串 "batch_type=task, result=success"
	parts := strings.Split(labels, ", ")
	var translatedParts []string

	for _, part := range parts {
		kv := strings.Split(part, "=")
		if len(kv) != 2 {
			translatedParts = append(translatedParts, part)
			continue
		}

		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		// 翻译键
		if cnKey, ok := labelsNameMapping[key]; ok {
			key = cnKey
		}

		// 翻译值（支持冒号分隔的复合值，如 task:payout）
		if strings.Contains(value, ":") {
			parts := strings.Split(value, ":")
			for i, part := range parts {
				if cnValue, ok := labelsValueMapping[part]; ok {
					parts[i] = cnValue
				}
			}
			value = strings.Join(parts, ":")
		} else if cnValue, ok := labelsValueMapping[value]; ok {
			value = cnValue
		}

		translatedParts = append(translatedParts, fmt.Sprintf("%s=%s", key, value))
	}

	return strings.Join(translatedParts, ", ")
}

// displayWidth 计算字符串的显示宽度
// 中文字符算 2 个宽度，英文字符算 1 个宽度
func displayWidth(s string) int {
	width := 0
	for _, r := range s {
		if r >= 0x4E00 && r <= 0x9FFF || // CJK 统一表意文字
			r >= 0x3040 && r <= 0x309F || // 日文平假名
			r >= 0x30A0 && r <= 0x30FF { // 日文片假名
			width += 2
		} else {
			width += 1
		}
	}
	return width
}

// truncateByWidth 按显示宽度截断字符串
func truncateByWidth(s string, maxWidth int) string {
	if displayWidth(s) <= maxWidth {
		return s
	}

	width := 0
	var result strings.Builder
	for _, r := range s {
		charWidth := 1
		if r >= 0x4E00 && r <= 0x9FFF ||
			r >= 0x3040 && r <= 0x309F ||
			r >= 0x30A0 && r <= 0x30FF {
			charWidth = 2
		}

		if width+charWidth > maxWidth-3 { // 留出 3 个字符空间给 "..."
			break
		}

		result.WriteRune(r)
		width += charWidth
	}

	return result.String() + "..."
}

// padRight 右对齐填充字符串到指定显示宽度
func padRight(s string, width int) string {
	currentWidth := displayWidth(s)
	if currentWidth >= width {
		return s
	}

	// 计算需要添加多少个空格
	paddingCount := width - currentWidth
	return s + strings.Repeat(" ", paddingCount)
}

// PrintMetricsPeriodicly 定时打印性能指标
// interval: 打印间隔时间
func PrintMetricsPeriodicly(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			printMetricsSnapshot()
		}
	}()
}

// printMetricsSnapshot 打印当前指标快照
func printMetricsSnapshot() {
	// 收集所有 Prometheus 指标
	gatherer := prometheus.DefaultGatherer
	mfs, err := gatherer.Gather()
	if err != nil {
		logx.Errorf("❌ 收集指标失败: %v", err)
		return
	}

	// 构建输出
	var buf strings.Builder
	buf.WriteString("\n")
	buf.WriteString("╔════════════════════════════════════════════════════════════════╗\n")
	buf.WriteString("║                    📊 任务性能指标快照                            ║\n")
	buf.WriteString("╠════════════════════════════════════════════════════════════════╣\n")

	// 过滤并打印指标
	metricsOutput := make(map[string]string)

	for _, mf := range mfs {
		name := mf.GetName()

		// 只打印我们关心的指标
		if !strings.HasPrefix(name, "batch_") &&
			!strings.HasPrefix(name, "event_") &&
			!strings.HasPrefix(name, "stored_") &&
			!strings.HasPrefix(name, "duplicate_") &&
			!strings.HasPrefix(name, "active_") &&
			!strings.HasPrefix(name, "redis_") &&
			!strings.HasPrefix(name, "db_") &&
			!strings.HasPrefix(name, "payout_") {
			continue
		}

		for _, m := range mf.Metric {
			labels := make([]string, 0)
			for _, label := range m.Label {
				labels = append(labels, fmt.Sprintf("%s=%s", label.GetName(), label.GetValue()))
			}
			labelStr := strings.Join(labels, ", ")

			// 提取指标值
			var value string
			if m.Gauge != nil {
				value = fmt.Sprintf("%.2f", m.Gauge.GetValue())
			} else if m.Counter != nil {
				value = fmt.Sprintf("%.0f", m.Counter.GetValue())
			} else if m.Histogram != nil {
				// 对直方图显示 count 和 sum
				count := m.Histogram.GetSampleCount()
				sum := m.Histogram.GetSampleSum()
				if count > 0 {
					avg := sum / float64(count)
					value = fmt.Sprintf("count=%d, sum=%.4f, avg=%.4f", count, sum, avg)
				} else {
					value = "count=0"
				}
			}

			key := fmt.Sprintf("%s{%s}", name, labelStr)
			metricsOutput[key] = value
		}
	}

	// 分类打印指标
	printMetricsGroup(&buf, metricsOutput, "batch_", "│ 📦 批次处理指标")
	printMetricsGroup(&buf, metricsOutput, "event_", "│ 🔧 事件处理指标")
	printMetricsGroup(&buf, metricsOutput, "stored_", "│ 💾 存储指标")
	printMetricsGroup(&buf, metricsOutput, "duplicate_", "│ 🔄 重复数据指标")
	printMetricsGroup(&buf, metricsOutput, "active_", "│ ⚡ 活跃指标")
	printMetricsGroup(&buf, metricsOutput, "redis_", "│ 🔑 Redis 指标")
	printMetricsGroup(&buf, metricsOutput, "db_", "│ 🗄️  数据库指标")
	printMetricsGroup(&buf, metricsOutput, "payout_", "│ 💰 派发指标")

	buf.WriteString("╚════════════════════════════════════════════════════════════════╝\n")
	buf.WriteString(fmt.Sprintf("│ 📅 时间: %s\n", time.Now().Format("2006-01-02 15:04:05")))

	// 生成指标汇总和建议
	printMetricsSummary(&buf, mfs)

	output := buf.String()
	fmt.Print(output)
}

// printMetricsGroup 按类别打印指标组
func printMetricsGroup(buf *strings.Builder, metrics map[string]string, prefix, title string) {
	hasMetrics := false
	for k := range metrics {
		if strings.Contains(k, prefix) {
			if !hasMetrics {
				buf.WriteString("├─────────────────────────────────────────────────────────────┤\n")
				buf.WriteString(fmt.Sprintf("%s\n", title))
				hasMetrics = true
			}
		}
	}

	if hasMetrics {
		for k, v := range metrics {
			if strings.Contains(k, prefix) {
				// 提取指标名和标签
				parts := strings.Split(k, "{")
				metricName := parts[0]
				labelsPart := ""
				if len(parts) > 1 {
					labelsPart = "{" + parts[1]
				}

				// 翻译指标名和标签
				cnMetricName := translateMetricName(metricName)
				if labelsPart != "" {
					// 移除末尾的 }
					labelsContent := strings.TrimSuffix(labelsPart[1:], "}")
					cnLabels := translateLabels(labelsContent)
					labelsPart = "{" + cnLabels + "}"
				}

				// 组合翻译后的指标名
				metric := cnMetricName + labelsPart

				// 按显示宽度截断（最多 55 个显示宽度，只在超长时才截断）
				if displayWidth(metric) > 55 {
					metric = truncateByWidth(metric, 55)
				}

				// 右对齐到 55 个显示宽度
				metricPadded := padRight(metric, 55)
				buf.WriteString(fmt.Sprintf("│   %s : %s\n", metricPadded, v))
			}
		}
	}
}

// MetricsSummary 指标汇总数据
type MetricsSummary struct {
	TaskBatchSuccessRate    float64
	PayoutBatchSuccessRate  float64
	AvgTaskProcessingTime   float64
	AvgPayoutProcessingTime float64
	AvgPayoutRetries        float64
	ActiveTaskBatches       float64
	ActivePayoutBatches     float64
	TaskThroughput          float64
	PayoutThroughput        float64
	TaskStoredCount         float64 // 实际入库的任务数
	PayoutStoredCount       float64 // 实际入库的派发数
	TaskBatchTotal          float64 // 任务批次总数（用于成功率计算）
	TaskBatchSuccess        float64 // 任务批次成功数
	PayoutBatchTotal        float64 // 派发批次总数
	PayoutBatchSuccess      float64 // 派发批次成功数
}

// extractMetricsSummary 从 Prometheus 指标中提取汇总数据
func extractMetricsSummary(mfs []*dto.MetricFamily) MetricsSummary {
	summary := MetricsSummary{}

	for _, mf := range mfs {
		name := mf.GetName()

		switch name {
		case "batch_processing_total":
			// 批次成功率和吞吐量计算
			var taskSuccess, taskTotal, payoutSuccess, payoutTotal float64
			for _, m := range mf.Metric {
				// 获取标签值
				batchType := ""
				result := ""
				for _, label := range m.Label {
					if label.GetName() == "batch_type" {
						batchType = label.GetValue()
					} else if label.GetName() == "result" {
						result = label.GetValue()
					}
				}

				value := m.Counter.GetValue()
				// 精确匹配"task:task"类型（原始任务处理批次）
				if batchType == "task:task" && result == "success" {
					taskSuccess += value
				}
				if batchType == "task:task" {
					taskTotal += value
				}
				// 精确匹配"task:payout"类型（派发处理批次）
				if batchType == "task:payout" && result == "success" {
					payoutSuccess += value
				}
				if batchType == "task:payout" {
					payoutTotal += value
				}

				// 累计吞吐量（成功数）
				if result == "success" {
					if batchType == "task:task" {
						summary.TaskThroughput += value
					} else if batchType == "task:payout" {
						summary.PayoutThroughput += value
					}
				}
			}
			// 保存批次总数和成功数用于后续成功率计算
			summary.TaskBatchTotal = taskTotal
			summary.TaskBatchSuccess = taskSuccess
			summary.PayoutBatchTotal = payoutTotal
			summary.PayoutBatchSuccess = payoutSuccess

		case "event_processing_duration_seconds":
			// 事件处理平均耗时
			for _, m := range mf.Metric {
				eventType := ""
				for _, label := range m.Label {
					if label.GetName() == "event_type" {
						eventType = label.GetValue()
					}
				}

				if m.Histogram != nil && m.Histogram.GetSampleCount() > 0 {
					avg := m.Histogram.GetSampleSum() / float64(m.Histogram.GetSampleCount())
					if eventType == "task" {
						summary.AvgTaskProcessingTime = avg
					} else if eventType == "payout" {
						summary.AvgPayoutProcessingTime = avg
					}
				}
			}

		case "payout_attempts":
			// 派发平均重试次数
			for _, m := range mf.Metric {
				if m.Histogram != nil && m.Histogram.GetSampleCount() > 0 {
					summary.AvgPayoutRetries = m.Histogram.GetSampleSum() / float64(m.Histogram.GetSampleCount())
				}
			}

		case "active_batches":
			// 活跃批次数
			for _, m := range mf.Metric {
				batchType := ""
				for _, label := range m.Label {
					if label.GetName() == "batch_type" {
						batchType = label.GetValue()
					}
				}
				if m.Gauge != nil {
					if strings.Contains(batchType, "task") {
						summary.ActiveTaskBatches = m.Gauge.GetValue()
					} else if strings.Contains(batchType, "payout") {
						summary.ActivePayoutBatches = m.Gauge.GetValue()
					}
				}
			}

		case "stored_successfully_total":
			// 实际入库数（真实成功指标）
			for _, m := range mf.Metric {
				eventType := ""
				for _, label := range m.Label {
					if label.GetName() == "event_type" {
						eventType = label.GetValue()
					}
				}
				if m.Counter != nil {
					if eventType == "task" {
						summary.TaskStoredCount = m.Counter.GetValue()
					} else if eventType == "payout" {
						summary.PayoutStoredCount = m.Counter.GetValue()
					}
				}
			}
		}
	}

	// 成功率基于批次数（不是事件数）
	// 这样能准确反映各处理阶段的成功率，不会混淆
	if summary.TaskBatchTotal > 0 {
		summary.TaskBatchSuccessRate = (summary.TaskBatchSuccess / summary.TaskBatchTotal) * 100
	} else {
		summary.TaskBatchSuccessRate = 100 // 无数据时显示 100%
	}

	if summary.PayoutBatchTotal > 0 {
		summary.PayoutBatchSuccessRate = (summary.PayoutBatchSuccess / summary.PayoutBatchTotal) * 100
	} else {
		summary.PayoutBatchSuccessRate = 100 // 无数据时显示 100%
	}

	return summary
}

// printMetricsSummary 打印指标汇总和建议
func printMetricsSummary(buf *strings.Builder, mfs []*dto.MetricFamily) {
	summary := extractMetricsSummary(mfs)

	buf.WriteString("\n╔════════════════════════════════════════════════════════════════╗\n")
	buf.WriteString("║                    📋 任务指标汇总与优化建议                            ║\n")
	buf.WriteString("╠════════════════════════════════════════════════════════════════╣\n")

	// 打印关键指标
	buf.WriteString("├─────────────────────────────────────────────────────────────┤\n")
	buf.WriteString("│ 📊 关键指标汇总\n")
	buf.WriteString(fmt.Sprintf("│   任务批次成功率               : %.2f%%\n", summary.TaskBatchSuccessRate))
	buf.WriteString(fmt.Sprintf("│   派发批次成功率               : %.2f%%\n", summary.PayoutBatchSuccessRate))
	buf.WriteString(fmt.Sprintf("│   任务处理平均耗时(秒)         : %.4f\n", summary.AvgTaskProcessingTime))
	buf.WriteString(fmt.Sprintf("│   派发处理平均耗时(秒)         : %.4f\n", summary.AvgPayoutProcessingTime))
	buf.WriteString(fmt.Sprintf("│   派发平均重试次数             : %.2f\n", summary.AvgPayoutRetries))
	buf.WriteString(fmt.Sprintf("│   活跃任务批次数               : %.0f\n", summary.ActiveTaskBatches))
	buf.WriteString(fmt.Sprintf("│   活跃派发批次数               : %.0f\n", summary.ActivePayoutBatches))

	// 生成建议
	buf.WriteString("├─────────────────────────────────────────────────────────────┤\n")
	buf.WriteString("│ 💡 优化建议\n")

	suggestions := generateSuggestions(summary)
	for _, suggestion := range suggestions {
		// 换行处理长文本
		lines := strings.Split(suggestion, "\n")
		for i, line := range lines {
			if i == 0 {
				buf.WriteString(fmt.Sprintf("│   %s\n", line))
			} else {
				buf.WriteString(fmt.Sprintf("│       %s\n", line))
			}
		}
	}

	buf.WriteString("╚════════════════════════════════════════════════════════════════╝\n")
}

// generateSuggestions 根据指标生成优化建议
// 只在有实际处理数据时才提示成功率建议
func generateSuggestions(summary MetricsSummary) []string {
	var suggestions []string

	// 成功率建议（只在有实际任务处理时提示）
	if summary.TaskThroughput > 0 || (summary.TaskBatchSuccessRate < 100) {
		if summary.TaskBatchSuccessRate < 95 {
			suggestions = append(suggestions, fmt.Sprintf("⚠️  任务批次成功率低于95%% (当前: %.2f%%), 建议检查任务处理异常", summary.TaskBatchSuccessRate))
		} else if summary.TaskBatchSuccessRate < 99 {
			suggestions = append(suggestions, fmt.Sprintf("⚡ 任务批次成功率良好 (%.2f%%), 可继续优化", summary.TaskBatchSuccessRate))
		} else if summary.TaskBatchSuccessRate == 100 && summary.TaskThroughput > 0 {
			suggestions = append(suggestions, fmt.Sprintf("✅ 任务批次成功率优秀 (%.2f%%)", summary.TaskBatchSuccessRate))
		}
	}

	// 派发成功率建议（只在有实际派发处理时提示）
	if summary.PayoutThroughput > 0 || (summary.PayoutBatchSuccessRate < 100) {
		if summary.PayoutBatchSuccessRate < 95 {
			suggestions = append(suggestions, fmt.Sprintf("⚠️  派发批次成功率低于95%% (当前: %.2f%%), 需要重点关注", summary.PayoutBatchSuccessRate))
		} else if summary.PayoutBatchSuccessRate < 99 {
			suggestions = append(suggestions, fmt.Sprintf("⚡ 派发批次成功率良好 (%.2f%%)", summary.PayoutBatchSuccessRate))
		} else if summary.PayoutBatchSuccessRate == 100 && summary.PayoutThroughput > 0 {
			suggestions = append(suggestions, fmt.Sprintf("✅ 派发批次成功率优秀 (%.2f%%)", summary.PayoutBatchSuccessRate))
		}
	}

	// 处理延迟建议
	if summary.AvgTaskProcessingTime > 0.05 {
		suggestions = append(suggestions, fmt.Sprintf("⚠️  任务处理平均耗时过长 (%.4f秒), 可尝试:\n       • 增加并发度 (maxConcurrency)\n       • 优化数据库查询", summary.AvgTaskProcessingTime))
	}

	if summary.AvgPayoutProcessingTime > 0.1 {
		suggestions = append(suggestions, fmt.Sprintf("⚠️  派发处理平均耗时过长 (%.4f秒), 建议检查外部服务响应时间", summary.AvgPayoutProcessingTime))
	}

	// 重试次数建议
	if summary.AvgPayoutRetries > 2 {
		suggestions = append(suggestions, fmt.Sprintf("⚠️  派发平均重试次数较高 (%.2f次), 建议:\n       • 检查外部派发服务状态\n       • 增加重试超时时间", summary.AvgPayoutRetries))
	}

	// 活跃批次建议
	if summary.ActiveTaskBatches > 5 {
		suggestions = append(suggestions, fmt.Sprintf("ℹ️  当前活跃任务批次较多 (%.0f), 系统处理能力良好", summary.ActiveTaskBatches))
	}

	// 如果没有任何建议，添加默认的良好状态消息
	if len(suggestions) == 0 {
		suggestions = append(suggestions, "✅ 系统状态良好，所有指标均在正常范围内")
	}

	return suggestions
}
