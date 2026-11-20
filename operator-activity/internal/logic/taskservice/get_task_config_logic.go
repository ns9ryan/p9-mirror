package taskservicelogic

import (
	"context"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-activity/ent"
	"oa.98ent.com/p9/operator-activity/ent/taskconfig"
	"oa.98ent.com/p9/operator-activity/pkg/i18n"

	"oa.98ent.com/p9/operator-activity/internal/svc"
	"oa.98ent.com/p9/operator-activity/pb/operatoractivityrpc/taskpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTaskConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTaskConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTaskConfigLogic {
	return &GetTaskConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// group: task_config 获取单个任务配置
func (l *GetTaskConfigLogic) GetTaskConfig(in *taskpb.GetTaskConfigReq) (*taskpb.GetTaskConfigResp, error) {
	// 验证操作员权限
	claims, ok := ctxdata.OperatorClaimsFromCtx(l.ctx)
	if !ok {
		return nil, xerr.RpcErr(xerr.Unauthorized(i18n.Unauthorized))
	}

	// 验证操作员代码一致
	if in.OperatorCode != claims.OperatorCode {
		l.Infof("starting task config get for operator: %s, operator code in claims: %v", in.OperatorCode, claims.OperatorCode)
		return nil, xerr.RpcErr(xerr.Unauthorized(i18n.Unauthorized))
	}

	// 查询任务配置
	config, err := l.svcCtx.DB.TaskConfig.Query().
		Where(
			taskconfig.OperatorCode(in.OperatorCode),
			taskconfig.TaskCode(in.TaskCode),
		).
		Only(l.ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			l.Infof("task config not found: operator_code=%s, task_code=%s", in.OperatorCode, in.TaskCode)
			return nil, xerr.RpcErr(xerr.NotFound(i18n.NotFound))
		}
		l.Errorf("failed to query task config: %v", err)
		return nil, xerr.RpcErr(err)
	}

	// 转换为 proto 消息
	resp := &taskpb.TaskConfig{
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
	}

	return &taskpb.GetTaskConfigResp{
		TaskConfig: resp,
	}, nil
}
