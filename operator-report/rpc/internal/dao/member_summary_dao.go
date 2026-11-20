package dao

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/operator-report/rpc/ent"
	"oa.98ent.com/p9/operator-report/rpc/ent/dwdmemberloginevent"
	"oa.98ent.com/p9/operator-report/rpc/ent/dwdmemberregisterevent"
)

// MemberSummaryDao 会员概览数据访问对象
type MemberSummaryDao struct {
	db    *ent.Client
	sqlDB *sql.DB
}

// MemberSummaryStat 会员概览统计
type MemberSummaryStat struct {
	TotalMembers           int64 // 会员总数
	TodayNewMembers        int64 // 今日新增
	YesterdayNewMembers    int64 // 昨日新增
	Last24hNewMembers      int64 // 过去24小时新增
	Prev24hNewMembers      int64 // 前一个24小时新增
	ActiveMembers          int64 // 活跃会员总数
	TodayActiveMembers     int64 // 今日活跃
	YesterdayActiveMembers int64 // 昨日活跃
	Last24hActiveMembers   int64 // 过去24小时活跃
	Prev24hActiveMembers   int64 // 前一个24小时活跃
}

// NewMemberSummaryDao 创建会员概览DAO实例
func NewMemberSummaryDao(db *ent.Client, sqlDB *sql.DB) *MemberSummaryDao {
	return &MemberSummaryDao{db: db, sqlDB: sqlDB}
}

// GetMemberSummary 获取会员概览统计
func (d *MemberSummaryDao) GetMemberSummary(ctx context.Context, opCode, agentID string) (*MemberSummaryStat, error) {
	stat := &MemberSummaryStat{}

	// 获取总会员数
	totalMembers, err := d.getTotalMembersCount(ctx, opCode, agentID)
	if err != nil {
		logx.Errorf("[GetMemberSummary] getTotalMembersCount failed: %v", err)
		return nil, err
	}
	stat.TotalMembers = totalMembers

	// 获取今日新增
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayEnd := todayStart.AddDate(0, 0, 1)
	todayNewMembers, err := d.getNewMembersCountByTimeRange(ctx, opCode, agentID, todayStart, todayEnd)
	if err != nil {
		logx.Errorf("[GetMemberSummary] getNewMembersCountByTimeRange for today failed: %v", err)
		return nil, err
	}
	stat.TodayNewMembers = todayNewMembers

	// 获取昨日新增
	yesterdayStart := todayStart.AddDate(0, 0, -1)
	yesterdayNewMembers, err := d.getNewMembersCountByTimeRange(ctx, opCode, agentID, yesterdayStart, todayStart)
	if err != nil {
		logx.Errorf("[GetMemberSummary] getNewMembersCountByTimeRange for yesterday failed: %v", err)
		return nil, err
	}
	stat.YesterdayNewMembers = yesterdayNewMembers

	// 获取过去24小时新增
	last24hStart := now.Add(-24 * time.Hour)
	last24hNewMembers, err := d.getNewMembersCountByTimeRange(ctx, opCode, agentID, last24hStart, now)
	if err != nil {
		logx.Errorf("[GetMemberSummary] getNewMembersCountByTimeRange for last 24h failed: %v", err)
		return nil, err
	}
	stat.Last24hNewMembers = last24hNewMembers

	// 获取前一个24小时新增
	prev24hEnd := last24hStart
	prev24hStart := prev24hEnd.Add(-24 * time.Hour)
	prev24hNewMembers, err := d.getNewMembersCountByTimeRange(ctx, opCode, agentID, prev24hStart, prev24hEnd)
	if err != nil {
		logx.Errorf("[GetMemberSummary] getNewMembersCountByTimeRange for prev 24h failed: %v", err)
		return nil, err
	}
	stat.Prev24hNewMembers = prev24hNewMembers

	// 获取活跃会员总数
	activeMembers, err := d.getActiveMembersCount(ctx, opCode, agentID)
	if err != nil {
		logx.Errorf("[GetMemberSummary] getActiveMembersCount failed: %v", err)
		return nil, err
	}
	stat.ActiveMembers = activeMembers

	// 获取今日活跃
	todayActiveMembers, err := d.getActiveMembersCountByTimeRange(ctx, opCode, agentID, todayStart, todayEnd)
	if err != nil {
		logx.Errorf("[GetMemberSummary] getActiveMembersCountByTimeRange for today failed: %v", err)
		return nil, err
	}
	stat.TodayActiveMembers = todayActiveMembers

	// 获取昨日活跃
	yesterdayActiveMembers, err := d.getActiveMembersCountByTimeRange(ctx, opCode, agentID, yesterdayStart, todayStart)
	if err != nil {
		logx.Errorf("[GetMemberSummary] getActiveMembersCountByTimeRange for yesterday failed: %v", err)
		return nil, err
	}
	stat.YesterdayActiveMembers = yesterdayActiveMembers

	// 获取过去24小时活跃
	last24hActiveMembers, err := d.getActiveMembersCountByTimeRange(ctx, opCode, agentID, last24hStart, now)
	if err != nil {
		logx.Errorf("[GetMemberSummary] getActiveMembersCountByTimeRange for last 24h failed: %v", err)
		return nil, err
	}
	stat.Last24hActiveMembers = last24hActiveMembers

	// 获取前一个24小时活跃
	prev24hActiveMembers, err := d.getActiveMembersCountByTimeRange(ctx, opCode, agentID, prev24hStart, prev24hEnd)
	if err != nil {
		logx.Errorf("[GetMemberSummary] getActiveMembersCountByTimeRange for prev 24h failed: %v", err)
		return nil, err
	}
	stat.Prev24hActiveMembers = prev24hActiveMembers

	return stat, nil
}

