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

// registerRequest 汇总菜单、接口及翻译数据；这里只构造请求，实际登记由 Register 调用 Core 完成。
func registerRequest() *coreclient.RegisterCatalogReq {
	return &coreclient.RegisterCatalogReq{
		Menus: append(append(menus(), avatarResourceMenus()...), nicknameMenus()...),          // 菜单层级和按钮权限。
		Apis:  append(append(apis(), avatarResourceApis()...), nicknameAPIs()...),             // 需要登记到 Core 的 HTTP 接口目录。
		I18N:  append(append(platformBaseI18n(), avatarResourceI18n()...), nicknameI18n()...), // 菜单和接口说明的多语言文本。
	}
}

// menus 返回Platform Base菜单目录
func menus() []*coreclient.RegisterMenuReq {
	return []*coreclient.RegisterMenuReq{
		// 基础数据
		{Name: "BaseData", Title: "menu.route.baseData", Path: "/base-data", MenuType: menuTypeDir, Sort: 20},

		// 时区管理
		{Name: "Timezone", Title: "menu.route.timezone", Path: "/base-data/timezone", MenuType: menuTypeMenu, Component: "base-data/timezone/index", ParentName: "BaseData", Sort: 21},
		{Name: "TimezoneUpdate", Title: "menu.route.timezoneUpdate", MenuType: menuTypeButton, Permission: "timezone:update", ParentName: "Timezone", Sort: 211},
		{Name: "TimezoneReorder", Title: "menu.route.timezoneReorder", MenuType: menuTypeButton, Permission: "timezone:reorder", ParentName: "Timezone", Sort: 212},

		// 货币管理
		{Name: "Currency", Title: "menu.route.currency", Path: "/base-data/currency", MenuType: menuTypeMenu, Component: "base-data/currency/index", ParentName: "BaseData", Sort: 22},
		{Name: "CurrencyUpdate", Title: "menu.route.currencyUpdate", MenuType: menuTypeButton, Permission: "currency:update", ParentName: "Currency", Sort: 221},
		{Name: "CurrencyReorder", Title: "menu.route.currencyReorder", MenuType: menuTypeButton, Permission: "currency:reorder", ParentName: "Currency", Sort: 222},

		// 国家地区管理
		{Name: "Region", Title: "menu.route.region", Path: "/base-data/region", MenuType: menuTypeMenu, Component: "base-data/region/index", ParentName: "BaseData", Sort: 23},
		{Name: "RegionUpdate", Title: "menu.route.regionUpdate", MenuType: menuTypeButton, Permission: "region:update", ParentName: "Region", Sort: 231},
		{Name: "RegionReorder", Title: "menu.route.regionReorder", MenuType: menuTypeButton, Permission: "region:reorder", ParentName: "Region", Sort: 232},

		// 默认 VIP 图标
		{
			Name:       "VipDefaultIcon",                   // 菜单或按钮的名称。
			Title:      "menu.route.vipDefaultIcon",        // 菜单或按钮标题的翻译键。
			Path:       "/base-data/vip-default-icon",      // 前端页面路由。
			MenuType:   menuTypeMenu,                       // 页面菜单。
			HideMenu:   1,                                  // 不下发前端导航，保留原页面及按钮权限。
			Component:  "base-data/vip-default-icon/index", // 前端页面组件。
			ParentName: "BaseData",                         // 归属的上级菜单名称。
			Sort:       24,                                 // 同级菜单或按钮的排序值。
		},
		{
			Name:       "VipDefaultIconSave",            // 菜单或按钮的名称。
			Title:      "menu.route.vipDefaultIconSave", // 菜单或按钮标题的翻译键。
			MenuType:   menuTypeButton,                  // 按钮权限项。
			Permission: "vip-default-icon:save",         // 按钮操作的权限标识。
			ParentName: "VipDefaultIcon",                // 归属的上级菜单名称。
			Sort:       241,                             // 同级菜单或按钮的排序值。
		},
		{
			Name:       "VipDefaultIconUpload",            // 菜单或按钮的名称。
			Title:      "menu.route.vipDefaultIconUpload", // 菜单或按钮标题的翻译键。
			MenuType:   menuTypeButton,                    // 按钮权限项。
			Permission: "vip-default-icon:upload",         // 按钮操作的权限标识。
			ParentName: "VipDefaultIcon",                  // 归属的上级菜单名称。
			Sort:       242,                               // 同级菜单或按钮的排序值。
		},
	}
}

