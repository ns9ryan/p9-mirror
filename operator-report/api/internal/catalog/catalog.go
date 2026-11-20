package catalog

import (
	"net/http"

	"oa.98ent.com/p9/core/rpc/coreclient"
)

const (
	menuTypeDir    int32 = 0
	menuTypeMenu   int32 = 1
	menuTypeButton int32 = 2
)

// AdminReq 管理员请求(菜单相关种子数据)
func AdminReq(code string) *coreclient.RegisterCatalogReq {
	return &coreclient.RegisterCatalogReq{
		Menus: []*coreclient.RegisterMenuReq{
			{Name: "Dashboard", Title: "menu.route.dashboard", Path: "/dashboard", MenuType: menuTypeDir, Sort: 50},
		},
		Apis: []*coreclient.CreateApiReq{
			{Path: "/dashboard/get-member-growth-trend", Method: http.MethodPost, Description: "api.getMemberGrowthTrend", ApiGroup: "dashboard", ServiceName: "operator-report-api"},
			{Path: "/dashboard/get-member-summary", Method: http.MethodGet, Description: "api.getMemberSummary", ApiGroup: "dashboard", ServiceName: "operator-report-api"},
		},
		I18N: append(append(menuI18n(code), apiI18n(code)...)),
		// I18NLangs: langSeeds(),
	}
}