// getTotalMembersCount 获取总会员数（历史注册总数，去重）
func (d *MemberSummaryDao) getTotalMembersCount(ctx context.Context, opCode, agentID string) (int64, error) {
	query := d.db.DwdMemberRegisterEvent.Query()

	if opCode != "" {
		query = query.Where(dwdmemberregisterevent.OpCodeEQ(opCode))
	}
	if agentID != "" {
		query = query.Where(dwdmemberregisterevent.AgentIDEQ(agentID))
	}

	// 使用Unique(true)实现DISTINCT去重
	count, err := query.Select(dwdmemberregisterevent.FieldUserID).
		Unique(true).
		Count(ctx)
	if err != nil {
		logx.Errorf("[getTotalMembersCount] query failed: %v", err)
		return 0, err
	}

	return int64(count), nil
}

// getNewMembersCountByTimeRange 获取时间范围内的新增会员数
func (d *MemberSummaryDao) getNewMembersCountByTimeRange(ctx context.Context, opCode, agentID string, startTime, endTime time.Time) (int64, error) {
	query := d.db.DwdMemberRegisterEvent.Query()

	if opCode != "" {
		query = query.Where(dwdmemberregisterevent.OpCodeEQ(opCode))
	}
	if agentID != "" {
		query = query.Where(dwdmemberregisterevent.AgentIDEQ(agentID))
	}

	query = query.Where(
		dwdmemberregisterevent.RegisterTimeGTE(startTime),
		dwdmemberregisterevent.RegisterTimeLT(endTime),
	)

	count, err := query.Count(ctx)
	if err != nil {
		logx.Errorf("[getNewMembersCountByTimeRange] query failed: %v", err)
		return 0, err
	}

	return int64(count), nil
}

// getActiveMembersCount 获取活跃会员总数（历史登录去重总数）
func (d *MemberSummaryDao) getActiveMembersCount(ctx context.Context, opCode, agentID string) (int64, error) {
	query := d.db.DwdMemberLoginEvent.Query()

	if opCode != "" {
		query = query.Where(dwdmemberloginevent.OpCodeEQ(opCode))
	}
	if agentID != "" {
		query = query.Where(dwdmemberloginevent.AgentIDEQ(agentID))
	}

	// 使用Unique(true)实现DISTINCT去重
	count, err := query.Select(dwdmemberloginevent.FieldUserID).
		Unique(true).
		Count(ctx)
	if err != nil {
		logx.Errorf("[getActiveMembersCount] query failed: %v", err)
		return 0, err
	}

	return int64(count), nil
}

// getActiveMembersCountByTimeRange 获取时间范围内的活跃会员数（去重）
func (d *MemberSummaryDao) getActiveMembersCountByTimeRange(ctx context.Context, opCode, agentID string, startTime, endTime time.Time) (int64, error) {
	query := d.db.DwdMemberLoginEvent.Query()

	if opCode != "" {
		query = query.Where(dwdmemberloginevent.OpCodeEQ(opCode))
	}
	if agentID != "" {
		query = query.Where(dwdmemberloginevent.AgentIDEQ(agentID))
	}

	query = query.Where(
		dwdmemberloginevent.LoginTimeGTE(startTime),
		dwdmemberloginevent.LoginTimeLT(endTime),
	)

	// 使用Unique(true)实现DISTINCT去重
	count, err := query.Select(dwdmemberloginevent.FieldUserID).
		Unique(true).
		Count(ctx)
	if err != nil {
		logx.Errorf("[getActiveMembersCountByTimeRange] query failed: %v", err)
		return 0, err
	}

	return int64(count), nil
}
