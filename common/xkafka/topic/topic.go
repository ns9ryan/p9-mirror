package topic

const (
	AgentApplicationCreated       = "agent-application.created"        // 会员提交分成代理申请（用于通知分站后台）
	AgentApplicationStatusChanged = "agent-application.status-changed" // 分成代理申请状态变更（自动审核或人工审核后的结果通知）

	AgentCreate = "agent.create" // 新增代理事件

	AgentUserLoginFail    = "agent-user.login.fail"
	AgentUserLoginSuccess = "agent-user.login.success"

	CryptoTransactionDetected = "crypto.transaction.detected" // 加密货币交易监控: 检测到交易

	GameUserActive                = "game-user.active"                  // 会员上线
	GameUserAlertRaise            = "game-user.alert.raise"             // 会员提醒触发(登录/存款/提款)
	GameUserAutoRankReCalculate   = "game-user.auto-rank.recalculate"   // 会员自动层级需要重新计算
	GameUserDailyFirstActive      = "game-user.daily.first-active"      // 会员每日首次活动(登录,心跳等活动)
	GameUserDepositOrderApproved  = "game-user.deposit-order.approved"  // 会员存款订单审核通过
	GameUserDepositOrderCreated   = "game-user.deposit-order.created"   // 会员存款订单生成
	GameUserGameOrderCreated      = "game-user.game-order.created"      // 会员游戏订单创建(只标记已下注,不管是否结算)
	GameUserGameOrderSettled      = "game-user.game-order.settled"      // 会员游戏订单结算
	GameUserHeartbeat             = "game-user.heartbeat"               // 会员心跳
	GameUserLoginFail             = "game-user.login.fail"              // 会员登录失败
	GameUserLoginSuccess          = "game-user.login.success"           // 会员登录成功
	GameUserAutoUnfreeze          = "game-user.auto-unfreeze"           // 会员自动解冻
	GameUserRegister              = "game-user.register"                // 会员账号注册
	GameUserSecurityCodeSendAck   = "game-user.security-code.send-ack"  // 会员验证码发送任务确认
	GameUserSecurityCodeSendReq   = "game-user.security-code.send-req"  // 会员验证码发送任务请求
	GameUserTaskClaimed           = "game-user.task.claimed"            // 会员任务奖励已手动领取
	GameUserTaskDone              = "game-user.task.done"               // 会员任务完成
	GameUserTaskPayout            = "game-user.task.payout"             // 会员任务奖励派发
	GameUserTaskPayoutDLQ         = "game-user.task.payout.dlq"         // 会员任务奖励派发失败队列
	GameUserWithdrawOrderApproved = "game-user.withdraw-order.approved" // 会员提款订单审核通过
	GameUserWithdrawOrderCreated  = "game-user.withdraw-order.created"  // 会员提款订单生成

	MessageTaskCreate = "message-task.create" // 创建通知任务

	NoticeSend = "notice.send" // 通用发送消息通知

	AgentCommissionPayout = "agent.commission.payout" // 代理佣金发放

	PartnerAdvertisementClose = "partner.advertisement.close" // 分站广告停用
	PartnerAdvertisementOpen  = "partner.advertisement.open"  // 分站广告开启

	PartnerCreated = "partner.created" // 分站创建成功

	PartnerMaintenanceAlertQuit = "partner.maintenance.alert-quit" // 分站踢掉所有会员提醒
	PartnerMaintenanceEnd       = "partner.maintenance.end"        // 分站停止维护
	PartnerMaintenanceForceQuit = "partner.maintenance.force-quit" // 分站踢掉所有会员
	PartnerMaintenanceStart     = "partner.maintenance.start"      // 分站开始维护

	PartnerNoticeRecallStatus = "partner-user.notice.recall.status" // 分站回访通知状态修改

	PartnerPromotionAutoApply = "partner.promotion.auto-apply" // 会员自动参与活动
	PartnerPromotionClose     = "partner.promotion.close"      // 分站活动关闭
	PartnerPromotionCycle     = "partner.promotion.cycle"      // 分站活动循环
	PartnerPromotionOpen      = "partner.promotion.open"       // 分站活动开启

	PartnerPromotionOrderApproved = "partner.promotion-order.approved" // 活动订单审核通过(人工或机器审核)
	PartnerPromotionOrderClaimed  = "partner.promotion-order.claimed"  // 活动订单奖励手动领取
	PartnerPromotionOrderCreated  = "partner.promotion-order.created"  // 活动订单创建通知：创建活动订单触发
	PartnerPromotionOrderPayout   = "partner.promotion-order.payout"   // 活动订单奖励发放

	PartnerPromotionRefVisit = "partner.promotion-ref.visit" // 推广链接访问(访问推广链接后触发)

	PartnerSettlementDaily       = "partner.settlement.daily"        // 分站每日结算
	PartnerSettlementDailyCancel = "partner.settlement.daily.cancel" // 取消分站每日结算(该事件是用于删除分站每日结算的幂等锁，只需要重新再发一个分站每日结算事件，就可以重新再计算)

	PartnerUserLoginFail    = "partner-user.login.fail"
	PartnerUserLoginSuccess = "partner-user.login.success"

	PaymentGatewayDepositOrderDetected  = "payment-gateway.deposit-order.detected"  // 支付网关发现存款订单(订单状态改变,订单处理结果在payload中定义)
	PaymentGatewayWithdrawOrderDetected = "payment-gateway.withdraw-order.detected" // 支付网关发现提款订单(订单状态改变,订单处理结果在payload中定义)

	RiskControlRiskAlert = "risk-control.risk.alert" // 风险包括风控规则触发

	SystemError = "system-error"

	// 任务相关
	DepositSucceeded          = "operator.payment.deposit_succeeded"          // 充值成功事件
	BirthdateUpdated          = "operator.member.birthdate_updated"           // 生日更新事件
	RealnameVerified          = "operator.member.realname_verified"           // 实名认证事件
	AvatarNicknameUpdated     = "operator.member.avatar_nickname_updated"     // 头像与昵称更新事件
	MobileBound               = "operator.member.mobile_bound"                // 手机绑定事件
	IDCPFVerified             = "operator.member.id_cpf_verified"             // ID/CPF 认证事件
	EmailVerified             = "operator.member.email_verified"              // 邮箱验证事件
	SocialMediaUpdated        = "operator.member.social_media_updated"        // 社交媒体更新事件
	ValidBetSettled           = "operator.bet.valid_bet_settled"              // 有效下注结算事件
	ReferralRegisteredDeposit = "operator.member.referral_registered_deposit" // 下线会员注册并充值指定金额事件
	AgentApplicationApproved  = "operator.agent.application_approved"         // 代理申请通过事件
	WithdrawalSucceeded       = "operator.withdrawal.succeeded"               // 提现成功事件
	ChatMessageAccepted       = "operator.chat.message_accepted"              // 聊天消息被接受事件
)
