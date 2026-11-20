package seed

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"oa.98ent.com/p9/operator-activity/ent"
	"oa.98ent.com/p9/operator-activity/ent/activity"
	"oa.98ent.com/p9/operator-activity/modules/activity/repository"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"
)

const (
	OperatorCode = "OP_617E72"
)

// Run 写入活动配置与测试数据（幂等：按 ID 存在则跳过）
func Run(client *ent.Client) error {
	if client == nil {
		return fmt.Errorf("ent client is nil")
	}
	ctx := context.Background()
	repo := repository.New(client)

	now := time.Now()
	promos := []repository.ActivityModel{
		firstDepositPromo(now),
		dailyBettingPromo(now),
		referralPromo(now),
		luckyWheelPromo(now),
	}
	for _, p := range promos {
		exists, err := client.Activity.Query().Where(activity.IDEQ(p.ID)).Exist(ctx)
		if err != nil {
			return fmt.Errorf("check promo %d: %w", p.ID, err)
		}
		if exists {
			continue
		}
		c := client.Activity.Create().
			SetOperatorCode(p.OperatorCode).
			SetType(int(p.Type)).
			SetStatus(int(p.Status)).
			SetApplyMode(int(p.ApplyMode)).
			SetCycleMode(int(p.CycleMode)).
			SetAllowedFixedLevel(p.AllowedFixedLevel).
			SetAllowedAutoLevel(p.AllowedAutoLevel).
			SetAllowedClients(p.AllowedClients).
			SetAllowedGameIds(p.AllowedGameIds).
			SetAllowedAgentIds(p.AllowedAgentIds).
			SetApplyLimit(p.ApplyLimit).
			SetClaimLimit(p.ClaimLimit).
			SetNillableOpenAt(p.OpenAt).
			SetNillableCloseAt(p.CloseAt).
			SetNillableReleaseAt(p.ReleaseAt).
			SetShowInList(p.ShowInList).
			SetShowInSidebar(p.ShowInSidebar).
			SetAutoAudit(p.AutoAudit).
			SetAutoAuditMaxAmt(p.AutoAuditMaxAmt).
			SetWageringMultiplier(p.WageringMultiplier).
			SetWageringBase(int(p.WageringBase)).
			SetPayoutMode(int(p.PayoutMode)).
			SetPayoutTiming(int(p.PayoutTiming)).
			SetClaimExpireSec(p.ClaimExpireSec).
			SetSettings(p.Settings)
		if _, err := c.Save(ctx); err != nil {
			return fmt.Errorf("create promo %d: %w", p.ID, err)
		}
		repo.DelOngoingActivitiesCache(p.OperatorCode, types.ActivityType(p.Type))
	}

	return nil
}

// firstDepositPromo 首充奖励：充 100 元以上返 50 元，自动派发
func firstDepositPromo(now time.Time) repository.ActivityModel {
	settings := types.ActivitySettings{
		DepositRewardType: types.DepositSingle,
		BonusMode:         types.BonusFixed, // 固定金额多档
		RewardSchemas: []types.RewardSchema{
			{ID: 1, Threshold: 100, Amount: 10},
			{ID: 2, Threshold: 300, Amount: 50},
			{ID: 3, Threshold: 500, Amount: 80},
			{ID: 4, Threshold: 1000, Amount: 160},
		},
	}
	p := buildActivity(1, types.ActivityTypeDeposit, now, settings, true, 100)
	p.PayoutMode = types.PayoutManualExpireAuto // 手动领取，过期自动领取
	p.PayoutTiming = types.PayoutNextDay        // 次日到账
	p.ClaimExpireSec = 600                      // 领取奖励过期时间 单位秒
	p.ApplyMode = types.ApplyAuto
	p.CycleMode = types.CycleNone
	p.WageringMultiplier = 1
	p.WageringBase = types.WageringBaseBonusOnly
	p.AllowedFixedLevel = []int64{1, 2, 3}   // 固定等级
	p.AllowedAutoLevel = []int64{0, 1, 2, 3} // 自动等级
	p.AllowedClients = []int64{1, 2, 3}      // 客户端类型
	p.AllowedGameIds = []int64{}             // 指定游戏ID 空表示所有游戏
	p.AllowedAgentIds = []int64{}            // 指定代理ID 空表示所有代理
	p.ApplyLimit = []int64{}                 // 参与活动限制 空表示无限制
	p.ClaimLimit = []int64{}                 // 领取奖励限制 空表示无限制
	return p
}

// dailyBettingPromo 每日投注活动：累计投注 500 元以上奖 30 元
func dailyBettingPromo(now time.Time) repository.ActivityModel {
	settings := types.ActivitySettings{
		BonusMode: types.BonusFixed,
		RewardSchemas: []types.RewardSchema{
			{ID: 1, Threshold: 100, Amount: 5},
			{ID: 2, Threshold: 200, Amount: 16},
			{ID: 3, Threshold: 300, Amount: 50},
			{ID: 4, Threshold: 500, Amount: 80},
			{ID: 5, Threshold: 1000, Amount: 160},
		},
	}
	p := buildActivity(2, types.ActivityTypeWagering, now, settings, true, 100)
	p.PayoutMode = types.PayoutManualExpireVoid // 手动领取，过期作废
	p.PayoutTiming = types.PayoutNextDay        // 次日到账
	p.ClaimExpireSec = 600
	p.WageringMultiplier = 1
	p.WageringBase = types.WageringBaseBonusOnly
	p.ApplyMode = types.ApplyAuto
	p.CycleMode = types.CycleNone
	p.AllowedFixedLevel = []int64{1, 2}   // 固定等级
	p.AllowedAutoLevel = []int64{0, 1, 2} // 自动等级
	p.AllowedClients = []int64{1, 2}      // 客户端类型
	p.AllowedGameIds = []int64{}          // 指定游戏ID 空表示所有游戏
	p.AllowedAgentIds = []int64{}         // 指定代理ID 空表示所有代理
	p.ApplyLimit = []int64{}              // 参与活动限制 空表示无限制
	p.ClaimLimit = []int64{}              // 领取奖励限制 空表示无限制
	return p
}

