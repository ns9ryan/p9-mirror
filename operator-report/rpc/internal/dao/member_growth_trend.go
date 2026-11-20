package dao

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/operator-report/rpc/ent"
	"oa.98ent.com/p9/operator-report/rpc/ent/dwdmemberloginevent"
	"oa.98ent.com/p9/operator-report/rpc/ent/dwdmemberregisterevent"
	"oa.98ent.com/p9/operator-report/rpc/internal/utils"
)

// MemberGrowthTrendDao 会员增长趋势数据访问对象
type MemberGrowthTrendDao struct {
	db    *ent.Client
	sqlDB *sql.DB
}

// MemberGrowthTrendParam 会员增长趋势查询参数
type MemberGrowthTrendParam struct {
	OpCode      string    // 分站编码
	AgentID     string    // 代理ID
	Source      string    // 来源
	StartDate   time.Time // 开始日期
	EndDate     time.Time // 结束日期
	Granularity string    // 粒度：day、hour
	Type        string    // 统计类型：register、login、retention1d、retention2d、retention3d、retention7d、retention15d、retention30d、retention60d
}

// GrowthTrendItem 增长趋势项
type GrowthTrendItem struct {
	Date  string // 日期或小时 (YYYY-MM-DD 或 YYYY-MM-DD HH:00)
	Count int64  // 统计数量 (根据type字段表示：新增数、登录数、或留存率*100)
}

// NewMemberGrowthTrendDao 创建会员增长趋势DAO实例
func NewMemberGrowthTrendDao(db *ent.Client, sqlDB *sql.DB) *MemberGrowthTrendDao {
	return &MemberGrowthTrendDao{db: db, sqlDB: sqlDB}
}

// GetGrowthTrend 获取会员增长趋势数据
func (d *MemberGrowthTrendDao) GetGrowthTrend(ctx context.Context, param MemberGrowthTrendParam) ([]*GrowthTrendItem, error) {
	var result []*GrowthTrendItem
	var err error

	switch param.Type {
	case "register":
		result, err = d.getRegisterTrend(ctx, param)
		if err != nil {
			return nil, err
		}
		// 填充缺失的时间数据
		result = d.fillMissingTimeData(result, param.StartDate, param.EndDate, param.Granularity, "register")
	case "login":
		result, err = d.getLoginTrend(ctx, param)
		if err != nil {
			return nil, err
		}
		// 填充缺失的时间数据
		result = d.fillMissingTimeData(result, param.StartDate, param.EndDate, param.Granularity, "login")
	case "retention1d", "retention2d", "retention3d", "retention7d", "retention15d", "retention30d", "retention60d":
		result, err = d.getRetentionTrend(ctx, param)
		if err != nil {
			return nil, err
		}
		// 填充缺失的时间数据
		result = d.fillMissingTimeData(result, param.StartDate, param.EndDate, param.Granularity, param.Type)
	case "retention1dRate", "retention2dRate", "retention3dRate", "retention7dRate", "retention15dRate", "retention30dRate", "retention60dRate":
		result, err = d.getRetentionRateTrend(ctx, param)
		if err != nil {
			return nil, err
		}
		// 填充缺失的时间数据
		result = d.fillMissingTimeData(result, param.StartDate, param.EndDate, param.Granularity, param.Type)
	default:
		return nil, fmt.Errorf("unsupported trend type: %s", param.Type)
	}

	return result, nil
}

