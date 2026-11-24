// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package manage

import (
	"context"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-activity/pb/operatoractivityrpc/taskpb"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTaskGlobalConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取任务全局配置（包含当前生效和待生效配置）
func NewListTaskGlobalConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTaskGlobalConfigLogic {
	return &ListTaskGlobalConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListTaskGlobalConfigLogic) ListTaskGlobalConfig(req *types.ListTaskGlobalConfigReq) (resp *types.ListTaskGlobalConfigResp, err error) {
	l.Infof("starting to list task global config for operator: %s", req.OperatorCode)

	// 验证操作员权限
	claims, ok := ctxdata.OperatorClaimsFromCtx(l.ctx)
	if !ok {
		return nil, xerr.Unauthorized("unauthorized")
	}

	// 验证操作员代码一致
	if req.OperatorCode != claims.OperatorCode {
		l.Infof("operator code mismatch: request=%s, claims=%s", req.OperatorCode, claims.OperatorCode)
		return nil, xerr.Unauthorized("unauthorized")
	}

	// 调用 RPC 获取任务全局配置
	rpcResp, err := l.svcCtx.TaskServiceRpc.TaskGlobalConfigList(l.ctx, &taskpb.ListTaskGlobalConfigReq{
		OperatorCode: req.OperatorCode,
	})
	if err != nil {
		l.Errorf("failed to list task global config: %v", err)
		return nil, err
	}

	// 转换 proto 消息为 API 类型
	var list []types.TaskGlobalConfig
	for _, config := range rpcResp.List {
		list = append(list, types.TaskGlobalConfig{
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
		})
	}

	resp = &types.ListTaskGlobalConfigResp{
		List: list,
	}

	l.Infof("task global config list retrieved successfully: operator_code=%s, count=%d", req.OperatorCode, len(list))
	return resp, nil
}