// referralPromo 推荐活动：邀请 3 人各奖 100，邀请 5 人奖 300（需被推荐人充值达标）
func referralPromo(now time.Time) repository.ActivityModel {
	settings := types.ActivitySettings{
		BonusMode: types.BonusFixed,
		ReferConfig: &types.ReferConfig{
			MinDepositAmount: 100,
			MinBetAmount:     50,
			IPLimit:          0,
			DeviceLimit:      0,
		},
		RewardSchemas: []types.RewardSchema{
			{ID: 1, Count: 1, Amount: 10},
			{ID: 2, Count: 2, Amount: 30},
			{ID: 3, Count: 3, Amount: 50},
			{ID: 4, Count: 4, Amount: 80},
			{ID: 5, Count: 5, Amount: 160},
		},
	}
	p := buildActivity(3, types.ActivityTypeReferralMember, now, settings, true, 100)
	p.PayoutMode = types.PayoutAuto
	p.PayoutTiming = types.PayoutRealtime
	p.ClaimExpireSec = 0
	p.WageringMultiplier = 1
	p.WageringBase = types.WageringBaseBonusOnly
	p.ApplyMode = types.ApplyAuto
	p.CycleMode = types.CycleNone
	p.AllowedFixedLevel = []int64{1, 2, 3}   // 固定等级
	p.AllowedAutoLevel = []int64{0, 1, 2, 3} // 自动等级
	p.AllowedClients = []int64{1, 2, 3}      // 客户端类型
	p.AllowedGameIds = []int64{}             // 指定游戏ID 空表示所有游戏
	p.AllowedAgentIds = []int64{}            // 指定代理ID 空表示所有代理
	p.ApplyLimit = []int64{}                 // 参与活动限制 空表示无限制
	p.ClaimLimit = []int64{}                 // 领取奖励限制 空表示无限制
	return p
}

// luckyWheelPromo 幸运转盘：累计有效投注 200/500/1000 各送 1 次抽奖；奖池按权重抽奖；次日到账+手动过期自动领
func luckyWheelPromo(now time.Time) repository.ActivityModel {
	settings := types.ActivitySettings{
		BonusMode: types.BonusFixed,
		LuckWheel: &types.LuckWheelConfig{
			WheelTriggerMode: types.WheelTriggerBet | types.WheelTriggerDeposit,
			BetSchemas: []types.RewardSchema{
				{ID: 1, Threshold: 100, Amount: 1},
				{ID: 2, Threshold: 200, Amount: 1},
				{ID: 3, Threshold: 500, Amount: 1},
			},
			DepositSchemas: []types.RewardSchema{
				{ID: 1, Threshold: 200, Amount: 1},
				{ID: 2, Threshold: 500, Amount: 1},
				{ID: 3, Threshold: 1000, Amount: 1},
			},
		},
		RewardSchemas: []types.RewardSchema{
			{ID: 1, Amount: 10, Rate: 45},
			{ID: 2, Amount: 20, Rate: 25},
			{ID: 3, Amount: 50, Rate: 15},
			{ID: 4, Amount: 100, Rate: 8},
			{ID: 5, Amount: 200, Rate: 5},
			{ID: 6, Amount: 500, Rate: 2},
		},
	}
	p := buildActivity(4, types.ActivityTypeLuckyWheel, now, settings, true, 100)
	p.PayoutMode = types.PayoutAuto
	p.PayoutTiming = types.PayoutNextDay
	p.ClaimExpireSec = 0
	p.WageringMultiplier = 1
	p.WageringBase = types.WageringBaseBonusOnly
	p.ApplyMode = types.ApplyAuto
	p.CycleMode = types.CycleDaily
	p.AllowedFixedLevel = []int64{1, 2, 3}   // 固定等级
	p.AllowedAutoLevel = []int64{0, 1, 2, 3} // 自动等级
	p.AllowedClients = []int64{1, 2, 3}      // 客户端类型
	p.AllowedGameIds = []int64{}             // 指定游戏ID 空表示所有游戏
	p.AllowedAgentIds = []int64{}            // 指定代理ID 空表示所有代理
	p.ApplyLimit = []int64{}                 // 参与活动限制 空表示无限制
	p.ClaimLimit = []int64{}                 // 领取奖励限制 空表示无限制
	return p
}

func buildActivity(id int64, promoType types.ActivityType, now time.Time, settings types.ActivitySettings, autoAudit bool, autoAuditMaxAmt int64) repository.ActivityModel {
	raw, _ := json.Marshal(settings)
	openAt := now.Add(-30 * 24 * time.Hour)
	closeAt := now.Add(30 * 24 * time.Hour)
	releaseAt := now
	return repository.ActivityModel{
		ID:              id,
		OperatorCode:    OperatorCode,
		Type:            promoType,
		Status:          types.ActivityStatusOngoing,
		OpenAt:          &openAt,
		CloseAt:         &closeAt,
		ReleaseAt:       &releaseAt,
		ShowInList:      true,
		AutoAudit:       autoAudit,
		AutoAuditMaxAmt: autoAuditMaxAmt,
		Settings:        string(raw),
	}
}
