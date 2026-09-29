package catalog

import (
	"net/http"

	"oa.98ent.com/p9/core/rpc/coreclient"
)

const (
	menuTypeDir    int32 = 0 // 目录
	menuTypeMenu   int32 = 1 // 菜单
	menuTypeButton int32 = 2 // 按钮
)

// registerRequest 创建Node Dispatch目录注册请求
func registerRequest() *coreclient.RegisterCatalogReq {
	return &coreclient.RegisterCatalogReq{
		Menus: menus(),
		Apis:  apis(),
		I18N:  nodeDispatchI18n(),
	}
}

// menus 返回Node Dispatch菜单目录
func menus() []*coreclient.RegisterMenuReq {
	return []*coreclient.RegisterMenuReq{
		// 调度中心
		{Name: "DispatchCenter", Title: "menu.route.dispatchCenter", Path: "/dispatch", MenuType: menuTypeDir, Sort: 50},

		// 节点管理
		{Name: "NodeManagement", Title: "menu.route.nodeManagement", Path: "/dispatch/node", MenuType: menuTypeMenu, Component: "dispatch/node/index", ParentName: "DispatchCenter", Sort: 51},
		{Name: "NodeDetail", Title: "menu.route.nodeDetail", MenuType: menuTypeButton, Permission: "node:detail", ParentName: "NodeManagement", Sort: 511},
		{Name: "NodeCreate", Title: "menu.route.nodeCreate", MenuType: menuTypeButton, Permission: "node:create", ParentName: "NodeManagement", Sort: 512},
		{Name: "NodeUpdate", Title: "menu.route.nodeUpdate", MenuType: menuTypeButton, Permission: "node:update", ParentName: "NodeManagement", Sort: 513},
		{Name: "NodeResetAuthSecret", Title: "menu.route.nodeResetAuthSecret", MenuType: menuTypeButton, Permission: "node:resetAuthSecret", ParentName: "NodeManagement", Sort: 514},

		// 调度任务
		{Name: "DispatchTask", Title: "menu.route.dispatchTask", Path: "/dispatch/task", MenuType: menuTypeMenu, Component: "dispatch/task/index", ParentName: "DispatchCenter", Sort: 52},
		{Name: "DispatchTaskDetail", Title: "menu.route.dispatchTaskDetail", MenuType: menuTypeButton, Permission: "dispatchTask:detail", ParentName: "DispatchTask", Sort: 521},
	}
}

// apis 返回Node Dispatch API目录
func apis() []*coreclient.CreateApiReq {
	return []*coreclient.CreateApiReq{
		// 节点
		{Path: "/admin/node/create", Method: http.MethodPost, Description: "api.nodeCreate", ApiGroup: "api.group.node", ServiceName: "node-dispatch-api"},
		{Path: "/admin/node/update", Method: http.MethodPost, Description: "api.nodeUpdate", ApiGroup: "api.group.node", ServiceName: "node-dispatch-api"},
		{Path: "/admin/node/get", Method: http.MethodGet, Description: "api.nodeGet", ApiGroup: "api.group.node", ServiceName: "node-dispatch-api"},
		{Path: "/admin/node/list", Method: http.MethodGet, Description: "api.nodeList", ApiGroup: "api.group.node", ServiceName: "node-dispatch-api"},
		{Path: "/admin/node/reset-auth-secret", Method: http.MethodPost, Description: "api.nodeResetAuthSecret", ApiGroup: "api.group.node", ServiceName: "node-dispatch-api"},

		// 调度任务
		{Path: "/admin/dispatch-task/get", Method: http.MethodGet, Description: "api.dispatchTaskGet", ApiGroup: "api.group.dispatch_task", ServiceName: "node-dispatch-api"},
		{Path: "/admin/dispatch-task/list", Method: http.MethodGet, Description: "api.dispatchTaskList", ApiGroup: "api.group.dispatch_task", ServiceName: "node-dispatch-api"},
	}
}
