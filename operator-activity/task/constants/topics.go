package constants

import "oa.98ent.com/p9/common/xkafka/topic"

// // EventType 定义事件类型常量
// const (
//
//	EventTypeMemberRegistered          = "task.member.registered"                  // 会员注册事件 ✓
//	EventTypeDepositSucceeded          = "task.payment.deposit_succeeded"          // 充值成功事件 ✓ 目前使用充值模拟表
//	EventTypeBirthdateUpdated          = "task.member.birthdate_updated"           // 生日更新事件 ✓
//	EventTypeRealnameVerified          = "task.member.realname_verified"           // 实名认证事件 ✓
//	EventTypeAvatarNicknameUpdated     = "task.member.avatar_nickname_updated"     // 头像与昵称更新事件 ✓
//	EventTypeMobileBound               = "task.member.mobile_bound"                // 手机绑定事件 ✓
//	EventTypeIDCPFVerified             = "task.member.id_cpf_verified"             // ID/CPF 认证事件 ✓
//	EventTypeEmailVerified             = "task.member.email_verified"              // 邮箱验证事件 ✓
//	EventTypeSocialMediaUpdated        = "task.member.social_media_updated"        // 社交媒体更新事件 ✓
//	EventTypeValidBetSettled           = "task.bet.valid_bet_settled"              // 有效下注结算事件 ✓ 每日定时结算一次 目前使用游戏订单模拟表
//	EventTypeReferralRegisteredDeposit = "task.member.referral_registered_deposit" // 下线会员注册并充值指定金额事件 ✓ 每周定时结算一次，目前使用虚拟RPC查询
//	EventTypeAgentApplicationApproved  = "task.agent.application_approved"         // 代理申请通过事件 ✓
//	EventTypeWithdrawalSucceeded       = "task.withdrawal.succeeded"               // 提现成功事件 ✓ 目前使用提现模拟表
//	EventTypeChatMessageAccepted       = "task.chat.message_accepted"              // 聊天消息被接受事件 ✓ 目前使用聊天消息模拟表
//
// )
// EventType 定义事件类型常量
const (
	EventTypeMemberRegistered          = topic.GameUserRegister          // 会员注册事件 ✓
	EventTypeDepositSucceeded          = topic.DepositSucceeded          // 充值成功事件 ✓ 目前使用充值模拟表
	EventTypeBirthdateUpdated          = topic.BirthdateUpdated          // 生日更新事件 ✓
	EventTypeRealnameVerified          = topic.RealnameVerified          // 实名认证事件 ✓
	EventTypeAvatarNicknameUpdated     = topic.AvatarNicknameUpdated     // 头像与昵称更新事件 ✓
	EventTypeMobileBound               = topic.MobileBound               // 手机绑定事件 ✓
	EventTypeIDCPFVerified             = topic.IDCPFVerified             // ID/CPF 认证事件 ✓
	EventTypeEmailVerified             = topic.EmailVerified             // 邮箱验证事件 ✓
	EventTypeSocialMediaUpdated        = topic.SocialMediaUpdated        // 社交媒体更新事件 ✓
	EventTypeValidBetSettled           = topic.ValidBetSettled           // 有效下注结算事件 ✓ 每日定时结算一次 目前使用游戏订单模拟表
	EventTypeReferralRegisteredDeposit = topic.ReferralRegisteredDeposit // 下线会员注册并充值指定金额事件 ✓ 每周定时结算一次，目前使用虚拟RPC查询
	EventTypeAgentApplicationApproved  = topic.AgentApplicationApproved  // 代理申请通过事件 ✓
	EventTypeWithdrawalSucceeded       = topic.WithdrawalSucceeded       // 提现成功事件 ✓ 目前使用提现模拟表
	EventTypeChatMessageAccepted       = topic.ChatMessageAccepted       // 聊天消息被接受事件 ✓ 目前使用聊天消息模拟表
)

// Topics
const (
	TopicMemberRegistered          = topic.GameUserRegister
	TopicDepositSucceeded          = topic.DepositSucceeded
	TopicBirthdateUpdated          = topic.BirthdateUpdated
	TopicRealnameVerified          = topic.RealnameVerified
	TopicAvatarNicknameUpdated     = topic.AvatarNicknameUpdated
	TopicMobileBound               = topic.MobileBound
	TopicIDCPFVerified             = topic.IDCPFVerified
	TopicEmailVerified             = topic.EmailVerified
	TopicSocialMediaUpdated        = topic.SocialMediaUpdated
	TopicValidBetSettled           = topic.ValidBetSettled
	TopicReferralRegisteredDeposit = topic.ReferralRegisteredDeposit
	TopicAgentApplicationApproved  = topic.AgentApplicationApproved
	TopicWithdrawalSucceeded       = topic.WithdrawalSucceeded
	TopicChatMessageAccepted       = topic.ChatMessageAccepted
)

// topic对应的事件类型
var TopicToEventType = map[string]string{
	TopicMemberRegistered:          EventTypeMemberRegistered,
	TopicDepositSucceeded:          EventTypeDepositSucceeded,
	TopicBirthdateUpdated:          EventTypeBirthdateUpdated,
	TopicRealnameVerified:          EventTypeRealnameVerified,
	TopicAvatarNicknameUpdated:     EventTypeAvatarNicknameUpdated,
	TopicMobileBound:               EventTypeMobileBound,
	TopicIDCPFVerified:             EventTypeIDCPFVerified,
	TopicEmailVerified:             EventTypeEmailVerified,
	TopicSocialMediaUpdated:        EventTypeSocialMediaUpdated,
	TopicValidBetSettled:           EventTypeValidBetSettled,
	TopicReferralRegisteredDeposit: EventTypeReferralRegisteredDeposit,
	TopicAgentApplicationApproved:  EventTypeAgentApplicationApproved,
	TopicWithdrawalSucceeded:       EventTypeWithdrawalSucceeded,
	TopicChatMessageAccepted:       EventTypeChatMessageAccepted,
}

// AllTopics 所有订阅的 topic 列表
var AllTopics = []string{
	TopicMemberRegistered,
	TopicDepositSucceeded,
	TopicBirthdateUpdated,
	TopicRealnameVerified,
	TopicAvatarNicknameUpdated,
	TopicMobileBound,
	TopicIDCPFVerified,
	TopicEmailVerified,
	TopicSocialMediaUpdated,
	TopicValidBetSettled,
	TopicReferralRegisteredDeposit,
	TopicAgentApplicationApproved,
	TopicWithdrawalSucceeded,
	TopicChatMessageAccepted,
}