// getRegisterTrend 获取注册数趋势
func (d *MemberGrowthTrendDao) getRegisterTrend(ctx context.Context, param MemberGrowthTrendParam) ([]*GrowthTrendItem, error) {
	query := d.db.DwdMemberRegisterEvent.Query()

	// 条件过滤
	logx.Infof("[getRegisterTrend] 查询参数 - OpCode: %s, AgentID: %s, Source: %s, StartDate: %v, EndDate: %v",
		param.OpCode, param.AgentID, param.Source, param.StartDate, param.EndDate)

	if param.OpCode != "" {
		logx.Infof("[getRegisterTrend] 添加 OpCode 条件: %s", param.OpCode)
		query = query.Where(dwdmemberregisterevent.OpCodeEQ(param.OpCode))
	}
	if param.AgentID != "" {
		logx.Infof("[getRegisterTrend] 添加 AgentID 条件: %s", param.AgentID)
		query = query.Where(dwdmemberregisterevent.AgentIDEQ(param.AgentID))
	}
	if param.Source != "" {
		logx.Infof("[getRegisterTrend] 添加 Source 条件: %s", param.Source)
		query = query.Where(dwdmemberregisterevent.RegisterSourceEQ(param.Source))
	}
	if !param.StartDate.IsZero() {
		logx.Infof("[getRegisterTrend] 添加 StartDate 条件: %v", param.StartDate)
		query = query.Where(dwdmemberregisterevent.RegisterTimeGTE(param.StartDate))
	} else {
		logx.Errorf("[getRegisterTrend] StartDate 为零值，跳过开始日期条件")
	}
	if !param.EndDate.IsZero() {
		logx.Infof("[getRegisterTrend] 添加 EndDate 条件: %v", param.EndDate)
		query = query.Where(dwdmemberregisterevent.RegisterTimeLTE(param.EndDate))
	} else {
		logx.Errorf("[getRegisterTrend] EndDate 为零值，跳过结束日期条件")
	}

	// 按注册时间排序
	query = query.Order(ent.Asc(dwdmemberregisterevent.FieldRegisterTime))

	events, err := query.
		Select(dwdmemberregisterevent.FieldRegisterTime, dwdmemberregisterevent.FieldUserID).
		All(ctx)
	if err != nil {
		logx.Errorf("[getRegisterTrend] 查询失败: %v", err)
		return nil, err
	}
	logx.Infof("[getRegisterTrend] 查询成功，返回 %d 条注册事件", len(events))
	if len(events) > 0 {
		logx.Infof("[getRegisterTrend] 事件样本: %s", utils.JSON(events[0]))
	}

	return d.aggregateRegisterEvents(events, param.Granularity), nil
}

// aggregateRegisterEvents 聚合注册事件数据
// 返回新增数（count 字段）
func (d *MemberGrowthTrendDao) aggregateRegisterEvents(events []*ent.DwdMemberRegisterEvent, granularity string) []*GrowthTrendItem {
	var result []*GrowthTrendItem

	if granularity == "hour" {
		// 按小时聚合
		result = d.aggregateRegisterEventsByHour(events)
	} else {
		// 按日期聚合
		result = d.aggregateRegisterEventsByDate(events)
	}

	return result
}

// aggregateRegisterEventsByDate 按日期聚合注册事件
func (d *MemberGrowthTrendDao) aggregateRegisterEventsByDate(events []*ent.DwdMemberRegisterEvent) []*GrowthTrendItem {
	// 按日期聚合，计算新增数
	dateMap := make(map[string]map[string]bool) // date -> {userId -> bool}
	dateOrder := []string{}
	seenDate := make(map[string]bool)

	for _, event := range events {
		dateStr := event.RegisterTime.Format("2006-01-02")
		if dateMap[dateStr] == nil {
			dateMap[dateStr] = make(map[string]bool)
			if !seenDate[dateStr] {
				dateOrder = append(dateOrder, dateStr)
				seenDate[dateStr] = true
			}
		}
		dateMap[dateStr][event.UserID] = true
	}

	var result []*GrowthTrendItem

	for _, date := range dateOrder {
		newCount := int64(len(dateMap[date]))

		result = append(result, &GrowthTrendItem{
			Date:  date,
			Count: newCount,
		})
	}

	return result
}

