package operatornodeservicelogic

import (
	"context"
	"strings"

	"oa.98ent.com/p9/node-dispatch/rpc/ent"
	"oa.98ent.com/p9/node-dispatch/rpc/ent/operatornode"
	"oa.98ent.com/p9/node-dispatch/rpc/internal/svc"
	"oa.98ent.com/p9/node-dispatch/rpc/pb/nodedispatchrpc/operatornodepb"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLogic {
	return &GetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Get 获取分站部署节点
func (l *GetLogic) Get(in *operatornodepb.GetOperatorNodeRequest) (*operatornodepb.GetOperatorNodeResponse, error) {
	// 整理请求参数
	operatorCode := strings.TrimSpace(in.OperatorCode)

	// 校验请求参数
	if operatorCode == "" {
		return nil, status.Error(codes.InvalidArgument, "operator_code is required")
	}

	// 获取分站部署节点关系
	data, err := l.svcCtx.DB.OperatorNode.
		Query().
		Where(operatornode.OperatorCodeEQ(operatorCode)).
		WithNode().
		Only(l.ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, "operator node not found")
		}

		l.Logger.Errorw(
			"获取分站部署节点失败",
			logx.Field("operator_code", operatorCode),
			logx.Field("error", err.Error()),
		)
		return nil, err
	}

	// 获取节点信息
	nodeData, err := data.Edges.NodeOrErr()
	if err != nil {
		l.Logger.Errorw(
			"获取分站部署节点信息失败",
			logx.Field("operator_code", operatorCode),
			logx.Field("error", err.Error()),
		)
		return nil, err
	}

	// 获取节点在线状态
	online := l.svcCtx.Connections.IsOnline(nodeData.Code)

	// 返回分站部署节点
	return &operatornodepb.GetOperatorNodeResponse{
		OperatorNode: toOperatorNodeInfo(data, nodeData, online), // 分站部署节点信息
	}, nil
}
