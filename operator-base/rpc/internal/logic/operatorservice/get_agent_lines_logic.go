package operatorservicelogic

import (
	"context"
	"strings"

	"oa.98ent.com/p9/operator-base/rpc/ent"
	"oa.98ent.com/p9/operator-base/rpc/ent/operator"
	"oa.98ent.com/p9/operator-base/rpc/ent/operatoragentline"
	"oa.98ent.com/p9/operator-base/rpc/internal/svc"
	"oa.98ent.com/p9/operator-base/rpc/pb/operatorbaserpc/operatorpb"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GetAgentLinesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAgentLinesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAgentLinesLogic {
	return &GetAgentLinesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取分站已分配代理子线路
func (l *GetAgentLinesLogic) GetAgentLines(in *operatorpb.GetAgentLinesRequest) (*operatorpb.GetAgentLinesResponse, error) {
	operatorCode := strings.TrimSpace(in.GetOperatorCode())
	if operatorCode == "" {
		return nil, status.Error(codes.InvalidArgument, "operator_code is required")
	}

	operatorData, err := l.svcCtx.DB.Operator.Query().
		Where(operator.CodeEQ(operatorCode)).
		First(l.ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, "operator not found")
		}

		l.Logger.Errorw(
			"查询分站失败",
			logx.Field("operator_code", operatorCode),
			logx.Field("error", err.Error()),
		)
		return nil, err
	}

	agentLines, err := l.svcCtx.DB.OperatorAgentLine.Query().
		Where(operatoragentline.OperatorIDEQ(operatorData.ID)).
		Order(operatoragentline.ByCreatedAt(), operatoragentline.ByID()).
		All(l.ctx)
	if err != nil {
		l.Logger.Errorw(
			"查询分站代理子线路失败",
			logx.Field("operator_code", operatorCode),
			logx.Field("error", err.Error()),
		)
		return nil, err
	}

	agentLineCodes := make([]string, 0, len(agentLines))
	for _, item := range agentLines {
		agentLineCodes = append(agentLineCodes, item.AgentLineCode)
	}

	return &operatorpb.GetAgentLinesResponse{
		AgentLineCodes: agentLineCodes,
	}, nil
}
