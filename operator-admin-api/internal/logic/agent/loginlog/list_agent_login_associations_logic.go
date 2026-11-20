package loginlog

import (
	"context"
	"strconv"

	"oa.98ent.com/p9/common/i18n"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAgentLoginAssociationsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 关联检索登录日志
func NewListAgentLoginAssociationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAgentLoginAssociationsLogic {
	return &ListAgentLoginAssociationsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListAgentLoginAssociationsLogic) ListAgentLoginAssociations(req *types.ListAgentLoginAssociationsReq) (*types.ListAgentLoginAssociationsResp, error) {
	if req == nil {
		req = &types.ListAgentLoginAssociationsReq{}
	}
	rpcResp, err := l.svcCtx.AgentRpc.ListAgentLoginAssociations(l.ctx, &agentpb.ListAgentLoginAssociationsReq{
		Page:        req.Page,
		PageSize:    req.PageSize,
		TimeStart:   req.TimeStart,
		TimeEnd:     req.TimeEnd,
		AgentId:     req.AgentId,
		DeviceType:  req.DeviceType,
		KeywordType: req.KeywordType,
		Keyword:     req.Keyword,
	})
	if err != nil {
		return nil, err
	}
	list := make([]types.AgentLoginAssociationItem, 0, len(rpcResp.GetList()))
	for _, it := range rpcResp.GetList() {
		if it == nil {
			continue
		}
		list = append(list, types.AgentLoginAssociationItem{
			Id:                   it.Id,
			AgentId:              it.AgentId,
			Account:              it.Account,
			LoginIp:              it.LoginIp,
			LoginLocation:        it.LoginLocation,
			Udid:                 it.Udid,
			LoginDevice:          it.LoginDevice,
			DeviceType:           it.DeviceType,
			DeviceTypeLabel:      i18n.TG(l.ctx, i18n.CodeOperator, i18n.GroupFront, "agent.deviceType."+strconv.FormatInt(int64(it.DeviceType), 10)),
			LoginAt:              it.LoginAt,
			SameIpAgentCount:     it.SameIpAgentCount,
			SameDeviceAgentCount: it.SameDeviceAgentCount,
		})
	}
	return &types.ListAgentLoginAssociationsResp{List: list, Total: rpcResp.GetTotal()}, nil
}
