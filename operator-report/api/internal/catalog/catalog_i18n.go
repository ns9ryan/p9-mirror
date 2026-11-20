package catalog

import (
	"oa.98ent.com/p9/common/i18n"
	"oa.98ent.com/p9/core/rpc/coreclient"
)

const i18nGroupGame = "game" // 游戏管理多语言分组

func langSeeds() []*coreclient.CreateI18NLangReq {
	return []*coreclient.CreateI18NLangReq{
		{Lang: i18n.LangZH, Name: "简体中文", SortNo: 1},
		{Lang: i18n.LangHK, Name: "繁體中文", SortNo: 2},
		{Lang: i18n.LangEN, Name: "English", SortNo: 3},
	}
}

func menuI18n(i18nCode string) []*coreclient.I18NItem {
	var out []*coreclient.I18NItem
	add := func(key, zh, hk, en string) {
		out = append(out,
			&coreclient.I18NItem{I18NCode: i18nCode, I18NGroup: i18n.GroupMenu, TransKey: key, Lang: i18n.LangZH, Value: zh},
			&coreclient.I18NItem{I18NCode: i18nCode, I18NGroup: i18n.GroupMenu, TransKey: key, Lang: i18n.LangHK, Value: hk},
			&coreclient.I18NItem{I18NCode: i18nCode, I18NGroup: i18n.GroupMenu, TransKey: key, Lang: i18n.LangEN, Value: en},
		)
	}
	// add("menu.route.cashWinReport", "输赢管理", "輸贏管理", "Cash Win Management")
	// add("menu.route.gameReport", "游戏报表", "遊戲報表", "Game Report")
	// add("menu.route.operationReport", "运营报表", "運營報表", "Operation Report")
	// add("menu.route.userReport", "用户报表", "用戶報表", "User Report")
	// add("menu.route.gameDetailReport", "游戏明细报表", "遊戲明細報表", "Game Detail Report")
	// add("menu.route.sportDetailReport", "体育明细报表", "體育明細報表", "Sport Detail Report")
	add("menu.route.dashboard", "仪表盘", "儀表盤", "Dashboard")
	return out
}

func apiI18n(i18nCode string) []*coreclient.I18NItem {
	var out []*coreclient.I18NItem
	add := func(key, zh, hk, en string) {
		out = append(out,
			&coreclient.I18NItem{I18NCode: i18nCode, I18NGroup: i18n.GroupAPI, TransKey: key, Lang: i18n.LangZH, Value: zh},
			&coreclient.I18NItem{I18NCode: i18nCode, I18NGroup: i18n.GroupAPI, TransKey: key, Lang: i18n.LangHK, Value: hk},
			&coreclient.I18NItem{I18NCode: i18nCode, I18NGroup: i18n.GroupAPI, TransKey: key, Lang: i18n.LangEN, Value: en},
		)
	}
	// add("api.cashWinReportList", "输赢报表列表", "輸贏報表列表", "Cash Win Report list")
	// add("api.gameReportList", "游戏报表列表", "遊戲報表列表", "Game Report list")
	// add("api.operationReportList", "运营报表列表", "運營報表列表", "Operation Report list")
	// add("api.userReportList", "用户报表列表", "用戶報表列表", "User Report list")
	// add("api.gameDetailReportList", "游戏明细报表列表", "遊戲明細報表列表", "Game Detail Report list")
	// add("api.sportDetailReportList", "体育明细报表列表", "體育明細報表列表", "Sport Detail Report list")
	add("api.dashboardGetMemberGrowthTrend", "获取会员增长趋势", "獲取會員增長趨勢", "Get Member Growth Trend")
	add("api.dashboardGetMemberSummary", "获取会员汇总", "獲取會員匯總", "Get Member Summary")
	return out
}
