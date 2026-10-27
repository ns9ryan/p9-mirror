package loginlog

import (
	"context"
	"fmt"
	"slices"

	"oa.98ent.com/p9/common/i18n"
	"oa.98ent.com/p9/integration/rpc/types/integration"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAgentLoginLogsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListAgentLoginLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAgentLoginLogsLogic {
	return &ListAgentLoginLogsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListAgentLoginLogsLogic) ListAgentLoginLogs(req *types.ListAgentLoginLogsReq) (*types.ListAgentLoginLogsResp, error) {
	if req == nil {
		req = &types.ListAgentLoginLogsReq{}
	}
	rpcResp, err := l.svcCtx.AgentRpc.ListAgentLoginLogs(l.ctx, &agentpb.ListAgentLoginLogsReq{
		Page:        req.Page,
		PageSize:    req.PageSize,
		TimeStart:   req.TimeStart,
		TimeEnd:     req.TimeEnd,
		LoginType:   req.LoginType,
		DeviceType:  req.DeviceType,
		KeywordType: req.KeywordType,
		Keyword:     req.Keyword,
	})
	if err != nil {
		return nil, err
	}
	// 获取IP地理信息
	ipList := []string{}
	for _, it := range rpcResp.GetItems() {
		if it != nil && !slices.Contains(ipList, it.LoginIp) {
			ipList = append(ipList, it.LoginIp)
		}
	}
	ipInfoMap, err := l.getIpsLocation(ipList)
	if err != nil {
		return nil, err
	}

	items := make([]types.AgentLoginLogItem, 0, len(rpcResp.GetItems()))
	for _, it := range rpcResp.GetItems() {
		if it == nil {
			continue
		}
		if ipLocation, ok := ipInfoMap[it.LoginIp]; ok {
			it.LoginLocation = ipLocation
		}
		items = append(items, types.AgentLoginLogItem{
			Id:              it.Id,
			AgentId:         it.AgentId,
			Account:         it.Account,
			LoginType:       it.LoginType,
			LoginTypeLabel:  i18n.TG(l.ctx, i18n.CodeOperator, i18n.GroupFront, fmt.Sprintf("agent.loginType.%d", it.LoginType)),
			LoginIp:         it.LoginIp,
			LoginLocation:   it.LoginLocation,
			Udid:            it.Udid,
			LoginDevice:     it.LoginDevice,
			DeviceType:      it.DeviceType,
			DeviceTypeLabel: i18n.TG(l.ctx, i18n.CodeOperator, i18n.GroupFront, fmt.Sprintf("agent.deviceType.%d", it.DeviceType)),
			LoginAt:         it.LoginAt,
			Result:          it.Result,
			ResultLabel:     i18n.TG(l.ctx, i18n.CodeOperator, i18n.GroupFront, fmt.Sprintf("agent.loginResult.%d", it.Result)),
		})
	}
	return &types.ListAgentLoginLogsResp{Items: items, Total: rpcResp.GetTotal()}, nil
}

// 获取IP地理信息
func (l *ListAgentLoginLogsLogic) getIpsLocation(ipList []string) (map[string]string, error) {
	ipInfo, err := l.svcCtx.IntegrationRpc.SearchIp(l.ctx, &integration.SearchIpReq{
		Ips: ipList, Lang: "zh-CN",
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
