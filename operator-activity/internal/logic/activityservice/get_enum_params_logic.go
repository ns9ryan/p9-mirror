package activityservicelogic

import (
	"context"

	"oa.98ent.com/p9/operator-activity/internal/svc"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"
	"oa.98ent.com/p9/operator-activity/pb/operatoractivityrpc/activitypb"
	"oa.98ent.com/p9/operator-activity/pkg/i18n"

	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
)

type GetEnumParamsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetEnumParamsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetEnumParamsLogic {
	return &GetEnumParamsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetEnumParams 活动枚举参数（多语言 label）
func (l *GetEnumParamsLogic) GetEnumParams(in *activitypb.GetActivityEnumParamsReq) (*activitypb.GetActivityEnumParamsResp, error) {
	if _, ok := ctxdata.OperatorClaimsFromCtx(l.ctx); !ok {
		return nil, xerr.RpcErr(xerr.Unauthorized(i18n.Unauthorized))
	}

	params := types.GetEnumParams()
	return &activitypb.GetActivityEnumParamsResp{
		Type:              toActivityEnumItems(params.Type),
		Status:            toActivityEnumItems(params.Status),
		ApplyMode:         toActivityEnumItems(params.ApplyMode),
		CycleMode:         toActivityEnumItems(params.CycleMode),
		PayoutMode:        toActivityEnumItems(params.PayoutMode),
		PayoutTiming:      toActivityEnumItems(params.PayoutTiming),
		BonusMode:         toActivityEnumItems(params.BonusMode),
		BetClaimMode:      toActivityEnumItems(params.BetClaimMode),
		DepositRewardType: toActivityEnumItems(params.DepositRewardType),
		WheelTriggerMode:  toActivityEnumItems(params.WheelTriggerMode),
		WageringBase:      toActivityEnumItems(params.WageringBase),
		ClientType:        toActivityEnumItems(params.ClientType),
		ApplyLimitType:    toActivityEnumItems(params.ApplyLimitType),
		ClaimLimitType:    toActivityEnumItems(params.ClaimLimitType),
		Section:           toActivityEnumItems(params.Section),
	}, nil
}

// toActivityEnumItems 转换活动枚举参数为 protobuf 格式
func toActivityEnumItems(list []types.EnumItem) []*activitypb.ActivityEnumItem {
	if len(list) == 0 {
		return nil
	}
	out := make([]*activitypb.ActivityEnumItem, 0, len(list))
	for _, item := range list {
		out = append(out, &activitypb.ActivityEnumItem{
			Key:   item.Key,
			Value: item.Value,
		})
	}
	return out
}
