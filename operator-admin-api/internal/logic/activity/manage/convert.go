package manage

import (
	"oa.98ent.com/p9/operator-activity/pb/operatoractivityrpc/activitypb"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
)

func toCreateActivityReq(req *types.CreateActivityReq) *activitypb.CreateActivityReq {
	if req == nil {
		return nil
	}
	out := &activitypb.CreateActivityReq{
		Type:               req.Type,
		Status:             req.Status,
		ApplyMode:          req.ApplyMode,
		Section:            req.Section,
		CycleMode:          req.CycleMode,
		AllowedFixedLevel:  req.AllowedFixedLevel,
		AllowedAutoLevel:   req.AllowedAutoLevel,
		AllowedClients:     req.AllowedClients,
		AllowedGameIds:     req.AllowedGameIds,
		AllowedAgentIds:    req.AllowedAgentIds,
		ApplyLimit:         req.ApplyLimit,
		ClaimLimit:         req.ClaimLimit,
		IsPermanent:        req.IsPermanent,
		OpenAtSec:          req.OpenAtSec,
		CloseAtSec:         req.CloseAtSec,
		ReleaseAtSec:       req.ReleaseAtSec,
		AutoAudit:          req.AutoAudit,
		AutoAuditMaxAmt:    req.AutoAuditMaxAmt,
		PayoutMode:         req.PayoutMode,
		PayoutTiming:       req.PayoutTiming,
		ClaimExpireSec:     req.ClaimExpireSec,
		Settings:           toPBActivitySettings(req.Settings),
		WageringMultiplier: req.WageringMultiplier,
		WageringBase:       req.WageringBase,
		ShowInList:         req.ShowInList,
		ShowInSidebar:      req.ShowInSidebar,
		SortWeight:         req.SortWeight,
		CreateUid:          req.CreateUid,
		I18N:               toPBActivityI18nItems(req.I18n),
	}
	return out
}

func toUpdateActivityReq(req *types.UpdateActivityReq) *activitypb.UpdateActivityReq {
	if req == nil {
		return nil
	}
	return &activitypb.UpdateActivityReq{
		Id:                 req.Id,
		Type:               req.Type,
		ApplyMode:          req.ApplyMode,
		Section:            req.Section,
		CycleMode:          req.CycleMode,
		AllowedFixedLevel:  req.AllowedFixedLevel,
		AllowedAutoLevel:   req.AllowedAutoLevel,
		AllowedClients:     req.AllowedClients,
		AllowedGameIds:     req.AllowedGameIds,
		AllowedAgentIds:    req.AllowedAgentIds,
		ApplyLimit:         req.ApplyLimit,
		ClaimLimit:         req.ClaimLimit,
		IsPermanent:        req.IsPermanent,
		OpenAtSec:          req.OpenAtSec,
		CloseAtSec:         req.CloseAtSec,
		ReleaseAtSec:       req.ReleaseAtSec,
		AutoAudit:          req.AutoAudit,
		AutoAuditMaxAmt:    req.AutoAuditMaxAmt,
		PayoutMode:         req.PayoutMode,
		PayoutTiming:       req.PayoutTiming,
		ClaimExpireSec:     req.ClaimExpireSec,
		Settings:           toPBActivitySettings(req.Settings),
		WageringMultiplier: req.WageringMultiplier,
		WageringBase:       req.WageringBase,
		ShowInList:         req.ShowInList,
		ShowInSidebar:      req.ShowInSidebar,
		SortWeight:         req.SortWeight,
		I18N:               toPBActivityI18nItems(req.I18n),
	}
}

