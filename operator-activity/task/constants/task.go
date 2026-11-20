package constants

// TaskCode 定义所有任务编码
const (
	// REGISTER 注册账号
	TaskCodeRegister = "REGISTER"
	// DAILY_DEPOSIT 每日充值
	TaskCodeDailyDeposit = "DAILY_DEPOSIT"
	// SET_BIRTHDATE 设置生日
	TaskCodeSetBirthdate = "SET_BIRTHDATE"
	// SET_REALNAME 设置真实姓名
	TaskCodeSetRealname = "SET_REALNAME"
	// SET_AVATAR_NICKNAME 设置头像与昵称
	TaskCodeSetAvatarNickname = "SET_AVATAR_NICKNAME"
	// BIND_MOBILE 绑定手机号
	TaskCodeBindMobile = "BIND_MOBILE"
	// BIND_BANK_CARD 绑定银行卡
	TaskCodeBindBankCard = "BIND_BANK_CARD"
	// BIND_ID_CPF 绑定 ID/CPF
	TaskCodeBindIDCPF = "BIND_ID_CPF"
	// VERIFY_EMAIL 验证邮箱
	TaskCodeVerifyEmail = "VERIFY_EMAIL"
	// SET_SOCIAL_MEDIA 填写社交媒体账号
	TaskCodeSetSocialMedia = "SET_SOCIAL_MEDIA"
	// BET_ANY_GAME 任意游戏下注
	TaskCodeBetAnyGame = "BET_ANY_GAME"
	// PROMO_REGISTER 拓展一名下线会员
	TaskCodePromoRegister = "PROMO_REGISTER"
	// APPLY_AGENT 申请并成为代理
	TaskCodeApplyAgent = "APPLY_AGENT"
	// FIRST_WITHDRAWAL 首次提现
	TaskCodeFirstWithdrawal = "FIRST_WITHDRAWAL"
	// JOIN_CHAT 参与聊天
	TaskCodeJoinChat = "JOIN_CHAT"
)

// TaskCategory 定义任务分类
const (
	TaskCategoryOnce    = "ONCE"    // 一次性任务
	TaskCategoryDaily   = "DAILY"   // 每日任务
	TaskCategoryWeekly  = "WEEKLY"  // 每周任务
	TaskCategoryMonthly = "MONTHLY" // 每月任务
)

// 1=任务进行中|2=待派发|11=任务已拒绝|12=派发被拒绝|99=已派发
const (
	TaskStatusPendingClaim     = 1  // 待领取
	TaskStatusPendingDispatch  = 2  // 待派发
	TaskStatusRejected         = 11 // 任务已拒绝
	TaskStatusDispatchRejected = 12 // 派发被拒绝
	TaskStatusDispatched       = 99 // 已派发
)

// PayoutMode 定义派发方式
const (
	PayoutModeManual = "MANUAL"
	PayoutModeAuto   = "AUTO"
)

// redis 前缀（数据存储）
const (
	RedisPrefixTask        = "task:task"
	RedisPrefixPayout      = "task:payout"
	RedisPrefixTaskRetry   = "task:taskretry"
	RedisPrefixPayoutRetry = "task:payoutretry"
)

// redis 前缀（分布式锁）
const (
	RedisPrefixLockTask        = "task:lock:task"
	RedisPrefixLockPayout      = "task:lock:payout"
	RedisPrefixLockTaskRetry   = "task:lock:taskretry"
	RedisPrefixLockPayoutRetry = "task:lock:payoutretry"
)

// Batch 处理相关常量
const (
	// BatchMessageSize 单个批次最多包含的消息数
	BatchMessageSize = 100
	// BatchTimeWindowFormat 批次时间窗口格式（YYYYMMDDhhmmss）
	BatchTimeWindowFormat = "20060102150405"
)

// 分布式锁相关常量
const (
	// LockTTL 锁的过期时间（秒）
	LockTTL = 30
	// LockRenewalInterval 锁续期间隔（秒）
	LockRenewalInterval = 10
)

// 队列扫描间隔常量
const (
	// NormalQueueScanInterval 正常队列扫描间隔（秒）
	NormalQueueScanInterval = 1
	// RetryQueueScanInterval 重试队列扫描间隔（秒）
	RetryQueueScanInterval = 10
)

