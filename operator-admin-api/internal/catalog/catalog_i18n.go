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
	add("api.agentTreeNode", "代理树节点", "代理樹節點", "Agent tree node")
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
	add("common.column.operations", "操作", "操作", "Operations")
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
	add("agent.levelExceeded", "代理层级超限", "代理層級超限", "Agent level exceeded")
	add("agent.shareRateExceedParent", "占成不能超过上级", "佔成不能超過上級", "Share rate exceeds parent")
	add("agent.shareRateExceedMax", "占成不能超过上限", "佔成不能超過上限", "Share rate exceeds max")
	add("agent.fieldsRequired", "代理字段不能为空", "代理字段不能為空", "Agent fields required")
	add("agent.conflict", "代理账号冲突", "代理賬號衝突", "Agent conflict")
	add("agent.lineCodeInvalid", "代理线路编码无效", "代理線路編碼無效", "Agent line code invalid")
	return out
}
