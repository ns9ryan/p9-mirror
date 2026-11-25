package catalog

import (
	"oa.98ent.com/p9/common/i18n"
	"oa.98ent.com/p9/core/rpc/coreclient"
)

// 多语言种子数据
func langSeeds(i18nCode string) []*coreclient.CreateI18NLangReq {
	// 分站语言由建站分配迁移，此处不注册
	return []*coreclient.CreateI18NLangReq{}
}

// 多语言词典种子数据
func i18nSeeds(i18nCode string) []*coreclient.I18NItem {
	items := make([]*coreclient.I18NItem, 0)
	items = append(items, menuI18n(i18nCode)...)
	items = append(items, apiI18n(i18nCode)...)
	items = append(items, frontI18n(i18nCode)...)
	items = append(items, errorI18n(i18nCode)...)
	return items
}

// 菜单多语言数据
func menuI18n(i18nCode string) []*coreclient.I18NItem {
	var out []*coreclient.I18NItem
	add := func(key, zh, hk, en string) {
		out = append(out,
			&coreclient.I18NItem{I18NCode: i18nCode, I18NGroup: i18n.GroupMenu, TransKey: key, Lang: i18n.LangZH, Value: zh},
			&coreclient.I18NItem{I18NCode: i18nCode, I18NGroup: i18n.GroupMenu, TransKey: key, Lang: i18n.LangHK, Value: hk},
			&coreclient.I18NItem{I18NCode: i18nCode, I18NGroup: i18n.GroupMenu, TransKey: key, Lang: i18n.LangEN, Value: en},
		)
	}
	add("menu.route.agent", "代理管理", "代理管理", "Agents")
	add("menu.route.agentList", "代理列表", "代理列表", "Agent list")
	add("menu.route.agentTree", "代理树", "代理樹", "Agent tree")
	add("menu.route.agentConfig", "代理配置", "代理配置", "Agent config")
	add("menu.route.agentLevel", "代理等级", "代理等級", "Agent levels")
	add("menu.route.agentCreate", "添加代理", "添加代理", "Create agent")
	add("menu.route.agentSearch", "搜索代理", "搜索代理", "Search agents")
	add("menu.route.agentConfigUpdate", "更新代理配置", "更新代理配置", "Update agent config")
	add("menu.route.agentLevelUpdate", "更新代理等级", "更新代理等級", "Update agent levels")
	add("menu.route.agentApply", "代理申请", "代理申請", "Agent applies")
	add("menu.route.agentApplyReview", "审批代理申请", "審批代理申請", "Review agent apply")
	add("menu.route.agentProfileGet", "查看代理资料", "查看代理資料", "Get agent profile")
	add("menu.route.agentPasswordReset", "重置登录密码", "重置登錄密碼", "Reset login password")
	add("menu.route.agentWithdrawPasswordReset", "重置提现密码", "重置提現密碼", "Reset withdraw password")
	add("menu.route.agentShareRateUpdate", "修改占成", "修改佔成", "Update share rate")
	add("menu.route.agentPayoutUpdate", "修改出款限制", "修改出款限制", "Update payout restriction")
	add("menu.route.agentStatusUpdate", "修改代理状态", "修改代理狀態", "Update agent status")
	add("menu.route.agentRemarkUpdate", "修改代理备注", "修改代理備註", "Update agent remark")
	add("menu.route.agentProfileUpdate", "修改代理资料", "修改代理資料", "Update agent profile")
	add("menu.route.agentOpLogs", "资料变更记录", "資料變更記錄", "Profile change logs")
	add("menu.route.agentLoginLog", "代理日志", "代理日誌", "Agent login logs")
	add("menu.route.agentLoginAssociation", "关联检索", "關聯檢索", "Login associations")
	add("menu.route.agentDepositOrder", "人工存款", "人工存款", "Agent deposit orders")
	add("menu.route.agentDepositOrderAudit", "存款订单审核", "存款訂單審核", "Deposit Order Audit")
	add("menu.route.agentDepositOrderManage", "存款订单管理", "存款訂單管理", "Deposit Order Management")
	add("menu.route.agentTurnover", "流水管理", "流水管理", "Turnover")
	add("menu.route.agentDepositOrderAdd", "添加人工存款", "添加人工存款", "Add deposit order")
	add("menu.route.agentDepositOrderList", "人工存款订单列表", "人工存款訂單列表", "Deposit order list")
	add("menu.route.agentTurnoverAdd", "添加流水", "添加流水", "Add turnover")
	add("menu.route.agentTurnoverList", "流水列表", "流水列表", "Turnover list")

	// 会员管理
	add("menu.memberManage", "会员管理", "會員管理", "Member management")
	add("menu.memberManage.list", "会员列表", "會員列表", "Member list")
	add("menu.route.memberLevel", "层级管理", "層級管理", "Member levels")
	add("menu.route.memberVip", "VIP管理", "VIP管理", "VIP management")
	add("menu.route.memberContactBindingCodeList", "联系方式绑定验证码记录", "聯繫方式綁定驗證碼記錄", "Contact binding code records")

	// 层级管理操作权限
	add("menu.route.memberAutoLevelCreate", "创建自动层级", "建立自動層級", "Create auto level")
	add("menu.route.memberFixedLevelCreate", "创建固定层级", "建立固定層級", "Create fixed level")
	add("menu.route.memberFixedLevelUpdate", "修改固定层级", "修改固定層級", "Update fixed level")
	add("menu.route.memberFixedLevelDelete", "删除固定层级", "刪除固定層級", "Delete fixed level")
	add("menu.route.memberSearchByCode", "搜索会员", "搜索會員", "Search members")

	// 财务中心
	add("menu.route.financeCenter", "财务中心", "財務中心", "Finance")
	add("menu.route.financeDepositAudit", "存款审核", "存款審核", "Deposit audit")
	add("menu.route.financeDepositManage", "存款管理", "存款管理", "Deposit manage")
	add("menu.route.financeDepositOrderManage", "下载报表", "下載報表", "Export report")
	add("menu.route.financeMemberDepositAudit", "会员存款审核", "會員存款審核", "Member deposit audit")
	add("menu.route.financeSimulatedTurnover", "模拟流水", "模擬流水", "Simulated turnover")
	add("menu.route.financeMemberDepositAuditList", "会员存款审核列表", "會員存款審核列表", "Member deposit audit list")
	add("menu.route.financeSimulatedTurnoverList", "模拟流水列表", "模擬流水列表", "Simulated turnover list")
	add("menu.route.FinanceDepositOrderClaim", "领取审核", "領取審核", "Claim order")
	add("menu.route.FinanceDepositOrderRelease", "释放审核", "釋放審核", "Release order")
	add("menu.route.FinanceDepositOrderApprove", "审核通过", "審核通過", "Approve order")
	add("menu.route.FinanceDepositOrderReject", "审核拒绝", "審核拒絕", "Reject order")
	add("menu.route.FinanceDepositOrderFreeze", "冻结订单", "凍結訂單", "Freeze order")
	add("menu.route.FinanceDepositOrderUnfreeze", "解冻订单", "解凍訂單", "Unfreeze order")

	// 优惠中心
	add("menu.route.promoCenter", "优惠中心", "優惠中心", "Promotions")
	add("menu.route.activityCenter", "活动中心", "活動中心", "Activity Center")
	add("menu.route.taskCenter", "任务中心", "任務中心", "Task Center")
	return out
}

