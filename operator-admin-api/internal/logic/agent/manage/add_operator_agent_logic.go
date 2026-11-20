package manage

import (
	"context"
	"strings"

	"oa.98ent.com/p9/common/utils"
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
	// 设置agent code
	agentCode, err := utils.GenLongID()
	if err != nil {
		l.Logger.Errorw("GenLongID error", logx.Field("error", err))
		return nil, err
	}
	// 设置promo code
	promoCode, err := utils.GenShortID()
	if err != nil {
		l.Logger.Errorw("GenShortID error", logx.Field("error", err))
		return nil, err
	}
	// 设置agent code和promo code
	req.Agent.Code = strings.ToUpper(agentCode)
	req.Agent.PromoCode = strings.ToUpper(promoCode)

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