// aggregateRegisterEventsByHour 按小时聚合注册事件
func (d *MemberGrowthTrendDao) aggregateRegisterEventsByHour(events []*ent.DwdMemberRegisterEvent) []*GrowthTrendItem {
	// 按小时聚合，计算新增数
	hourMap := make(map[string]map[string]bool) // hour -> {userId -> bool}
	hourOrder := []string{}
	seenHour := make(map[string]bool)

	for _, event := range events {
		hourStr := event.RegisterTime.Format("2006-01-02 15:00")
		if hourMap[hourStr] == nil {
			hourMap[hourStr] = make(map[string]bool)
			if !seenHour[hourStr] {
				hourOrder = append(hourOrder, hourStr)
				seenHour[hourStr] = true
			}
		}
		hourMap[hourStr][event.UserID] = true
	}

	var result []*GrowthTrendItem

	for _, hour := range hourOrder {
		newCount := int64(len(hourMap[hour]))

		result = append(result, &GrowthTrendItem{
			Date:  hour,
			Count: newCount,
		})
	}

	return result
}

// getLoginTrend 获取登录数趋势
func (d *MemberGrowthTrendDao) getLoginTrend(ctx context.Context, param MemberGrowthTrendParam) ([]*GrowthTrendItem, error) {
	query := d.db.DwdMemberLoginEvent.Query()

	// 条件过滤
	logx.Infof("[getLoginTrend] 查询参数 - OpCode: %s, AgentID: %s, Source: %s, StartDate: %v, EndDate: %v",
		param.OpCode, param.AgentID, param.Source, param.StartDate, param.EndDate)

	if param.OpCode != "" {
		logx.Infof("[getLoginTrend] 添加 OpCode 条件: %s", param.OpCode)
		query = query.Where(dwdmemberloginevent.OpCodeEQ(param.OpCode))
	}
	if param.AgentID != "" {
		logx.Infof("[getLoginTrend] 添加 AgentID 条件: %s", param.AgentID)
		query = query.Where(dwdmemberloginevent.AgentIDEQ(param.AgentID))
	}
	if param.Source != "" {
		logx.Infof("[getLoginTrend] 添加 Source 条件: %s", param.Source)
		query = query.Where(dwdmemberloginevent.LoginSourceEQ(param.Source))
	}
	if !param.StartDate.IsZero() {
		logx.Infof("[getLoginTrend] 添加 StartDate 条件: %v", param.StartDate)
		query = query.Where(dwdmemberloginevent.LoginTimeGTE(param.StartDate))
	} else {
		logx.Errorf("[getLoginTrend] StartDate 为零值，跳过开始日期条件")
	}
	if !param.EndDate.IsZero() {
		logx.Infof("[getLoginTrend] 添加 EndDate 条件: %v", param.EndDate)
		query = query.Where(dwdmemberloginevent.LoginTimeLTE(param.EndDate))
	} else {
		logx.Errorf("[getLoginTrend] EndDate 为零值，跳过结束日期条件")
	}

	// 按登录时间排序
	query = query.Order(ent.Asc(dwdmemberloginevent.FieldLoginTime))

	events, err := query.
		Select(dwdmemberloginevent.FieldLoginTime, dwdmemberloginevent.FieldUserID).
		All(ctx)
	if err != nil {
		logx.Errorf("[getLoginTrend] 查询失败: %v", err)
		return nil, err
	}
	logx.Infof("[getLoginTrend] 查询成功，返回 %d 条登录事件", len(events))
	if len(events) > 0 {
		logx.Infof("[getLoginTrend] 事件样本: %s", utils.JSON(events[0]))
	}

	return d.aggregateLoginEvents(events, param.Granularity), nil
}

// aggregateLoginEvents 聚合登录事件数据
func (d *MemberGrowthTrendDao) aggregateLoginEvents(events []*ent.DwdMemberLoginEvent, granularity string) []*GrowthTrendItem {
	var result []*GrowthTrendItem

	if granularity == "hour" {
		// 按小时聚合
		result = d.aggregateLoginEventsByHour(events)
	} else {
		// 按日期聚合
		result = d.aggregateLoginEventsByDate(events)
	}

	return result
}

