// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package dashboard

import (
	"context"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/operator-report/api/internal/svc"
	"oa.98ent.com/p9/operator-report/api/internal/types"
	"oa.98ent.com/p9/operator-report/rpc/pb/operator_report"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMemberSummaryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMemberSummaryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMemberSummaryLogic {
	return &GetMemberSummaryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMemberSummaryLogic) GetMemberSummary(req *types.MemberSummaryRequest) (resp *types.MemberSummaryResponse, err error) {
	// 获取分站编码
	opCode := ctxdata.OperatorCodeFromCtx(l.ctx)
	if opCode == "" {
		// 测试
		opCode = "OP_TEST_001"
	}

	// 构建 RPC 请求
	rpcReq := &operator_report.MemberSummaryRequest{
		OpCode:  opCode,
		AgentId: req.AgentId,
	}

	l.Infof("[GetMemberSummary] calling RPC with opCode=%s, agentId=%s",
		opCode, req.AgentId)

	// 调用 RPC 服务
	rpcResp, err := l.svcCtx.OperatorReportGrpcClient.GetStatisticsServiceClient().GetMemberSummary(l.ctx, rpcReq)
	if err != nil {
		l.Errorf("[GetMemberSummary] RPC call failed: %v", err)
		return &types.MemberSummaryResponse{
			Code:    500,
			Message: "failed to get member summary: " + err.Error(),
			Data:    types.MemberSummary{},
		}, nil
	}

	if rpcResp == nil || rpcResp.Data == nil {
		l.Errorf("[GetMemberSummary] RPC returned nil response")
		return &types.MemberSummaryResponse{
			Code:    500,
			Message: "rpc returned nil response",
			Data:    types.MemberSummary{},
		}, nil
	}

	// 转换响应数据
	data := types.MemberSummary{
		TotalMembers:            rpcResp.Data.TotalMembers,
		MembersGrowthRate:       rpcResp.Data.MembersGrowthRate,
		TodayNewMembers:         rpcResp.Data.TodayNewMembers,
		Last24HNewMembers:       rpcResp.Data.Last_24HNewMembers,
		Last24HNewGrowthRate:    rpcResp.Data.Last_24HNewGrowthRate,
		TodayDepositAmount:      rpcResp.Data.TodayDepositAmount,
		DepositGrowthRate:       rpcResp.Data.DepositGrowthRate,
		ActiveMembers:           rpcResp.Data.ActiveMembers,
		TodayActiveMembers:      rpcResp.Data.TodayActiveMembers,
		Last24HActiveMembers:    rpcResp.Data.Last_24HActiveMembers,
		Last24HActiveGrowthRate: rpcResp.Data.Last_24HActiveGrowthRate,
		// YesterdayNewMembers	yesterday_new_members	昨日新增会员数
		// NewMembersGrowthRate	new_members_growth_rate	今日新增较昨日增长比例
		// WeekNewMembers	week_new_members	本周新增会员数
		// MonthNewMembers	month_new_members	本月新增会员数
		// TotalRegisteredMembers	total_registered_members	累计注册会员数
		// TodayFirstLoginMembers	today_first_login_members	今日首次登录会员数
		// NeverLoginMembers	never_login_members	注册后从未登录的会员数
		// NeverLoginRate	never_login_rate	注册后从未登录的会员占比
	}

	l.Infof("[GetMemberSummary] success, totalMembers=%d, todayNewMembers=%d, activeMembers=%d, todayActiveMembers=%d",
		data.TotalMembers, data.TodayNewMembers, data.ActiveMembers, data.TodayActiveMembers)

	return &types.MemberSummaryResponse{
		Code:    0,
		Message: "ok",
		Data:    data,
	}, nil
}
