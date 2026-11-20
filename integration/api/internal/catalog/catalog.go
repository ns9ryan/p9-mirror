package catalog

import (
	"net/http"

	"oa.98ent.com/p9/core/rpc/coreclient"
)

const (
	menuTypeDir  int32 = 0
	menuTypeMenu int32 = 1

	serviceName = "integration-api"
	apiGroup    = "api.group.ipgeo"
	prefix      = "/p9/integration/ipgeo"
)

// CatalogReq integration-api 启动时写入 core 的菜单 + 需 Casbin 鉴权的 HTTP path。
// path 必须是完整 URL（含 prefix），与 routes / r.URL.Path 一致。
// Component 须与前端 views 一致。
func CatalogReq() *coreclient.RegisterCatalogReq {
	return &coreclient.RegisterCatalogReq{
		Menus: []*coreclient.RegisterMenuReq{
			{Name: "Integration", Title: "menu.route.integration", MenuType: menuTypeDir, Path: "/integration", Sort: 30},
			{Name: "IpGeo", Title: "menu.route.ipGeoLocal", MenuType: menuTypeMenu, Path: "/integration/ipgeo", Component: "integration/ipgeo/index", ParentName: "Integration", Sort: 31},
			{Name: "IpGeoPolicy", Title: "menu.route.ipGeoPolicy", MenuType: menuTypeMenu, Path: "/integration/strategy", Component: "integration/strategy/index", ParentName: "Integration", Sort: 32},
		},
		Apis: []*coreclient.CreateApiReq{
			api("api.ipgeoSearch", http.MethodPost, "/ip/search"),
			api("api.ipgeoPolicyList", http.MethodPost, "/policy/list"),
			api("api.ipgeoPolicySave", http.MethodPost, "/policy/save"),
			api("api.ipgeoPolicyDefaultSave", http.MethodPost, "/policy/default/save"),
			api("api.ipgeoLocalList", http.MethodPost, "/local/list"),
			api("api.ipgeoLocalCreate", http.MethodPost, "/local/create"),
			api("api.ipgeoLocalUpdate", http.MethodPost, "/local/update"),
			api("api.ipgeoLocalDelete", http.MethodPost, "/local/delete"),
			api("api.ipgeoLocalDetail", http.MethodGet, "/local/detail"),
		},
		I18N: append(menuI18n(), apiI18n()...),
	}
}

func api(desc, method, path string) *coreclient.CreateApiReq {
	return &coreclient.CreateApiReq{
		Description: desc,
		ApiGroup:    apiGroup,
		Method:      method,
		Path:        prefix + path,
		ServiceName: serviceName,
	}
}