// 接口多语言数据
func apiI18n(i18nCode string) []*coreclient.I18NItem {
	var out []*coreclient.I18NItem
	add := func(key, zh, hk, en string) {
		out = append(out,
			&coreclient.I18NItem{I18NCode: i18nCode, I18NGroup: i18n.GroupAPI, TransKey: key, Lang: i18n.LangZH, Value: zh},
			&coreclient.I18NItem{I18NCode: i18nCode, I18NGroup: i18n.GroupAPI, TransKey: key, Lang: i18n.LangHK, Value: hk},
			&coreclient.I18NItem{I18NCode: i18nCode, I18NGroup: i18n.GroupAPI, TransKey: key, Lang: i18n.LangEN, Value: en},
		)
	}
	add("api.memberExport", "导出会员 CSV", "匯出會員 CSV", "Export members CSV")
	// 初始化 API
	add("api.initStatus", "初始化状态检测", "初始化狀態檢測", "Init status check")

	// 代理管理 API
	add("api.group.agent", "代理管理", "代理管理", "Agents")

	// 会员管理 API
	add("api.group.member", "会员管理", "會員管理", "Members")

	// 会员管理 API
	add("api.memberList", "会员列表", "會員列表", "Member list")
	add("api.memberAutoLevelUpdate", "修改会员自动层级", "修改會員自動層級", "Update member auto level")
	add("api.memberFixedLevelsUpdate", "修改会员固定层级", "修改會員固定層級", "Update member fixed levels")
	add("api.memberUpdateStatus", "修改会员状态", "修改会员状态", "UpdateStatus")
	add("api.memberResetAuthErrors", "重置会员认证错误次数", "重置会员认证错误次数", "ResetAuthErrors")
	add("api.memberContactBindingCodeList", "查询会员联系方式绑定验证码记录", "查詢會員聯繫方式綁定驗證碼記錄", "List member contact binding codes")
	add("api.memberContactBindingCodeSend", "发送会员联系方式验证码", "發送會員聯繫方式驗證碼", "Send member contact binding code")
	add("member.contactBindingCodeInvalid", "验证码无效或已过期", "驗證碼無效或已過期", "Invalid or expired verification code")
	add("member.contactBindingCodeTooFrequent", "验证码请求过于频繁", "驗證碼請求過於頻繁", "Verification code requested too frequently")
	add("member.contactBindingCodeSendFailed", "验证码发送失败", "驗證碼發送失敗", "Could not send verification code")
	add("member.emailAlreadyBound", "该邮箱已被其他会员绑定，请更换邮箱", "該郵箱已被其他會員綁定，請更換郵箱", "This email is already bound to another member. Use a different email.")
	add("member.phoneAlreadyBound", "该手机号已被其他会员绑定，请更换手机号", "該手機號已被其他會員綁定，請更換手機號", "This phone number is already bound to another member. Use a different phone number.")
	add("api.memberUpdateContact", "修改会员联系方式", "修改會員聯繫方式", "Update member contact")
	add("api.memberListImages", "获取会员头像资源", "取得會員頭像資源", "List member image resources")
	add("api.memberUpdateRemark", "修改会员备注", "修改会员备注", "UpdateRemark")
	add("api.memberUpdateChat", "修改会员禁言开关", "修改会员禁言开关", "UpdateChat")
	add("api.memberUpdateBetting", "修改会员禁止投注开关", "修改会员禁止投注开关", "UpdateBetting")
	add("api.memberUpdateWithdrawal", "修改会员禁止提款开关", "修改会员禁止提款开关", "UpdateWithdrawal")
	add("api.memberCreate", "创建会员", "建立會員", "Create member")
	add("api.memberBatchCreate", "批量创建会员", "批次建立會員", "Create members in batch")
	add("api.memberDetail", "会员详情", "會員詳情", "Member detail")
	add("api.memberSearchByCode", "按编号/账号/昵称搜索会员", "按編號/帳號/暱稱搜索會員", "Search members by code/account/nickname")

	// 会员自动层级 API
	add("api.memberAutoLevelList", "自动层级列表", "自動層級列表", "Auto level list")
	add("api.memberAutoLevelCreate", "创建自动层级", "建立自動層級", "Create auto level")

	// 会员固定层级 API
	add("api.memberFixedLevelList", "固定层级列表", "固定層級列表", "Fixed level list")
	add("api.memberFixedLevelCreate", "创建固定层级", "建立固定層級", "Create fixed level")
	add("api.memberFixedLevelUpdate", "修改固定层级", "修改固定層級", "Update fixed level")
	add("api.memberFixedLevelDelete", "删除固定层级", "刪除固定層級", "Delete fixed level")

	// VIP 基础配置 API
	add("api.memberVipConfigGet", "获取VIP基础配置", "獲取VIP基礎配置", "Get VIP base config")
	add("api.memberVipConfigUpdate", "修改VIP基础配置", "修改VIP基礎配置", "Update VIP base config")

	// VIP 等级 API
	add("api.memberVipLevelList", "VIP等级列表", "VIP等級列表", "VIP level list")
	add("api.memberVipLevelGet", "获取VIP等级详情", "獲取VIP等級詳情", "Get VIP level detail")
	add("api.memberVipLevelCreate", "创建VIP等级", "建立VIP等級", "Create VIP level")
	add("api.memberVipLevelUpdate", "修改VIP等级", "修改VIP等級", "Update VIP level")
	add("api.memberVipLevelDelete", "删除VIP等级", "刪除VIP等級", "Delete VIP level")

	// VIP 图标 API
	add("api.memberVipIconList", "VIP图标组列表", "VIP圖標組列表", "VIP icon group list")
	add("api.memberVipIconUpdate", "修改VIP图标", "修改VIP圖標", "Update VIP icon")
	add("api.memberVipIconReset", "恢复默认VIP图标", "恢復預設VIP圖標", "Reset VIP icon")

	// 活动管理 API
	add("api.group.activity", "活动管理", "活動管理", "Activities")
	add("api.activityCreate", "创建活动", "建立活動", "Create activity")
	add("api.activityUpdate", "编辑活动", "編輯活動", "Update activity")
	add("api.activityDetail", "活动详情", "活動詳情", "Activity detail")
	add("api.activityList", "活动列表", "活動列表", "Activity list")
	add("api.activityClose", "关闭活动", "關閉活動", "Close activity")
	add("api.activityPublish", "发布活动", "發佈活動", "Publish activity")
	add("api.activityEnumParams", "活动枚举参数", "活動枚舉參數", "Activity enum params")
	add("api.group.activityOrder", "活动订单", "活動訂單", "Activity orders")
	add("api.activityOrderList", "活动订单列表", "活動訂單列表", "Activity order list")
	add("api.activityOrderAuditList", "活动订单审核列表", "活動訂單審核列表", "Activity order audit list")
	add("api.activityOrderAudit", "活动订单审核", "活動訂單審核", "Audit activity order")

	// 任务管理 API
	add("api.group.task", "任务中心", "任務中心", "Task Center")
	add("api.taskConfigList", "任务配置列表", "任務配置列表", "Task config list")
	add("api.taskConfigUpdate", "修改任务配置", "修改任務配置", "Update task config")
	add("api.taskRewardList", "会员任务奖励订单查询", "會員任務獎勵訂單查詢", "Member task reward list")

	add("api.agentList", "代理列表", "代理列表", "Agent list")
	add("api.agentCreate", "分站添加代理", "分站添加代理", "Add operator agent")
	add("api.agentSearch", "搜索代理", "搜索代理", "Search agents")
	add("api.agentOperatorList", "分站代理分组列表", "分站代理分組列表", "Operator agent groups")
	add("api.agentLevelList", "代理等级列表", "代理等級列表", "Agent level list")
	add("api.agentLevelUpdate", "批量更新代理等级", "批量更新代理等級", "Update agent levels")
	add("resource", "图片素材", "圖片素材", "Image resources")
	add("api.resourceUpload", "上传图片素材", "上傳圖片素材", "Upload image resource")

	add("api.agentConfigGet", "获取代理配置", "獲取代理配置", "Get agent config")
	add("api.agentConfigUpdate", "更新代理配置", "更新代理配置", "Update agent config")
	add("api.agentLineDefaults", "代理线默认代理", "代理線默認代理", "Default agents by line")
	add("api.agentTreeNode", "代理树节点", "代理樹節點", "Agent tree node")
	add("api.agentTreeAncestors", "代理上级链", "代理上級鏈", "Agent ancestor chain")
	add("api.agentApplyList", "代理申请列表", "代理申請列表", "Agent apply list")
	add("api.agentApplyReview", "审批代理申请", "審批代理申請", "Review agent apply")
	add("api.agentProfileGet", "获取代理资料", "獲取代理資料", "Get agent profile")
	add("api.agentPasswordReset", "重置登录密码", "重置登錄密碼", "Reset login password")
	add("api.agentWithdrawPasswordReset", "重置提现密码", "重置提現密碼", "Reset withdraw password")
	add("api.agentShareRateUpdate", "修改占成", "修改佔成", "Update share rate")
	add("api.agentPayoutUpdate", "修改出款限制", "修改出款限制", "Update payout restriction")
	add("api.agentStatusUpdate", "修改代理状态", "修改代理狀態", "Update agent status")
	add("api.agentRemarkUpdate", "修改代理备注", "修改代理備註", "Update agent remark")
	add("api.agentProfileUpdate", "修改代理资料", "修改代理資料", "Update agent profile")
	add("api.agentOpLogs", "代理资料变更操作日志", "代理資料變更操作日誌", "Agent profile op logs")
	add("api.agentLoginLogList", "代理登录日志列表", "代理登錄日誌列表", "Agent login log list")
	add("api.agentLoginAssociations", "关联检索登录日志", "關聯檢索登錄日誌", "Login association search")
	// 财务管理 API
	add("api.group.finance", "财务管理", "財務管理", "Finance")

	add("api.agentDepositOrderAdd", "添加人工存款订单", "添加人工存款訂單", "Add agent deposit order")
	add("api.agentDepositOrderList", "人工存款订单列表", "人工存款訂單列表", "Agent deposit order list")
	add("api.agentDepositOrderClaim", "领取审核", "領取審核", "Claim Review")
	add("api.agentDepositOrderRelease", "释放审核", "釋放審核", "Release Review")
	add("api.agentDepositOrderApprove", "审核通过", "審核通過", "Approve")
	add("api.agentDepositOrderReject", "审核拒绝", "審核拒絕", "Reject")
	add("api.agentDepositOrderFreeze", "冻结", "凍結", "Freeze")
	add("api.agentDepositOrderUnfreeze", "解冻", "解凍", "Unfreeze")
	add("api.exportAgentDepositOrders", "导出存款订单", "匯出存款訂單", "Export Deposit Orders")
	add("api.agentTurnoverAdd", "添加流水", "添加流水", "Add turnover")
	add("api.agentTurnoverList", "流水列表", "流水列表", "Turnover list")

	add("api.taskConfigList", "任务配置列表", "任務配置列表", "Task config list")
	add("api.taskConfigUpdate", "更新任务配置", "更新任務配置", "Update task config")
	add("api.taskRewardList", "任务奖励列表", "任務獎勵列表", "Task reward list")
	add("api.taskRewardUpdateStatus", "更新任务奖励状态", "更新任務獎勵狀態", "Update task reward status")
	add("api.taskGlobalConfigList", "全局任务配置列表", "全局任務配置列表", "Global task config list")
	add("api.taskGlobalConfigUpdate", "更新全局任务配置", "更新全局任務配置", "Update global task config")
	return out
}

