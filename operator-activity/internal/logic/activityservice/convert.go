package activityservicelogic

import (
	"encoding/json"
	"time"

	"oa.98ent.com/p9/operator-activity/modules/activity/repository"
	"oa.98ent.com/p9/operator-activity/modules/activity/repository/admin"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"
	"oa.98ent.com/p9/operator-activity/pb/operatoractivityrpc/activitypb"
)

func toActivityDetail(m *repository.ActivityModel, i18nList []admin.ActivityI18nInput) (*activitypb.ActivityDetail, error) {
	if m == nil {
		return nil, nil
	}
	settings, err := fromActivitySettingsJSON(m.Settings)
	if err != nil {
		return nil, err
	}

	out := &activitypb.ActivityDetail{
		Id:                 m.ID,
		Type:               int32(m.Type),
		Status:             int32(m.Status),
		ApplyMode:          int32(m.ApplyMode),
		CycleMode:          int32(m.CycleMode),
		AllowedFixedLevel:  m.AllowedFixedLevel,
		AllowedAutoLevel:   m.AllowedAutoLevel,
		AllowedClients:     m.AllowedClients,
		AllowedGameIds:     m.AllowedGameIds,
		AllowedAgentIds:    m.AllowedAgentIds,
		ApplyLimit:         m.ApplyLimit,
		ClaimLimit:         m.ClaimLimit,
		IsPermanent:        m.IsPermanent,
		OpenAtSec:          timePtrToSec(m.OpenAt),
		CloseAtSec:         timePtrToSec(m.CloseAt),
		ReleaseAtSec:       timePtrToSec(m.ReleaseAt),
		AutoAudit:          m.AutoAudit,
		AutoAuditMaxAmt:    m.AutoAuditMaxAmt,
		PayoutMode:         int32(m.PayoutMode),
		PayoutTiming:       int32(m.PayoutTiming),
		Section:            int32(m.Section),
		ClaimExpireSec:     m.ClaimExpireSec,
		Settings:           settings,
		WageringMultiplier: int32(m.WageringMultiplier),
		WageringBase:       int32(m.WageringBase),
		ShowInList:         m.ShowInList,
		ShowInSidebar:      m.ShowInSidebar,
		SortWeight:         int32(m.SortWeight),
		CreateUid:          m.CreateUid,
		UpdateUid:          m.UpdateUid,
		Version:            int32(m.Version),
		CreatedAtSec:       m.CreatedAt.Unix(),
		UpdatedAtSec:       m.UpdatedAt.Unix(),
		I18N:               toActivityI18nItems(i18nList),
	}
	return out, nil
}

func toActivityListItem(row admin.ActivityListRow) *activitypb.ActivityListItem {
	return &activitypb.ActivityListItem{
		Id:                 row.ID,
		Type:               int32(row.Type),
		Status:             int32(row.Status),
		Name:               row.Name,
		Image:              row.Image,
		AllowedClients:     row.AllowedClients,
		OpenAtSec:          timePtrToSec(row.OpenAt),
		CloseAtSec:         timePtrToSec(row.CloseAt),
		ReleaseAtSec:       timePtrToSec(row.ReleaseAt),
		SortWeight:         int32(row.SortWeight),
		ShowInList:         row.ShowInList,
		UpdatedAtSec:       row.UpdatedAt.Unix(),
		AutoAudit:          row.AutoAudit,
		AutoAuditMaxAmt:    row.AutoAuditMaxAmt,
		AllowedFixedLevel:  row.AllowedFixedLevel,
		AllowedAutoLevel:   row.AllowedAutoLevel,
		WageringMultiplier: int32(row.WageringMultiplier),
		AllowedGameIds:     row.AllowedGameIds,
		IsPermanent:        row.IsPermanent,
		CreatedAtSec:       row.CreatedAt.Unix(),
		CreateUid:          row.CreateUid,
		UpdateUid:          row.UpdateUid,
		Section:            int32(row.Section),
	}
}

func toActivityI18nItems(list []admin.ActivityI18nInput) []*activitypb.ActivityI18NItem {
	if len(list) == 0 {
		return nil
	}
	out := make([]*activitypb.ActivityI18NItem, 0, len(list))
	for _, item := range list {
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

func fromActivitySettingsJSON(raw string) (*activitypb.ActivitySettings, error) {
	if raw == "" || raw == "{}" {
		return &activitypb.ActivitySettings{}, nil
	}
	var settings types.ActivitySettings
	if err := json.Unmarshal([]byte(raw), &settings); err != nil {
		return nil, err
	}
	return fromActivitySettings(settings), nil
}

func fromActivitySettings(in types.ActivitySettings) *activitypb.ActivitySettings {
	out := &activitypb.ActivitySettings{
		DepositRewardType: int32(in.DepositRewardType),
		BetClaimMode:      int32(in.BetClaimMode),
		BonusMode:         int32(in.BonusMode),
		RewardSchemas:     fromRewardSchemas(in.RewardSchemas),
	}
	if in.ReferConfig != nil {
		out.ReferConfig = &activitypb.ReferConfig{
			MinDepositAmount: in.ReferConfig.MinDepositAmount,
			MinBetAmount:     in.ReferConfig.MinBetAmount,
			IpLimit:          int32(in.ReferConfig.IPLimit),
			DeviceLimit:      int32(in.ReferConfig.DeviceLimit),
		}
	}
	if in.LuckWheel != nil {
		out.LuckWheel = &activitypb.LuckWheelConfig{
			WheelTriggerMode: int32(in.LuckWheel.WheelTriggerMode),
			BetSchemas:       fromRewardSchemas(in.LuckWheel.BetSchemas),
			DepositSchemas:   fromRewardSchemas(in.LuckWheel.DepositSchemas),
		}
	}
	return out
}

func fromRewardSchemas(items []types.RewardSchema) []*activitypb.RewardSchema {
	if len(items) == 0 {
		return nil
	}
	out := make([]*activitypb.RewardSchema, 0, len(items))
	for _, item := range items {
		out = append(out, &activitypb.RewardSchema{
			Id:        item.ID,
			Threshold: item.Threshold,
			Amount:    item.Amount,
			Rate:      int32(item.Rate),
			MaxBonus:  item.MaxBonus,
			MinBonus:  item.MinBonus,
			Count:     int32(item.Count),
		})
	}
	return out
}

func timePtrToSec(t *time.Time) *int64 {
	if t == nil || t.IsZero() {
		return nil
	}
	sec := t.Unix()
	return &sec
}
