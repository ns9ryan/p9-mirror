package statisticsservicelogic

import (
	"context"
	"fmt"

	"oa.98ent.com/p9/operator-report/rpc/internal/svc"
	"oa.98ent.com/p9/operator-report/rpc/pb/operator_report"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMemberSummaryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMemberSummaryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMemberSummaryLogic {
	return &GetMemberSummaryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取会员概览（总数、新增、活跃等）
func (l *GetMemberSummaryLogic) GetMemberSummary(in *operator_report.MemberSummaryRequest) (*operator_report.MemberSummaryResponse, error) {
	if in.OpCode == "" {
		l.Errorf("[GetMemberSummary] op_code is required")
		return nil, fmt.Errorf("op_code is required")
	}

	// 获取会员统计数据
	stat, err := l.svcCtx.DAOManager.MemberSummary.GetMemberSummary(l.ctx, in.OpCode, in.AgentId)
	if err != nil {
		l.Errorf("[GetMemberSummary] GetMemberSummary failed: %v", err)
		return nil, err
	}

	// 计算增长比例
	var membersGrowthRate float64 = 0.0
	if stat.YesterdayNewMembers > 0 {
		membersGrowthRate = float64(stat.TodayNewMembers-stat.YesterdayNewMembers) / float64(stat.YesterdayNewMembers) * 100
	}

	var last24hNewGrowthRate float64 = 0.0
	if stat.Prev24hNewMembers > 0 {
		last24hNewGrowthRate = float64(stat.Last24hNewMembers-stat.Prev24hNewMembers) / float64(stat.Prev24hNewMembers) * 100
	}

	var last24hActiveGrowthRate float64 = 0.0
	if stat.Prev24hActiveMembers > 0 {
		last24hActiveGrowthRate = float64(stat.Last24hActiveMembers-stat.Prev24hActiveMembers) / float64(stat.Prev24hActiveMembers) * 100
	}

	l.Infof("[GetMemberSummary] stat - totalMembers: %d, todayNewMembers: %d, yesterdayNewMembers: %d, "+
		"last24hNewMembers: %d, prev24hNewMembers: %d, activeMembers: %d, todayActiveMembers: %d, "+
		"yesterdayActiveMembers: %d, last24hActiveMembers: %d, prev24hActiveMembers: %d",
		stat.TotalMembers, stat.TodayNewMembers, stat.YesterdayNewMembers,
		stat.Last24hNewMembers, stat.Prev24hNewMembers, stat.ActiveMembers, stat.TodayActiveMembers,
		stat.YesterdayActiveMembers, stat.Last24hActiveMembers, stat.Prev24hActiveMembers)

	return &operator_report.MemberSummaryResponse{
		Data: &operator_report.MemberSummary{
			TotalMembers:             stat.TotalMembers,
			MembersGrowthRate:        membersGrowthRate,
			TodayNewMembers:          stat.TodayNewMembers,
			Last_24HNewMembers:       stat.Last24hNewMembers,
			Last_24HNewGrowthRate:    last24hNewGrowthRate,
			TodayDepositAmount:       0, // 充值相关先填默认值
			DepositGrowthRate:        0.0,
			ActiveMembers:            stat.ActiveMembers,
			TodayActiveMembers:       stat.TodayActiveMembers,
			Last_24HActiveMembers:    stat.Last24hActiveMembers,
			Last_24HActiveGrowthRate: last24hActiveGrowthRate,
		},
	}, nil
}
