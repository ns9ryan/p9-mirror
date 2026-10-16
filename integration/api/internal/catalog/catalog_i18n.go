package catalog

import (
	"oa.98ent.com/p9/common/i18n"
	"oa.98ent.com/p9/core/rpc/coreclient"
)

func menuI18n() []*coreclient.I18NItem {
	var out []*coreclient.I18NItem
	add := func(key, zh, hk, en string) {
		out = append(out,
			&coreclient.I18NItem{I18NGroup: i18n.GroupMenu, TransKey: key, Lang: i18n.LangZH, Value: zh},
			&coreclient.I18NItem{I18NGroup: i18n.GroupMenu, TransKey: key, Lang: i18n.LangHK, Value: hk},
			&coreclient.I18NItem{I18NGroup: i18n.GroupMenu, TransKey: key, Lang: i18n.LangEN, Value: en},
		)
	}
	add("menu.route.integration", "IP 地理", "IP 地理", "IP Geo")
	add("menu.route.ipGeoLocal", "本地 IP", "本地 IP", "Local IP")
	add("menu.route.ipGeoPolicy", "查询策略", "查詢策略", "Query Policy")
	return out
}

func apiI18n() []*coreclient.I18NItem {
	var out []*coreclient.I18NItem
	add := func(key, zh, hk, en string) {
		out = append(out,
			&coreclient.I18NItem{I18NGroup: i18n.GroupAPI, TransKey: key, Lang: i18n.LangZH, Value: zh},
			&coreclient.I18NItem{I18NGroup: i18n.GroupAPI, TransKey: key, Lang: i18n.LangHK, Value: hk},
			&coreclient.I18NItem{I18NGroup: i18n.GroupAPI, TransKey: key, Lang: i18n.LangEN, Value: en},
		)
	}
	add("api.ipgeoSearch", "IP 地理查询", "IP 地理查詢", "IP geo search")
	add("api.ipgeoPolicyList", "查询策略列表", "查詢策略列表", "Query policy list")
	add("api.ipgeoPolicySave", "保存查询策略", "保存查詢策略", "Save query policy")
	add("api.ipgeoPolicyDefaultSave", "保存默认查询策略", "保存默認查詢策略", "Save default query policy")
	add("api.ipgeoLocalList", "本地 IP 列表", "本地 IP 列表", "Local IP list")
	add("api.ipgeoLocalCreate", "新增本地 IP", "新增本地 IP", "Create local IP")
	add("api.ipgeoLocalUpdate", "更新本地 IP", "更新本地 IP", "Update local IP")
	add("api.ipgeoLocalDelete", "删除本地 IP", "刪除本地 IP", "Delete local IP")
	add("api.ipgeoLocalDetail", "本地 IP 详情", "本地 IP 詳情", "Local IP detail")

	// 接口组多语言翻译
	add("api.group.ipgeo", "IP查询", "IP查詢", "IP query interface")
	return out
}
