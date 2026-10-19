package catalog

import (
	"oa.98ent.com/p9/common/i18n"
	"oa.98ent.com/p9/core/rpc/coreclient"
)

const i18nGroupBase = "base" // 基础数据多语言分组

// platformBaseI18n 返回Platform Base全部多语言数据
func platformBaseI18n() []*coreclient.I18NItem {
	var out []*coreclient.I18NItem

	out = append(out, menuI18n()...)
	out = append(out, apiI18n()...)
	out = append(out, baseDataI18n()...)

	return out
}

// menuI18n 返回菜单多语言数据
func menuI18n() []*coreclient.I18NItem {
	var out []*coreclient.I18NItem

	// 基础数据
	addI18n(&out, i18n.GroupMenu, "menu.route.baseData", "基础数据", "基礎數據", "Base data")

	// 时区
	addI18n(&out, i18n.GroupMenu, "menu.route.timezone", "时区管理", "時區管理", "Timezones")
	addI18n(&out, i18n.GroupMenu, "menu.route.timezoneUpdate", "编辑时区", "編輯時區", "Edit timezone")
	addI18n(&out, i18n.GroupMenu, "menu.route.timezoneReorder", "调整时区排序", "調整時區排序", "Reorder timezones")

	// 货币
	addI18n(&out, i18n.GroupMenu, "menu.route.currency", "货币管理", "貨幣管理", "Currencies")
	addI18n(&out, i18n.GroupMenu, "menu.route.currencyUpdate", "编辑货币", "編輯貨幣", "Edit currency")
	addI18n(&out, i18n.GroupMenu, "menu.route.currencyReorder", "调整货币排序", "調整貨幣排序", "Reorder currencies")

	// 国家地区
	addI18n(&out, i18n.GroupMenu, "menu.route.region", "国家地区", "國家地區", "Countries and regions")
	addI18n(&out, i18n.GroupMenu, "menu.route.regionUpdate", "编辑国家地区", "編輯國家地區", "Edit country or region")
	addI18n(&out, i18n.GroupMenu, "menu.route.regionReorder", "调整国家地区排序", "調整國家地區排序", "Reorder countries and regions")

	return out
}

// apiI18n 返回API多语言数据
func apiI18n() []*coreclient.I18NItem {
	var out []*coreclient.I18NItem

	// 时区
	addI18n(&out, i18n.GroupAPI, "api.timezoneUpdate", "更新时区", "更新時區", "Update timezone")
	addI18n(&out, i18n.GroupAPI, "api.timezoneGet", "时区详情", "時區詳情", "Timezone detail")
	addI18n(&out, i18n.GroupAPI, "api.timezoneList", "时区列表", "時區列表", "Timezone list")
	addI18n(&out, i18n.GroupAPI, "api.timezoneListAll", "全部时区", "全部時區", "All timezones")
	addI18n(&out, i18n.GroupAPI, "api.timezoneReorder", "调整时区排序", "調整時區排序", "Reorder timezones")

	// 货币
	addI18n(&out, i18n.GroupAPI, "api.currencyUpdate", "更新货币", "更新貨幣", "Update currency")
	addI18n(&out, i18n.GroupAPI, "api.currencyGet", "货币详情", "貨幣詳情", "Currency detail")
	addI18n(&out, i18n.GroupAPI, "api.currencyList", "货币列表", "貨幣列表", "Currency list")
	addI18n(&out, i18n.GroupAPI, "api.currencyListAll", "全部货币", "全部貨幣", "All currencies")
	addI18n(&out, i18n.GroupAPI, "api.currencyReorder", "调整货币排序", "調整貨幣排序", "Reorder currencies")

	// 国家地区
	addI18n(&out, i18n.GroupAPI, "api.regionUpdate", "更新国家地区", "更新國家地區", "Update country or region")
	addI18n(&out, i18n.GroupAPI, "api.regionGet", "国家地区详情", "國家地區詳情", "Country or region detail")
	addI18n(&out, i18n.GroupAPI, "api.regionList", "国家地区列表", "國家地區列表", "Country and region list")
	addI18n(&out, i18n.GroupAPI, "api.regionListAll", "全部国家地区", "全部國家地區", "All countries and regions")
	addI18n(&out, i18n.GroupAPI, "api.regionReorder", "调整国家地区排序", "調整國家地區排序", "Reorder countries and regions")

	return out
}