// 前端多语言数据
func frontI18n(i18nCode string) []*coreclient.I18NItem {
	var out []*coreclient.I18NItem
	add := func(key, zh, hk, en string) {
		out = append(out,
			&coreclient.I18NItem{I18NCode: i18nCode, I18NGroup: i18n.GroupFront, TransKey: key, Lang: i18n.LangZH, Value: zh},
			&coreclient.I18NItem{I18NCode: i18nCode, I18NGroup: i18n.GroupFront, TransKey: key, Lang: i18n.LangHK, Value: hk},
			&coreclient.I18NItem{I18NCode: i18nCode, I18NGroup: i18n.GroupFront, TransKey: key, Lang: i18n.LangEN, Value: en},
		)
	}
	add("agent.loginType.1", "登录", "登錄", "Login")
	add("agent.loginType.2", "注册", "註冊", "Register")
	add("agent.deviceType.1", "PC网页", "PC網頁", "PC Web")
	add("agent.deviceType.2", "H5", "H5", "H5")
	add("agent.deviceType.3", "Android", "Android", "Android")
	add("agent.deviceType.4", "iOS", "iOS", "iOS")
	add("agent.deviceType.5", "PC客户端", "PC客戶端", "PC Client")
	add("agent.loginResult.1", "成功", "成功", "Success")
	add("agent.loginResult.2", "密码无效", "密碼無效", "Invalid password")
	add("agent.loginResult.3", "账号停用", "帳號停用", "Account disabled")
	add("agent.loginResult.4", "账号不存在", "帳號不存在", "Account not found")
	add("agent.opResult.1", "成功", "成功", "Success")
	add("agent.opResult.2", "失败", "失敗", "Failed")
	add("agent.opTargetType.1", "代理账号", "代理帳號", "Agent account")
	add("agent.opTargetType.2", "会员账号", "會員帳號", "Member account")
	add("agent.adminType.1", "分站管理员", "分站管理員", "Operator admin")
	add("agent.adminType.2", "分站代理", "分站代理", "Operator agent")
	add("agent.opAction.create", "创建代理", "創建代理", "Create agent")
	add("agent.opAction.delete", "删除代理", "刪除代理", "Delete agent")
	add("agent.opAction.share_rate", "更新占成", "更新佔成", "Update share rate")
	add("agent.opAction.payout_restricted", "更新出款限制", "更新出款限制", "Update payout restriction")
	add("agent.opAction.status", "更新状态", "更新狀態", "Update status")
	add("agent.opAction.remark", "更新备注", "更新備註", "Update remark")
	add("agent.opAction.reset_login_password", "重置登录密码", "重置登錄密碼", "Reset login password")
	add("agent.opAction.update_login_password", "修改登录密码", "修改登錄密碼", "Update login password")
	add("agent.opAction.update_withdraw_password", "修改提款密码", "修改提款密碼", "Update withdraw password")
	add("agent.opAction.profile", "更新资料", "更新資料", "Update profile")
	add("agent.opAction.profile.avatar", "更新头像", "更新頭像", "Update avatar")
	add("agent.opAction.profile.realname", "更新真实姓名", "更新真實姓名", "Update real name")
	add("agent.opAction.profile.telegram", "更新 Telegram", "更新 Telegram", "Update Telegram")
	add("agent.opAction.profile.facebook", "更新 Facebook", "更新 Facebook", "Update Facebook")
	add("agent.opAction.profile.whatsapp", "更新 WhatsApp", "更新 WhatsApp", "Update WhatsApp")
	add("agent.opAction.profile.mobile", "更新手机号", "更新手機號", "Update mobile")
	add("agent.opAction.profile.nickname", "更新昵称", "更新暱稱", "Update nickname")
	add("agent.opAction.profile.id_document", "更新证件号", "更新證件號", "Update ID document")
	add("agent.opAction.profile.zalo", "更新 Zalo", "更新 Zalo", "Update Zalo")
	add("agent.opAction.profile.twitter", "更新 Twitter", "更新 Twitter", "Update Twitter")
	add("agent.opAction.profile.line", "更新 Line", "更新 Line", "Update Line")
	add("agent.opAction.profile.email", "更新邮箱", "更新郵箱", "Update email")

	// 活动枚举
	add("activity.type.1", "充值", "充值", "Recharge")
	add("activity.type.2", "洗码", "洗碼", "Wagering")
	add("activity.type.3", "推荐会员", "推薦會員", "Recommended Member")
	add("activity.type.4", "幸运转盘", "幸運轉盤", "Lucky Wheel")
	add("activity.type.5", "助力", "助力", "Assist")
	add("activity.type.6", "盲盒", "盲盒", "Blind Box")
	add("activity.type.7", "新人彩金", "新人彩金", "Newcomer Bonus")
	add("activity.type.8", "全民代理", "全民代理", "Universal Agent")
	add("activity.type.9", "救援金", "救援金", "Rescue Fund")
	add("activity.type.10", "幸运注单", "幸運注單", "Lucky Bet")
	add("activity.type.11", "VIP奖金", "VIP獎金", "VIP Reward")
	add("activity.type.12", "红包雨", "紅包雨", "Red Envelope Rain")
	add("activity.type.13", "图文(广告)", "圖文(廣告)", "Image/Text Ad")
	add("activity.type.14", "签到", "簽到", "Attendance")
	add("activity.status.1", "草稿", "草稿", "Draft")
	add("activity.status.2", "进行中", "進行中", "Ongoing")
	add("activity.status.3", "已结束", "已結束", "Closed")
	add("activity.applyMode.1", "自动参与", "自動參與", "Auto apply")
	add("activity.applyMode.2", "手动参与", "手動參與", "Manual apply")
	add("activity.section.1", "大厅", "大廳", "Lobby")
	add("activity.section.2", "电子", "電子", "Electronic")
	add("activity.section.3", "体育", "體育", "Sports")
	add("activity.cycleMode.1", "不循环", "不循環", "No cycle")
	add("activity.cycleMode.2", "每日循环", "每日循環", "Daily")
	add("activity.cycleMode.3", "每周循环", "每週循環", "Weekly")
	add("activity.cycleMode.4", "每月循环", "每月循環", "Monthly")
	add("activity.cycleMode.5", "无限循环", "無限循環", "Infinite")
	add("activity.payoutMode.1", "手动领取(过期作废)", "手動領取(過期作廢)", "Manual claim(expire void)")
	add("activity.payoutMode.2", "手动领取(过期自动领取)", "手動領取(過期自動領取)", "Manual claim(expire auto)")
	add("activity.payoutMode.3", "自动派发", "自動派發", "Auto payout")
	add("activity.payoutTiming.1", "实时(影响留存)", "實時(影響留存)", "Realtime(affect retention)")
	add("activity.payoutTiming.2", "次日(财务过账后)", "次日(財務過賬後)", "Next day(after financial posting)")
	add("activity.bonusMode.1", "固定金额多档", "固定金額多檔", "Fixed multi-tier")
	add("activity.bonusMode.2", "按比例多档", "按比例多檔", "Percentage multi-tier")
	add("activity.bonusMode.3", "固定金额阶梯最高下发", "固定金額階梯最高下發", "Fixed tiered highest")
	add("activity.bonusMode.4", "按比例阶梯最高下发", "按比例階梯最高下發", "Percentage tiered highest")
	add("activity.bonusMode.5", "单独奖励", "單獨獎勵", "Separate reward")
	add("activity.betClaimMode.1", "逐条领取", "逐條領取", "Claim by item")
	add("activity.betClaimMode.2", "最高档领取", "最高檔領取", "Claim highest")
	add("activity.betClaimMode.3", "累计领取", "累計領取", "Claim cumulative")
	add("activity.depositRewardType.1", "首充", "首充", "First deposit")
	add("activity.depositRewardType.2", "次存", "次存", "Second deposit")
	add("activity.depositRewardType.3", "累计", "累計", "Cumulative")
	add("activity.depositRewardType.4", "单笔", "單筆", "Single")
	add("activity.wheelTriggerMode.1", "仅投注", "僅投注", "Bet only")
	add("activity.wheelTriggerMode.2", "仅充值", "僅充值", "Deposit only")
	add("activity.wheelTriggerMode.3", "投注或充值", "投注或充值", "Bet or deposit")
	add("activity.wageringBase.1", "本金+赠送", "本金+贈送", "Principal + bonus")
	add("activity.wageringBase.2", "仅赠送", "僅贈送", "Bonus only")
	add("activity.clientType.1", "PC网页", "PC網頁", "PC Web")
	add("activity.clientType.2", "移动网页", "移動網頁", "Mobile Web")
	add("activity.clientType.3", "安卓APP", "安卓APP", "Android App")
	add("activity.clientType.4", "iOS APP", "iOS APP", "iOS App")
	add("activity.applyLimitType.1", "同设备号不限参与次数", "同設備號不限參與次數", "Same device unlimited")
	add("activity.applyLimitType.2", "同设备IP限参与一次", "同設備IP限參與一次", "Same device IP once")
	add("activity.claimLimitType.1", "完成充值才能领取", "完成充值才能領取", "Require deposit")
	add("activity.claimLimitType.2", "完成银行卡绑定才能领取", "完成銀行卡綁定才能領取", "Require bank binding")
	add("activity.claimLimitType.3", "完成短信验证才能领取", "完成短信驗證才能領取", "Require SMS verification")
	add("activity.claimLimitType.4", "完成邮箱验证才能领取", "完成郵箱驗證才能領取", "Require email verification")
	add("activity.claimLimitType.5", "完成姓名填写才能领取", "完成姓名填寫才能領取", "Require real name")
	return out
}

