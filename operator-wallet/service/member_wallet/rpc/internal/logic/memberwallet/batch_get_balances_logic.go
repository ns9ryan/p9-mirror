package memberwalletlogic

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/svc"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc/managepb"
)

const maxBatchGetBalances = 100

// BatchGetBalancesLogic 处理批量查询会员钱包余额
type BatchGetBalancesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewBatchGetBalancesLogic 创建批量查询会员钱包余额逻辑
func NewBatchGetBalancesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchGetBalancesLogic {
	return &BatchGetBalancesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// BatchGetBalances 批量查询会员钱包余额, 只读取数据库持久化数据
func (l *BatchGetBalancesLogic) BatchGetBalances(in *managepb.BatchGetBalancesRequest) (*managepb.BatchGetBalancesResponse, error) {
	// 校验请求参数
	memberGlobalIDs, err := parseBatchGetBalancesRequest(in)
	if err != nil {
		return nil, err
	}

	// 校验可信厅身份
	if err := validateOperatorCode(l.ctx); err != nil {
		return nil, err
	}

	resp := &managepb.BatchGetBalancesResponse{}
	if len(memberGlobalIDs) == 0 {
		return resp, nil
	}

	// 一次性查询当前厅已有的钱包余额
	balances, err := l.svcCtx.DAO.MemberWallet.ListBalancesByMemberGlobalIDs(l.ctx, memberGlobalIDs)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			return nil, status.Error(codes.Canceled, "批量查询会员钱包余额请求已取消")

		case errors.Is(err, context.DeadlineExceeded):
			return nil, status.Error(codes.DeadlineExceeded, "批量查询会员钱包余额超时")

		default:
			l.Errorf("batch get member wallet balances failed: %v", err)
			return nil, status.Error(codes.Internal, "批量查询会员钱包余额失败")
		}
	}

	// 组装查询结果, 不存在钱包的会员不返回
	resp.Balances = make([]*managepb.MemberWalletBalanceItem, 0, len(balances))
	for _, balance := range balances {
		resp.Balances = append(resp.Balances, &managepb.MemberWalletBalanceItem{
			MemberGlobalId:   balance.MemberGlobalID.String(), // CRM会员全局唯一标识
			Balance:          balance.Balance,                 // 可用余额
			FrozenBalance:    balance.FrozenBalance,           // 冻结余额
			UnsettledBalance: balance.UnsettledBalance,        // 游戏未结算余额
		})
	}

	return resp, nil
}

// parseBatchGetBalancesRequest 校验并解析批量余额查询请求
func parseBatchGetBalancesRequest(in *managepb.BatchGetBalancesRequest) ([]uuid.UUID, error) {
	if in == nil {
		return nil, status.Error(codes.InvalidArgument, "批量余额查询请求不能为空")
	}

	ids := in.GetMemberGlobalIds()
	if len(ids) > maxBatchGetBalances {
		return nil, status.Error(codes.InvalidArgument, "单次最多查询100个会员")
	}

	memberGlobalIDs := make([]uuid.UUID, 0, len(ids))
	seen := make(map[uuid.UUID]struct{}, len(ids))

	for _, id := range ids {
		memberGlobalID, err := uuid.Parse(id)
		if err != nil || memberGlobalID == uuid.Nil {
			return nil, status.Error(codes.InvalidArgument, "会员全局标识无效")
		}

		// 去重, 避免重复查询同一个会员
		if _, exists := seen[memberGlobalID]; exists {
			continue
		}

		seen[memberGlobalID] = struct{}{}
		memberGlobalIDs = append(memberGlobalIDs, memberGlobalID)
	}

	return memberGlobalIDs, nil
}
