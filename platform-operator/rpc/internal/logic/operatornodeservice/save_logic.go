package operatornodeservicelogic

import (
	"context"
	"strings"

	"oa.98ent.com/p9/common/xerr"
	nodedispatchoperatornodepb "oa.98ent.com/p9/node-dispatch/rpc/pb/nodedispatchrpc/operatornodepb"
	"oa.98ent.com/p9/platform-operator/pkg/i18nkey"
	"oa.98ent.com/p9/platform-operator/rpc/internal/enterror"
	"oa.98ent.com/p9/platform-operator/rpc/internal/svc"
	"oa.98ent.com/p9/platform-operator/rpc/pb/platformoperatorrpc/operatornodepb"

	"github.com/zeromicro/go-zero/core/logx"
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
	// 校验请求参数
	if in.OperatorId <= 0 {
		return nil, xerr.RpcErr(xerr.BadRequest(i18nkey.ValidationError))
	}

	nodeCode := strings.TrimSpace(in.NodeCode)
	if nodeCode == "" {
		return nil, xerr.RpcErr(xerr.BadRequest(i18nkey.ValidationError))
	}

	// 获取当前分站
	current, err := l.svcCtx.DB.Operator.Get(l.ctx, in.OperatorId)
	if err != nil {
		// 转换Ent错误为gRPC错误
		return nil, enterror.Handle(l.Logger, err)
	}

	// 发布中或已经首次发布成功的分站不能修改部署节点
	if current.PublishStatus == 2 || current.PublishedAt != nil {
		return nil, xerr.RpcErr(xerr.BadRequest(i18nkey.ValidationError))
	}

	// 保存分站部署节点
	_, err = l.svcCtx.NodeDispatchOperatorNodeRpc.Save(
		l.ctx,
		&nodedispatchoperatornodepb.SaveOperatorNodeRequest{
			OperatorCode: current.Code, // 分站全局唯一业务编码
			NodeCode:     nodeCode,     // 节点业务编码
		},
	)
	if err != nil {
		l.Logger.Errorw(
			"保存分站部署节点失败",
			logx.Field("operator_id", current.ID),
			logx.Field("operator_code", current.Code),
			logx.Field("node_code", nodeCode),
			logx.Field("error", err.Error()),
		)
		return nil, err
	}

	// 返回保存结果
	return &operatornodepb.SaveOperatorNodeResponse{}, nil
}
