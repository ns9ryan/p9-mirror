package agent

import (
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"
)

func ConfigFromPB(in *agentpb.AgentConfigResp) types.AgentConfig {
	if in == nil {
		return types.AgentConfig{}
	}
	return types.AgentConfig{
		Id:                    in.Id,
		MaxShareRate:          in.MaxShareRate,
		ApplyMinShareRate:     in.ApplyMinShareRate,
		ApplyMaxShareRate:     in.ApplyMaxShareRate,
		CashMaxMemberCount:    in.CashMaxMemberCount,
		CashMaxAgentPerIp:     in.CashMaxAgentPerIp,
		CashProductionAgentId: in.CashProductionAgentId,
		CashDemoAgentId:       in.CashDemoAgentId,
		CashTestAgentId:       in.CashTestAgentId,
		CreditDemoAgentId:     in.CreditDemoAgentId,
		CreditTestAgentId:     in.CreditTestAgentId,
	}
}

func LevelItemFromPB(in *agentpb.AgentLevelItem) types.AgentLevelItem {
	if in == nil {
		return types.AgentLevelItem{}
	}
	return types.AgentLevelItem{
		Id:    in.Id,
		Level: in.Level,
		Name:  in.Name,
		Alias: in.Alias,
	}
}

func LevelItemsFromPB(in []*agentpb.AgentLevelItem) []types.AgentLevelItem {
	out := make([]types.AgentLevelItem, 0, len(in))
	for _, it := range in {
		if it == nil {
			continue
		}
		out = append(out, LevelItemFromPB(it))
	}
	return out
}

func LevelItemsToPB(in []types.AgentLevelItem) []*agentpb.AgentLevelItem {
	out := make([]*agentpb.AgentLevelItem, 0, len(in))
	for i := range in {
		it := in[i]
		out = append(out, &agentpb.AgentLevelItem{
			Id:    it.Id,
			Level: it.Level,
			Name:  it.Name,
			Alias: it.Alias,
		})
	}
	return out
}

func CreateFieldsToPB(in *types.AgentCreateFields) *agentpb.AgentCreateFields {
	if in == nil {
		return nil
	}
	return &agentpb.AgentCreateFields{
		Account:       in.Account,
		Code:          in.Code,
		Password:      in.Password,
		AgentType:     in.AgentType,
		AgentLineCode: in.AgentLineCode,
		ShareRate:     in.ShareRate,
		PromoCode:     in.PromoCode,
		RegisterIp:    in.RegisterIp,
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

func SearchItemFromPB(in *agentpb.AgentSearchItem) types.AgentSearchItem {
	if in == nil {
		return types.AgentSearchItem{}
	}
	return types.AgentSearchItem{
		Id:         in.Id,
		Account:    in.Account,
		Level:      in.Level,
		LevelAlias: in.LevelAlias,
		LineCode:   in.LineCode,
		ShareRate:  in.ShareRate,
	}
}

func SearchItemsFromPB(in []*agentpb.AgentSearchItem) []types.AgentSearchItem {
	out := make([]types.AgentSearchItem, 0, len(in))
	for _, it := range in {
		if it == nil {
			continue
		}
		out = append(out, SearchItemFromPB(it))
	}
	return out
}

func ListItemFromPB(in *agentpb.AgentListItem) types.AgentListItem {
	if in == nil {
		return types.AgentListItem{}
	}
	return types.AgentListItem{
		Id:                in.Id,
		Account:           in.Account,
		Level:             in.Level,
		LevelAlias:        in.LevelAlias,
		ParentId:          in.ParentId,
		ParentAccount:     in.ParentAccount,
		TopId:             in.TopId,
		TopAccount:        in.TopAccount,
		RegisterSource:    in.RegisterSource,
		DirectMemberCount: in.DirectMemberCount,
		DirectAgentCount:  in.DirectAgentCount,
		TotalMemberCount:  in.TotalMemberCount,
		TotalAgentCount:   in.TotalAgentCount,
		ShareRate:         in.ShareRate,
		WalletBalance:     in.WalletBalance,
		TotalCommission:   in.TotalCommission,
		UnpaidCommission:  in.UnpaidCommission,
		Status:            in.Status,
		PayoutRestricted:  in.PayoutRestricted,
		Remark:            in.Remark,
		PromoLink:         in.PromoLink,
		VisitUv:           in.VisitUv,
		RegisterAt:        in.RegisterAt,
		LastLoginAt:       in.LastLoginAt,
	}
}

func ListItemsFromPB(in []*agentpb.AgentListItem) []types.AgentListItem {
	out := make([]types.AgentListItem, 0, len(in))
	for _, it := range in {
		if it == nil {
			continue
		}
		out = append(out, ListItemFromPB(it))
	}
	return out
}

func ApplyItemFromPB(in *agentpb.AgentApplyItem) types.AgentApplyItem {
	if in == nil {
		return types.AgentApplyItem{}
	}
	return types.AgentApplyItem{
		Id:             in.Id,
		Account:        in.Account,
		Realname:       in.Realname,
		Mobile:         in.Mobile,
		RegisterSource: in.RegisterSource,
		ApplyShareRate: in.ApplyShareRate,
		MemberAccount:  in.MemberAccount,
		MemberCode:     in.MemberCode,
		AuditStatus:    in.AuditStatus,
		RejectReason:   in.RejectReason,
		ApplyRemark:    in.ApplyRemark,
		RegisterAt:     in.RegisterAt,
		AuditedAt:      in.AuditedAt,
		Auditor:        in.Auditor,
		AgentId:        in.AgentId,
	}
}

func ProfileFromPB(in *agentpb.AgentProfile) types.AgentProfile {
	if in == nil {
		return types.AgentProfile{}
	}
	return types.AgentProfile{
		AgentId:           in.AgentId,
		Remark:            in.Remark,
		Status:            in.Status,
		RegisterSource:    in.RegisterSource,
		PromoCode:         in.PromoCode,
		ParentAccount:     in.ParentAccount,
		RegisterAt:        in.RegisterAt,
		LastLoginAt:       in.LastLoginAt,
		LastLoginIp:       in.LastLoginIp,
		LastLoginUdid:     in.LastLoginUdid,
		ShareRate:         in.ShareRate,
		LoginCount:        in.LoginCount,
		InviteMemberCount: in.InviteMemberCount,
		PayoutRestricted:  in.PayoutRestricted,
		AgentType:         in.AgentType,
		AvatarUrl:         in.AvatarUrl,
		Realname:          in.Realname,
		Telegram:          in.Telegram,
		Facebook:          in.Facebook,
		Whatsapp:          in.Whatsapp,
		Mobile:            in.Mobile,
		Nickname:          in.Nickname,
		IdDocumentNo:      in.IdDocumentNo,
		Zalo:              in.Zalo,
		Twitter:           in.Twitter,
		Line:              in.Line,
		Email:             in.Email,
	}
}

func ApplyItemsFromPB(in []*agentpb.AgentApplyItem) []types.AgentApplyItem {
	out := make([]types.AgentApplyItem, 0, len(in))
	for _, it := range in {
		if it == nil {
			continue
		}
		out = append(out, ApplyItemFromPB(it))
	}
	return out
}
