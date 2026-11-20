package manage

import (
	"context"

	"oa.98ent.com/p9/common/i18n"
	"oa.98ent.com/p9/operator-activity/pb/operatoractivityrpc/activitypb"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetEnumParamsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetEnumParamsLogic 活动枚举参数
func NewGetEnumParamsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetEnumParamsLogic {
	return &GetEnumParamsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetEnumParamsLogic) GetEnumParams() (resp *types.GetActivityEnumParamsResp, err error) {
	rpcResp, err := l.svcCtx.ActivityRpc.GetEnumParams(l.ctx, &activitypb.GetActivityEnumParamsReq{})
	if err != nil {
		return nil, err
	}
	// 获取分站信息
	opInfo, err := l.svcCtx.OperatorBaseRpc.GetOperatorInfo(l.ctx, nil)
	if err != nil {
		logx.Errorw("get operator info failed", logx.Field("error", err))
		return nil, err
	}

	return &types.GetActivityEnumParamsResp{
		Type:              l.toActivityEnumItems(rpcResp.GetType()),
		Status:            l.toActivityEnumItems(rpcResp.GetStatus()),
		ApplyMode:         l.toActivityEnumItems(rpcResp.GetApplyMode()),
		CycleMode:         l.toActivityEnumItems(rpcResp.GetCycleMode()),
		PayoutMode:        l.toActivityEnumItems(rpcResp.GetPayoutMode()),
		PayoutTiming:      l.toActivityEnumItems(rpcResp.GetPayoutTiming()),
		BonusMode:         l.toActivityEnumItems(rpcResp.GetBonusMode()),
		BetClaimMode:      l.toActivityEnumItems(rpcResp.GetBetClaimMode()),
		DepositRewardType: l.toActivityEnumItems(rpcResp.GetDepositRewardType()),
		WheelTriggerMode:  l.toActivityEnumItems(rpcResp.GetWheelTriggerMode()),
		WageringBase:      l.toActivityEnumItems(rpcResp.GetWageringBase()),
		ClientType:        l.toActivityEnumItems(rpcResp.GetClientType()),
		ApplyLimitType:    l.toActivityEnumItems(rpcResp.GetApplyLimitType()),
		ClaimLimitType:    l.toActivityEnumItems(rpcResp.GetClaimLimitType()),
		Section:           l.toActivityEnumItems(rpcResp.GetSection()),
		Currency:          opInfo.SettlementCurrencyCode,
		Timezone:          opInfo.TimezoneCode,
	}, nil
}

func (l *GetEnumParamsLogic) toActivityEnumItems(list []*activitypb.ActivityEnumItem) []types.ActivityEnumItem {
	if len(list) == 0 {
		return []types.ActivityEnumItem{}
	}
	out := make([]types.ActivityEnumItem, 0, len(list))
	for _, item := range list {
		if item == nil {
			continue
		}
		out = append(out, types.ActivityEnumItem{
			Key:   item.GetKey(),
			Value: item.GetValue(),
			Label: i18n.TG(l.ctx, i18n.CodeOperator, i18n.GroupFront, item.GetKey()),
		})
	}
	return out
}