// aggregateLoginEventsByHour 按小时聚合登录事件
func (d *MemberGrowthTrendDao) aggregateLoginEventsByHour(events []*ent.DwdMemberLoginEvent) []*GrowthTrendItem {
	// 按小时聚合，计算登录数（distinct user_id）
	hourMap := make(map[string]map[string]bool) // hour -> {userId -> bool}
	hourOrder := []string{}
	seenHour := make(map[string]bool)

	for _, event := range events {
		hourStr := event.LoginTime.Format("2006-01-02 15:00")
		if hourMap[hourStr] == nil {
			hourMap[hourStr] = make(map[string]bool)
			if !seenHour[hourStr] {
				hourOrder = append(hourOrder, hourStr)
				seenHour[hourStr] = true
			}
		}
		hourMap[hourStr][event.UserID] = true
	}

	var result []*GrowthTrendItem

	for _, hour := range hourOrder {
		loginCount := int64(len(hourMap[hour]))

		result = append(result, &GrowthTrendItem{
			Date:  hour,
			Count: loginCount,
		})
	}

	return result
}

// aggregateLoginEventsByDate 按日期聚合登录事件
func (d *MemberGrowthTrendDao) aggregateLoginEventsByDate(events []*ent.DwdMemberLoginEvent) []*GrowthTrendItem {
	// 按日期聚合，计算登录数（distinct user_id）
	dateMap := make(map[string]map[string]bool) // date -> {userId -> bool}
	dateOrder := []string{}
	seenDate := make(map[string]bool)

	for _, event := range events {
		dateStr := event.LoginTime.Format("2006-01-02")
		if dateMap[dateStr] == nil {
			dateMap[dateStr] = make(map[string]bool)
			if !seenDate[dateStr] {
				dateOrder = append(dateOrder, dateStr)
				seenDate[dateStr] = true
			}
		}
		dateMap[dateStr][event.UserID] = true
	}

	var result []*GrowthTrendItem

	for _, date := range dateOrder {
		loginCount := int64(len(dateMap[date]))

		result = append(result, &GrowthTrendItem{
			Date:  date,
			Count: loginCount,
		})
	}

	return result
}

// getRetentionTrend 获取留存人数趋势（从视图查询，不支持小时维度）
func (d *MemberGrowthTrendDao) getRetentionTrend(ctx context.Context, param MemberGrowthTrendParam) ([]*GrowthTrendItem, error) {
	// 留存数据从 mv_user_retention_daily 视图查询
	// 根据统计类型查询不同的留存人数字段
	// 注意：留存数据不支持小时维度，只支持 day

	// 根据类型确定查询的字段
	var retentionField string
	switch param.Type {
	case "retention1d":
		retentionField = "retention_1d_users"
	case "retention2d":
		retentionField = "retention_2d_users"
	case "retention3d":
		retentionField = "retention_3d_users"
	case "retention7d":
		retentionField = "retention_7d_users"
	case "retention15d":
		retentionField = "retention_15d_users"
	case "retention30d":
		retentionField = "retention_30d_users"
	case "retention60d":
		retentionField = "retention_60d_users"
	default:
		retentionField = "retention_1d_users" // 默认为1日留存
	}

	query := fmt.Sprintf(`
	SELECT 
		DATE_FORMAT(register_date, '%%Y-%%m-%%d') as date,
		%s
	FROM mv_user_retention_daily
	WHERE 1=1
	`, retentionField)

	args := []interface{}{}

	if param.OpCode != "" {
		query += " AND op_code = ?"
		args = append(args, param.OpCode)
	}
	if param.AgentID != "" {
		query += " AND agent_id = ?"
		args = append(args, param.AgentID)
	}
	if !param.StartDate.IsZero() {
		query += " AND register_date >= ?"
		args = append(args, param.StartDate)
	}
	if !param.EndDate.IsZero() {
		query += " AND register_date <= ?"
		args = append(args, param.EndDate)
	}

	query += " ORDER BY register_date ASC"

	logx.Infof("[getRetentionTrend] Type: %s, SQL Query: %s", param.Type, query)
	logx.Infof("[getRetentionTrend] Query Args: %v", args)
	logx.Infof("[getRetentionTrend] Param - OpCode: %s, AgentID: %s, StartDate: %v, EndDate: %v",
		param.OpCode, param.AgentID, param.StartDate, param.EndDate)

	var results []*GrowthTrendItem

	// 如果没有 SQL DB，返回空结果
	if d.sqlDB == nil {
		logx.Errorf("[getRetentionTrend] SQL DB not available for retention trend query")
		return []*GrowthTrendItem{}, fmt.Errorf("SQL DB not available for retention trend query")
	}

	// 使用 raw SQL 查询
	rows, err := d.sqlDB.QueryContext(ctx, query, args...)
	if err != nil {
		logx.Errorf("[getRetentionTrend] Query failed: %v", err)
		return []*GrowthTrendItem{}, err
	}
	defer rows.Close()

	rowCount := 0
	for rows.Next() {
		rowCount++
		var (
			date           string
			retentionUsers int64
		)
		if err := rows.Scan(&date, &retentionUsers); err != nil {
			logx.Errorf("[getRetentionTrend] Scan row %d failed: %v", rowCount, err)
			continue
		}

		logx.Infof("[getRetentionTrend] Scanned row %d - date: %s, retentionUsers: %d",
			rowCount, date, retentionUsers)

		// 直接返回留存人数
		results = append(results, &GrowthTrendItem{
			Date:  date,
			Count: retentionUsers,
		})
	}

	if rows.Err() != nil {
		logx.Errorf("[getRetentionTrend] Row iteration error: %v", rows.Err())
		return []*GrowthTrendItem{}, rows.Err()
	}

	logx.Infof("[getRetentionTrend] Total rows scanned: %d, results count: %d", rowCount, len(results))
	logx.Infof("[getRetentionTrend] Results: %s", utils.JSON(results))

	return results, nil
}

