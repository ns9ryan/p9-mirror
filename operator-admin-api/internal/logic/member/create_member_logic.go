package member

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"
)

type CreateMemberLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateMemberLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateMemberLogic {
	return &CreateMemberLogic{ctx:ctx, svcCtx:svcCtx, Logger:logx.WithContext(ctx)}
}

// CreateMember 将后台表单转换为会员 RPC 请求。
func (l *CreateMemberLogic) CreateMember(req *types.CreateMemberReq) (*types.CreateMemberResp, error) {
	resp, err := l.svcCtx.MemberRpc.CreateMember(l.ctx, &memberpb.CreateMemberRequest{
		Account: req.Account, Password: req.Password, Nickname: req.Nickname,
		AgentId: req.AgentId, Remark: req.Remark, AvatarUrl: req.AvatarUrl,
		AvatarFrameUrl: req.AvatarFrameUrl,
	})
	if err != nil {
		return nil, err
	}
	return CreateMemberFromPB(resp), nil
}
