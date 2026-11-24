// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package manage

import (
	"context"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-base/rpc/pb/operatorbaserpc/operatorpb"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc/managepb"
)

// CreateLogic 会员创建逻辑
type CreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewCreateLogic 创建会员逻辑
func NewCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateLogic {
	return &CreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Create 创建会员
func (l *CreateLogic) Create(req *types.CreateMemberRequest) (resp *types.CreateMemberResponse, err error) {
	// 获取当前厅的结算币种
	siteResp, err := l.svcCtx.OperatorBaseRpc.GetSiteConfig(l.ctx, &operatorpb.GetSiteConfigRequest{})
	if err != nil {
		return nil, err
	}

	currencyCode := strings.TrimSpace(siteResp.GetSiteConfig().GetSettlementCurrencyCode())
	if currencyCode == "" {
		return nil, status.Error(codes.FailedPrecondition, "当前厅未配置结算币种")
	}

	// 创建CRM会员
	memberResp, err := l.svcCtx.MemberRpc.Create(l.ctx, &memberpb.CreateMemberRequest{
		Account:  req.Account,  // 会员登录账号
		Password: req.Password, // 初始密码
		Nickname: req.Nickname, // 会员昵称
		AgentId:  req.AgentId,  // 所属分层代理ID
		Remark:   req.Remark,   // 后台会员备注
	})
	if err != nil {
		return nil, err
	}

	// CRM创建成功后初始化会员钱包
	_, err = l.svcCtx.MemberWalletRpc.CreateWallet(l.ctx, &managepb.CreateWalletRequest{
		MemberId:       memberResp.GetId(),             // CRM会员ID
		MemberGlobalId: memberResp.GetMemberGlobalId(), // CRM会员全局标识
		CurrencyCode:   currencyCode,                   // 当前厅结算币种
	})
	if err != nil {
		// CRM已提交, 钱包失败不能视为整个创建流程已回滚
		l.Errorf("create member wallet failed, member_id=%d, member_global_id=%s: %v",
			memberResp.GetId(), memberResp.GetMemberGlobalId(), err)

		return nil, status.Errorf(
			codes.FailedPrecondition,
			"会员已创建（ID=%d），但钱包初始化失败，请勿重复创建",
			memberResp.GetId(),
		)
	}

	// 会员与钱包均创建成功
	return &types.CreateMemberResponse{
		Id:   memberResp.GetId(),   // 会员数据库主键
		Code: memberResp.GetCode(), // 会员对外编号
	}, nil
}
