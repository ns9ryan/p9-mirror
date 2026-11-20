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

type GetMemberGrowthTrendLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMemberGrowthTrendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMemberGrowthTrendLogic {
	return &GetMemberGrowthTrendLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMemberGrowthTrendLogic) GetMemberGrowthTrend(req *types.MemberGrowthTrendRequest) (resp *types.MemberGrowthTrendResponse, err error) {
	// 获取分站编码
	opCode := ctxdata.OperatorCodeFromCtx(l.ctx)
	if opCode == "" {
		// 测试
		opCode = "OP_TEST_001"
		// l.Errorf("[GetMemberGrowthTrend] operator code not found in context")
		// return &types.MemberGrowthTrendResponse{
		// 	Code:    400,
		// 	Message: "operator code not found",
		// 	List:    []types.GrowthTrendStatItem{},
		// }, nil
	}

	// 构建 RPC 请求
	rpcReq := &operator_report.MemberGrowthTrendRequest{
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		Granularity: req.Granularity,
		AgentId:     req.AgentId,
		Source:      req.Source,
		Type:        req.Type,
		OpCode:      opCode,
	}

	l.Infof("[GetMemberGrowthTrend] calling RPC with opCode=%s, startDate=%s, endDate=%s, type=%s, granularity=%s, agentId=%s, source=%s",
		opCode, req.StartDate, req.EndDate, req.Type, req.Granularity, req.AgentId, req.Source)

	// 调用 RPC 服务
	rpcResp, err := l.svcCtx.OperatorReportGrpcClient.GetStatisticsServiceClient().GetDailyGrowthTrend(l.ctx, rpcReq)
	if err != nil {
		l.Errorf("[GetMemberGrowthTrend] RPC call failed: %v", err)
		return &types.MemberGrowthTrendResponse{
			Code:    500,
			Message: "failed to get member growth trend: " + err.Error(),
			List:    []types.GrowthTrendStatItem{},
		}, nil
	}

	// 转换响应数据
	var items []types.GrowthTrendStatItem
	if rpcResp != nil && rpcResp.List != nil {
		for _, item := range rpcResp.List {
			items = append(items, types.GrowthTrendStatItem{
				Date:  item.Date,
				Count: item.Count,
			})
		}
	}

	l.Infof("[GetMemberGrowthTrend] success, returned %d items", len(items))

	return &types.MemberGrowthTrendResponse{
		Code:    0,
		Message: "ok",
		List:    items,
	}, nil
}
