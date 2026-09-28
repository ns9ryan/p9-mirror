package operatorservicelogic

import (
	"context"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"oa.98ent.com/p9/common/xerr"
	nodedispatchdispatchpb "oa.98ent.com/p9/node-dispatch/rpc/pb/nodedispatchrpc/dispatchpb"
	"oa.98ent.com/p9/platform-operator/pkg/i18nkey"
	"oa.98ent.com/p9/platform-operator/rpc/ent/operator"
	"oa.98ent.com/p9/platform-operator/rpc/internal/enterror"
	"oa.98ent.com/p9/platform-operator/rpc/internal/svc"
	"oa.98ent.com/p9/platform-operator/rpc/pb/platformoperatorrpc/operatorpb"
)

type SyncPublishStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSyncPublishStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncPublishStatusLogic {
	return &SyncPublishStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SyncPublishStatus 同步分站发布状态
func (l *SyncPublishStatusLogic) SyncPublishStatus(in *operatorpb.SyncPublishStatusRequest) (*operatorpb.SyncPublishStatusResponse, error) {
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

	// 已存在最终发布结果时直接返回
	if current.PublishStatus == 3 || current.PublishStatus == 4 {
		return &operatorpb.SyncPublishStatusResponse{
			OperatorId:    current.ID,            // 分站ID
			OperatorCode:  current.Code,          // 分站业务编码
			PublishStatus: current.PublishStatus, // 发布状态
		}, nil
	}

	// 只有发布中的分站需要同步
	if current.PublishStatus != 2 {
		return nil, xerr.RpcErr(xerr.BadRequest(i18nkey.ConstraintError))
	}

	// 当前发布必须存在请求编号
	if current.PublishRequestNo == nil || strings.TrimSpace(*current.PublishRequestNo) == "" {
		l.Logger.Errorw(
			"分站发布请求编号为空",
			logx.Field("operator_id", current.ID),
			logx.Field("operator_code", current.Code),
		)
		return nil, xerr.RpcErr(xerr.InternalServerError(i18nkey.InternalError))
	}

	requestNo := strings.TrimSpace(*current.PublishRequestNo)

	// 优先使用已记录任务编号查询
	var currentTaskNo string
	if current.PublishTaskNo != nil {
		currentTaskNo = strings.TrimSpace(*current.PublishTaskNo)
	}

	taskRequest := &nodedispatchdispatchpb.GetTaskRequest{}
	if currentTaskNo != "" {
		taskRequest.TaskNo = new(currentTaskNo)
	} else {
		taskRequest.RequestNo = new(requestNo)
	}

	// 从调度中心获取实际任务状态
	taskResult, err := l.svcCtx.NodeDispatchDispatchRpc.GetTask(l.ctx, taskRequest)
	if err != nil {
		// 未找到任务时保持发布中状态，避免误判仍在提交中的任务
		if status.Code(err) == codes.NotFound {
			l.Logger.Infow(
				"同步分站发布状态未找到调度任务",
				logx.Field("operator_id", current.ID),
				logx.Field("operator_code", current.Code),
				logx.Field("request_no", requestNo),
				logx.Field("task_no", currentTaskNo),
			)
			return nil, err
		}

		l.Logger.Errorw(
			"同步分站发布状态获取调度任务失败",
			logx.Field("operator_id", current.ID),
			logx.Field("operator_code", current.Code),
			logx.Field("request_no", requestNo),
			logx.Field("task_no", currentTaskNo),
			logx.Field("error", err.Error()),
		)
		return nil, err
	}

	// 调度任务信息必须完整
	if taskResult == nil || taskResult.Task == nil {
		l.Logger.Errorw(
			"同步分站发布状态任务结果为空",
			logx.Field("operator_id", current.ID),
			logx.Field("operator_code", current.Code),
			logx.Field("request_no", requestNo),
		)
		return nil, xerr.RpcErr(xerr.InternalServerError(i18nkey.InternalError))
	}

	// 校验当前发布任务
	taskNo, _, err := validatePublishTask(
		taskResult.Task,
		currentTaskNo,
		requestNo,
	)
	if err != nil {
		return nil, err
	}

	// 待执行或执行中时只补充实际任务编号
	if taskResult.Task.Status == 1 || taskResult.Task.Status == 2 {
		// 已经记录当前任务编号时无需重复更新
		if currentTaskNo == taskNo {
			return &operatorpb.SyncPublishStatusResponse{
				OperatorId:    current.ID,            // 分站ID
				OperatorCode:  current.Code,          // 分站业务编码
				PublishStatus: current.PublishStatus, // 发布状态: 2发布中
			}, nil
		}

		// 记录当前发布实际任务编号
		affected, updateErr := l.svcCtx.DB.Operator.
			Update().
			Where(
				operator.IDEQ(current.ID),
				operator.PublishRequestNoEQ(requestNo),
				operator.PublishStatusEQ(2),
			).
			SetPublishTaskNo(taskNo). // 当前发布任务编号
			Save(l.ctx)
		if updateErr != nil {
			// 转换Ent错误为gRPC错误
			return nil, enterror.Handle(l.Logger, updateErr)
		}

		// 状态已被其他流程修改时返回最新状态
		if affected == 0 {
			latest, getErr := l.svcCtx.DB.Operator.Get(l.ctx, current.ID)
			if getErr != nil {
				return nil, enterror.Handle(l.Logger, getErr)
			}

			return &operatorpb.SyncPublishStatusResponse{
				OperatorId:    latest.ID,            // 分站ID
				OperatorCode:  latest.Code,          // 分站业务编码
				PublishStatus: latest.PublishStatus, // 当前发布状态
			}, nil
		}

		return &operatorpb.SyncPublishStatusResponse{
			OperatorId:    current.ID,   // 分站ID
			OperatorCode:  current.Code, // 分站业务编码
			PublishStatus: 2,            // 发布状态: 2发布中
		}, nil
	}

	// 应用最终发布任务结果
	updated, _, err := applyFinalPublishTask(
		l.ctx,
		l.svcCtx,
		l.Logger,
		current,
		taskResult.Task,
	)
	if err != nil {
		return nil, err
	}

	// 返回同步后的发布状态
	return &operatorpb.SyncPublishStatusResponse{
		OperatorId:    updated.ID,            // 分站ID
		OperatorCode:  updated.Code,          // 分站业务编码
		PublishStatus: updated.PublishStatus, // 当前发布状态
	}, nil
}
