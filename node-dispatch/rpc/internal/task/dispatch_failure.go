package task

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/node-dispatch/rpc/ent/dispatchtask"
)

// markDispatchFailed 标记任务下发失败
func (s *Service) markDispatchFailed(ctx context.Context, taskID int64, errorMessage string) {
	// 只有待执行任务才能标记为下发失败
	_, err := s.db.DispatchTask.
		Update().
		Where(
			dispatchtask.IDEQ(taskID),            // 任务ID
			dispatchtask.StatusEQ(StatusPending), // 任务状态: 1待执行
		).
		SetStatus(StatusFailed).       // 任务状态: 4失败
		SetErrorMessage(errorMessage). // 任务执行失败原因
		SetFinishedAt(time.Now()).     // 执行结束时间
		Save(ctx)
	if err != nil {
		logx.WithContext(ctx).Errorw(
			"更新调度任务下发失败状态失败",
			logx.Field("task_id", taskID),
			logx.Field("error", err.Error()),
		)
	}
}
