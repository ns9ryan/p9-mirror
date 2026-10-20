package manage

import (
	"context"
	"strings"

	"oa.98ent.com/p9/common/ctxdata"
	agentconv "oa.98ent.com/p9/operator-admin-api/internal/logic/agent"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddOperatorAgentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddOperatorAgentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddOperatorAgentLogic {
	return &AddOperatorAgentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddOperatorAgentLogic) AddOperatorAgent(req *types.AddOperatorAgentReq) (*types.AddAgentResp, error) {
	if req == nil {
		req = &types.AddOperatorAgentReq{}
	}
	// 设置code, promo_code
	uniCode := ctxdata.OperatorCodeFromCtx(l.ctx) + "_" + strings.ToUpper(req.Agent.Account)
	if req.Agent.Code == "" {
		req.Agent.Code = uniCode
	}
	if req.Agent.PromoCode == "" {
		req.Agent.PromoCode = uniCode
	}
	rpcResp, err := l.svcCtx.AgentRpc.AddOperatorAgent(l.ctx, &agentpb.AddOperatorAgentReq{
		Agent:    agentconv.CreateFieldsToPB(&req.Agent),
		IsInit:   req.IsInit,
		ParentId: req.ParentId,
	})
	if err != nil {
		return nil, err
	}
	return agentconv.AddAgentFromPB(rpcResp), nil
}
