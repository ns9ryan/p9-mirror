package taskservicelogic

import (
	"context"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-activity/ent"
	"oa.98ent.com/p9/operator-activity/ent/taskconfig"
	"oa.98ent.com/p9/operator-activity/internal/svc"
	"oa.98ent.com/p9/operator-activity/pb/operatoractivityrpc/taskpb"
	"oa.98ent.com/p9/operator-activity/pkg/i18n"

	"github.com/zeromicro/go-zero/core/logx"
)

type TaskConfigListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTaskConfigListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TaskConfigListLogic {
	return &TaskConfigListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// group: task_config 任务配置列表（分页查询，支持排序）
func (l *TaskConfigListLogic) TaskConfigList(in *taskpb.ListTaskConfigReq) (*taskpb.ListTaskConfigResp, error) {
	// 验证操作员权限
	claims, ok := ctxdata.OperatorClaimsFromCtx(l.ctx)
	if !ok {
		return nil, xerr.RpcErr(xerr.Unauthorized(i18n.Unauthorized))
	}

	// 验证操作员代码一致
	if in.OperatorCode != claims.OperatorCode {
		l.Infof("starting task config list for operator: %s, operator code in claims: %v", in.OperatorCode, claims.OperatorCode)
		return nil, xerr.RpcErr(xerr.Unauthorized(i18n.Unauthorized))
	}

	// 设置分页参数
	page := in.GetPage()
	if page <= 0 {
		page = 1
	}
	pageSize := in.GetPageSize()
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// 构建查询条件
	query := l.svcCtx.DB.TaskConfig.Query().
		Where(taskconfig.OperatorCode(in.OperatorCode))

	// 添加分类过滤
	if in.Category != nil && *in.Category != "" {
		query = query.Where(taskconfig.Category(*in.Category))
	}

	// 添加状态过滤
	if in.Status != nil {
		query = query.Where(taskconfig.Status(int(*in.Status)))
	}

	// 获取总数
	total, err := query.Count(l.ctx)
	if err != nil {
		l.Errorf("failed to count task config: %v", err)
		return nil, xerr.RpcErr(err)
	}

	// 按 sort 字段升序排序
	configs, err := query.
		Order(ent.Asc(taskconfig.FieldSort)).
		Offset(int((page - 1) * pageSize)).
		Limit(int(pageSize)).
		All(l.ctx)
	if err != nil {
		l.Errorf("failed to query task config: %v", err)
		return nil, xerr.RpcErr(err)
	}

	// 转换为 proto 消息
	list := make([]*taskpb.TaskConfig, 0, len(configs))
	for _, config := range configs {
		list = append(list, &taskpb.TaskConfig{
			Id:               config.ID,
			OperatorCode:     config.OperatorCode,
			TaskCode:         config.TaskCode,
			Category:         config.Category,
			Name:             config.Name,
			NameI18N:         config.NameI18n,
			Description:      config.Description,
			DescriptionI18N:  config.DescriptionI18n,
			Timezone:         config.Timezone,
			Status:           int32(config.Status),
			CompletionAmount: int32(config.CompletionAmount),
			RewardAmount:     config.RewardAmount,
			PayoutMode:       config.PayoutMode,
			Sort:             config.Sort,
			CreatedAt:        config.CreatedAt.UnixMilli(),
			UpdatedAt:        config.UpdatedAt.UnixMilli(),
		})
	}

	return &taskpb.ListTaskConfigResp{
		List:     list,
		Total:    int64(total),
		Page:     page,
		PageSize: pageSize,
	}, nil
}
