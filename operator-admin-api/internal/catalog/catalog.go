package catalog

import (
	"net/http"

	"oa.98ent.com/p9/core/rpc/coreclient"
)

const (
	menuTypeDir    int32 = 0
	menuTypeMenu   int32 = 1
	menuTypeButton int32 = 2

	serviceName = "operator-admin-api"
)

// CatalogReq 菜单、API目录、多语言数据种子数据
func catalogReq(code string) *coreclient.RegisterCatalogReq {
	return &coreclient.RegisterCatalogReq{
		Menus: []*coreclient.RegisterMenuReq{
			{Name: "Dashboard", Title: "menu.route.dashboard", Path: "/dashboard", MenuType: menuTypeMenu, Component: "dashboard/index", Sort: 1},
			{Name: "Agent", Title: "menu.route.agent", Path: "/agent", MenuType: menuTypeDir, Sort: 20},
			{Name: "AgentList", Title: "menu.route.agentList", Path: "/agent/list", MenuType: menuTypeMenu, Component: "agent/list/index", ParentName: "Agent", Sort: 21},
			{Name: "AgentTree", Title: "menu.route.agentTree", Path: "/agent/tree", MenuType: menuTypeMenu, Component: "agent/tree/index", ParentName: "Agent", Sort: 22},
			{Name: "AgentConfig", Title: "menu.route.agentConfig", Path: "/agent/config", MenuType: menuTypeMenu, Component: "agent/config/index", ParentName: "Agent", Sort: 23},
			{Name: "AgentLevel", Title: "menu.route.agentLevel", Path: "/agent/level", MenuType: menuTypeMenu, Component: "agent/level/index", ParentName: "Agent", Sort: 24},
			{Name: "AgentCreate", Title: "menu.route.agentCreate", MenuType: menuTypeButton, Permission: "agent:create", ParentName: "AgentList", Sort: 211},
			{Name: "AgentSearch", Title: "menu.route.agentSearch", MenuType: menuTypeButton, Permission: "agent:search", ParentName: "AgentList", Sort: 212},
			{Name: "AgentConfigUpdate", Title: "menu.route.agentConfigUpdate", MenuType: menuTypeButton, Permission: "agent:config:update", ParentName: "AgentConfig", Sort: 231},
			{Name: "AgentLevelUpdate", Title: "menu.route.agentLevelUpdate", MenuType: menuTypeButton, Permission: "agent:level:update", ParentName: "AgentLevel", Sort: 241},
		},
		Apis: []*coreclient.CreateApiReq{
			{Path: "/agent/list", Method: http.MethodPost, Description: "api.agentList", ApiGroup: "api.group.agent", ServiceName: serviceName},
			{Path: "/agent/create", Method: http.MethodPost, Description: "api.agentCreate", ApiGroup: "api.group.agent", ServiceName: serviceName},
			{Path: "/agent/search", Method: http.MethodPost, Description: "api.agentSearch", ApiGroup: "api.group.agent", ServiceName: serviceName},
			{Path: "/agent/operator/list", Method: http.MethodGet, Description: "api.agentOperatorList", ApiGroup: "api.group.agent", ServiceName: serviceName},
			{Path: "/agent/level/list", Method: http.MethodGet, Description: "api.agentLevelList", ApiGroup: "api.group.agent", ServiceName: serviceName},
			{Path: "/agent/level/update", Method: http.MethodPost, Description: "api.agentLevelUpdate", ApiGroup: "api.group.agent", ServiceName: serviceName},
			{Path: "/agent/config", Method: http.MethodGet, Description: "api.agentConfigGet", ApiGroup: "api.group.agent", ServiceName: serviceName},
			{Path: "/agent/config/update", Method: http.MethodPost, Description: "api.agentConfigUpdate", ApiGroup: "api.group.agent", ServiceName: serviceName},
			{Path: "/agent/tree/node", Method: http.MethodGet, Description: "api.agentTreeNode", ApiGroup: "api.group.agent", ServiceName: serviceName},
		},
		I18N: i18nSeeds(code),
	}
}
