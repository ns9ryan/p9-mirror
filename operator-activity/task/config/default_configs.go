package config

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"oa.98ent.com/p9/operator-activity/ent"
	"oa.98ent.com/p9/operator-activity/ent/taskconfig"
	"oa.98ent.com/p9/operator-activity/task/constants"
	"oa.98ent.com/p9/operator-base/rpc/client/operatorservice"
)

// DefaultTaskConfigs 定义默认的任务配置
type DefaultTaskConfigs struct {
	opCode string
}

// NewDefaultTaskConfigs 创建默认配置管理器
func NewDefaultTaskConfigs(opCode string) *DefaultTaskConfigs {
	return &DefaultTaskConfigs{
		opCode: opCode,
	}
}

func InitializeAllDefaultConfigs(ctx context.Context, db *ent.Client, operatorBaseClient zrpc.Client) {
	operatorResp, err := operatorservice.NewOperatorService(operatorBaseClient).GetAllOperatorTimezoneInfo(ctx, nil)
	if err != nil {
		logx.Errorf("获取所有分站时区信息失败: %v", err)
	}
	operatorTimezoneList := operatorResp.GetOperators()
	logx.Infof("获取到的所有分站时区信息: %v", operatorTimezoneList)
	for _, operatorTimezoneInfo := range operatorTimezoneList {
		opcode := operatorTimezoneInfo.Code
		operatorTimezone := operatorTimezoneInfo.TimezoneCode
		defaultConfigs := NewDefaultTaskConfigs(opcode)
		if err := defaultConfigs.InitializeDefaultConfigs(ctx, db, opcode, operatorTimezone); err != nil {
			logx.Errorf("初始化默认任务配置失败 (opcode=%s): %v", opcode, err)
		}
	}
}

// InitializeDefaultConfigs 初始化默认任务配置
// 如果任务配置表为空，则插入默认配置
func (d *DefaultTaskConfigs) InitializeDefaultConfigs(ctx context.Context, db *ent.Client, operatorCode string, timezone string) error {
	// 检查是否已有任务配置
	count, err := db.TaskConfig.
		Query().
		Where(taskconfig.OperatorCodeEQ(operatorCode)).
		Count(ctx)
	if err != nil {
		return err
	}

	if count > 0 {
		// 已有配置，无需初始化
		return nil
	}

	// 获取默认配置列表
	configs := d.getDefaultConfigs()

	// 批量创建配置
	for _, cfg := range configs {
		_, err := db.TaskConfig.Create().
			SetOperatorCode(operatorCode).
			SetTaskCode(cfg.TaskCode).
			SetCategory(cfg.Category).
			SetName(cfg.Name).
			SetNameI18n(cfg.NameI18n).
			SetDescription(cfg.Description).
			SetDescriptionI18n(cfg.DescriptionI18n).
			SetStatus(constants.TaskConfigStatusEnabled).
			SetCompletionAmount(cfg.CompletionAmount).
			SetRewardAmount(cfg.RewardAmount).
			SetPayoutMode(cfg.PayoutMode).
			SetSort(cfg.Sort).
			SetTimezone(timezone).
			Save(ctx)
		if err != nil {
			return err
		}
	}
	logx.Infof("分站 %s 默认任务配置初始化完成", operatorCode)
	return nil
}

func (d *DefaultTaskConfigs) GetTaskConfigMap(ctx context.Context, db *ent.Client, operatorCode string) (map[string]*ent.TaskConfig, error) {
	configs, err := db.TaskConfig.Query().
		Where(
			taskconfig.StatusEQ(constants.TaskConfigStatusEnabled),
			taskconfig.OperatorCodeEQ(operatorCode),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}
	if len(configs) == 0 {
		return nil, nil
	}
	configMap := make(map[string]*ent.TaskConfig)
	for _, cfg := range configs {
		configMap[cfg.TaskCode] = cfg
	}
	return configMap, nil
}

