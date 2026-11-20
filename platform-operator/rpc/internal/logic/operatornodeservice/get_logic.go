package operatornodeservicelogic

import (
	"context"

	"oa.98ent.com/p9/common/xerr"
	nodedispatchoperatornodepb "oa.98ent.com/p9/node-dispatch/rpc/pb/nodedispatchrpc/operatornodepb"
	"oa.98ent.com/p9/platform-operator/pkg/i18nkey"
	"oa.98ent.com/p9/platform-operator/rpc/internal/enterror"
	"oa.98ent.com/p9/platform-operator/rpc/internal/svc"
	"oa.98ent.com/p9/platform-operator/rpc/pb/platformoperatorrpc/operatornodepb"

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
	// 校验分站ID
	if in.OperatorId <= 0 {
		return nil, xerr.RpcErr(xerr.BadRequest(i18nkey.ValidationError))
	}

	// 获取当前分站
	current, err := l.svcCtx.DB.Operator.Get(l.ctx, in.OperatorId)
	if err != nil {
		// 转换Ent错误为gRPC错误
		return nil, enterror.Handle(l.Logger, err)
	}

	// 获取分站部署节点
	result, err := l.svcCtx.NodeDispatchOperatorNodeRpc.Get(
		l.ctx,
		&nodedispatchoperatornodepb.GetOperatorNodeRequest{
			OperatorCode: current.Code, // 分站全局唯一业务编码
		},
	)
	if err != nil {
		// 草稿阶段尚未选择部署节点时正常返回空结果
		if status.Code(err) == codes.NotFound {
			return &operatornodepb.GetOperatorNodeResponse{}, nil
		}

		l.Logger.Errorw(
			"获取分站部署节点失败",
			logx.Field("operator_id", current.ID),
			logx.Field("operator_code", current.Code),
			logx.Field("error", err.Error()),
		)
		return nil, err
	}

	// 返回分站部署节点
	return &operatornodepb.GetOperatorNodeResponse{
		OperatorNode: toOperatorNodeInfo(current.ID, result.OperatorNode), // 分站部署节点信息
	}, nil
}
