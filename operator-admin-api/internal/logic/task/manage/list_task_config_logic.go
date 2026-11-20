package manage

import (
	"context"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/operator-activity/pb/operatoractivityrpc/taskpb"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTaskConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewListTaskConfigLogic 任务配置列表
func NewListTaskConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTaskConfigLogic {
	return &ListTaskConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListTaskConfigLogic) ListTaskConfig(req *types.ListTaskConfigReq) (resp *types.ListTaskConfigResp, err error) {
	// 获取操作员信息
	claims := ctxdata.ClaimsFromCtx(l.ctx)
	operatorCode := claims.OperatorCode

	// 构建 RPC 请求
	rpcReq := &taskpb.ListTaskConfigReq{
		OperatorCode: operatorCode,
		Category:     req.Category,
		Status:       req.Status,
		Page:         &req.Page,
		PageSize:     &req.PageSize,
	}

	// 调用 RPC
	rpcResp, err := l.svcCtx.TaskServiceRpc.TaskConfigList(l.ctx, rpcReq)
	if err != nil {
		return nil, err
	}

	// 转换响应
	list := make([]types.TaskConfig, 0, len(rpcResp.GetList()))
	for _, item := range rpcResp.GetList() {
		list = append(list, types.TaskConfig{
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
		})
	}

	return &types.ListTaskConfigResp{
		Total:    rpcResp.GetTotal(),
		List:     list,
		Page:     rpcResp.GetPage(),
		PageSize: rpcResp.GetPageSize(),
	}, nil
}
