// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package manage

import (
	"context"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"

	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateContactLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateContactLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateContactLogic {
	return &UpdateContactLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateContactLogic) UpdateContact(req *types.UpdateMemberContactRequest) (resp *types.EmptyResponse, err error) {
	_, err = l.svcCtx.MemberRpc.UpdateContact(l.ctx, &memberpb.UpdateMemberContactRequest{MemberId: req.MemberId, AvatarId: req.AvatarId, AvatarFrameId: req.AvatarFrameId, Nickname: req.Nickname, RealName: req.RealName, Phone: req.Phone, IdentityNumber: req.IdentityNumber, Email: req.Email, PhoneVerification: contactBindingCodeProof(req.PhoneVerification), EmailVerification: contactBindingCodeProof(req.EmailVerification)})
	if err != nil {
		return nil, err
	}
	return &types.EmptyResponse{}, nil
}

func contactBindingCodeProof(proof *types.ContactBindingCodeProof) *memberpb.ContactBindingCodeProof {
	if proof == nil {
		return nil
	}
	return &memberpb.ContactBindingCodeProof{RecordId: proof.RecordId, VerificationCode: proof.VerificationCode}
}
