package manage

import (
	"context"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/operator-activity/pb/operatoractivityrpc/taskpb"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateTaskConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewUpdateTaskConfigLogic 更新任务配置
func NewUpdateTaskConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTaskConfigLogic {
	return &UpdateTaskConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateTaskConfigLogic) UpdateTaskConfig(req *types.UpdateTaskConfigReq) (resp *types.UpdateTaskConfigResp, err error) {
	// 获取操作员信息
	claims := ctxdata.ClaimsFromCtx(l.ctx)
	operatorCode := claims.OperatorCode

	// 构建 RPC 请求
	rpcReq := &taskpb.UpdateTaskConfigReq{
		Id:               req.Id,
		OperatorCode:     operatorCode,
		Timezone:         req.Timezone,
		Status:           req.Status,
		CompletionAmount: req.CompletionAmount,
		RewardAmount:     req.RewardAmount,
		PayoutMode:       req.PayoutMode,
		Sort:             req.Sort,
	}

	// 调用 RPC
	rpcResp, err := l.svcCtx.TaskServiceRpc.TaskConfigUpdate(l.ctx, rpcReq)
	if err != nil {
		return nil, err
	}
	item := rpcResp.GetTaskConfig()
	// 转换响应
	return &types.UpdateTaskConfigResp{
		TaskConfig: types.TaskConfig{
			Id:               item.Id,
			OperatorCode:     item.OperatorCode,
			TaskCode:         item.TaskCode,
			Category:         item.Category,
			Name:             item.Name,
			NameI18n:         item.NameI18N,
			Description:      item.Description,
			DescriptionI18n:  item.DescriptionI18N,
			Timezone:         item.Timezone,
			Status:           item.Status,
			CompletionAmount: item.CompletionAmount,
			RewardAmount:     item.RewardAmount,
			PayoutMode:       item.PayoutMode,
			Sort:             item.Sort,
			CreatedAt:        item.CreatedAt,
			UpdatedAt:        item.UpdatedAt,
		},
	}, nil
}
