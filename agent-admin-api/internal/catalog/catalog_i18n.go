package catalog

import (
	"oa.98ent.com/p9/common/i18n"
	"oa.98ent.com/p9/core/rpc/coreclient"
)

// 多语言种子数据
func langSeeds(i18nCode string) []*coreclient.CreateI18NLangReq {
	// 代理不注册多语言，直接使用分站的多语言
	return []*coreclient.CreateI18NLangReq{}
}

// 多语言词典种子数据
func i18nSeeds(i18nCode string) []*coreclient.I18NItem {
	items := make([]*coreclient.I18NItem, 0)

	items = append(items, menuI18n(i18nCode)...)
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
	add("menu.route.dashboard", "数据看板", "數據看板", "Dashboard")
	add("menu.route.account", "账号管理", "賬號管理", "Accounts")
	add("menu.route.member", "会员列表", "會員列表", "Members")
	add("menu.route.agent", "代理列表", "代理列表", "Agents")
	add("menu.route.report", "报表统计", "報表統計", "Reports")
	add("menu.route.winLoss", "输赢报表", "輸贏報表", "Win-loss report")
	add("menu.route.gameReport", "游戏报表", "遊戲報表", "Game report")
	add("menu.route.commission", "佣金报表", "佣金報表", "Commission report")
	add("menu.route.finance", "财务管理", "財務管理", "Finance")
	add("menu.route.withdrawApply", "提款申请", "提款申請", "Withdrawal requests")
	add("menu.route.depositOrder", "存款订单管理", "存款訂單管理", "Deposit orders")
	add("menu.route.withdrawOrder", "提款订单管理", "提款訂單管理", "Withdrawal orders")
	add("menu.route.system", "系统管理", "系統管理", "System")
	add("menu.route.profile", "个人资料", "個人資料", "Profile")
	add("menu.route.adminLog", "后台日志", "後台日誌", "Admin logs")
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
	addLogin := func(key, zh, hk, en string) {
		out = append(out,
			&coreclient.I18NItem{I18NCode: i18nCode, I18NGroup: i18n.GroupFront, TransKey: key, Lang: i18n.LangZH, Value: zh},
			&coreclient.I18NItem{I18NCode: i18nCode, I18NGroup: i18n.GroupFront, TransKey: key, Lang: i18n.LangHK, Value: hk},
			&coreclient.I18NItem{I18NCode: i18nCode, I18NGroup: i18n.GroupFront, TransKey: key, Lang: i18n.LangEN, Value: en},
		)
	}
	add("common.column.operations", "操作", "操作", "Operations")
	add("common.search.reset", "重置", "重置", "Reset")
	add("common.search.search", "搜索", "搜尋", "Search")
	add("common.placeholder.select", "请选择", "請選擇", "Please select")
	add("common.action.delete", "删除", "刪除", "Delete")
	add("common.column.createdAt", "创建时间", "建立時間", "Created At")
	add("common.column.index", "序号", "序號", "No.")
	add("common.okText.save", "保存", "儲存", "Save")
	add("common.message.saveSuccess", "保存成功", "儲存成功", "Saved successfully")
	add("common.action.edit", "编辑", "編輯", "Edit")
	add("common.message.deleteSuccess", "删除成功", "刪除成功", "Deleted successfully")
	add("common.delete.title", "确认删除", "確認刪除", "Confirm Delete")
	add("common.okText.create", "创建", "建立", "Create")
	add("common.message.createSuccess", "创建成功", "建立成功", "Created successfully")
	add("common.column.status", "状态", "狀態", "Status")
	add("common.toolbar.refresh", "刷新", "重新整理", "Refresh")
	add("common.toolbar.add", "添加", "新增", "Add")
	add("common.action.back", "返回", "返回", "Back")
	add("common.status.enabled", "启用", "啟用", "Enabled")
	add("common.message.enabled", "已启用", "已啟用", "Enabled")
	add("common.placeholder.input", "请输入", "請輸入", "Please input")
	add("common.search.statusDisabled", "关闭", "關閉", "Close")
	add("common.all", "全部", "全部", "All")
	add("common.action.view", "查看", "檢視", "View")
	add("common.column.updatedAt", "更新时间", "更新時間", "Updated At")
	add("common.message.submitSuccess", "提交成功", "提交成功", "Submitted successfully")
	add("common.status.disabled", "停用", "停用", "Disabled")
	add("common.delete.okText", "删除", "刪除", "Delete")
	add("common.message.updateSuccess", "更新成功", "更新成功", "Updated successfully")
	add("common.switch.on", "开", "開", "On")
	add("common.switch.off", "关", "關", "Off")
	add("common.yes", "是", "是", "Yes")
	add("common.no", "否", "否", "No")
	addLogin("login.brand.title", "智能科技平台", "智能科技平台", "Intelligent Technology Platform")
	addLogin("login.brand.subtitle", "ZHINENGKEJIPINGTAI", "ZHINENGKEJIPINGTAI", "ZHINENGKEJIPINGTAI")
	addLogin("login.form.title", "用户登录", "用戶登錄", "Sign in")
	addLogin("login.form.userName", "用户名", "用戶名", "Username")
	addLogin("login.form.userName.placeholder", "请输入用户名", "請輸入用戶名", "Enter username")
	addLogin("login.form.userName.errMsg", "请输入用户名", "請輸入用戶名", "Please enter username")
	addLogin("login.form.password", "密码", "密碼", "Password")
	addLogin("login.form.password.placeholder", "请输入密码", "請輸入密碼", "Enter password")
	addLogin("login.form.password.errMsg", "请输入密码", "請輸入密碼", "Please enter password")
	addLogin("login.form.rememberPassword", "记住我", "記住我", "Remember me")
	addLogin("login.form.forgetPassword", "忘记密码?", "忘記密碼?", "Forgot password?")
	addLogin("login.form.login", "登录", "登錄", "Sign in")
	addLogin("login.form.login.success", "登录成功", "登錄成功", "Signed in successfully")
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
	add("auth.ipMismatch", "登录 IP 已变化，请重新登录", "登錄 IP 已變化，請重新登錄", "Login IP has changed, please login again")
	add("auth.ipNotAllowed", "当前 IP 不在白名单内", "當前 IP 不在白名單內", "Current IP is not in the whitelist")
	add("auth.invalidToken", "登录凭证无效", "登錄憑證無效", "Invalid login credentials")
	add("auth.userSaltMismatch", "用户盐不匹配", "用戶鹽不匹配", "User salt mismatch")
	add("auth.userDisabled", "账号已停用", "賬號已停用", "Account disabled")
	add("auth.invalidCredentials", "登录凭证无效", "登錄憑證無效", "Invalid login credentials")
	add("auth.userNotFound", "登录凭证无效，用户不存在", "登錄憑證無效，用戶不存在", "Invalid login credentials, user not found")
	return out
}