// getRetentionRateTrend 获取留存率趋势（从视图查询）
// 留存率 = 留存人数 / 新增用户总数 * 100，返回整数百分比
func (d *MemberGrowthTrendDao) getRetentionRateTrend(ctx context.Context, param MemberGrowthTrendParam) ([]*GrowthTrendItem, error) {
	// 根据类型确定查询的字段
	var retentionField string
	switch param.Type {
	case "retention1dRate":
		retentionField = "retention_1d_users"
	case "retention2dRate":
		retentionField = "retention_2d_users"
	case "retention3dRate":
		retentionField = "retention_3d_users"
	case "retention7dRate":
		retentionField = "retention_7d_users"
	case "retention15dRate":
		retentionField = "retention_15d_users"
	case "retention30dRate":
		retentionField = "retention_30d_users"
	case "retention60dRate":
		retentionField = "retention_60d_users"
	default:
		retentionField = "retention_1d_users" // 默认为1日留存
	}

	query := fmt.Sprintf(`
	SELECT 
		DATE_FORMAT(register_date, '%%Y-%%m-%%d') as date,
		total_new_users,
		%s
	FROM mv_user_retention_daily
	WHERE 1=1
	`, retentionField)

	args := []interface{}{}

	if param.OpCode != "" {
		query += " AND op_code = ?"
		args = append(args, param.OpCode)
	}
	if param.AgentID != "" {
		query += " AND agent_id = ?"
		args = append(args, param.AgentID)
	}
	if !param.StartDate.IsZero() {
		query += " AND register_date >= ?"
		args = append(args, param.StartDate)
	}
	if !param.EndDate.IsZero() {
		query += " AND register_date <= ?"
		args = append(args, param.EndDate)
	}

	query += " ORDER BY register_date ASC"

	logx.Infof("[getRetentionRateTrend] Type: %s, SQL Query: %s", param.Type, query)
	logx.Infof("[getRetentionRateTrend] Query Args: %v", args)
	logx.Infof("[getRetentionRateTrend] Param - OpCode: %s, AgentID: %s, StartDate: %v, EndDate: %v",
		param.OpCode, param.AgentID, param.StartDate, param.EndDate)

	var results []*GrowthTrendItem

	// 如果没有 SQL DB，返回空结果
	if d.sqlDB == nil {
		logx.Errorf("[getRetentionRateTrend] SQL DB not available for retention rate trend query")
		return []*GrowthTrendItem{}, fmt.Errorf("SQL DB not available for retention rate trend query")
	}

	// 使用 raw SQL 查询
	rows, err := d.sqlDB.QueryContext(ctx, query, args...)
	if err != nil {
		logx.Errorf("[getRetentionRateTrend] Query failed: %v", err)
		return []*GrowthTrendItem{}, err
	}
	defer rows.Close()

	rowCount := 0
	for rows.Next() {
		rowCount++
		var (
			date           string
			totalNewUsers  int64
			retentionUsers int64
		)
		if err := rows.Scan(&date, &totalNewUsers, &retentionUsers); err != nil {
			logx.Errorf("[getRetentionRateTrend] Scan row %d failed: %v", rowCount, err)
			continue
		}

		logx.Infof("[getRetentionRateTrend] Scanned row %d - date: %s, totalNewUsers: %d, retentionUsers: %d",
			rowCount, date, totalNewUsers, retentionUsers)

		// 计算留存率：retention_rate = retention_users / total_new_users * 100
		// 返回整数百分比，比如 1.23% 返回为 1
		var retentionRate int64 = 0
		if totalNewUsers > 0 {
			retentionRate = (retentionUsers * 100) / totalNewUsers
		}

		results = append(results, &GrowthTrendItem{
			Date:  date,
			Count: retentionRate,
		})
	}

	if rows.Err() != nil {
		logx.Errorf("[getRetentionRateTrend] Row iteration error: %v", rows.Err())
		return []*GrowthTrendItem{}, rows.Err()
	}

	logx.Infof("[getRetentionRateTrend] Total rows scanned: %d, results count: %d", rowCount, len(results))
	logx.Infof("[getRetentionRateTrend] Results: %s", utils.JSON(results))

	return results, nil
}

