package dao

import (
	"context"
	"time"

	"oa.98ent.com/p9/operator-report/rpc/ent"
	"oa.98ent.com/p9/operator-report/rpc/ent/userretentiondaily"
)

// UserRetentionDailyDao 用户留存数据访问对象
type UserRetentionDailyDao struct {
	db *ent.Client
}

// UserRetentionDailyParam 用户留存查询参数
type UserRetentionDailyParam struct {
	OpCode    string
	AgentId   int64
	StartDate time.Time
	EndDate   time.Time
}

// UserRetentionDailyResult 用户留存查询结果
type UserRetentionDailyResult struct {
	OpCode              string
	AgentId             int64
	RegisterDate        time.Time
	TotalNewUsers       int64
	TodayLoginCount     int64
	TodayLoginRate      float64
	YesterdayLoginCount int64
	Range7dLoginCount   int64
	Range7dLoginRate    float64
	Range15dLoginCount  int64
	Range15dLoginRate   float64
	Range30dLoginCount  int64
	Range30dLoginRate   float64
	Range60dLoginCount  int64
	Range60dLoginRate   float64
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// NewUserRetentionDailyDao 创建用户留存DAO实例
func NewUserRetentionDailyDao(db *ent.Client) *UserRetentionDailyDao {
	return &UserRetentionDailyDao{db: db}
}

// GetByParam 根据参数查询用户留存数据
func (d *UserRetentionDailyDao) GetByParam(ctx context.Context, param UserRetentionDailyParam) ([]*UserRetentionDailyResult, error) {
	query := d.db.UserRetentionDaily.Query()

	// 按分站编码过滤
	if param.OpCode != "" {
		query = query.Where(userretentiondaily.OpCodeEQ(param.OpCode))
	}

	// 按代理ID过滤
	if param.AgentId > 0 {
		query = query.Where(userretentiondaily.AgentIDEQ(param.AgentId))
	}

	// 按注册日期范围过滤
	if !param.StartDate.IsZero() {
		query = query.Where(userretentiondaily.RegisterDateGTE(param.StartDate))
	}
	if !param.EndDate.IsZero() {
		query = query.Where(userretentiondaily.RegisterDateLTE(param.EndDate))
	}

	// 按注册日期排序
	query = query.Order(ent.Asc(userretentiondaily.FieldRegisterDate))

	results, err := query.All(ctx)
	if err != nil {
		return nil, err
	}

	// 转换为结果对象
	var retentionData []*UserRetentionDailyResult
	for _, item := range results {
		retentionData = append(retentionData, &UserRetentionDailyResult{
			OpCode:        item.OpCode,
			AgentId:       item.AgentID,
			RegisterDate:  item.RegisterDate,
			TotalNewUsers: item.TotalNewUsers,
			// TodayLoginCount:     item.TodayLoginCount,
			// TodayLoginRate:      item.TodayLoginRate,
			// YesterdayLoginCount: item.YesterdayLoginCount,
			// Range7dLoginCount:   item.Range7dLoginCount,
			// Range7dLoginRate:    item.Range7dLoginRate,
			// Range15dLoginCount:  item.Range15dLoginCount,
			// Range15dLoginRate:   item.Range15dLoginRate,
			// Range30dLoginCount:  item.Range30dLoginCount,
			// Range30dLoginRate:   item.Range30dLoginRate,
			// Range60dLoginCount:  item.Range60dLoginCount,
			// Range60dLoginRate:   item.Range60dLoginRate,
			// CreatedAt:           item.CreatedAt,
			// UpdatedAt:           item.UpdatedAt,
		})
	}

	return retentionData, nil
}

// GetByOpCodeAndAgentId 按分站编码和代理ID查询用户留存数据
func (d *UserRetentionDailyDao) GetByOpCodeAndAgentId(ctx context.Context, opCode string, agentId int64, startDate, endDate time.Time) ([]*UserRetentionDailyResult, error) {
	return d.GetByParam(ctx, UserRetentionDailyParam{
		OpCode:    opCode,
		AgentId:   agentId,
		StartDate: startDate,
		EndDate:   endDate,
	})
}

// GetByOpCode 按分站编码查询用户留存数据
func (d *UserRetentionDailyDao) GetByOpCode(ctx context.Context, opCode string, startDate, endDate time.Time) ([]*UserRetentionDailyResult, error) {
	return d.GetByParam(ctx, UserRetentionDailyParam{
		OpCode:    opCode,
		StartDate: startDate,
		EndDate:   endDate,
	})
}

// GetByDateRange 按日期范围查询用户留存数据
func (d *UserRetentionDailyDao) GetByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*UserRetentionDailyResult, error) {
	return d.GetByParam(ctx, UserRetentionDailyParam{
		StartDate: startDate,
		EndDate:   endDate,
	})
}
