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
	add("menu.route.agentLoginLog", "代理日志", "代理日誌", "Agent login logs")
	add("menu.route.agentLoginAssociation", "关联检索", "關聯檢索", "Login associations")
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
	add("api.group.agent", "代理管理", "代理管理", "Agents")

	add("api.agentList", "代理列表", "代理列表", "Agent list")
	add("api.agentCreate", "分站添加代理", "分站添加代理", "Add operator agent")
	add("api.agentSearch", "搜索代理", "搜索代理", "Search agents")
	add("api.agentOperatorList", "分站代理分组列表", "分站代理分組列表", "Operator agent groups")
	add("api.agentLevelList", "代理等级列表", "代理等級列表", "Agent level list")
	add("api.agentLevelUpdate", "批量更新代理等级", "批量更新代理等級", "Update agent levels")
	add("api.agentConfigGet", "获取代理配置", "獲取代理配置", "Get agent config")
	add("api.agentConfigUpdate", "更新代理配置", "更新代理配置", "Update agent config")
	add("api.agentLineDefaults", "代理线默认代理", "代理線默認代理", "Default agents by line")
	add("api.agentInitStatus", "代理初始化检测", "代理初始化檢測", "Agent init status")
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
	add("api.agentLoginLogList", "代理登录日志列表", "代理登錄日誌列表", "Agent login log list")
	add("api.agentLoginAssociations", "关联检索登录日志", "關聯檢索登錄日誌", "Login association search")
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
	add("common.invalidParam", "参数无效", "參數無效", "Invalid parameter")
	add("common.invalidPhone", "手机号无效", "手機號無效", "Invalid phone number")
	add("auth.ipMismatch", "登录 IP 已变化，请重新登录", "登錄 IP 已變化，請重新登錄", "Login IP has changed, please login again")
	add("auth.ipNotAllowed", "当前 IP 不在白名单内", "當前 IP 不在白名單內", "Current IP is not in the whitelist")
	add("auth.invalidToken", "登录凭证无效", "登錄憑證無效", "Invalid login credentials")
	add("auth.userSaltMismatch", "用户盐不匹配", "用戶鹽不匹配", "User salt mismatch")
	add("auth.userDisabled", "账号已停用", "賬號已停用", "Account disabled")
	add("auth.invalidCredentials", "登录凭证无效", "登錄憑證無效", "Invalid login credentials")
	add("auth.userNotFound", "登录凭证无效，用户不存在", "登錄憑證無效，用戶不存在", "Invalid login credentials, user not found")
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
	add("agent.notFound", "代理不存在", "代理不存在", "Agent not found")
	return out
}
