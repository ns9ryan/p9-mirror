package operatorservicelogic

import (
	"context"
	"strings"
	"time"

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
	// 整理请求参数
	taskNo := strings.TrimSpace(in.TaskNo)
	if taskNo == "" {
		return nil, xerr.RpcErr(xerr.BadRequest(i18nkey.ValidationError))
	}

	// 从调度中心获取最终任务状态
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

	taskData := taskResult.Task

	// 校验调度任务基本信息
	actualTaskNo := strings.TrimSpace(taskData.TaskNo)
	requestNo := strings.TrimSpace(taskData.RequestNo)

	if actualTaskNo == "" || actualTaskNo != taskNo || requestNo == "" {
		l.Logger.Errorw(
			"分站发布任务信息无效",
			logx.Field("task_no", taskNo),
			logx.Field("actual_task_no", actualTaskNo),
			logx.Field("request_no", requestNo),
		)
		return nil, xerr.RpcErr(xerr.InternalServerError(i18nkey.InternalError))
	}

	// 只处理分站创建任务
	if taskData.Target != dispatchTargetOperatorBase || taskData.TaskType != dispatchTaskTypeCreateOperator {
		return nil, xerr.RpcErr(xerr.BadRequest(i18nkey.ConstraintError))
	}

	// 只处理最终任务状态
	if taskData.Status != 3 && taskData.Status != 4 {
		return nil, xerr.RpcErr(xerr.BadRequest(i18nkey.ConstraintError))
	}

	// 最终任务必须包含结束时间
	if taskData.FinishedAt == nil || *taskData.FinishedAt <= 0 {
		l.Logger.Errorw(
			"分站发布任务结束时间为空",
			logx.Field("task_no", taskNo),
			logx.Field("request_no", requestNo),
		)
		return nil, xerr.RpcErr(xerr.InternalServerError(i18nkey.InternalError))
	}

	// 根据发布请求编号获取当前分站
	current, err := l.svcCtx.DB.Operator.
		Query().
		Where(operator.PublishRequestNoEQ(requestNo)).
		Only(l.ctx)
	if err != nil {
		// 当前分站已进入新的发布轮次时，忽略旧任务的迟到回调
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

	// 相同任务结果重复回调时直接返回
	if current.PublishStatus == taskData.Status {
		if current.PublishTaskNo != nil && *current.PublishTaskNo == taskNo {
			return &operatorpb.HandlePublishResultResponse{
				OperatorId:    current.ID,            // 分站ID
				OperatorCode:  current.Code,          // 分站业务编码
				PublishStatus: current.PublishStatus, // 发布状态
			}, nil
		}

		return nil, xerr.RpcErr(xerr.BadRequest(i18nkey.ConstraintError))
	}

	// 只有发布中的分站允许接收最终发布结果
	if current.PublishStatus != 2 {
		return nil, xerr.RpcErr(xerr.BadRequest(i18nkey.ConstraintError))
	}

	// 创建发布结果更新
	update := l.svcCtx.DB.Operator.
		Update().
		Where(
			operator.IDEQ(current.ID),
			operator.PublishRequestNoEQ(requestNo),
			operator.PublishStatusEQ(2),
		).
		SetPublishTaskNo(taskNo).         // 当前发布任务编号
		SetPublishStatus(taskData.Status) // 发布状态: 3已发布, 4发布失败

	// 首次发布成功时记录实际任务完成时间
	if taskData.Status == 3 && current.PublishedAt == nil {
		update.SetPublishedAt(time.UnixMilli(*taskData.FinishedAt))
	}

	// 保存发布结果
	affected, err := update.Save(l.ctx)
	if err != nil {
		// 转换Ent错误为gRPC错误
		return nil, enterror.Handle(l.Logger, err)
	}

	// 发布轮次在更新前已经变化时，不处理旧任务结果
	if affected == 0 {
		latest, getErr := l.svcCtx.DB.Operator.Get(l.ctx, current.ID)
		if getErr != nil {
			// 转换Ent错误为gRPC错误
			return nil, enterror.Handle(l.Logger, getErr)
		}

		// 当前已经不是本次发布请求时，忽略迟到结果
		if latest.PublishRequestNo == nil || *latest.PublishRequestNo != requestNo {
			return &operatorpb.HandlePublishResultResponse{}, nil
		}

		// 相同结果已被其他流程处理时直接返回
		if latest.PublishStatus == taskData.Status &&
			latest.PublishTaskNo != nil &&
			*latest.PublishTaskNo == taskNo {
			return &operatorpb.HandlePublishResultResponse{
				OperatorId:    latest.ID,            // 分站ID
				OperatorCode:  latest.Code,          // 分站业务编码
				PublishStatus: latest.PublishStatus, // 发布状态
			}, nil
		}

		return nil, xerr.RpcErr(xerr.BadRequest(i18nkey.ConstraintError))
	}

	// 返回发布结果
	return &operatorpb.HandlePublishResultResponse{
		OperatorId:    current.ID,      // 分站ID
		OperatorCode:  current.Code,    // 分站业务编码
		PublishStatus: taskData.Status, // 发布状态: 3已发布, 4发布失败
	}, nil
}
