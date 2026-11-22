package memberwalletlogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/enum"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/svc"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/wallet"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc/rewardpb"
)

// VipRewardLogic 处理会员 VIP 奖励入账
type VipRewardLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewVipRewardLogic 创建 VIP 奖励入账逻辑
func NewVipRewardLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VipRewardLogic {
	return &VipRewardLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// VipReward 发放会员 VIP 奖励
func (l *VipRewardLogic) VipReward(in *rewardpb.VipRewardRequest) (*rewardpb.VipRewardResponse, error) {
	// 校验请求参数
	params, err := parseRewardOperationParams(
		in.GetMemberGlobalId(),   // 会员全局唯一标识
		in.GetOperationId(),      // 钱包操作唯一标识, UUID
		in.GetVipRewardOrderNo(), // VIP 奖励订单号
		in.GetCurrencyCode(),     // 钱包结算币种编码
		in.GetAmount(),           // 奖励金额, 币种最小单位
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

	// 提交 VIP 奖励入账操作
	_, err = l.svcCtx.WalletEngine.Submit(
		l.ctx,
		walletGlobalID,
		wallet.RecordParams{
			// 操作与业务来源
			OperationID:    params.operationID,            // 钱包操作唯一标识
			BusinessType:   enum.BusinessTypeSystemReward, // 一级业务类型: 系统奖励
			BusinessAction: enum.BusinessActionVipReward,  // 具体资金操作类型: VIP 奖励入账
			BusinessNo:     params.businessNo,             // 奖励业务订单号

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
		return nil, handleWalletSubmitError(err, "vip reward", params.operationID, l.Logger)
	}

	return &rewardpb.VipRewardResponse{}, nil
}
