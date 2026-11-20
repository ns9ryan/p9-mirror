package types

import "fmt"

// EnumItem 枚举选项；Label 为 front 词典 trans_key（如 activity.type.1）
type EnumItem struct {
	Key   string
	Value int32
}

// EnumParams 活动表单枚举参数（按字段分组）
type EnumParams struct {
	Type              []EnumItem
	Status            []EnumItem
	ApplyMode         []EnumItem
	CycleMode         []EnumItem
	PayoutMode        []EnumItem
	PayoutTiming      []EnumItem
	BonusMode         []EnumItem
	BetClaimMode      []EnumItem
	DepositRewardType []EnumItem
	WheelTriggerMode  []EnumItem
	WageringBase      []EnumItem
	ClientType        []EnumItem
	ApplyLimitType    []EnumItem
	ClaimLimitType    []EnumItem
	Section           []EnumItem
}

// GetEnumParams 返回活动表单枚举参数；Label 为 i18n trans_key，由调用方词典翻译。
func GetEnumParams() EnumParams {
	return EnumParams{
		Type: buildEnumItems("type",
			int32(ActivityTypeDeposit),
			int32(ActivityTypeWagering),
			int32(ActivityTypeReferralMember),
			int32(ActivityTypeLuckyWheel),
			int32(ActivityTypeAssist),
			int32(ActivityTypeBlindBox),
			int32(ActivityTypeNewbieBonus),
			int32(ActivityTypeUniversalAgent),
			int32(ActivityTypeRescueFund),
			int32(ActivityTypeLuckyBet),
			int32(ActivityTypeVIPBonus),
			int32(ActivityTypeRedPacketRain),
			int32(ActivityTypeImageTextAd),
			int32(ActivityTypeCheckIn),
		),
		Status: buildEnumItems("status",
			int32(ActivityStatusDraft),
			int32(ActivityStatusOngoing),
			int32(ActivityStatusClosed),
		),
		ApplyMode: buildEnumItems("applyMode",
			int32(ApplyAuto),
			int32(ApplyManual),
		),
		CycleMode: buildEnumItems("cycleMode",
			int32(CycleNone),
			int32(CycleDaily),
			int32(CycleWeekly),
			int32(CycleMonthly),
			int32(CycleInfinite),
		),
		PayoutMode: buildEnumItems("payoutMode",
			int32(PayoutManualExpireVoid),
			int32(PayoutManualExpireAuto),
			int32(PayoutAuto),
		),
		PayoutTiming: buildEnumItems("payoutTiming",
			int32(PayoutRealtime),
			int32(PayoutNextDay),
		),
		BonusMode: buildEnumItems("bonusMode",
			int32(BonusFixed),
			int32(BonusPercentage),
			int32(BonusFixedTiered),
			int32(BonusPercentageTiered),
			int32(BonusSeparate),
		),
		BetClaimMode: buildEnumItems("betClaimMode",
			int32(BetClaimModeByItem),
			int32(BetClaimModeByHighest),
			int32(BetClaimModeByCumulative),
		),
		DepositRewardType: buildEnumItems("depositRewardType",
			int32(DepositFirst),
			int32(DepositSecond),
			int32(DepositCumulative),
			int32(DepositSingle),
		),
		WheelTriggerMode: buildEnumItems("wheelTriggerMode",
			int32(WheelTriggerBet),
			int32(WheelTriggerDeposit),
		),
		WageringBase: buildEnumItems("wageringBase",
			int32(WageringBasePrincipalAndBonus),
			int32(WageringBaseBonusOnly),
		),
		ClientType: buildEnumItems("clientType",
			int32(ClientTypeWeb),
			int32(ClientTypeMobile),
			int32(ClientTypeAndroid),
			int32(ClientTypeIOS),
		),
		ApplyLimitType: buildEnumItems("applyLimitType",
			int32(ApplyLimitTypeUdid),
			int32(ApplyLimitTypeIP),
		),
		ClaimLimitType: buildEnumItems("claimLimitType",
			int32(ClaimLimitTypeDeposit),
			int32(ClaimLimitTypeBank),
			int32(ClaimLimitTypeSms),
			int32(ClaimLimitTypeEmail),
			int32(ClaimLimitTypeRealname),
		),
		Section: buildEnumItems("section",
			int32(SectionLobby),
			int32(SectionElectronic),
			int32(SectionSports),
		),
	}
}

func buildEnumItems(group string, values ...int32) []EnumItem {
	out := make([]EnumItem, 0, len(values))
	for _, v := range values {
		out = append(out, EnumItem{
			Key:   fmt.Sprintf("activity.%s.%d", group, v),
			Value: v,
		})
	}
	return out
}
