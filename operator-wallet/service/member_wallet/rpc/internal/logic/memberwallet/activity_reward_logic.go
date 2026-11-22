package memberwalletlogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/enum"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/svc"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/wallet"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc/rewardpb"
)

// ActivityRewardLogic 处理会员活动奖励入账
type ActivityRewardLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewActivityRewardLogic 创建活动奖励入账逻辑
func NewActivityRewardLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ActivityRewardLogic {
	return &ActivityRewardLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ActivityReward 发放会员活动奖励
func (l *ActivityRewardLogic) ActivityReward(in *rewardpb.ActivityRewardRequest) (*rewardpb.ActivityRewardResponse, error) {
	// 校验请求参数
	params, err := parseRewardOperationParams(
		in.GetMemberGlobalId(),  // 会员全局唯一标识
		in.GetOperationId(),     // 钱包操作唯一标识, UUID
		in.GetActivityOrderNo(), // 活动奖励订单号
		in.GetCurrencyCode(),    // 钱包结算币种编码
		in.GetAmount(),          // 奖励金额, 币种最小单位
	)
	if err != nil {
		return nil, err
	}

	// 确保会员钱包已加载并注册到 Engine
	walletGlobalID, err := ensureWalletActorForRPC(
		l.ctx,
		l.svcCtx,
		walletActorParams{
			memberGlobalID: params.memberGlobalID, // 会员全局唯一标识
			currencyCode:   params.currencyCode,   // 请求币种编码
		},
		params.operationID, // 钱包操作唯一标识
		l.Logger,
	)
	if err != nil {
		return nil, err
	}

	// 提交活动奖励入账操作
	_, err = l.svcCtx.WalletEngine.Submit(
		l.ctx,
		walletGlobalID,
		wallet.RecordParams{
			// 操作与业务来源
			OperationID:    params.operationID,                // 钱包操作唯一标识
			BusinessType:   enum.BusinessTypeActivity,         // 一级业务类型: 活动
			BusinessAction: enum.BusinessActionActivityReward, // 具体资金操作类型: 活动奖励入账
			BusinessNo:     params.businessNo,                 // 奖励业务订单号

			// 增加可用余额
			Changes: []wallet.BalanceChange{
				{
					BalanceType:  enum.BalanceTypeAvailable, // 可用余额
					ChangeAmount: params.amount,             // 增加
				},
			},
		},
	)
	if err != nil {
		return nil, handleWalletSubmitError(err, "activity reward", params.operationID, l.Logger)
	}

	return &rewardpb.ActivityRewardResponse{}, nil
}