// 重试相关常量
const (
	// MaxTaskRetries 任务最大重试次数（包含初始尝试）
	MaxTaskRetries = 3
	// MaxPayoutRetries 奖励派发最大重试次数（包含初始尝试）
	MaxPayoutRetries = 6
)

// EventTypeToTaskCode 事件类型对应的任务编码，目前一个事件对应一个任务编码，如果要改为动态配置，后续可以考虑从数据库或配置中心加载
var EventTypeToTaskCode = map[string]string{
	EventTypeMemberRegistered:      TaskCodeRegister,
	EventTypeDepositSucceeded:      TaskCodeDailyDeposit,
	EventTypeBirthdateUpdated:      TaskCodeSetBirthdate,
	EventTypeRealnameVerified:      TaskCodeSetRealname,
	EventTypeAvatarNicknameUpdated: TaskCodeSetAvatarNickname,
	EventTypeMobileBound:           TaskCodeBindMobile,
	// EventTypeBankCardBound:            TaskCodeBindBankCard,
	EventTypeIDCPFVerified:             TaskCodeBindIDCPF,
	EventTypeEmailVerified:             TaskCodeVerifyEmail,
	EventTypeSocialMediaUpdated:        TaskCodeSetSocialMedia,
	EventTypeValidBetSettled:           TaskCodeBetAnyGame,
	EventTypeReferralRegisteredDeposit: TaskCodePromoRegister,
	EventTypeAgentApplicationApproved:  TaskCodeApplyAgent,
	EventTypeWithdrawalSucceeded:       TaskCodeFirstWithdrawal,
	EventTypeChatMessageAccepted:       TaskCodeJoinChat,
}

// TaskConfig 相关常量
const (
	// TaskConfigStatusEnabled 任务配置启用状态
	TaskConfigStatusEnabled = 1
	// TaskConfigStatusDisabled 任务配置禁用状态
	TaskConfigStatusDisabled = 2
)

// Progress 相关常量
const (
	// InitialProgressValue 进度初始值
	InitialProgressValue = int64(0)
	// DefaultTargetValue 进度默认目标值
	DefaultTargetValue = int64(1)
)

// Completion 相关常量
const (
	// InitialCompletionSeq 完成序号初始值
	InitialCompletionSeq = 1
)

// PeriodKey 相关常量
const (
	// PeriodKeyOnce 一次性任务周期键
	PeriodKeyOnce = "ALL"
	// PeriodKeyUnknown 未知周期键
	PeriodKeyUnknown = "UNKNOWN"
	// DateFormatDaily 每日周期日期格式
	DateFormatDaily = "20060102"
	// DateFormatMonthly 每月周期日期格式
	DateFormatMonthly = "200601"
	// DateFormatWeekly 每周周期日期格式
	DateFormatWeekly = "%d-W%02d"
)

const (
	// 会员表
	TableMember = "member"
	// 会员资料表
	TableMemberProfile = "member_profile"
	// 会员第三方身份表，包含社媒授权
	TableMemberOAuthIdentity = "member_oauth_identity"
	// 会员认证表
	TableMemberAuth = "member_auth"
	// 游戏充值表
	TableGameDeposit = "game_deposit"
	// 游戏提现表
	TableGameWithdraw = "game_withdraw"
	// 代理申请表
	TableAgentApply = "agent_apply"
	// 聊天日志表
	TableGameChatLog = "game_chat_log"
)

// 监听的表
var ObservedTables = []string{TableMember, TableMemberProfile, TableMemberOAuthIdentity, TableMemberAuth, TableGameDeposit, TableGameWithdraw, TableAgentApply, TableGameChatLog}

// TaskCodeList 所有任务编码列表
var TaskCodeList = []string{
	TaskCodeRegister,
	TaskCodeDailyDeposit,
	TaskCodeSetBirthdate,
	TaskCodeSetRealname,
	TaskCodeSetAvatarNickname,
	TaskCodeBindMobile,
	TaskCodeBindIDCPF,
	TaskCodeVerifyEmail,
	TaskCodeSetSocialMedia,
	TaskCodeBetAnyGame,
	TaskCodePromoRegister,
	TaskCodeApplyAgent,
	TaskCodeFirstWithdrawal,
	TaskCodeJoinChat,
}
