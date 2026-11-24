package manage

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/operator-activity/pb/operatoractivityrpc/taskpb"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
)

type UpdateTaskGlobalConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewUpdateTaskGlobalConfigLogic 更新任务全局配置
func NewUpdateTaskGlobalConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTaskGlobalConfigLogic {
	return &UpdateTaskGlobalConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateTaskGlobalConfigLogic) UpdateTaskGlobalConfig(req *types.UpdateTaskGlobalConfigReq) (resp *types.UpdateTaskGlobalConfigResp, err error) {
	// 获取操作员信息
	claims := ctxdata.ClaimsFromCtx(l.ctx)
	operatorCode := claims.OperatorCode

	l.Infof("UpdateTaskGlobalConfig for operator: %s", operatorCode)

	// 构建 RPC 请求
	rpcReq := &taskpb.UpdateTaskGlobalConfigReq{
		OperatorCode: operatorCode,
	}

	// 处理 AutoLevelIds
	if req.AutoLevelIds != nil {
		rpcReq.AutoLevelIds = []byte(*req.AutoLevelIds)
	}

	// 处理 FixedLevelIds
	if req.FixedLevelIds != nil {
		rpcReq.FixedLevelIds = []byte(*req.FixedLevelIds)
	}

	// 处理 MemberTagIds
	if req.MemberTagIds != nil {
		rpcReq.MemberTagIds = []byte(*req.MemberTagIds)
	}

	// 处理 TurnoverMultiple
	if req.TurnoverMultiple != nil {
		rpcReq.TurnoverMultiple = req.TurnoverMultiple
	}

	// 处理 ClientTypes
	if req.ClientTypes != nil {
		rpcReq.ClientTypes = req.ClientTypes
	}

	// 处理 LimitsApply
	if req.LimitsApply != nil {
		rpcReq.LimitsApply = req.LimitsApply
	}

	// 处理 CostIncluded
	if req.CostIncluded != nil {
		rpcReq.CostIncluded = req.CostIncluded
	}

	// 处理 EffectTime
	if req.EffectTime != nil {
		rpcReq.EffectTime = req.EffectTime
	}

	// 处理 Status
	if req.Status != nil {
		rpcReq.Status = req.Status
	}

	// 调用 RPC
	rpcResp, err := l.svcCtx.TaskServiceRpc.UpdateTaskGlobalConfig(l.ctx, rpcReq)
	if err != nil {
		l.Errorf("failed to update task global config: %v", err)
		return nil, err
	}

	// 转换响应
	config := rpcResp.GetTaskGlobalConfig()
	if config == nil {
		l.Errorf("empty task global config in response")
		return nil, err
	}

	resp = &types.UpdateTaskGlobalConfigResp{
		TaskGlobalConfig: types.TaskGlobalConfig{
			Id:               config.Id,
			OperatorCode:     config.OperatorCode,
			AutoLevelIds:     config.AutoLevelIds,
			FixedLevelIds:    config.FixedLevelIds,
			MemberTagIds:     config.MemberTagIds,
			TurnoverMultiple: config.TurnoverMultiple,
			ClientTypes:      config.ClientTypes,
			LimitsApply:      config.LimitsApply,
			CostIncluded:     config.CostIncluded,
			EffectTime:       config.EffectTime,
			Status:           config.Status,
			CreatedAt:        config.CreatedAt,
			UpdatedAt:        config.UpdatedAt,
			Active:           config.Active,
		},
	}

	l.Infof("UpdateTaskGlobalConfig successfully: operator_code=%s, effect_time=%d", operatorCode, config.EffectTime)
	return resp, nil
}
