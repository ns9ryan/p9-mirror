package catalog

import (
	"net/http"

	"oa.98ent.com/p9/core/rpc/coreclient"
)

const (
	// menuTypeDir  int32 = 0
	// menuTypeMenu int32 = 1

	serviceName = "oss-api"
	prefix      = "/p9/oss"
)

// CatalogReq oss-api 启动时写入 core 的菜单 + 需 Casbin 鉴权的 HTTP path。
func CatalogReq() *coreclient.RegisterCatalogReq {
	return &coreclient.RegisterCatalogReq{
		// 暂未使用菜单
		// Menus: []*coreclient.RegisterMenuReq{
		// 	{Name: "Oss", Title: "menu.route.oss", MenuType: menuTypeDir, Path: "/oss", Sort: 40},
		// 	{Name: "OssFile", Title: "menu.route.ossFile", MenuType: menuTypeMenu, Path: "/oss/file", Component: "oss/file/index", ParentName: "Oss", Sort: 42},
		// 	{Name: "OssProvider", Title: "menu.route.ossProvider", MenuType: menuTypeMenu, Path: "/oss/provider", Component: "oss/provider/index", ParentName: "Oss", Sort: 43},
		// },
		Apis: []*coreclient.CreateApiReq{
			api("api.group.oss", "api.ossUrlDetail", http.MethodGet, "/url/detail"),
			api("api.group.oss", "api.ossFileUpload", http.MethodPost, "/file/upload"),
			api("api.group.oss", "api.ossUploadPresign", http.MethodPost, "/upload/presign"),
			api("api.group.oss", "api.ossUploadComplete", http.MethodPost, "/upload/complete"),
			api("api.group.oss", "api.ossFileList", http.MethodPost, "/file/list"),
			api("api.group.oss", "api.ossFileDelete", http.MethodPost, "/file/delete"),
			api("api.group.oss", "api.ossProviderList", http.MethodPost, "/provider/list"),
			api("api.group.oss", "api.ossProviderCreate", http.MethodPost, "/provider/create"),
			api("api.group.oss", "api.ossProviderActive", http.MethodGet, "/provider/active"),
			api("api.group.oss", "api.ossProviderDetail", http.MethodGet, "/provider/detail"),
			api("api.group.oss", "api.ossProviderUpdate", http.MethodPost, "/provider/update"),
			api("api.group.oss", "api.ossProviderDelete", http.MethodPost, "/provider/delete"),
			api("api.group.oss", "api.ossProviderEnable", http.MethodPost, "/provider/enable"),
		},
		I18N: append(menuI18n(), apiI18n()...),
	}
}

func api(group, desc, method, path string) *coreclient.CreateApiReq {
	return &coreclient.CreateApiReq{
		Description: desc,
		ApiGroup:    group,
		Method:      method,
		Path:        prefix + path,
		ServiceName: serviceName,
	}
}
