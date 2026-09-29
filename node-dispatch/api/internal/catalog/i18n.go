package catalog

import (
	"oa.98ent.com/p9/common/i18n"
	"oa.98ent.com/p9/core/rpc/coreclient"
)

// nodeDispatchI18n 返回Node Dispatch全部多语言数据
func nodeDispatchI18n() []*coreclient.I18NItem {
	var out []*coreclient.I18NItem

	out = append(out, menuI18n()...)
	out = append(out, apiI18n()...)

	return out
}

// menuI18n 返回菜单多语言数据
func menuI18n() []*coreclient.I18NItem {
	var out []*coreclient.I18NItem

	// 调度中心
	addI18n(&out, i18n.GroupMenu, "menu.route.dispatchCenter", "调度中心", "調度中心", "Dispatch center")

	// 节点管理
	addI18n(&out, i18n.GroupMenu, "menu.route.nodeManagement", "节点管理", "節點管理", "Node management")
	addI18n(&out, i18n.GroupMenu, "menu.route.nodeDetail", "节点详情", "節點詳情", "Node details")
	addI18n(&out, i18n.GroupMenu, "menu.route.nodeCreate", "新增节点", "新增節點", "Create node")
	addI18n(&out, i18n.GroupMenu, "menu.route.nodeUpdate", "编辑节点", "編輯節點", "Edit node")
	addI18n(&out, i18n.GroupMenu, "menu.route.nodeResetAuthSecret", "重置认证密钥", "重設認證密鑰", "Reset authentication secret")

	// 调度任务
	addI18n(&out, i18n.GroupMenu, "menu.route.dispatchTask", "调度任务", "調度任務", "Dispatch tasks")
	addI18n(&out, i18n.GroupMenu, "menu.route.dispatchTaskDetail", "调度任务详情", "調度任務詳情", "Dispatch task details")

	return out
}

// apiI18n 返回API多语言数据
func apiI18n() []*coreclient.I18NItem {
	var out []*coreclient.I18NItem

	// 节点
	addI18n(&out, i18n.GroupAPI, "api.nodeCreate", "创建节点", "建立節點", "Create node")
	addI18n(&out, i18n.GroupAPI, "api.nodeUpdate", "修改节点", "修改節點", "Update node")
	addI18n(&out, i18n.GroupAPI, "api.nodeGet", "节点详情", "節點詳情", "Node details")
	addI18n(&out, i18n.GroupAPI, "api.nodeList", "节点列表", "節點列表", "Node list")
	addI18n(&out, i18n.GroupAPI, "api.nodeResetAuthSecret", "重置节点认证密钥", "重設節點認證密鑰", "Reset node authentication secret")

	// 调度任务
	addI18n(&out, i18n.GroupAPI, "api.dispatchTaskGet", "调度任务详情", "調度任務詳情", "Dispatch task details")
	addI18n(&out, i18n.GroupAPI, "api.dispatchTaskList", "调度任务列表", "調度任務列表", "Dispatch task list")

	// 接口组多语言翻译
	addI18n(&out, i18n.GroupAPI, "api.group.node", "节点管理", "節點管理", "Node management")
	addI18n(&out, i18n.GroupAPI, "api.group.dispatch_task", "调度任务", "調度任務", "Dispatch tasks")

	return out
}

// addI18n 添加Platform站点的简体中文、繁体中文和英文翻译
func addI18n(out *[]*coreclient.I18NItem, group, key, zh, hk, en string) {
	*out = append(*out,
		&coreclient.I18NItem{I18NCode: i18n.CodePlatform, I18NGroup: group, TransKey: key, Lang: i18n.LangZH, Value: zh},
		&coreclient.I18NItem{I18NCode: i18n.CodePlatform, I18NGroup: group, TransKey: key, Lang: i18n.LangHK, Value: hk},
		&coreclient.I18NItem{I18NCode: i18n.CodePlatform, I18NGroup: group, TransKey: key, Lang: i18n.LangEN, Value: en},
	)
}
