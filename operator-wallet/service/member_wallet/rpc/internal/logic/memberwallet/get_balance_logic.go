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

// GetBalanceLogic 处理会员钱包余额查询
type GetBalanceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewGetBalanceLogic 创建会员钱包余额查询逻辑
func NewGetBalanceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBalanceLogic {
	return &GetBalanceLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetBalance 查询会员钱包余额
func (l *GetBalanceLogic) GetBalance(in *managepb.GetBalanceRequest) (*managepb.GetBalanceResponse, error) {
	// 校验请求参数
	memberGlobalID, err := parseGetBalanceRequest(in)
	if err != nil {
		return nil, err
	}

	// 确保会员钱包已加载并注册到 Engine, 查询不要求调用方提供币种, 使用钱包自身的结算币种
	walletGlobalID, err := ensureWalletActor(
		l.ctx,
		l.svcCtx,
		walletActorParams{
			memberGlobalID: memberGlobalID, // 会员全局唯一标识
		},
	)
	if err != nil {
		return nil, mapGetBalanceError(err, memberGlobalID, l.Logger)
	}

	// 从 Actor 获取当前运行时余额
	snapshot, err := l.svcCtx.WalletEngine.GetBalance(l.ctx, walletGlobalID)
	if err != nil {
		return nil, mapGetBalanceError(err, memberGlobalID, l.Logger)
	}

	return &managepb.GetBalanceResponse{
		WalletGlobalId:   snapshot.WalletGlobalID.String(), // 钱包全局唯一标识
		CurrencyCode:     snapshot.CurrencyCode,            // 钱包结算币种
		Balance:          snapshot.Balance,                 // 可用余额
		FrozenBalance:    snapshot.FrozenBalance,           // 冻结余额
		UnsettledBalance: snapshot.UnsettledBalance,        // 游戏未结算余额
		Version:          snapshot.Version,                 // 钱包状态版本
	}, nil
}

// parseGetBalanceRequest 校验并解析余额查询请求
func parseGetBalanceRequest(in *managepb.GetBalanceRequest) (uuid.UUID, error) {
	if in == nil {
		return uuid.Nil, status.Error(codes.InvalidArgument, "余额查询请求不能为空")
	}

	memberGlobalID, err := uuid.Parse(in.GetMemberGlobalId())
	if err != nil || memberGlobalID == uuid.Nil {
		return uuid.Nil, status.Error(codes.InvalidArgument, "会员全局标识无效")
	}

	return memberGlobalID, nil
}

// mapGetBalanceError 统一处理余额查询错误
func mapGetBalanceError(err error, memberGlobalID uuid.UUID, logger logx.Logger) error {
	if err == nil {
		return nil
	}

	// 保留已经转换过的 RPC 错误
	if _, ok := status.FromError(err); ok {
		return err
	}

	switch {
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, "查询会员钱包余额请求已取消")

	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, "查询会员钱包余额超时")
	}

	// 复用钱包统一错误映射
	mappedErr := mapWalletError(err)

	if status.Code(mappedErr) == codes.Internal {
		logger.Errorf("get member wallet balance failed, member_global_id=%s: %v", memberGlobalID, err)
	}

	return mappedErr
}
