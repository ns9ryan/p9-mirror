// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package deposit_order

import (
	"context"
	"strconv"

	"oa.98ent.com/p9/common/ctxdata"
	agentconv "oa.98ent.com/p9/operator-admin-api/internal/logic/agent"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentdepositorderpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAgentDepositOrdersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 人工存款订单列表
func NewListAgentDepositOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAgentDepositOrdersLogic {
	return &ListAgentDepositOrdersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListAgentDepositOrdersLogic) ListAgentDepositOrders(req *types.ListAgentDepositOrdersReq) (*types.ListAgentDepositOrdersResp, error) {
	if req == nil {
		req = &types.ListAgentDepositOrdersReq{}
	}
	pbReq := &agentdepositorderpb.ListAgentDepositOrdersReq{
		Page:        req.Page,
		PageSize:    req.PageSize,
		DepositType: req.DepositType,
		TimeStart:   req.TimeStart,
		TimeEnd:     req.TimeEnd,
		TimeType:    req.TimeType,
	}
	// member_id 整数转字符串，作为 member.code 模糊搜索条件
	if req.MemberId > 0 {
		mc := strconv.FormatInt(req.MemberId, 10)
		pbReq.MemberCode = &mc
	}
	// 从分站配置读取结算货币，作为存款货币过滤条件。货币未配置时直接报错。
	depositCurrency, _, err := agentconv.GetSiteCurrency(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}
	pbReq.DepositCurrency = &depositCurrency
	if req.OrderNo != "" {
		pbReq.OrderNo = &req.OrderNo
	}
	if req.MemberAccount != "" {
		pbReq.MemberAccount = &req.MemberAccount
	}
	if req.ThirdPartyOrderNo != "" {
		pbReq.ThirdPartyOrderNo = &req.ThirdPartyOrderNo
	}
	if req.ReviewerId > 0 {
		pbReq.ReviewerId = &req.ReviewerId
	}
	if req.ClaimStatus > 0 {
		pbReq.ClaimStatus = &req.ClaimStatus
	}

	// mode 未传时默认为 1（审核列表）。
	// Statuses 多选优先于 Status 单选；两者均受 mode 状态范围约束，越界值静默剔除。
	if req.Mode == 0 {
		req.Mode = 1
	}
	allowed := modeAllowedStatuses(req.Mode)
	if len(req.Statuses) > 0 {
		// 多选：裁剪到 mode 允许范围；全部越界则兜底用 mode 全量
		filtered := filterStatuses(req.Statuses, allowed)
		if len(filtered) == 0 && allowed != nil {
			pbReq.Statuses = modeAllowedKeys(allowed)
		} else {
			pbReq.Statuses = filtered
		}
	} else if req.Status != 0 {
		// 单选：不在 mode 范围内则兜底用 mode 全量
		filtered := filterStatus(req.Status, allowed)
		if filtered == 0 && allowed != nil {
			pbReq.Statuses = modeAllowedKeys(allowed)
		} else {
			pbReq.Status = filtered
		}
	} else if allowed != nil {
		// 未传任何状态，限定在 mode 允许的状态范围
		pbReq.Statuses = modeAllowedKeys(allowed)
	}

	rpcResp, err := l.svcCtx.AgentDepositOrderRpc.ListAgentDepositOrders(l.ctx, pbReq)
	if err != nil {
		return nil, err
	}
	// 从 JWT claims 获取当前用户 ID，用于计算 ClaimedBySelf
	var selfUID int64
	if claims := ctxdata.ClaimsFromCtx(l.ctx); claims != nil {
		selfUID = claims.UserID
	}
	return &types.ListAgentDepositOrdersResp{
		List:  agentconv.DepositOrderListItemsFromPB(rpcResp.GetList(), selfUID),
		Total: rpcResp.GetTotal(),
	}, nil
}
