package profile

import (
	"context"
	"strconv"
	"strings"

	"oa.98ent.com/p9/common/i18n"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAgentOpLogsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListAgentOpLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAgentOpLogsLogic {
	return &ListAgentOpLogsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListAgentOpLogsLogic) ListAgentOpLogs(req *types.ListAgentOpLogsReq) (*types.ListAgentOpLogsResp, error) {
	if req == nil {
		req = &types.ListAgentOpLogsReq{}
	}
	rpcResp, err := l.svcCtx.AgentRpc.ListAgentOpLogs(l.ctx, &agentpb.ListAgentOpLogsReq{
		AgentId:   req.AgentId,
		Page:      req.Page,
		PageSize:  req.PageSize,
		TimeStart: req.TimeStart,
		TimeEnd:   req.TimeEnd,
		Result:    req.Result,
	})
	if err != nil {
		return nil, err
	}
	items := make([]types.AgentOpLogItem, 0, len(rpcResp.GetItems()))
	for _, it := range rpcResp.GetItems() {
		if it == nil {
			continue
		}
		items = append(items, types.AgentOpLogItem{
			Id:           it.Id,
			AgentId:      it.AgentId,
			AdminId:      it.AdminId,
			AdminAccount: it.AdminAccount,
			Action:       it.Action,
			ActionLabel:  i18n.TG(l.ctx, i18n.CodeOperator, i18n.GroupFront, "agent.opAction."+it.Action),
			BeforeValue:  it.BeforeValue,
			AfterValue:   it.AfterValue,
			ClientIp:     it.ClientIp,
			LoginDevice:  it.LoginDevice,
			Udid:         it.Udid,
			OperatedAt:   it.OperatedAt,
			Result:       it.Result,
			ResultLabel:  i18n.TG(l.ctx, i18n.CodeOperator, i18n.GroupFront, "agent.opResult."+strconv.FormatInt(int64(it.Result), 10)),
			FailReason:   localizeFailReason(l.ctx, it.FailReason),
		})
	}
	return &types.ListAgentOpLogsResp{Items: items, Total: rpcResp.GetTotal()}, nil
}

func localizeFailReason(ctx context.Context, reason string) string {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return ""
	}
	if strings.Contains(reason, ".") && !strings.ContainsAny(reason, " \t\n") {
		if label := i18n.TG(ctx, i18n.CodeOperator, i18n.GroupError, reason); label != "" && label != reason {
			return label
		}
	}
	return reason
}
