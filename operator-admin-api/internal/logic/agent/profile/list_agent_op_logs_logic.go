package profile

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"oa.98ent.com/p9/common/i18n"
	"oa.98ent.com/p9/integration/rpc/types/integration"
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
		OpTargetId:   req.TargetId,
		OpTargetType: req.TargetType,
		Page:         req.Page,
		PageSize:     req.PageSize,
		TimeStart:    req.TimeStart,
		TimeEnd:      req.TimeEnd,
		Result:       req.Result,
	})
	if err != nil {
		return nil, err
	}
	// 获取IP地理信息
	ipList := []string{}
	for _, it := range rpcResp.GetItems() {
		if it != nil && !slices.Contains(ipList, it.ClientIp) {
			ipList = append(ipList, it.ClientIp)
		}
	}
	ipInfoMap, err := l.getIpsLocation(ipList)
	if err != nil {
		// return nil, err
		l.Logger.Errorf("getIpsLocation error: %v", err)
	}

	items := make([]types.AgentOpLogItem, 0, len(rpcResp.GetItems()))
	for _, it := range rpcResp.GetItems() {
		if it == nil {
			continue
		}
		// 设置IP地理信息到返回数据
		clientIpLocation := ""
		if ipLocation, ok := ipInfoMap[it.ClientIp]; ok {
			clientIpLocation = ipLocation
		}
		items = append(items, types.AgentOpLogItem{
			Id:                it.Id,
			OpTargetId:        it.OpTargetId,
			OpTargetType:      it.OpTargetType,
			OpTargetTypeLabel: i18n.TG(l.ctx, i18n.CodeOperator, i18n.GroupFront, fmt.Sprintf("agent.opTargetType.%d", it.OpTargetType)),
			OpTargetAccount:   it.OpTargetAccount,
			AdminId:           it.AdminId,
			AdminType:         it.AdminType,
			AdminTypeLabel:    i18n.TG(l.ctx, i18n.CodeOperator, i18n.GroupFront, fmt.Sprintf("agent.adminType.%d", it.AdminType)),
			AdminAccount:      it.AdminAccount,
			Action:            it.Action,
			ActionLabel:       i18n.TG(l.ctx, i18n.CodeOperator, i18n.GroupFront, it.Action),
			BeforeValue:       it.BeforeValue,
			AfterValue:        it.AfterValue,
			ClientIp:          it.ClientIp,
			ClientIpLocation:  clientIpLocation,
			LoginDevice:       it.LoginDevice,
			Udid:              it.Udid,
			OperatedAt:        it.OperatedAt,
			Result:            it.Result,
			ResultLabel:       i18n.TG(l.ctx, i18n.CodeOperator, i18n.GroupFront, fmt.Sprintf("agent.opResult.%d", it.Result)),
			FailReason:        localizeFailReason(l.ctx, it.FailReason),
		})
	}
	return &types.ListAgentOpLogsResp{Items: items, Total: rpcResp.GetTotal()}, nil
}

// 获取IP地理信息
func (l *ListAgentOpLogsLogic) getIpsLocation(ipList []string) (map[string]string, error) {
	ipInfo, err := l.svcCtx.IntegrationRpc.SearchIp(l.ctx, &integration.SearchIpReq{
		Ips: ipList, Lang: i18n.Lang(l.ctx, nil),
	})
	if err != nil {
		return nil, err
	}
	ipInfoMap := make(map[string]string)
	for _, it := range ipInfo.GetList() {
		ipInfoMap[it.GetQuery()] = it.GetCountry()
	}
	return ipInfoMap, nil
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
