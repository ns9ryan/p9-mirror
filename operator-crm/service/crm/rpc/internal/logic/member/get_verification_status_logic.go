package memberlogic

import (
	"context"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent"
	memberdao "oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/dao/member"

	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetVerificationStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetVerificationStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetVerificationStatusLogic {
	return &GetVerificationStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 仅返回认证字段是否设置，不返回密文或哈希
func (l *GetVerificationStatusLogic) GetVerificationStatus(in *memberpb.GetMemberContextRequest) (*memberpb.GetMemberVerificationStatusResponse, error) {
	if in.GetMemberId() <= 0 && in.GetMemberCode() == "" {
		return nil, xerr.BadRequest("common.invalidParam")
	}

	var (
		m   *ent.Member
		err error
	)

	// id > 0 优先使用会员ID查询，否则使用会员编码查询
	if in.GetMemberId() > 0 {
		m, err = l.svcCtx.Member.GetMemberWithContextData(l.ctx, in.GetMemberId(), memberdao.ContextKindVerification)
		if err != nil {
			return nil, xerr.RpcErr(err)
		}
	} else {
		m, err = l.svcCtx.Member.GetMemberWithContextDataByCode(l.ctx, in.GetMemberCode(), memberdao.ContextKindVerification)
		if err != nil {
			return nil, xerr.RpcErr(err)
		}
	}
	auth, err := m.Edges.AuthOrErr()
	if err != nil {
		return nil, xerr.RpcErr(err)
	}

	return &memberpb.GetMemberVerificationStatusResponse{
		HasEmail:              auth.EmailCiphertext != nil && *auth.EmailCiphertext != "",
		HasPhone:              auth.PhoneCiphertext != nil && *auth.PhoneCiphertext != "",
		HasWithdrawalPassword: auth.WithdrawalPasswordHash != nil && *auth.WithdrawalPasswordHash != "",
		RegisterIp:            m.RegisterIP, RegisterDeviceCode: m.RegisterDeviceCode,
	}, nil
}