// baseDataI18n 返回基础数据多语言数据
func baseDataI18n() []*coreclient.I18NItem {
	var out []*coreclient.I18NItem

	// 时区
	addI18n(&out, i18nGroupBase, "base.timezone.etc_utc.name", "协调世界时", "協調世界時", "Coordinated Universal Time")
	addI18n(&out, i18nGroupBase, "base.timezone.asia_tokyo.name", "日本标准时间", "日本標準時間", "Japan Standard Time")
	addI18n(&out, i18nGroupBase, "base.timezone.asia_ho_chi_minh.name", "越南时间", "越南時間", "Vietnam Time")
	addI18n(&out, i18nGroupBase, "base.timezone.asia_manila.name", "菲律宾时间", "菲律賓時間", "Philippine Time")
	addI18n(&out, i18nGroupBase, "base.timezone.asia_bangkok.name", "泰国时间", "泰國時間", "Thailand Time")
	addI18n(&out, i18nGroupBase, "base.timezone.asia_kuala_lumpur.name", "马来西亚时间", "馬來西亞時間", "Malaysia Time")
	addI18n(&out, i18nGroupBase, "base.timezone.asia_singapore.name", "新加坡标准时间", "新加坡標準時間", "Singapore Standard Time")
	addI18n(&out, i18nGroupBase, "base.timezone.asia_jakarta.name", "印度尼西亚西部时间", "印度尼西亞西部時間", "Western Indonesia Time")
	addI18n(&out, i18nGroupBase, "base.timezone.asia_seoul.name", "韩国标准时间", "韓國標準時間", "Korea Standard Time")
	addI18n(&out, i18nGroupBase, "base.timezone.america_sao_paulo.name", "巴西利亚时间", "巴西利亞時間", "Brasilia Time")

	// 货币
	addI18n(&out, i18nGroupBase, "base.currency.usd.name", "美元", "美元", "US Dollar")
	addI18n(&out, i18nGroupBase, "base.currency.php.name", "菲律宾比索", "菲律賓比索", "Philippine Peso")
	addI18n(&out, i18nGroupBase, "base.currency.vnd.name", "越南盾", "越南盾", "Vietnamese Dong")
	addI18n(&out, i18nGroupBase, "base.currency.jpy.name", "日元", "日圓", "Japanese Yen")
	addI18n(&out, i18nGroupBase, "base.currency.myr.name", "马来西亚林吉特", "馬來西亞林吉特", "Malaysian Ringgit")
	addI18n(&out, i18nGroupBase, "base.currency.thb.name", "泰铢", "泰銖", "Thai Baht")
	addI18n(&out, i18nGroupBase, "base.currency.idr.name", "印度尼西亚卢比", "印度尼西亞盧比", "Indonesian Rupiah")
	addI18n(&out, i18nGroupBase, "base.currency.sgd.name", "新加坡元", "新加坡元", "Singapore Dollar")
	addI18n(&out, i18nGroupBase, "base.currency.krw.name", "韩元", "韓元", "South Korean Won")
	addI18n(&out, i18nGroupBase, "base.currency.brl.name", "巴西雷亚尔", "巴西雷亞爾", "Brazilian Real")
	addI18n(&out, i18nGroupBase, "base.currency.usdt.name", "泰达币", "泰達幣", "Tether")

	// 国家地区

	// 北美洲
	addI18n(&out, i18nGroupBase, "base.region.us.name", "美国", "美國", "United States")
	addI18n(&out, i18nGroupBase, "base.region.ca.name", "加拿大", "加拿大", "Canada")
	addI18n(&out, i18nGroupBase, "base.region.mx.name", "墨西哥", "墨西哥", "Mexico")

	// 南美洲
	addI18n(&out, i18nGroupBase, "base.region.br.name", "巴西", "巴西", "Brazil")
	addI18n(&out, i18nGroupBase, "base.region.ar.name", "阿根廷", "阿根廷", "Argentina")
	addI18n(&out, i18nGroupBase, "base.region.co.name", "哥伦比亚", "哥倫比亞", "Colombia")
	addI18n(&out, i18nGroupBase, "base.region.cl.name", "智利", "智利", "Chile")
	addI18n(&out, i18nGroupBase, "base.region.pe.name", "秘鲁", "秘魯", "Peru")

	// 欧洲
	addI18n(&out, i18nGroupBase, "base.region.gb.name", "英国", "英國", "United Kingdom")
	addI18n(&out, i18nGroupBase, "base.region.de.name", "德国", "德國", "Germany")
	addI18n(&out, i18nGroupBase, "base.region.fr.name", "法国", "法國", "France")
	addI18n(&out, i18nGroupBase, "base.region.it.name", "意大利", "意大利", "Italy")
	addI18n(&out, i18nGroupBase, "base.region.es.name", "西班牙", "西班牙", "Spain")
	addI18n(&out, i18nGroupBase, "base.region.pt.name", "葡萄牙", "葡萄牙", "Portugal")
	addI18n(&out, i18nGroupBase, "base.region.nl.name", "荷兰", "荷蘭", "Netherlands")
	addI18n(&out, i18nGroupBase, "base.region.ch.name", "瑞士", "瑞士", "Switzerland")
	addI18n(&out, i18nGroupBase, "base.region.pl.name", "波兰", "波蘭", "Poland")
	addI18n(&out, i18nGroupBase, "base.region.se.name", "瑞典", "瑞典", "Sweden")

	// 亚洲
	addI18n(&out, i18nGroupBase, "base.region.jp.name", "日本", "日本", "Japan")
	addI18n(&out, i18nGroupBase, "base.region.kr.name", "韩国", "韓國", "South Korea")
	addI18n(&out, i18nGroupBase, "base.region.in.name", "印度", "印度", "India")
	addI18n(&out, i18nGroupBase, "base.region.vn.name", "越南", "越南", "Vietnam")
	addI18n(&out, i18nGroupBase, "base.region.ph.name", "菲律宾", "菲律賓", "Philippines")
	addI18n(&out, i18nGroupBase, "base.region.id.name", "印度尼西亚", "印度尼西亞", "Indonesia")
	addI18n(&out, i18nGroupBase, "base.region.my.name", "马来西亚", "馬來西亞", "Malaysia")
	addI18n(&out, i18nGroupBase, "base.region.th.name", "泰国", "泰國", "Thailand")
	addI18n(&out, i18nGroupBase, "base.region.sg.name", "新加坡", "新加坡", "Singapore")
	addI18n(&out, i18nGroupBase, "base.region.ae.name", "阿联酋", "阿聯酋", "United Arab Emirates")
	addI18n(&out, i18nGroupBase, "base.region.sa.name", "沙特阿拉伯", "沙特阿拉伯", "Saudi Arabia")

	// 非洲
	addI18n(&out, i18nGroupBase, "base.region.za.name", "南非", "南非", "South Africa")
	addI18n(&out, i18nGroupBase, "base.region.eg.name", "埃及", "埃及", "Egypt")
	addI18n(&out, i18nGroupBase, "base.region.ng.name", "尼日利亚", "尼日利亞", "Nigeria")
	addI18n(&out, i18nGroupBase, "base.region.ke.name", "肯尼亚", "肯尼亞", "Kenya")
	addI18n(&out, i18nGroupBase, "base.region.ma.name", "摩洛哥", "摩洛哥", "Morocco")

	// 大洋洲
	addI18n(&out, i18nGroupBase, "base.region.au.name", "澳大利亚", "澳大利亞", "Australia")
	addI18n(&out, i18nGroupBase, "base.region.nz.name", "新西兰", "新西蘭", "New Zealand")

	return out
}

// addI18n 添加Platform站点的简体中文、繁体中文和英文翻译
func addI18n(out *[]*coreclient.I18NItem, group, key, zh, hk, en string) {
	*out = append(*out,
		&coreclient.I18NItem{I18NCode: i18n.CodePlatform, I18NGroup: group, TransKey: key, Lang: i18n.LangZH, Value: zh},
		&coreclient.I18NItem{I18NCode: i18n.CodePlatform, I18NGroup: group, TransKey: key, Lang: i18n.LangHK, Value: hk},
		&coreclient.I18NItem{I18NCode: i18n.CodePlatform, I18NGroup: group, TransKey: key, Lang: i18n.LangEN, Value: en},
	)
}