func toActivityDetail(in *activitypb.ActivityDetail) *types.ActivityDetail {
	if in == nil {
		return nil
	}
	return &types.ActivityDetail{
		Id:                 in.GetId(),
		Type:               in.GetType(),
		Status:             in.GetStatus(),
		ApplyMode:          in.GetApplyMode(),
		Section:            in.GetSection(),
		CycleMode:          in.GetCycleMode(),
		AllowedFixedLevel:  in.GetAllowedFixedLevel(),
		AllowedAutoLevel:   in.GetAllowedAutoLevel(),
		AllowedClients:     in.GetAllowedClients(),
		AllowedGameIds:     in.GetAllowedGameIds(),
		AllowedAgentIds:    in.GetAllowedAgentIds(),
		ApplyLimit:         in.GetApplyLimit(),
		ClaimLimit:         in.GetClaimLimit(),
		IsPermanent:        in.GetIsPermanent(),
		OpenAtSec:          in.OpenAtSec,
		CloseAtSec:         in.CloseAtSec,
		ReleaseAtSec:       in.ReleaseAtSec,
		AutoAudit:          in.GetAutoAudit(),
		AutoAuditMaxAmt:    in.GetAutoAuditMaxAmt(),
		PayoutMode:         in.GetPayoutMode(),
		PayoutTiming:       in.GetPayoutTiming(),
		ClaimExpireSec:     in.GetClaimExpireSec(),
		Settings:           toTypesActivitySettings(in.GetSettings()),
		WageringMultiplier: in.GetWageringMultiplier(),
		WageringBase:       in.GetWageringBase(),
		ShowInList:         in.GetShowInList(),
		ShowInSidebar:      in.GetShowInSidebar(),
		SortWeight:         in.GetSortWeight(),
		CreateUid:          in.GetCreateUid(),
		UpdateUid:          in.GetUpdateUid(),
		Version:            in.GetVersion(),
		CreatedAtSec:       in.GetCreatedAtSec(),
		UpdatedAtSec:       in.GetUpdatedAtSec(),
		I18n:               toTypesActivityI18nItems(in.GetI18N()),
	}
}

func toActivityListItem(in *activitypb.ActivityListItem) types.ActivityListItem {
	if in == nil {
		return types.ActivityListItem{}
	}
	return types.ActivityListItem{
		Id:                 in.GetId(),
		Type:               in.GetType(),
		Status:             in.GetStatus(),
		Name:               in.GetName(),
		Image:              in.GetImage(),
		AllowedClients:     in.GetAllowedClients(),
		OpenAtSec:          in.OpenAtSec,
		CloseAtSec:         in.CloseAtSec,
		ReleaseAtSec:       in.ReleaseAtSec,
		SortWeight:         in.GetSortWeight(),
		ShowInList:         in.GetShowInList(),
		UpdatedAtSec:       in.GetUpdatedAtSec(),
		AutoAudit:          in.GetAutoAudit(),
		AutoAuditMaxAmt:    in.GetAutoAuditMaxAmt(),
		AllowedFixedLevel:  in.GetAllowedFixedLevel(),
		AllowedAutoLevel:   in.GetAllowedAutoLevel(),
		WageringMultiplier: in.GetWageringMultiplier(),
		AllowedGameIds:     in.GetAllowedGameIds(),
		IsPermanent:        in.GetIsPermanent(),
		CreatedAtSec:       in.GetCreatedAtSec(),
		CreateUid:          in.GetCreateUid(),
		UpdateUid:          in.GetUpdateUid(),
		Section:            in.GetSection(),
	}
}

func toPBActivitySettings(in *types.ActivitySettings) *activitypb.ActivitySettings {
	if in == nil {
		return nil
	}
	out := &activitypb.ActivitySettings{
		DepositRewardType: in.DepositRewardType,
		BetClaimMode:      in.BetClaimMode,
		BonusMode:         in.BonusMode,
		RewardSchemas:     toPBRewardSchemas(in.RewardSchemas),
	}
	if in.ReferConfig != nil {
		out.ReferConfig = &activitypb.ReferConfig{
			MinDepositAmount: in.ReferConfig.MinDepositAmount,
			MinBetAmount:     in.ReferConfig.MinBetAmount,
			IpLimit:          in.ReferConfig.IpLimit,
			DeviceLimit:      in.ReferConfig.DeviceLimit,
		}
	}
	if in.LuckWheel != nil {
		out.LuckWheel = &activitypb.LuckWheelConfig{
			WheelTriggerMode: in.LuckWheel.WheelTriggerMode,
			BetSchemas:       toPBRewardSchemas(in.LuckWheel.BetSchemas),
			DepositSchemas:   toPBRewardSchemas(in.LuckWheel.DepositSchemas),
		}
	}
	return out
}

