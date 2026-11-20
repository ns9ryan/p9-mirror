package catalog

import (
	"net/http"

	"oa.98ent.com/p9/common/i18n"
	"oa.98ent.com/p9/core/rpc/coreclient"
)

// registerRequest 汇总需要权限校验的四个网站管理接口，公开查询站点编码的接口不在此登记。
// 此处只构造登记数据，不注册 HTTP 路由、不校验 Token，也不向 Core 发送 RPC。
func registerRequest() *coreclient.RegisterCatalogReq {
	return &coreclient.RegisterCatalogReq{
		// Path 和 Method 共同标识接口；Description、ApiGroup 为翻译键，ServiceName 标识所属服务。
		Apis: []*coreclient.CreateApiReq{
			// 查询本站网站设置及语言、地区等配置详情。
			{
				Path:        "/operator/info",
				Method:      http.MethodGet,
				Description: "api.operatorSiteInfo",
				ApiGroup:    "api.group.operatorSite",
				ServiceName: "operator-base-api",
			},
			{
				// 完整保存八项设置，允许用空 Logo 清除图片。
				Path:        "/operator/settings",
				Method:      http.MethodPut,
				Description: "api.operatorSiteSettingsSave",
				ApiGroup:    "api.group.operatorSite",
				ServiceName: "operator-base-api",
			},
			{
				// 保存已分配地区中的完整勾选结果，不修改总网分配范围。
				Path:        "/operator/regions",
				Method:      http.MethodPut,
				Description: "api.operatorSiteRegionsUpdate",
				ApiGroup:    "api.group.operatorSite",
				ServiceName: "operator-base-api",
			},
			{
				// 保存已分配语言中的完整勾选结果，不新增总网语言资料。
				Path:        "/operator/languages",
				Method:      http.MethodPut,
				Description: "api.operatorSiteLanguagesUpdate",
				ApiGroup:    "api.group.operatorSite",
				ServiceName: "operator-base-api",
			},
		},
		// 为总站和分站的权限目录提供名称翻译；登记名称本身不向用户授予权限。
		I18N: append(apiI18n(i18n.CodePlatform), apiI18n(i18n.CodeOperator)...),
	}
}

// apiI18n 为指定业务编码生成简体中文、繁体中文和英文的接口分组及操作名称。
func apiI18n(code string) []*coreclient.I18NItem {
	var items []*coreclient.I18NItem
	// 每个翻译键写入三种语言；业务编码区分总站与分站，翻译内容不参与权限判断。
	add := func(key, zh, hk, en string) {
		items = append(items,
			&coreclient.I18NItem{
				I18NCode:  code,          // 所属业务：总站或分站
				I18NGroup: i18n.GroupAPI, // 接口权限目录的翻译分组
				TransKey:  key,           // 与接口 Description 或 ApiGroup 对应的翻译键
				Lang:      i18n.LangZH,   // 当前条目的语言
				Value:     zh,            // 对应语言下显示的名称
			},
			&coreclient.I18NItem{
				I18NCode:  code,
				I18NGroup: i18n.GroupAPI,
				TransKey:  key,
				Lang:      i18n.LangHK,
				Value:     hk,
			},
			&coreclient.I18NItem{
				I18NCode:  code,
				I18NGroup: i18n.GroupAPI,
				TransKey:  key,
				Lang:      i18n.LangEN,
				Value:     en,
			},
		)
	}
	add("api.group.operatorSite", "网站管理", "網站管理", "Website management")
	add("api.operatorSiteInfo", "查看网站配置", "查看網站配置", "View website configuration")
	add("api.operatorSiteSettingsSave", "保存网站配置", "儲存網站配置", "Save website configuration")
	add("api.operatorSiteRegionsUpdate", "修改经营地区", "修改經營地區", "Update operating regions")
	add("api.operatorSiteLanguagesUpdate", "修改开放语言", "修改開放語言", "Update website languages")

	return items
}
