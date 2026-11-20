package agent

import (
	"oa.98ent.com/p9/agent-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"
)

func CreateFieldsToPB(in *types.AddSubAgentReq) *agentpb.AgentCreateFields {
	if in == nil {
		return nil
	}
	return &agentpb.AgentCreateFields{
		Account:    in.Account,
		Password:   in.Password,
		AgentType:  in.AgentType,
		ShareRate:  in.ShareRate,
		RegisterIp: in.RegisterIp,
		Remark:     in.Remark,
	}
}

func AddAgentFromPB(in *agentpb.AddAgentResp) *types.AddAgentResp {
	if in == nil {
		return &types.AddAgentResp{}
	}
	return &types.AddAgentResp{
		Id:        in.Id,
		Account:   in.Account,
		Code:      in.Code,
		ParentId:  in.ParentId,
		TopId:     in.TopId,
		Level:     in.Level,
		ShareRate: in.ShareRate,
	}
}
