package memberwalletlogic

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"oa.98ent.com/p9/operator-wallet/ent"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/dao"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/svc"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc/managepb"
)

// CreateWalletLogic 处理会员钱包创建
type CreateWalletLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewCreateWalletLogic 创建会员钱包逻辑
func NewCreateWalletLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateWalletLogic {
	return &CreateWalletLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreateWallet 创建会员钱包, 重复请求返回已有钱包
func (l *CreateWalletLogic) CreateWallet(in *managepb.CreateWalletRequest) (*managepb.CreateWalletResponse, error) {
	// 校验请求参数
	params, err := parseCreateWalletRequest(in)
	if err != nil {
		return nil, err
	}

	// 校验可信厅身份
	if err := validateOperatorCode(l.ctx); err != nil {
		return nil, err
	}

	// 创建钱包, 已存在且参数一致时返回原钱包
	memberWallet, err := l.svcCtx.DAO.MemberWallet.CreateOrGet(l.ctx, params)
	if err != nil {
		switch {
		case errors.Is(err, dao.ErrWalletIdentityConflict):
			return nil, status.Error(codes.FailedPrecondition, "会员钱包信息冲突")

		case ent.IsValidationError(err):
			return nil, status.Error(codes.InvalidArgument, "会员钱包参数无效")

		case errors.Is(err, context.Canceled):
			return nil, status.Error(codes.Canceled, "创建会员钱包请求已取消")

		case errors.Is(err, context.DeadlineExceeded):
			return nil, status.Error(codes.DeadlineExceeded, "创建会员钱包超时")

		default:
			l.Errorf("create member wallet failed, member_global_id=%s: %v", params.MemberGlobalID, err)
			return nil, status.Error(codes.Internal, "创建会员钱包失败")
		}
	}

	return &managepb.CreateWalletResponse{
		WalletGlobalId: memberWallet.GlobalID.String(),
	}, nil
}

// parseCreateWalletRequest 校验并解析创建钱包请求
func parseCreateWalletRequest(in *managepb.CreateWalletRequest) (dao.CreateWalletParams, error) {
	if in == nil {
		return dao.CreateWalletParams{}, status.Error(codes.InvalidArgument, "创建钱包请求不能为空")
	}

	// 校验 CRM 会员 ID
	memberID := in.GetMemberId()
	if memberID <= 0 {
		return dao.CreateWalletParams{}, status.Error(codes.InvalidArgument, "会员ID无效")
	}

	// 校验 CRM 会员全局标识
	memberGlobalID, err := uuid.Parse(in.GetMemberGlobalId())
	if err != nil || memberGlobalID == uuid.Nil {
		return dao.CreateWalletParams{}, status.Error(codes.InvalidArgument, "会员全局标识无效")
	}

	// 校验钱包结算币种
	currencyCode := in.GetCurrencyCode()
	if currencyCode == "" || strings.TrimSpace(currencyCode) != currencyCode || utf8.RuneCountInString(currencyCode) > 16 {
		return dao.CreateWalletParams{}, status.Error(codes.InvalidArgument, "币种编码无效")
	}

	return dao.CreateWalletParams{
		MemberID:       memberID,       // CRM会员ID
		MemberGlobalID: memberGlobalID, // CRM会员全局唯一标识
		CurrencyCode:   currencyCode,   // 钱包结算币种编码
	}, nil
}
