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
	// ==================== 当前发布状态校验 ====================

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

	// ==================== 查询调度任务 ====================

	// 根据当前发布请求编号获取实际任务状态
	taskResult, err := l.svcCtx.NodeDispatchDispatchRpc.GetTask(
		l.ctx,
		&nodedispatchdispatchpb.GetTaskRequest{
			RequestNo: new(requestNo), // 当前发布请求编号
		},
	)
	if err != nil {
		// 未找到任务时保持发布中状态
		if status.Code(err) == codes.NotFound {
			l.Logger.Infow(
				"同步分站发布状态未找到调度任务",
				logx.Field("operator_id", current.ID),
				logx.Field("operator_code", current.Code),
				logx.Field("request_no", requestNo),
			)
			return nil, err
		}

		l.Logger.Errorw(
			"同步分站发布状态获取调度任务失败",
			logx.Field("operator_id", current.ID),
			logx.Field("operator_code", current.Code),
			logx.Field("request_no", requestNo),
			logx.Field("error", err.Error()),
		)
		return nil, err
	}

	// ==================== 校验调度任务 ====================

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
	_, _, err = validatePublishTask(
		taskResult.Task,
		"",
		requestNo,
	)
	if err != nil {
		return nil, err
	}

	// ==================== 同步发布状态 ====================

	// 待执行或执行中时保持当前发布状态
	if taskResult.Task.Status == 1 || taskResult.Task.Status == 2 {
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
