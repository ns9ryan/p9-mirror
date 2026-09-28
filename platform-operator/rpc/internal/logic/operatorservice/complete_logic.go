package operatorservicelogic

import (
	"context"

	"oa.98ent.com/p9/common/xerr"
	nodedispatchoperatornodepb "oa.98ent.com/p9/node-dispatch/rpc/pb/nodedispatchrpc/operatornodepb"
	"oa.98ent.com/p9/platform-operator/pkg/i18nkey"
	"oa.98ent.com/p9/platform-operator/rpc/internal/enterror"
	"oa.98ent.com/p9/platform-operator/rpc/internal/svc"
	"oa.98ent.com/p9/platform-operator/rpc/pb/platformoperatorrpc/operatorpb"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CompleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCompleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CompleteLogic {
	return &CompleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Complete 完成分站创建
func (l *CompleteLogic) Complete(in *operatorpb.CompleteOperatorRequest) (*operatorpb.CompleteOperatorResponse, error) {
	// 分站ID必须大于0
	if in.Id <= 0 {
		return nil, xerr.RpcErr(xerr.BadRequest(i18nkey.ValidationError))
	}

	// 获取当前分站
	current, err := l.svcCtx.DB.Operator.Get(l.ctx, in.Id)
	if err != nil {
		// 转换Ent错误为gRPC错误
		return nil, enterror.Handle(l.Logger, err)
	}

	// 已完成时直接返回
	if current.CreationStatus == 2 {
		return &operatorpb.CompleteOperatorResponse{}, nil
	}

	// 获取分站部署节点
	operatorNodeResult, err := l.svcCtx.NodeDispatchOperatorNodeRpc.Get(
		l.ctx,
		&nodedispatchoperatornodepb.GetOperatorNodeRequest{
			OperatorCode: current.Code, // 分站全局唯一业务编码
		},
	)
	if err != nil {
		// 未选择部署节点时不能完成创建
		if status.Code(err) == codes.NotFound {
			return nil, xerr.RpcErr(xerr.BadRequest(i18nkey.ValidationError))
		}

		l.Logger.Errorw(
			"检查分站部署节点失败",
			logx.Field("operator_id", current.ID),
			logx.Field("operator_code", current.Code),
			logx.Field("error", err.Error()),
		)
		return nil, err
	}

	// 部署节点信息必须完整
	if operatorNodeResult.OperatorNode == nil || operatorNodeResult.OperatorNode.Node == nil {
		return nil, xerr.RpcErr(xerr.BadRequest(i18nkey.ValidationError))
	}

	// 停用节点不能用于完成分站创建
	if operatorNodeResult.OperatorNode.Node.Status != 1 {
		return nil, xerr.RpcErr(xerr.BadRequest(i18nkey.ValidationError))
	}

	// 完成分站创建
	err = current.
		Update().
		SetCreationStatus(2). // 创建状态: 2已完成
		Exec(l.ctx)
	if err != nil {
		// 转换Ent错误为gRPC错误
		return nil, enterror.Handle(l.Logger, err)
	}

	// 返回完成结果
	return &operatorpb.CompleteOperatorResponse{}, nil
}