// DefaultConfig 默认任务配置
type DefaultConfig struct {
	TaskCode         string
	Category         string
	Name             string
	NameI18n         string
	Description      string
	DescriptionI18n  string
	CompletionAmount int
	RewardAmount     int64
	RewardCurrency   string
	PayoutMode       string
	Sort             int64
}

// getDefaultConfigs 获取所有默认配置
func (d *DefaultTaskConfigs) getDefaultConfigs() []DefaultConfig {
	return []DefaultConfig{
		// ============ 事实型任务 ============
		// REGISTER - 注册账号
		{
			TaskCode:         constants.TaskCodeRegister,
			Category:         constants.TaskCategoryOnce,
			Name:             "注册账号",
			NameI18n:         "{\"zh-CN\":\"注册账号\",\"en-US\":\"Register Account\"}",
			Description:      "完成会员注册即可获得奖励",
			DescriptionI18n:  "{\"zh-CN\":\"完成会员注册即可获得奖励\",\"en-US\":\"Complete member registration to receive rewards\"}",
			CompletionAmount: 0,
			RewardAmount:     5000, // 50元
			RewardCurrency:   "USD",
			PayoutMode:       constants.PayoutModeManual,
			Sort:             100,
		},

		// SET_BIRTHDATE - 设置生日
		{
			TaskCode:         constants.TaskCodeSetBirthdate,
			Category:         constants.TaskCategoryOnce,
			Name:             "设置生日",
			NameI18n:         "{\"zh-CN\":\"设置生日\",\"en-US\":\"Set Birthdate\"}",
			Description:      "在个人信息中设置生日",
			DescriptionI18n:  "{\"zh-CN\":\"在个人信息中设置生日\",\"en-US\":\"Set birthdate in personal information\"}",
			CompletionAmount: 0,
			RewardAmount:     1000, // 10元
			RewardCurrency:   "USD",
			PayoutMode:       constants.PayoutModeManual,
			Sort:             90,
		},

		// SET_REALNAME - 设置真实姓名
		{
			TaskCode:         constants.TaskCodeSetRealname,
			Category:         constants.TaskCategoryOnce,
			Name:             "设置真实姓名",
			NameI18n:         "{\"zh-CN\":\"设置真实姓名\",\"en-US\":\"Set Real Name\"}",
			Description:      "在个人信息中设置真实姓名",
			DescriptionI18n:  "{\"zh-CN\":\"在个人信息中设置真实姓名\",\"en-US\":\"Set real name in personal information\"}",
			CompletionAmount: 0,
			RewardAmount:     1000, // 10元
			RewardCurrency:   "USD",
			PayoutMode:       constants.PayoutModeManual,
			Sort:             80,
		},

		// SET_AVATAR_NICKNAME - 设置头像和昵称
		{
			TaskCode:         constants.TaskCodeSetAvatarNickname,
			Category:         constants.TaskCategoryOnce,
			Name:             "设置头像和昵称",
			NameI18n:         "{\"zh-CN\":\"设置头像和昵称\",\"en-US\":\"Set Avatar and Nickname\"}",
			Description:      "设置个人头像和昵称",
			DescriptionI18n:  "{\"zh-CN\":\"设置个人头像和昵称\",\"en-US\":\"Set personal avatar and nickname\"}",
			CompletionAmount: 0,
			RewardAmount:     800, // 8元
			RewardCurrency:   "USD",
			PayoutMode:       constants.PayoutModeManual,
			Sort:             70,
		},

		// BIND_MOBILE - 绑定手机号
		{
			TaskCode:         constants.TaskCodeBindMobile,
			Category:         constants.TaskCategoryOnce,
			Name:             "绑定手机号",
			NameI18n:         "{\"zh-CN\":\"绑定手机号\",\"en-US\":\"Bind Mobile\"}",
			Description:      "绑定有效的手机号码",
			DescriptionI18n:  "{\"zh-CN\":\"绑定有效的手机号码\",\"en-US\":\"Bind a valid mobile number\"}",
			CompletionAmount: 0,
			RewardAmount:     1500, // 15元
			RewardCurrency:   "USD",
			PayoutMode:       constants.PayoutModeManual,
			Sort:             85,
		},

		// BIND_BANK_CARD - 绑定银行卡
		{
			TaskCode:         constants.TaskCodeBindBankCard,
			Category:         constants.TaskCategoryOnce,
			Name:             "绑定银行卡",
			NameI18n:         "{\"zh-CN\":\"绑定银行卡\",\"en-US\":\"Bind Bank Card\"}",
			Description:      "绑定银行卡用于提现",
			DescriptionI18n:  "{\"zh-CN\":\"绑定银行卡用于提现\",\"en-US\":\"Bind bank card for withdrawal\"}",
			CompletionAmount: 0,
			RewardAmount:     2000, // 20元
			RewardCurrency:   "USD",
			PayoutMode:       constants.PayoutModeManual,
			Sort:             80,
		},

		// BIND_ID_CPF - 绑定ID/CPF认证
		{
			TaskCode:         constants.TaskCodeBindIDCPF,
			Category:         constants.TaskCategoryOnce,
			Name:             "身份认证",
			NameI18n:         "{\"zh-CN\":\"身份认证\",\"en-US\":\"Identity Verification\"}",
			Description:      "完成ID或CPF身份认证",
			DescriptionI18n:  "{\"zh-CN\":\"完成ID或CPF身份认证\",\"en-US\":\"Complete ID or CPF identity verification\"}",
			CompletionAmount: 0,
			RewardAmount:     2500, // 25元
			RewardCurrency:   "USD",
			PayoutMode:       constants.PayoutModeManual,
			Sort:             90,
		},

		// VERIFY_EMAIL - 验证邮箱
		{
			TaskCode:         constants.TaskCodeVerifyEmail,
			Category:         constants.TaskCategoryOnce,
			Name:             "验证邮箱",
			NameI18n:         "{\"zh-CN\":\"验证邮箱\",\"en-US\":\"Verify Email\"}",
			Description:      "验证邮箱地址",
			DescriptionI18n:  "{\"zh-CN\":\"验证邮箱地址\",\"en-US\":\"Verify email address\"}",
			CompletionAmount: 0,
			RewardAmount:     800, // 8元
			RewardCurrency:   "USD",
			PayoutMode:       constants.PayoutModeManual,
			Sort:             60,
		},

		// SET_SOCIAL_MEDIA - 设置社交媒体
		{
			TaskCode:         constants.TaskCodeSetSocialMedia,
			Category:         constants.TaskCategoryOnce,
			Name:             "关注社交账号",
			NameI18n:         "{\"zh-CN\":\"关注社交账号\",\"en-US\":\"Follow Social Media Account\"}",
			Description:      "关注我们的社交媒体账号",
			DescriptionI18n:  "{\"zh-CN\":\"关注我们的社交媒体账号\",\"en-US\":\"Follow our social media account\"}",
			CompletionAmount: 0,
			RewardAmount:     500, // 5元
			RewardCurrency:   "USD",
			PayoutMode:       constants.PayoutModeManual,
			Sort:             50,
		},

		// ============ 增量型任务 ============
		// DAILY_DEPOSIT - 每日充值
		{
			TaskCode:         constants.TaskCodeDailyDeposit,
			Category:         constants.TaskCategoryDaily,
			Name:             "每日充值",
			NameI18n:         "{\"zh-CN\":\"每日充值\",\"en-US\":\"Daily Deposit\"}",
			Description:      "每天充值满足条件即可获得奖励",
			DescriptionI18n:  "{\"zh-CN\":\"每天充值满足条件即可获得奖励\",\"en-US\":\"Deposit daily to receive rewards\"}",
			CompletionAmount: 0,
			RewardAmount:     1000, // 10元
			RewardCurrency:   "USD",
			PayoutMode:       constants.PayoutModeAuto,
			Sort:             200,
		},

		// BET_ANY_GAME - 任意游戏下注
		{
			TaskCode:         constants.TaskCodeBetAnyGame,
			Category:         constants.TaskCategoryDaily,
			Name:             "每日下注",
			NameI18n:         "{\"zh-CN\":\"每日下注\",\"en-US\":\"Daily Bet\"}",
			Description:      "每天有效投注满足条件即可获得奖励",
			DescriptionI18n:  "{\"zh-CN\":\"每天有效投注满足条件即可获得奖励\",\"en-US\":\"Place daily bets to receive rewards\"}",
			CompletionAmount: 0,
			RewardAmount:     500, // 5元
			RewardCurrency:   "USD",
			PayoutMode:       constants.PayoutModeAuto,
			Sort:             150,
		},

		// ============ 特殊任务 ============
		// FIRST_WITHDRAWAL - 首次提现
		{
			TaskCode:         constants.TaskCodeFirstWithdrawal,
			Category:         constants.TaskCategoryOnce,
			Name:             "首次提现",
			NameI18n:         "{\"zh-CN\":\"首次提现\",\"en-US\":\"First Withdrawal\"}",
			Description:      "完成首次提现申请",
			DescriptionI18n:  "{\"zh-CN\":\"完成首次提现申请\",\"en-US\":\"Complete the first withdrawal request\"}",
			CompletionAmount: 0,
			RewardAmount:     3000, // 30元
			RewardCurrency:   "USD",
			PayoutMode:       constants.PayoutModeAuto,
			Sort:             120,
		},

		// PROMO_REGISTER - 邀请下线会员
		{
			TaskCode:         constants.TaskCodePromoRegister,
			Category:         constants.TaskCategoryWeekly,
			Name:             "邀请好友",
			NameI18n:         "{\"zh-CN\":\"邀请好友\",\"en-US\":\"Invite Friends\"}",
			Description:      "每周邀请新会员注册,会员充值一定金额后，每邀请1人获得奖励",
			DescriptionI18n:  "{\"zh-CN\":\"每周邀请新会员注册,会员充值一定金额后，每邀请1人获得奖励\",\"en-US\":\"Invite new members to register each week, receive rewards for each invite after they deposit a certain amount\"}",
			CompletionAmount: 5000, // 邀请的好友充值 50 USD后才算完成任务
			RewardAmount:     2000, // 20元/人
			RewardCurrency:   "USD",
			PayoutMode:       constants.PayoutModeManual,
			Sort:             110,
		},

		// APPLY_AGENT - 申请成为代理
		{
			TaskCode:         constants.TaskCodeApplyAgent,
			Category:         constants.TaskCategoryOnce,
			Name:             "成为代理",
			NameI18n:         "{\"zh-CN\":\"成为代理\",\"en-US\":\"Become an Agent\"}",
			Description:      "成功申请并成为代理",
			DescriptionI18n:  "{\"zh-CN\":\"成功申请并成为代理\",\"en-US\":\"Successfully apply and become an agent\"}",
			CompletionAmount: 0,
			RewardAmount:     10000, // 100元
			RewardCurrency:   "USD",
			PayoutMode:       constants.PayoutModeManual,
			Sort:             130,
		},

		// CONTINUOUS_CHAT - 连续参与聊天
		{
			TaskCode:         constants.TaskCodeJoinChat,
			Category:         constants.TaskCategoryWeekly,
			Name:             "参与聊天",
			NameI18n:         "{\"zh-CN\":\"参与聊天\",\"en-US\":\"Participate in Chat\"}",
			Description:      "每周只要参与一次聊天",
			DescriptionI18n:  "{\"zh-CN\":\"每周只要参与一次聊天即可获得奖励\",\"en-US\":\"Participate in chat once each week to receive rewards\"}",
			CompletionAmount: 0,
			RewardAmount:     5000, // 50元
			RewardCurrency:   "USD",
			PayoutMode:       constants.PayoutModeManual,
			Sort:             95,
		},
	}
}
