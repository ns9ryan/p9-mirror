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
			{Name: "AgentApply", Title: "menu.route.agentApply", HideMenu: 1, Path: "/agent/apply", MenuType: menuTypeMenu, Component: "agent/apply/index", ParentName: "Agent", Sort: 25},
			{Name: "AgentCreate", Title: "menu.route.agentCreate", MenuType: menuTypeButton, Permission: "agent:create", ParentName: "AgentList", Sort: 211},
			{Name: "AgentSearch", Title: "menu.route.agentSearch", MenuType: menuTypeButton, Permission: "agent:search", ParentName: "AgentList", Sort: 212},
			{Name: "AgentProfileGet", Title: "menu.route.agentProfileGet", MenuType: menuTypeButton, Permission: "agent:profile:get", ParentName: "AgentList", Sort: 213},
			{Name: "AgentPasswordReset", Title: "menu.route.agentPasswordReset", MenuType: menuTypeButton, Permission: "agent:password:reset", ParentName: "AgentList", Sort: 214},
			{Name: "AgentWithdrawPasswordReset", Title: "menu.route.agentWithdrawPasswordReset", MenuType: menuTypeButton, Permission: "agent:withdraw-password:reset", ParentName: "AgentList", Sort: 215},
			{Name: "AgentShareRateUpdate", Title: "menu.route.agentShareRateUpdate", MenuType: menuTypeButton, Permission: "agent:share-rate:update", ParentName: "AgentList", Sort: 216},
			{Name: "AgentPayoutUpdate", Title: "menu.route.agentPayoutUpdate", MenuType: menuTypeButton, Permission: "agent:payout:update", ParentName: "AgentList", Sort: 217},
			{Name: "AgentRemarkUpdate", Title: "menu.route.agentRemarkUpdate", MenuType: menuTypeButton, Permission: "agent:remark:update", ParentName: "AgentList", Sort: 218},
			{Name: "AgentProfileUpdate", Title: "menu.route.agentProfileUpdate", MenuType: menuTypeButton, Permission: "agent:profile:update", ParentName: "AgentList", Sort: 219},
			{Name: "AgentConfigUpdate", Title: "menu.route.agentConfigUpdate", MenuType: menuTypeButton, Permission: "agent:config:update", ParentName: "AgentConfig", Sort: 231},
			{Name: "AgentLevelUpdate", Title: "menu.route.agentLevelUpdate", MenuType: menuTypeButton, Permission: "agent:level:update", ParentName: "AgentLevel", Sort: 241},
			{Name: "AgentApplyReview", Title: "menu.route.agentApplyReview", HideMenu: 1, MenuType: menuTypeButton, Permission: "agent:apply:review", ParentName: "AgentApply", Sort: 251},
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
			{Path: "/agent/line/defaults", Method: http.MethodGet, Description: "api.agentLineDefaults", ApiGroup: "api.group.agent", ServiceName: serviceName},
			{Path: "/agent/init/status", Method: http.MethodGet, Description: "api.agentInitStatus", ApiGroup: "api.group.agent", ServiceName: serviceName},
			{Path: "/agent/tree/node", Method: http.MethodGet, Description: "api.agentTreeNode", ApiGroup: "api.group.agent", ServiceName: serviceName},
			{Path: "/agent/tree/ancestors", Method: http.MethodGet, Description: "api.agentTreeAncestors", ApiGroup: "api.group.agent", ServiceName: serviceName},
			{Path: "/agent/apply/list", Method: http.MethodPost, Description: "api.agentApplyList", ApiGroup: "api.group.agent", ServiceName: serviceName},
			{Path: "/agent/apply/review", Method: http.MethodPost, Description: "api.agentApplyReview", ApiGroup: "api.group.agent", ServiceName: serviceName},
			{Path: "/agent/profile", Method: http.MethodGet, Description: "api.agentProfileGet", ApiGroup: "api.group.agent", ServiceName: serviceName},
			{Path: "/agent/password/reset", Method: http.MethodPost, Description: "api.agentPasswordReset", ApiGroup: "api.group.agent", ServiceName: serviceName},
			{Path: "/agent/withdraw-password/reset", Method: http.MethodPost, Description: "api.agentWithdrawPasswordReset", ApiGroup: "api.group.agent", ServiceName: serviceName},
			{Path: "/agent/share-rate/update", Method: http.MethodPost, Description: "api.agentShareRateUpdate", ApiGroup: "api.group.agent", ServiceName: serviceName},
			{Path: "/agent/payout-restricted/update", Method: http.MethodPost, Description: "api.agentPayoutUpdate", ApiGroup: "api.group.agent", ServiceName: serviceName},
			{Path: "/agent/remark/update", Method: http.MethodPost, Description: "api.agentRemarkUpdate", ApiGroup: "api.group.agent", ServiceName: serviceName},
			{Path: "/agent/profile/update", Method: http.MethodPost, Description: "api.agentProfileUpdate", ApiGroup: "api.group.agent", ServiceName: serviceName},
		},
		I18N: i18nSeeds(code),
	}
}
