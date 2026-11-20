package statisticsservicelogic

import (
	"context"
	"fmt"
	"time"

	"oa.98ent.com/p9/operator-report/rpc/internal/dao"
	"oa.98ent.com/p9/operator-report/rpc/internal/svc"
	"oa.98ent.com/p9/operator-report/rpc/pb/operator_report"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDailyGrowthTrendLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDailyGrowthTrendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDailyGrowthTrendLogic {
	return &GetDailyGrowthTrendLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取会员增长趋势
func (l *GetDailyGrowthTrendLogic) GetDailyGrowthTrend(in *operator_report.MemberGrowthTrendRequest) (*operator_report.MemberGrowthTrendResponse, error) {
	// 参数验证
	if err := l.validateRequest(in); err != nil {
		l.Errorf("[GetDailyGrowthTrend] validation error: %v", err)
		return &operator_report.MemberGrowthTrendResponse{}, err
	}
	opCode := in.OpCode

	l.Infof("[GetDailyGrowthTrend] request - opCode: %s, agentId: %s, type: %s, granularity: %s, source: %s",
		opCode, in.AgentId, in.Type, in.Granularity, in.Source)

	// 解析日期
	startDate, err := time.Parse("2006-01-02", in.StartDate)
	if err != nil {
		l.Errorf("[GetDailyGrowthTrend] invalid start date: %s", in.StartDate)
		return &operator_report.MemberGrowthTrendResponse{}, fmt.Errorf("invalid start date format")
	}

	endDate, err := time.Parse("2006-01-02", in.EndDate)
	if err != nil {
		l.Errorf("[GetDailyGrowthTrend] invalid end date: %s", in.EndDate)
		return &operator_report.MemberGrowthTrendResponse{}, fmt.Errorf("invalid end date format")
	}
	// 将 EndDate 设置为该天的最后一刻（23:59:59）
	// 这样可以包含 EndDate 这一整天的数据
	endDate = endDate.AddDate(0, 0, 1).Add(-time.Second)

	l.Infof("[GetDailyGrowthTrend] parsed dates - startDate: %v, endDate: %v", startDate, endDate)

	// 查询数据
	param := dao.MemberGrowthTrendParam{
		OpCode:      opCode,
		AgentID:     in.AgentId,
		Source:      in.Source,
		StartDate:   startDate,
		EndDate:     endDate,
		Granularity: in.Granularity,
		Type:        in.Type,
	}

	if l.svcCtx == nil || l.svcCtx.DAOManager == nil {
		l.Errorf("[GetDailyGrowthTrend] DAO Manager not available")
		return &operator_report.MemberGrowthTrendResponse{}, fmt.Errorf("DAO Manager not available")
	}

	trends, err := l.svcCtx.DAOManager.MemberGrowthTrend.GetGrowthTrend(l.ctx, param)
	if err != nil {
		l.Errorf("[GetDailyGrowthTrend] query failed: %v", err)
		return &operator_report.MemberGrowthTrendResponse{}, err
	}

	// 转换为响应格式
	var items []*operator_report.GrowthTrendStatItem
	for _, trend := range trends {
		item := &operator_report.GrowthTrendStatItem{
			Date:  trend.Date,
			Count: trend.Count,
		}
		items = append(items, item)
	}

	return &operator_report.MemberGrowthTrendResponse{
		List: items,
	}, nil
}

// validateRequest 验证请求参数
func (l *GetDailyGrowthTrendLogic) validateRequest(in *operator_report.MemberGrowthTrendRequest) error {
	// 验证日期格式
	if in.StartDate == "" {
		return fmt.Errorf("start_date is required")
	}
	if in.EndDate == "" {
		return fmt.Errorf("end_date is required")
	}

	// 验证日期格式
	_, err := time.Parse("2006-01-02", in.StartDate)
	if err != nil {
		return fmt.Errorf("start_date format should be YYYY-MM-DD")
	}

	_, err = time.Parse("2006-01-02", in.EndDate)
	if err != nil {
		return fmt.Errorf("end_date format should be YYYY-MM-DD")
	}

	// 验证粒度
	if in.Granularity == "" {
		return fmt.Errorf("granularity is required")
	}
	if in.Granularity != "day" && in.Granularity != "hour" {
		return fmt.Errorf("granularity should be 'day' or 'hour'")
	}

	// 验证统计类型
	if in.Type == "" {
		return fmt.Errorf("type is required")
	}
	// if in.Type != "register" && in.Type != "login" && in.Type != "retention" {
	// 	return fmt.Errorf("type should be 'register', 'login' or 'retention'")
	// }

	// 留存率不支持小时维度
	if in.Type == "retention" && in.Granularity == "hour" {
		return fmt.Errorf("retention statistics do not support hour granularity")
	}

	return nil
}
