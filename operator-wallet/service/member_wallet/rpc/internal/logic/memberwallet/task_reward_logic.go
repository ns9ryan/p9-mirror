package memberwalletlogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/enum"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/svc"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/wallet"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc/rewardpb"
)

// TaskRewardLogic 处理会员任务奖励入账
type TaskRewardLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewTaskRewardLogic 创建任务奖励入账逻辑
func NewTaskRewardLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TaskRewardLogic {
	return &TaskRewardLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// TaskReward 发放会员任务奖励
func (l *TaskRewardLogic) TaskReward(in *rewardpb.TaskRewardRequest) (*rewardpb.TaskRewardResponse, error) {
	// 校验请求参数
	params, err := parseRewardOperationParams(
		in.GetMemberGlobalId(), // 会员全局唯一标识
		in.GetOperationId(),    // 钱包操作唯一标识, UUID
		in.GetTaskOrderNo(),    // 任务奖励订单号
		in.GetCurrencyCode(),   // 钱包结算币种编码
		in.GetAmount(),
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

	// 提交任务奖励入账操作
	_, err = l.svcCtx.WalletEngine.Submit(
		l.ctx,
		walletGlobalID,
		wallet.RecordParams{
			// 操作与业务来源
			OperationID:    params.operationID,            // 钱包操作唯一标识
			BusinessType:   enum.BusinessTypeTask,         // 一级业务类型: 任务
			BusinessAction: enum.BusinessActionTaskReward, // 具体资金操作类型: 任务奖励入账
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
		return nil, handleWalletSubmitError(err, "task reward", params.operationID, l.Logger)
	}

	return &rewardpb.TaskRewardResponse{}, nil
}
