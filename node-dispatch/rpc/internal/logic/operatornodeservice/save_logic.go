package operatornodeservicelogic

import (
	"context"
	"strings"

	"oa.98ent.com/p9/node-dispatch/rpc/ent"
	"oa.98ent.com/p9/node-dispatch/rpc/ent/node"
	"oa.98ent.com/p9/node-dispatch/rpc/ent/operatornode"
	"oa.98ent.com/p9/node-dispatch/rpc/internal/svc"
	"oa.98ent.com/p9/node-dispatch/rpc/pb/nodedispatchrpc/operatornodepb"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SaveLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveLogic {
	return &SaveLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Save 保存分站部署节点
func (l *SaveLogic) Save(in *operatornodepb.SaveOperatorNodeRequest) (*operatornodepb.SaveOperatorNodeResponse, error) {
	// 整理请求参数
	operatorCode := strings.TrimSpace(in.OperatorCode)
	nodeCode := strings.TrimSpace(in.NodeCode)

	// 校验请求参数
	if operatorCode == "" {
		return nil, status.Error(codes.InvalidArgument, "operator_code is required")
	}
	if nodeCode == "" {
		return nil, status.Error(codes.InvalidArgument, "node_code is required")
	}

	// 获取部署节点
	nodeData, err := l.svcCtx.DB.Node.
		Query().
		Where(node.CodeEQ(nodeCode)).
		Only(l.ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, "node not found")
		}

		l.Logger.Errorw(
			"获取分站部署节点失败",
			logx.Field("operator_code", operatorCode),
			logx.Field("node_code", nodeCode),
			logx.Field("error", err.Error()),
		)
		return nil, err
	}

	// 只有启用节点才能作为分站部署节点
	if nodeData.Status != 1 {
		return nil, status.Error(codes.FailedPrecondition, "node is disabled")
	}

	// 保存分站部署节点
	err = l.svcCtx.DB.OperatorNode.
		Create().
		SetOperatorCode(operatorCode). // 分站全局唯一业务编码
		SetNodeID(nodeData.ID).        // 部署节点ID
		OnConflictColumns(operatornode.FieldOperatorCode).
		UpdateNewValues().
		Exec(l.ctx)
	if err != nil {
		l.Logger.Errorw(
			"保存分站部署节点失败",
			logx.Field("operator_code", operatorCode),
			logx.Field("node_code", nodeCode),
			logx.Field("error", err.Error()),
		)
		return nil, err
	}

	// 返回保存结果
	return &operatornodepb.SaveOperatorNodeResponse{}, nil
}