// fillMissingTimeData 填充缺失的时间数据（支持日和小时维度）
func (d *MemberGrowthTrendDao) fillMissingTimeData(items []*GrowthTrendItem, startDate, endDate time.Time, granularity, trendType string) []*GrowthTrendItem {
	if items == nil {
		items = []*GrowthTrendItem{}
	}

	if granularity == "hour" {
		return d.fillMissingHours(items, startDate, endDate, trendType)
	}

	// 默认按日期填充
	return d.fillMissingDates(items, startDate, endDate, trendType)
}

// fillMissingHours 填充缺失的小时数据
func (d *MemberGrowthTrendDao) fillMissingHours(items []*GrowthTrendItem, startDate, endDate time.Time, trendType string) []*GrowthTrendItem {
	// 建立小时到数据的映射
	hourMap := make(map[string]*GrowthTrendItem)
	for _, item := range items {
		hourMap[item.Date] = item
	}

	// 生成小时范围内的所有小时（格式：2006-01-02 15:00）
	var result []*GrowthTrendItem

	// 从 startDate 00:00 开始
	currentHour := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	// 到 endDate 23:59 结束
	endTime := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, endDate.Location())

	for !currentHour.After(endTime) {
		hourStr := currentHour.Format("2006-01-02 15:00")

		if item, exists := hourMap[hourStr]; exists {
			result = append(result, item)
		} else {
			// 填充缺失的小时：count为0
			result = append(result, &GrowthTrendItem{
				Date:  hourStr,
				Count: 0,
			})
		}

		currentHour = currentHour.Add(time.Hour)
	}

	return result
}

// fillMissingDates 填充缺失的日期数据
func (d *MemberGrowthTrendDao) fillMissingDates(items []*GrowthTrendItem, startDate, endDate time.Time, trendType string) []*GrowthTrendItem {
	// 建立日期到数据的映射
	dateMap := make(map[string]*GrowthTrendItem)
	for _, item := range items {
		dateMap[item.Date] = item
	}

	// 生成日期范围内的所有日期
	var result []*GrowthTrendItem
	currentDate := startDate

	for !currentDate.After(endDate) {
		dateStr := currentDate.Format("2006-01-02")

		if item, exists := dateMap[dateStr]; exists {
			result = append(result, item)
		} else {
			// 填充缺失的日期：count为0
			result = append(result, &GrowthTrendItem{
				Date:  dateStr,
				Count: 0,
			})
		}

		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return result
}
