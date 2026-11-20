package profile

import (
	"context"

	agentconv "oa.98ent.com/p9/operator-admin-api/internal/logic/agent"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateAgentProfileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateAgentProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAgentProfileLogic {
	return &UpdateAgentProfileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateAgentProfileLogic) UpdateAgentProfile(req *types.UpdateAgentProfileReq) (*types.AgentProfile, error) {
	if req == nil {
		req = &types.UpdateAgentProfileReq{}
	}
	rpcResp, err := l.svcCtx.AgentRpc.UpdateAgentProfile(l.ctx, &agentpb.UpdateAgentProfileReq{
		AgentId:      req.AgentId,
		FieldType:    req.FieldType,
		AvatarUrl:    req.AvatarUrl,
		Realname:     req.Realname,
		Telegram:     req.Telegram,
		Facebook:     req.Facebook,
		Whatsapp:     req.Whatsapp,
		Mobile:       req.Mobile,
		Nickname:     req.Nickname,
		IdDocumentNo: req.IdDocumentNo,
		Zalo:         req.Zalo,
		Twitter:      req.Twitter,
		Line:         req.Line,
		Email:        req.Email,
	})
	if err != nil {
		return nil, err
	}
	out := agentconv.ProfileFromPB(rpcResp)
	return &out, nil
}