// 错误多语言数据
func errorI18n(i18nCode string) []*coreclient.I18NItem {
	var out []*coreclient.I18NItem
	add := func(key, zh, hk, en string) {
		out = append(out,
			&coreclient.I18NItem{I18NCode: i18nCode, I18NGroup: i18n.GroupError, TransKey: key, Lang: i18n.LangZH, Value: zh},
			&coreclient.I18NItem{I18NCode: i18nCode, I18NGroup: i18n.GroupError, TransKey: key, Lang: i18n.LangHK, Value: hk},
			&coreclient.I18NItem{I18NCode: i18nCode, I18NGroup: i18n.GroupError, TransKey: key, Lang: i18n.LangEN, Value: en},
		)
	}
	add("common.success", "成功", "成功", "Success")
	add("common.unauthorized", "未授权", "未授權", "Unauthorized")
	add("common.forbidden", "无权限", "無權限", "Forbidden")
	add("common.tokenExpired", "登录已过期", "登錄已過期", "Login expired")
	add("common.internal", "内部错误", "內部錯誤", "Internal error")
	add("common.internalError", "内部错误", "內部錯誤", "Internal error")
	add("common.notFound", "资源不存在", "資源不存在", "Resource not found")
	add("common.badRequest", "请求错误", "請求錯誤", "Bad request")
	add("common.timeout", "请求超时", "請求超時", "Request timeout")
	add("common.serviceUnavailable", "服务不可用", "服務不可用", "Service unavailable")
	add("resource.invalidFile",
		"图片格式、尺寸或大小不符合场景要求", "圖片格式、尺寸或大小不符合場景要求",
		"Image does not meet the scene requirements")
	add("resource.uploadFailed", "图片上传失败", "圖片上傳失敗", "Image upload failed")
	add("common.invalidParam", "参数无效", "參數無效", "Invalid parameter")
	add("common.invalidPhone", "手机号无效", "手機號無效", "Invalid phone number")
	add("auth.ipMismatch", "登录 IP 已变化，请重新登录", "登錄 IP 已變化，請重新登錄", "Login IP has changed, please login again")
	add("auth.ipNotAllowed", "当前 IP 不在白名单内", "當前 IP 不在白名單內", "Current IP is not in the whitelist")
	add("auth.invalidToken", "登录凭证无效", "登錄憑證無效", "Invalid login credentials")
	add("auth.userSaltMismatch", "用户盐不匹配", "用戶鹽不匹配", "User salt mismatch")
	add("auth.userDisabled", "账号已停用", "賬號已停用", "Account disabled")
	add("auth.invalidCredentials", "登录凭证无效", "登錄憑證無效", "Invalid login credentials")
	add("auth.userNotFound", "登录凭证无效，用户不存在", "登錄憑證無效，用戶不存在", "Invalid login credentials, user not found")
	add("auth.passwordIncorrect", "密码错误", "密碼錯誤", "Incorrect password")
	add("auth.accountOrPasswordIncorrect", "账号或密码错误", "賬號或密碼錯誤", "Incorrect account or password")
	add("auth.usernamePasswordRequired", "用户名和密码不能为空", "用戶名和密碼不能為空", "Username and password required")
	add("auth.operatorCodeRequired", "分站编码不能为空", "分站編碼不能為空", "Operator code required")
	add("auth.refreshTokenRequired", "刷新令牌不能为空", "刷新令牌不能為空", "Refresh token required")
	add("agent.levelItemsRequired", "等级项不能为空", "等級項不能為空", "Level items required")
	add("agent.levelDuplicate", "等级重复", "等級重複", "Duplicate level")
	add("agent.levelOutOfRange", "等级超出范围", "等級超出範圍", "Level out of range")
	add("agent.shareRateInvalid", "占成比例无效", "佔成比例無效", "Invalid share rate")
	add("agent.shareRateOrderInvalid", "占成比例区间无效", "佔成比例區間無效", "Invalid share rate range")
	add("agent.configMissing", "代理配置不存在", "代理配置不存在", "Agent config missing")
	add("agent.designatedAgentMissing", "指定代理不存在", "指定代理不存在", "Designated agent missing")
	add("agent.designatedAgentInvalid", "指定代理无效", "指定代理無效", "Designated agent invalid")
	add("agent.parentNotFound", "上级代理不存在", "上級代理不存在", "Parent agent not found")
	add("agent.parentLevelInvalid", "上级代理层级无效", "上級代理層級無效", "Parent agent level invalid")
	add("agent.parentDisabled", "上级代理已停用", "上級代理已停用", "Parent agent disabled")
	add("agent.parentRequired", "必须指定上级代理", "必須指定上級代理", "Parent agent required")
	add("agent.levelExceeded", "代理层级超限，无法再添加下级", "代理層級超限，無法再添加下級", "Agent level exceeded, cannot add sub-agent")
	add("agent.shareRateExceedParent", "占成不能超过上级", "佔成不能超過上級", "Share rate exceeds parent")
	add("agent.shareRateExceedMax", "占成不能超过上限", "佔成不能超過上限", "Share rate exceeds max")
	add("agent.shareRateBelowChild", "占成不能低于下级", "佔成不能低於下級", "Share rate below child")
	add("agent.fieldsRequired", "代理字段不能为空", "代理字段不能為空", "Agent fields required")
	add("agent.conflict", "代理账号已存在", "代理賬號已存在", "Agent account already exists")
	add("agent.lineCodeInvalid", "代理线路编码无效", "代理線路編碼無效", "Agent line code invalid")
	add("agent.applyNotFound", "代理申请不存在", "代理申請不存在", "Agent apply not found")
	add("agent.applyNotPending", "申请不在待审核状态", "申請不在待審核狀態", "Agent apply is not pending")
	add("agent.applyRejectReasonRequired", "拒绝理由不能为空", "拒絕理由不能為空", "Reject reason required")
	add("agent.applyPendingExists", "该账号已有待审核申请", "該賬號已有待審核申請", "Pending apply already exists")
	add("agent.applyMemberCodeExists", "该会员已申请过代理", "該會員已申請過代理", "Member already applied")
	add("agent.applyAccountUsedByOther", "该账号已被其他代理使用", "該賬號已被其他代理使用", "Account already used by another agent")
	add("agent.depositTypeInvalid", "存款方式无效", "存款方式無效", "Invalid deposit type")
	add("agent.notFound", "代理不存在", "代理不存在", "Agent not found")
	add("agent.hasChildren", "代理有下级，无法删除", "代理有下級，無法刪除", "Agent has children, cannot delete")
	add("agent.cannotDeleteRoot", "根代理无法删除", "根代理無法刪除", "Root agent cannot delete")
	add("agent.isDesignated", "代理已被指定为默认代理线路，无法删除", "代理已被指定為默認代理線路，無法刪除", "Agent is designated as default agent line, cannot delete")
	add("activity.invalidType", "活动类型无效", "活動類型無效", "Invalid activity type")
	add("activity.invalidApplyMode", "参与模式无效", "參與模式無效", "Invalid apply mode")
	add("activity.invalidSection", "活动所属无效", "活動所屬無效", "Invalid activity section")
	add("activity.luckWheelSettingsRequired", "幸运转盘配置不能为空", "幸運轉盤配置不能為空", "Lucky wheel settings required")
	add("activity.notDraft", "仅草稿活动可执行此操作", "僅草稿活動可執行此操作", "Only draft activities allow this action")
	add("activity.notFound", "活动不存在", "活動不存在", "Activity not found")
	add("activity.orderNotFound", "订单不存在", "訂單不存在", "Order not found")
	add("activity.invalidId", "活动 ID 无效", "活動 ID 無效", "Invalid activity ID")
	add("activity.orderNotPending", "订单不在待审核状态", "訂單不在待審核狀態", "Order is not pending audit")
	add("activity.engineDisabled", "活动引擎未启动，无法通过审核", "活動引擎未啟動，無法通過審核", "Activity engine is disabled")
	add("agentDepositOrder.notFound", "存款订单不存在", "存款訂單不存在", "Deposit order not found")
	add("agentDepositOrder.notClaimable", "订单当前状态不可领取", "訂單當前狀態不可領取", "Order is not claimable")
	add("agentDepositOrder.alreadyClaimed", "订单已被其他审核人领取", "訂單已被其他審核人領取", "Order already claimed by another reviewer")
	add("agentDepositOrder.notClaimer", "您未领取该订单，无法操作", "您未領取該訂單，無法操作", "You have not claimed this order")
	add("agentDepositOrder.frozenOnlyUnfreeze", "订单已冻结，仅可解冻", "訂單已凍結，僅可解凍", "Order is frozen, can only be unfrozen")
	add("agentDepositOrder.notFrozen", "订单不在冻结状态", "訂單不在凍結狀態", "Order is not frozen")
	add("agentDepositOrder.rejectNotesRequired", "拒绝原因不能为空", "拒絕原因不能為空", "Reject reason is required")
	return out
}
