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

	// 只允许分站创建发布任务
	if taskData.Target != dispatchTargetOperatorBase || taskData.TaskType != dispatchTaskTypeCreateOperator {
		return "", "", xerr.RpcErr(xerr.BadRequest(i18nkey.ConstraintError))
	}

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
		return nil, false, err
	}

	// 任务必须属于分站当前发布轮次
	if current.PublishRequestNo == nil || strings.TrimSpace(*current.PublishRequestNo) != requestNo {
		return current, false, nil
	}

	// 只允许应用最终任务状态
	if taskData.Status != 3 && taskData.Status != 4 {
		return nil, false, xerr.RpcErr(xerr.BadRequest(i18nkey.ConstraintError))
	}

	// 最终任务必须存在结束时间
	if taskData.FinishedAt == nil || *taskData.FinishedAt <= 0 {
		return nil, false, xerr.RpcErr(xerr.InternalServerError(i18nkey.InternalError))
	}

	// 相同任务结果已经处理时直接返回
	if current.PublishStatus == taskData.Status {
		if current.PublishTaskNo != nil && strings.TrimSpace(*current.PublishTaskNo) == taskNo {
			return current, true, nil
		}

		return nil, false, xerr.RpcErr(xerr.BadRequest(i18nkey.ConstraintError))
	}

	// 只有发布中的分站允许进入最终状态
	if current.PublishStatus != 2 {
		return nil, false, xerr.RpcErr(xerr.BadRequest(i18nkey.ConstraintError))
	}

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

		return current, true, nil
	}

	// 获取并发更新后的最新分站状态
	latest, err := svcCtx.DB.Operator.Get(ctx, current.ID)
	if err != nil {
		return nil, false, enterror.Handle(logger, err)
	}

	// 当前已经进入其他发布轮次时忽略本次结果
	if latest.PublishRequestNo == nil || strings.TrimSpace(*latest.PublishRequestNo) != requestNo {
		return latest, false, nil
	}

	// 相同任务结果已被其他流程先一步处理
	if latest.PublishStatus == taskData.Status &&
		latest.PublishTaskNo != nil &&
		strings.TrimSpace(*latest.PublishTaskNo) == taskNo {
		return latest, true, nil
	}

	return nil, false, xerr.RpcErr(xerr.BadRequest(i18nkey.ConstraintError))
}
