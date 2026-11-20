package operatorservicelogic

import (
	"context"
	"fmt"
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

// trimOptionalString 整理可选字符串
func trimOptionalString(value *string) *string {
	if value == nil {
		return nil
	}

	return new(strings.TrimSpace(*value))
}

// toOperatorInfo 转换分站信息
func toOperatorInfo(data *ent.Operator) *operatorpb.OperatorInfo {
	// 转换首次发布时间
	var publishedAt *int64
	if data.PublishedAt != nil {
		value := data.PublishedAt.UnixMilli()
		publishedAt = &value
	}

	return &operatorpb.OperatorInfo{
		Id:                     data.ID,                     // 分站ID
		Code:                   data.Code,                   // 分站业务编码
		Name:                   data.Name,                   // 分站名称
		TimezoneCode:           data.TimezoneCode,           // 时区编码
		SettlementCurrencyCode: data.SettlementCurrencyCode, // 结算币种编码
		CreationStatus:         data.CreationStatus,         // 创建状态: 1草稿, 2已完成
		PublishStatus:          data.PublishStatus,          // 发布状态: 1未发布, 2发布中, 3已发布, 4发布失败
		Status:                 data.Status,                 // 分站状态: 1正常, 2暂停, 3关闭
		Remark:                 data.Remark,                 // 内部备注
		PublishedAt:            publishedAt,                 // 首次发布成功时间, Unix毫秒时间戳
		CreatedAt:              data.CreatedAt.UnixMilli(),  // 创建时间, Unix毫秒时间戳
		UpdatedAt:              data.UpdatedAt.UnixMilli(),  // 更新时间, Unix毫秒时间戳
	}
}

// validatePublishTask 校验分站发布调度任务
func validatePublishTask(
	taskData *nodedispatchdispatchpb.TaskInfo,
	expectedTaskNo string,
	expectedRequestNo string,
) (string, string, error) {
	// 调度任务信息不能为空
	if taskData == nil {
		return "", "", xerr.RpcErr(xerr.InternalServerError(i18nkey.InternalError))
	}

	// 整理任务编号
	taskNo := strings.TrimSpace(taskData.TaskNo)
	requestNo := strings.TrimSpace(taskData.RequestNo)

	// 任务编号必须完整
	if taskNo == "" || requestNo == "" {
		return "", "", xerr.RpcErr(xerr.InternalServerError(i18nkey.InternalError))
	}

	// 校验指定的调度任务编号
	expectedTaskNo = strings.TrimSpace(expectedTaskNo)
	if expectedTaskNo != "" && taskNo != expectedTaskNo {
		return "", "", xerr.RpcErr(xerr.InternalServerError(i18nkey.InternalError))
	}

	// 校验指定的发布请求编号
	expectedRequestNo = strings.TrimSpace(expectedRequestNo)
	if expectedRequestNo != "" && requestNo != expectedRequestNo {
		return "", "", xerr.RpcErr(xerr.InternalServerError(i18nkey.InternalError))
	}

	// 打印任务详细信息用于调试
	logx.Infof(
		"validatePublishTask: task_no=%s, request_no=%s, target=%s, task_type=%s, expected_target=%s, expected_request_no=%s",
		taskNo, requestNo, taskData.Target, taskData.TaskType, expectedTaskNo, expectedRequestNo,
	)

	// 只允许分站相关的发布任务
	if taskData.Target != dispatchTargetOperatorBase {
		logx.Errorf(
			"validatePublishTask: target验证失败，期望target=%s，实际target=%s，task_no=%s",
			dispatchTargetOperatorBase, taskData.Target, taskNo,
		)
		return "", "", xerr.RpcErr(xerr.BadRequest(i18nkey.ConstraintError))
	}

	// 允许的任务类型
	allowedTaskTypes := map[string]bool{
		dispatchTaskTypeCreateOperator: true, // CREATE_OPERATOR
		"PUBLISH_BASIC_ALLOCATION":     true, // 发布基础资源分配
		"PUBLISH_DOMAIN":               true, // 发布域名
	}

	// 校验任务类型是否被允许
	if !allowedTaskTypes[taskData.TaskType] {
		logx.Errorf(
			"validatePublishTask: task_type验证失败，task_type=%s不在允许列表中，task_no=%s，允许列表=%v",
			taskData.TaskType, taskNo, allowedTaskTypes,
		)
		return "", "", xerr.RpcErr(xerr.BadRequest(i18nkey.ConstraintError))
	}

	logx.Infof("validatePublishTask: 验证通过，task_no=%s, request_no=%s", taskNo, requestNo)

	return taskNo, requestNo, nil
}

// applyFinalPublishTask 应用分站最终发布任务结果
func applyFinalPublishTask(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	logger logx.Logger,
	current *ent.Operator,
	taskData *nodedispatchdispatchpb.TaskInfo,
) (*ent.Operator, bool, error) {
	// 校验发布任务
	taskNo, requestNo, err := validatePublishTask(taskData, "", "")
	if err != nil {
		logger.Errorw("发布任务校验失败", logx.Field("error", err.Error()))
		return nil, false, err
	}

	logger.Debugw(
		"开始应用最终发布任务结果",
		logx.Field("task_no", taskNo),
		logx.Field("request_no", requestNo),
		logx.Field("operator_id", current.ID),
		logx.Field("operator_code", current.Code),
		logx.Field("current_publish_status", current.PublishStatus),
		logx.Field("task_status", taskData.Status),
	)

	// 任务必须属于分站当前发布轮次
	if current.PublishRequestNo == nil || strings.TrimSpace(*current.PublishRequestNo) != requestNo {
		logger.Infow(
			"任务不属于分站当前发布轮次，忽略处理",
			logx.Field("task_no", taskNo),
			logx.Field("task_request_no", requestNo),
			logx.Field("current_request_no", current.PublishRequestNo),
		)
		return current, false, nil
	}

	// 只允许应用最终任务状态
	if taskData.Status != 3 && taskData.Status != 4 {
		logger.Errorw(
			"任务状态不是最终状态",
			logx.Field("task_no", taskNo),
			logx.Field("task_status", taskData.Status),
		)
		return nil, false, xerr.RpcErr(xerr.BadRequest(i18nkey.ConstraintError))
	}

	// 最终任务必须存在结束时间
	if taskData.FinishedAt == nil || *taskData.FinishedAt <= 0 {
		logger.Errorw(
			"任务结束时间不有效",
			logx.Field("task_no", taskNo),
			logx.Field("finished_at", taskData.FinishedAt),
		)
		return nil, false, xerr.RpcErr(xerr.InternalServerError(i18nkey.InternalError))
	}

	// 相同任务结果已经处理时直接返回
	if current.PublishStatus == taskData.Status {
		logger.Debugw(
			"分站已处于目标发布状态",
			logx.Field("task_no", taskNo),
			logx.Field("publish_status", current.PublishStatus),
		)

		if current.PublishTaskNo != nil && strings.TrimSpace(*current.PublishTaskNo) == taskNo {
			logger.Infow(
				"相同任务结果已处理，直接返回",
				logx.Field("task_no", taskNo),
			)
			return current, true, nil
		}

		logger.Errorw(
			"分站状态与任务状态相同，但任务编号不匹配",
			logx.Field("task_no", taskNo),
			logx.Field("current_task_no", current.PublishTaskNo),
		)
		return nil, false, xerr.RpcErr(xerr.BadRequest(i18nkey.ConstraintError))
	}

	// 只有发布中的分站允许进入最终状态
	if current.PublishStatus != 2 {
		logger.Errorw(
			"分站发布状态不是发布中，无法进入最终状态",
			logx.Field("task_no", taskNo),
			logx.Field("current_publish_status", current.PublishStatus),
		)
		return nil, false, xerr.RpcErr(xerr.BadRequest(i18nkey.ConstraintError))
	}

	logger.Infow(
		"准备更新分站发布状态",
		logx.Field("task_no", taskNo),
		logx.Field("operator_code", current.Code),
		logx.Field("new_status", taskData.Status),
		logx.Field("status_description", statusDescription(taskData.Status)),
	)

	// 创建最终发布状态更新
	update := svcCtx.DB.Operator.
		Update().
		Where(
			operator.IDEQ(current.ID),
			operator.PublishRequestNoEQ(requestNo),
			operator.PublishStatusEQ(2),
		).
		SetPublishTaskNo(taskNo).         // 当前发布任务编号
		SetPublishStatus(taskData.Status) // 发布状态: 3已发布, 4发布失败

	// 首次发布成功时记录任务实际完成时间
	if taskData.Status == 3 && current.PublishedAt == nil {
		update.SetPublishedAt(time.UnixMilli(*taskData.FinishedAt))
	}

	// 保存最终发布状态
	affected, err := update.Save(ctx)
	if err != nil {
		logger.Errorw(
			"更新分站发布状态失败",
			logx.Field("task_no", taskNo),
			logx.Field("operator_code", current.Code),
			logx.Field("error", err.Error()),
		)
		return nil, false, enterror.Handle(logger, err)
	}

	// 更新成功后同步当前内存数据
	if affected == 1 {
		current.PublishTaskNo = new(taskNo)
		current.PublishStatus = taskData.Status

		if taskData.Status == 3 && current.PublishedAt == nil {
			finishedAt := time.UnixMilli(*taskData.FinishedAt)
			current.PublishedAt = &finishedAt
		}

		logger.Infow(
			"分站发布状态更新成功（第一次写入）",
			logx.Field("task_no", taskNo),
			logx.Field("operator_code", current.Code),
			logx.Field("new_status", current.PublishStatus),
		)
		return current, true, nil
	}

	logger.Debugw(
		"并发更新检测，重新查询分站最新状态",
		logx.Field("task_no", taskNo),
		logx.Field("affected_rows", affected),
	)

	// 获取并发更新后的最新分站状态
	latest, err := svcCtx.DB.Operator.Get(ctx, current.ID)
	if err != nil {
		logger.Errorw(
			"查询分站最新状态失败",
			logx.Field("task_no", taskNo),
			logx.Field("error", err.Error()),
		)
		return nil, false, enterror.Handle(logger, err)
	}

	// 当前已经进入其他发布轮次时忽略本次结果
	if latest.PublishRequestNo == nil || strings.TrimSpace(*latest.PublishRequestNo) != requestNo {
		logger.Infow(
			"分站已进入新的发布轮次，忽略本次结果",
			logx.Field("task_no", taskNo),
			logx.Field("task_request_no", requestNo),
			logx.Field("latest_request_no", latest.PublishRequestNo),
		)
		return latest, false, nil
	}

	// 相同任务结果已被其他流程先一步处理
	if latest.PublishStatus == taskData.Status &&
		latest.PublishTaskNo != nil &&
		strings.TrimSpace(*latest.PublishTaskNo) == taskNo {
		logger.Infow(
			"相同任务结果已被其他流程处理",
			logx.Field("task_no", taskNo),
			logx.Field("operator_code", latest.Code),
		)
		return latest, true, nil
	}

	logger.Errorw(
		"分站发布状态冲突，无法处理",
		logx.Field("task_no", taskNo),
		logx.Field("operator_code", latest.Code),
		logx.Field("latest_status", latest.PublishStatus),
		logx.Field("task_status", taskData.Status),
	)
	return nil, false, xerr.RpcErr(xerr.BadRequest(i18nkey.ConstraintError))
}

// statusDescription 发布状态描述
func statusDescription(status int64) string {
	switch status {
	case 1:
		return "未发布"
	case 2:
		return "发布中"
	case 3:
		return "已发布"
	case 4:
		return "发布失败"
	default:
		return fmt.Sprintf("unknown(%d)", status)
	}
}