// apis 返回Platform Base API目录
func apis() []*coreclient.CreateApiReq {
	return []*coreclient.CreateApiReq{
		// 时区
		{Path: "/admin/timezone/update", Method: http.MethodPost, Description: "api.timezoneUpdate", ApiGroup: "timezone", ServiceName: "platform-base-api"},
		{Path: "/admin/timezone/get", Method: http.MethodGet, Description: "api.timezoneGet", ApiGroup: "timezone", ServiceName: "platform-base-api"},
		{Path: "/admin/timezone/list", Method: http.MethodGet, Description: "api.timezoneList", ApiGroup: "timezone", ServiceName: "platform-base-api"},
		{Path: "/admin/timezone/list-all", Method: http.MethodGet, Description: "api.timezoneListAll", ApiGroup: "timezone", ServiceName: "platform-base-api"},
		{Path: "/admin/timezone/reorder", Method: http.MethodPost, Description: "api.timezoneReorder", ApiGroup: "timezone", ServiceName: "platform-base-api"},

		// 货币
		{Path: "/admin/currency/update", Method: http.MethodPost, Description: "api.currencyUpdate", ApiGroup: "currency", ServiceName: "platform-base-api"},
		{Path: "/admin/currency/get", Method: http.MethodGet, Description: "api.currencyGet", ApiGroup: "currency", ServiceName: "platform-base-api"},
		{Path: "/admin/currency/list", Method: http.MethodGet, Description: "api.currencyList", ApiGroup: "currency", ServiceName: "platform-base-api"},
		{Path: "/admin/currency/list-all", Method: http.MethodGet, Description: "api.currencyListAll", ApiGroup: "currency", ServiceName: "platform-base-api"},
		{Path: "/admin/currency/reorder", Method: http.MethodPost, Description: "api.currencyReorder", ApiGroup: "currency", ServiceName: "platform-base-api"},

		// 国家地区
		{Path: "/admin/region/update", Method: http.MethodPost, Description: "api.regionUpdate", ApiGroup: "region", ServiceName: "platform-base-api"},
		{Path: "/admin/region/get", Method: http.MethodGet, Description: "api.regionGet", ApiGroup: "region", ServiceName: "platform-base-api"},
		{Path: "/admin/region/list", Method: http.MethodGet, Description: "api.regionList", ApiGroup: "region", ServiceName: "platform-base-api"},
		{Path: "/admin/region/list-all", Method: http.MethodGet, Description: "api.regionListAll", ApiGroup: "region", ServiceName: "platform-base-api"},
		{Path: "/admin/region/reorder", Method: http.MethodPost, Description: "api.regionReorder", ApiGroup: "region", ServiceName: "platform-base-api"},

		// 默认 VIP 图标
		{
			Path:        "/admin/vip-default-icon/list-all", // 查询全部默认图片的接口路径。
			Method:      http.MethodGet,                     // GET 查询请求。
			Description: "api.vipDefaultIconListAll",        // 接口说明的翻译键。
			ApiGroup:    "vip-default-icon",                 // 接口目录分组。
			ServiceName: "platform-base-api",                // 接口所属服务。
		},
		{
			Path:        "/admin/vip-default-icon/save", // 保存默认 VIP 图标配置的接口路径。
			Method:      http.MethodPost,                // POST 保存、修改或上传请求。
			Description: "api.vipDefaultIconSave",       // 接口说明的翻译键。
			ApiGroup:    "vip-default-icon",             // 接口目录分组。
			ServiceName: "platform-base-api",            // 接口所属服务。
		},
		{
			Path:        "/resource/upload",   // 统一图片上传入口。
			Method:      http.MethodPost,      // POST 保存、修改或上传请求。
			Description: "api.resourceUpload", // 接口说明的翻译键。
			ApiGroup:    "resource",           // 接口目录分组。
			ServiceName: "platform-base-api",  // 接口所属服务。
		},
	}
}