func toTypesActivitySettings(in *activitypb.ActivitySettings) *types.ActivitySettings {
	if in == nil {
		return nil
	}
	out := &types.ActivitySettings{
		DepositRewardType: in.GetDepositRewardType(),
		BetClaimMode:      in.GetBetClaimMode(),
		BonusMode:         in.GetBonusMode(),
		RewardSchemas:     toTypesRewardSchemas(in.GetRewardSchemas()),
	}
	if in.ReferConfig != nil {
		out.ReferConfig = &types.ActivityReferConfig{
			MinDepositAmount: in.GetReferConfig().GetMinDepositAmount(),
			MinBetAmount:     in.GetReferConfig().GetMinBetAmount(),
			IpLimit:          in.GetReferConfig().GetIpLimit(),
			DeviceLimit:      in.GetReferConfig().GetDeviceLimit(),
		}
	}
	if in.LuckWheel != nil {
		out.LuckWheel = &types.ActivityLuckWheelConfig{
			WheelTriggerMode: in.GetLuckWheel().GetWheelTriggerMode(),
			BetSchemas:       toTypesRewardSchemas(in.GetLuckWheel().GetBetSchemas()),
			DepositSchemas:   toTypesRewardSchemas(in.GetLuckWheel().GetDepositSchemas()),
		}
	}
	return out
}

func toPBRewardSchemas(items []types.ActivityRewardSchema) []*activitypb.RewardSchema {
	if len(items) == 0 {
		return nil
	}
	out := make([]*activitypb.RewardSchema, 0, len(items))
	for _, item := range items {
		out = append(out, &activitypb.RewardSchema{
			Id:        item.Id,
			Threshold: item.Threshold,
			Amount:    item.Amount,
			Rate:      item.Rate,
			MaxBonus:  item.MaxBonus,
			MinBonus:  item.MinBonus,
			Count:     item.Count,
		})
	}
	return out
}

func toTypesRewardSchemas(items []*activitypb.RewardSchema) []types.ActivityRewardSchema {
	if len(items) == 0 {
		return nil
	}
	out := make([]types.ActivityRewardSchema, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, types.ActivityRewardSchema{
			Id:        item.GetId(),
			Threshold: item.GetThreshold(),
			Amount:    item.GetAmount(),
			Rate:      item.GetRate(),
			MaxBonus:  item.GetMaxBonus(),
			MinBonus:  item.GetMinBonus(),
			Count:     item.GetCount(),
		})
	}
	return out
}

func toPBActivityI18nItems(items []types.ActivityI18nItem) []*activitypb.ActivityI18NItem {
	if len(items) == 0 {
		return nil
	}
	out := make([]*activitypb.ActivityI18NItem, 0, len(items))
	for _, item := range items {
		out = append(out, &activitypb.ActivityI18NItem{
			Lang:        item.Lang,
			Name:        item.Name,
			Rule:        item.Rule,
			Description: item.Description,
			Image:       item.Image,
		})
	}
	return out
}

func toTypesActivityI18nItems(items []*activitypb.ActivityI18NItem) []types.ActivityI18nItem {
	if len(items) == 0 {
		return nil
	}
	out := make([]types.ActivityI18nItem, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, types.ActivityI18nItem{
			Lang:        item.GetLang(),
			Name:        item.GetName(),
			Rule:        item.GetRule(),
			Description: item.GetDescription(),
			Image:       item.GetImage(),
		})
	}
	return out
}
