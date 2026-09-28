package operatorservicelogic

import (
	"context"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/common/xerr"
	nodedispatchdispatchpb "oa.98ent.com/p9/node-dispatch/rpc/pb/nodedispatchrpc/dispatchpb"
	"oa.98ent.com/p9/platform-operator/pkg/i18nkey"
	"oa.98ent.com/p9/platform-operator/rpc/ent"
	"oa.98ent.com/p9/platform-operator/rpc/ent/operator"
	"oa.98ent.com/p9/platform-operator/rpc/internal/enterror"
	"oa.98ent.com/p9/platform-operator/rpc/internal/svc"
	"oa.98ent.com/p9/platform-operator/rpc/pb/platformoperatorrpc/operatorpb"
)

type HandlePublishResultLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewHandlePublishResultLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandlePublishResultLogic {
	return &HandlePublishResultLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// HandlePublishResult 处理分站发布结果
func (l *HandlePublishResultLogic) HandlePublishResult(in *operatorpb.HandlePublishResultRequest) (*operatorpb.HandlePublishResultResponse, error) {
	// 整理调度任务编号
	taskNo := strings.TrimSpace(in.TaskNo)
	if taskNo == "" {
		return nil, xerr.RpcErr(xerr.BadRequest(i18nkey.ValidationError))
	}

	// 从调度中心获取实际任务状态
	taskResult, err := l.svcCtx.NodeDispatchDispatchRpc.GetTask(
		l.ctx,
		&nodedispatchdispatchpb.GetTaskRequest{
			TaskNo: new(taskNo), // 调度任务编号
		},
	)
	if err != nil {
		l.Logger.Errorw(
			"获取分站发布任务失败",
			logx.Field("task_no", taskNo),
			logx.Field("error", err.Error()),
		)
		return nil, err
	}

	// 调度任务信息必须完整
	if taskResult == nil || taskResult.Task == nil {
		l.Logger.Errorw(
			"获取分站发布任务结果为空",
			logx.Field("task_no", taskNo),
		)
		return nil, xerr.RpcErr(xerr.InternalServerError(i18nkey.InternalError))
	}

	// 校验发布任务并获取发布请求编号
	_, requestNo, err := validatePublishTask(taskResult.Task, taskNo, "")
	if err != nil {
		return nil, err
	}

	// 根据发布请求编号获取当前分站
	current, err := l.svcCtx.DB.Operator.
		Query().
		Where(operator.PublishRequestNoEQ(requestNo)).
		Only(l.ctx)
	if err != nil {
		// 当前分站已进入新的发布轮次时忽略旧任务回调
		if ent.IsNotFound(err) {
			l.Logger.Infow(
				"忽略非当前分站发布任务结果",
				logx.Field("task_no", taskNo),
				logx.Field("request_no", requestNo),
			)
			return &operatorpb.HandlePublishResultResponse{}, nil
		}

		// 转换Ent错误为gRPC错误
		return nil, enterror.Handle(l.Logger, err)
	}

	// 应用最终发布任务结果
	updated, handled, err := applyFinalPublishTask(
		l.ctx,
		l.svcCtx,
		l.Logger,
		current,
		taskResult.Task,
	)
	if err != nil {
		return nil, err
	}

	// 当前已经进入新的发布轮次时忽略旧任务结果
	if !handled {
		return &operatorpb.HandlePublishResultResponse{}, nil
	}

	// 返回发布结果
	return &operatorpb.HandlePublishResultResponse{
		OperatorId:    updated.ID,            // 分站ID
		OperatorCode:  updated.Code,          // 分站业务编码
		PublishStatus: updated.PublishStatus, // 发布状态: 3已发布, 4发布失败
	}, nil
}
