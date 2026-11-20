package dao

import (
	"database/sql"

	"oa.98ent.com/p9/operator-report/rpc/ent"
)

// Manager 数据库访问对象管理器
type Manager struct {
	DB                 *ent.Client
	SqlDB              *sql.DB
	UserRetentionDaily *UserRetentionDailyDao
	MemberGrowthTrend  *MemberGrowthTrendDao
	MemberSummary      *MemberSummaryDao
}

// NewManager 创建 DAO 管理器
func NewManager(db *ent.Client) *Manager {
	return &Manager{
		DB:                 db,
		UserRetentionDaily: NewUserRetentionDailyDao(db),
		MemberGrowthTrend:  NewMemberGrowthTrendDao(db, nil),
		MemberSummary:      NewMemberSummaryDao(db, nil),
	}
}

// NewManagerWithSqlDB 创建带 SQL DB 的 DAO 管理器
func NewManagerWithSqlDB(db *ent.Client, sqlDB *sql.DB) *Manager {
	return &Manager{
		DB:                 db,
		SqlDB:              sqlDB,
		UserRetentionDaily: NewUserRetentionDailyDao(db),
		MemberGrowthTrend:  NewMemberGrowthTrendDao(db, sqlDB),
		MemberSummary:      NewMemberSummaryDao(db, sqlDB),
	}
}
