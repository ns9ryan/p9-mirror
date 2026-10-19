package svc

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/platform-base/rpc/ent"
	"oa.98ent.com/p9/platform-base/rpc/ent/currency"
	"oa.98ent.com/p9/platform-base/rpc/ent/migrate"
	"oa.98ent.com/p9/platform-base/rpc/ent/region"
	"oa.98ent.com/p9/platform-base/rpc/ent/timezone"
)

// MustMigrate 执行数据库自动迁移
func (s *ServiceContext) MustMigrate() {
	ctx := context.Background()

	// 根据 Ent Schema 自动创建或更新数据库结构
	logx.Must(
		s.DB.Schema.Create(
			ctx,
			migrate.WithForeignKeys(false), // 不创建数据库外键
			migrate.WithDropIndex(true),    // 允许删除废弃索引
		),
	)

	// 初始化系统默认数据
	s.mustInitTimezone(ctx)
	s.mustInitCurrency(ctx)
	s.mustInitRegion(ctx)
}

// mustInitTimezone 初始化系统时区
func (s *ServiceContext) mustInitTimezone(ctx context.Context) {
	data := []struct {
		code    string // IANA时区编码
		nameKey string // 名称翻译 Key
		status  int64  // 状态: 1启用, 2停用
		sortNo  int64  // 排序值, 数值越小越靠前
	}{
		// 全球
		{code: "Etc/UTC", nameKey: "timezone.etc_utc.name", status: 1, sortNo: 10},

		// 北美洲
		{code: "America/New_York", nameKey: "timezone.america_new_york.name", status: 1, sortNo: 20},
		{code: "America/Chicago", nameKey: "timezone.america_chicago.name", status: 1, sortNo: 30},
		{code: "America/Denver", nameKey: "timezone.america_denver.name", status: 1, sortNo: 40},
		{code: "America/Los_Angeles", nameKey: "timezone.america_los_angeles.name", status: 1, sortNo: 50},
		{code: "America/Toronto", nameKey: "timezone.america_toronto.name", status: 1, sortNo: 60},
		{code: "America/Vancouver", nameKey: "timezone.america_vancouver.name", status: 1, sortNo: 70},
		{code: "America/Mexico_City", nameKey: "timezone.america_mexico_city.name", status: 1, sortNo: 80},
		{code: "America/Tijuana", nameKey: "timezone.america_tijuana.name", status: 1, sortNo: 90},

		// 南美洲
		{code: "America/Sao_Paulo", nameKey: "timezone.america_sao_paulo.name", status: 1, sortNo: 100},
		{code: "America/Manaus", nameKey: "timezone.america_manaus.name", status: 1, sortNo: 110},
		{code: "America/Argentina/Buenos_Aires", nameKey: "timezone.america_argentina_buenos_aires.name", status: 1, sortNo: 120},
		{code: "America/Bogota", nameKey: "timezone.america_bogota.name", status: 1, sortNo: 130},
		{code: "America/Santiago", nameKey: "timezone.america_santiago.name", status: 1, sortNo: 140},
		{code: "America/Lima", nameKey: "timezone.america_lima.name", status: 1, sortNo: 150},

		// 欧洲
		{code: "Europe/London", nameKey: "timezone.europe_london.name", status: 1, sortNo: 160},
		{code: "Europe/Berlin", nameKey: "timezone.europe_berlin.name", status: 1, sortNo: 170},
		{code: "Europe/Paris", nameKey: "timezone.europe_paris.name", status: 1, sortNo: 180},
		{code: "Europe/Rome", nameKey: "timezone.europe_rome.name", status: 1, sortNo: 190},
		{code: "Europe/Madrid", nameKey: "timezone.europe_madrid.name", status: 1, sortNo: 200},
		{code: "Europe/Lisbon", nameKey: "timezone.europe_lisbon.name", status: 1, sortNo: 210},
		{code: "Europe/Amsterdam", nameKey: "timezone.europe_amsterdam.name", status: 1, sortNo: 220},
		{code: "Europe/Zurich", nameKey: "timezone.europe_zurich.name", status: 1, sortNo: 230},
		{code: "Europe/Warsaw", nameKey: "timezone.europe_warsaw.name", status: 1, sortNo: 240},
		{code: "Europe/Stockholm", nameKey: "timezone.europe_stockholm.name", status: 1, sortNo: 250},

		// 亚洲
		{code: "Asia/Tokyo", nameKey: "timezone.asia_tokyo.name", status: 1, sortNo: 260},
		{code: "Asia/Seoul", nameKey: "timezone.asia_seoul.name", status: 1, sortNo: 270},
		{code: "Asia/Kolkata", nameKey: "timezone.asia_kolkata.name", status: 1, sortNo: 280},
		{code: "Asia/Ho_Chi_Minh", nameKey: "timezone.asia_ho_chi_minh.name", status: 1, sortNo: 290},
		{code: "Asia/Manila", nameKey: "timezone.asia_manila.name", status: 1, sortNo: 300},
		{code: "Asia/Jakarta", nameKey: "timezone.asia_jakarta.name", status: 1, sortNo: 310},
		{code: "Asia/Makassar", nameKey: "timezone.asia_makassar.name", status: 1, sortNo: 320},
		{code: "Asia/Jayapura", nameKey: "timezone.asia_jayapura.name", status: 1, sortNo: 330},
		{code: "Asia/Kuala_Lumpur", nameKey: "timezone.asia_kuala_lumpur.name", status: 1, sortNo: 340},
		{code: "Asia/Bangkok", nameKey: "timezone.asia_bangkok.name", status: 1, sortNo: 350},
		{code: "Asia/Singapore", nameKey: "timezone.asia_singapore.name", status: 1, sortNo: 360},
		{code: "Asia/Dubai", nameKey: "timezone.asia_dubai.name", status: 1, sortNo: 370},
		{code: "Asia/Riyadh", nameKey: "timezone.asia_riyadh.name", status: 1, sortNo: 380},

		// 非洲
		{code: "Africa/Johannesburg", nameKey: "timezone.africa_johannesburg.name", status: 1, sortNo: 390},
		{code: "Africa/Cairo", nameKey: "timezone.africa_cairo.name", status: 1, sortNo: 400},
		{code: "Africa/Lagos", nameKey: "timezone.africa_lagos.name", status: 1, sortNo: 410},
		{code: "Africa/Nairobi", nameKey: "timezone.africa_nairobi.name", status: 1, sortNo: 420},
		{code: "Africa/Casablanca", nameKey: "timezone.africa_casablanca.name", status: 1, sortNo: 430},

		// 大洋洲
		{code: "Australia/Sydney", nameKey: "timezone.australia_sydney.name", status: 1, sortNo: 440},
		{code: "Australia/Brisbane", nameKey: "timezone.australia_brisbane.name", status: 1, sortNo: 450},
		{code: "Australia/Adelaide", nameKey: "timezone.australia_adelaide.name", status: 1, sortNo: 460},
		{code: "Australia/Perth", nameKey: "timezone.australia_perth.name", status: 1, sortNo: 470},
		{code: "Pacific/Auckland", nameKey: "timezone.pacific_auckland.name", status: 1, sortNo: 480},
	}

	builders := make([]*ent.TimezoneCreate, 0, len(data))

	for _, item := range data {
		builders = append(builders,
			s.DB.Timezone.
				Create().
				SetCode(item.code).
				SetNameKey(item.nameKey).
				SetStatus(item.status).
				SetSortNo(item.sortNo),
		)
	}

	// 只补充缺失的系统时区, 已存在的数据保持不变
	logx.Must(
		s.DB.Timezone.
			CreateBulk(builders...).
			OnConflictColumns(timezone.FieldCode).
			Ignore().
			Exec(ctx),
	)
}

// mustInitCurrency 初始化系统货币
func (s *ServiceContext) mustInitCurrency(ctx context.Context) {
	data := []struct {
		code         string // 货币编码
		nameKey      string // 名称翻译 Key
		currencyType int64  // 货币类型: 1法币, 2数字货币
		symbol       string // 货币符号
		amountFactor int64  // 金额换算因子, USD=100, VND=1
		status       int64  // 状态: 1启用, 2停用
		sortNo       int64  // 排序值, 数值越小越靠前
	}{
		// 数字货币
		{code: "USDT", nameKey: "currency.usdt.name", currencyType: 2, symbol: "USDT", amountFactor: 100, status: 1, sortNo: 0},

		// 北美洲
		{code: "USD", nameKey: "currency.usd.name", currencyType: 1, symbol: "$", amountFactor: 100, status: 1, sortNo: 10},
		{code: "CAD", nameKey: "currency.cad.name", currencyType: 1, symbol: "C$", amountFactor: 100, status: 1, sortNo: 20},
		{code: "MXN", nameKey: "currency.mxn.name", currencyType: 1, symbol: "MX$", amountFactor: 100, status: 1, sortNo: 30},

		// 南美洲
		{code: "BRL", nameKey: "currency.brl.name", currencyType: 1, symbol: "R$", amountFactor: 100, status: 1, sortNo: 40},
		{code: "ARS", nameKey: "currency.ars.name", currencyType: 1, symbol: "AR$", amountFactor: 100, status: 1, sortNo: 50},
		{code: "COP", nameKey: "currency.cop.name", currencyType: 1, symbol: "COL$", amountFactor: 100, status: 1, sortNo: 60},
		{code: "CLP", nameKey: "currency.clp.name", currencyType: 1, symbol: "CLP$", amountFactor: 1, status: 1, sortNo: 70},
		{code: "PEN", nameKey: "currency.pen.name", currencyType: 1, symbol: "S/", amountFactor: 100, status: 1, sortNo: 80},

		// 欧洲
		{code: "GBP", nameKey: "currency.gbp.name", currencyType: 1, symbol: "£", amountFactor: 100, status: 1, sortNo: 90},
		{code: "EUR", nameKey: "currency.eur.name", currencyType: 1, symbol: "€", amountFactor: 100, status: 1, sortNo: 100},
		{code: "CHF", nameKey: "currency.chf.name", currencyType: 1, symbol: "CHF", amountFactor: 100, status: 1, sortNo: 110},
		{code: "PLN", nameKey: "currency.pln.name", currencyType: 1, symbol: "zł", amountFactor: 100, status: 1, sortNo: 120},
		{code: "SEK", nameKey: "currency.sek.name", currencyType: 1, symbol: "kr", amountFactor: 100, status: 1, sortNo: 130},

		// 亚洲
		{code: "JPY", nameKey: "currency.jpy.name", currencyType: 1, symbol: "¥", amountFactor: 1, status: 1, sortNo: 140},
		{code: "KRW", nameKey: "currency.krw.name", currencyType: 1, symbol: "₩", amountFactor: 1, status: 1, sortNo: 150},
		{code: "INR", nameKey: "currency.inr.name", currencyType: 1, symbol: "₹", amountFactor: 100, status: 1, sortNo: 160},
		{code: "VND", nameKey: "currency.vnd.name", currencyType: 1, symbol: "₫", amountFactor: 1, status: 1, sortNo: 170},
		{code: "PHP", nameKey: "currency.php.name", currencyType: 1, symbol: "₱", amountFactor: 100, status: 1, sortNo: 180},
		{code: "IDR", nameKey: "currency.idr.name", currencyType: 1, symbol: "Rp", amountFactor: 1, status: 1, sortNo: 190},
		{code: "MYR", nameKey: "currency.myr.name", currencyType: 1, symbol: "RM", amountFactor: 100, status: 1, sortNo: 200},
		{code: "THB", nameKey: "currency.thb.name", currencyType: 1, symbol: "฿", amountFactor: 100, status: 1, sortNo: 210},
		{code: "SGD", nameKey: "currency.sgd.name", currencyType: 1, symbol: "S$", amountFactor: 100, status: 1, sortNo: 220},
		{code: "AED", nameKey: "currency.aed.name", currencyType: 1, symbol: "د.إ", amountFactor: 100, status: 1, sortNo: 230},
		{code: "SAR", nameKey: "currency.sar.name", currencyType: 1, symbol: "ر.س", amountFactor: 100, status: 1, sortNo: 240},

		// 非洲
		{code: "ZAR", nameKey: "currency.zar.name", currencyType: 1, symbol: "R", amountFactor: 100, status: 1, sortNo: 250},
		{code: "EGP", nameKey: "currency.egp.name", currencyType: 1, symbol: "E£", amountFactor: 100, status: 1, sortNo: 260},
		{code: "NGN", nameKey: "currency.ngn.name", currencyType: 1, symbol: "₦", amountFactor: 100, status: 1, sortNo: 270},
		{code: "KES", nameKey: "currency.kes.name", currencyType: 1, symbol: "KSh", amountFactor: 100, status: 1, sortNo: 280},
		{code: "MAD", nameKey: "currency.mad.name", currencyType: 1, symbol: "DH", amountFactor: 100, status: 1, sortNo: 290},

		// 大洋洲
		{code: "AUD", nameKey: "currency.aud.name", currencyType: 1, symbol: "A$", amountFactor: 100, status: 1, sortNo: 300},
		{code: "NZD", nameKey: "currency.nzd.name", currencyType: 1, symbol: "NZ$", amountFactor: 100, status: 1, sortNo: 310},
	}

	builders := make([]*ent.CurrencyCreate, 0, len(data))

	for _, item := range data {
		builders = append(builders,
			s.DB.Currency.
				Create().
				SetCode(item.code).
				SetNameKey(item.nameKey).
				SetCurrencyType(item.currencyType).
				SetSymbol(item.symbol).
				SetAmountFactor(item.amountFactor).
				SetStatus(item.status).
				SetSortNo(item.sortNo),
		)
	}

	// 只补充缺失的系统货币, 已存在的数据保持不变
	logx.Must(
		s.DB.Currency.
			CreateBulk(builders...).
			OnConflictColumns(currency.FieldCode).
			Ignore().
			Exec(ctx),
	)
}

// mustInitRegion 初始化系统国家地区
func (s *ServiceContext) mustInitRegion(ctx context.Context) {
	data := []struct {
		code        string // 国家地区编码
		callingCode string // 国际电话区号
		nameKey     string // 名称翻译 Key
		status      int64  // 状态: 1启用, 2停用
		sortNo      int64  // 排序值, 数值越小越靠前
	}{
		// 北美洲
		{code: "US", callingCode: "1", nameKey: "region.us.name", status: 1, sortNo: 10},
		{code: "CA", callingCode: "1", nameKey: "region.ca.name", status: 1, sortNo: 20},
		{code: "MX", callingCode: "52", nameKey: "region.mx.name", status: 1, sortNo: 30},

		// 南美洲
		{code: "BR", callingCode: "55", nameKey: "region.br.name", status: 1, sortNo: 40},
		{code: "AR", callingCode: "54", nameKey: "region.ar.name", status: 1, sortNo: 50},
		{code: "CO", callingCode: "57", nameKey: "region.co.name", status: 1, sortNo: 60},
		{code: "CL", callingCode: "56", nameKey: "region.cl.name", status: 1, sortNo: 70},
		{code: "PE", callingCode: "51", nameKey: "region.pe.name", status: 1, sortNo: 80},

		// 欧洲
		{code: "GB", callingCode: "44", nameKey: "region.gb.name", status: 1, sortNo: 90},
		{code: "DE", callingCode: "49", nameKey: "region.de.name", status: 1, sortNo: 100},
		{code: "FR", callingCode: "33", nameKey: "region.fr.name", status: 1, sortNo: 110},
		{code: "IT", callingCode: "39", nameKey: "region.it.name", status: 1, sortNo: 120},
		{code: "ES", callingCode: "34", nameKey: "region.es.name", status: 1, sortNo: 130},
		{code: "PT", callingCode: "351", nameKey: "region.pt.name", status: 1, sortNo: 140},
		{code: "NL", callingCode: "31", nameKey: "region.nl.name", status: 1, sortNo: 150},
		{code: "CH", callingCode: "41", nameKey: "region.ch.name", status: 1, sortNo: 160},
		{code: "PL", callingCode: "48", nameKey: "region.pl.name", status: 1, sortNo: 170},
		{code: "SE", callingCode: "46", nameKey: "region.se.name", status: 1, sortNo: 180},

		// 亚洲
		{code: "JP", callingCode: "81", nameKey: "region.jp.name", status: 1, sortNo: 190},
		{code: "KR", callingCode: "82", nameKey: "region.kr.name", status: 1, sortNo: 200},
		{code: "IN", callingCode: "91", nameKey: "region.in.name", status: 1, sortNo: 210},
		{code: "VN", callingCode: "84", nameKey: "region.vn.name", status: 1, sortNo: 220},
		{code: "PH", callingCode: "63", nameKey: "region.ph.name", status: 1, sortNo: 230},
		{code: "ID", callingCode: "62", nameKey: "region.id.name", status: 1, sortNo: 240},
		{code: "MY", callingCode: "60", nameKey: "region.my.name", status: 1, sortNo: 250},
		{code: "TH", callingCode: "66", nameKey: "region.th.name", status: 1, sortNo: 260},
		{code: "SG", callingCode: "65", nameKey: "region.sg.name", status: 1, sortNo: 270},
		{code: "AE", callingCode: "971", nameKey: "region.ae.name", status: 1, sortNo: 280},
		{code: "SA", callingCode: "966", nameKey: "region.sa.name", status: 1, sortNo: 290},

		// 非洲
		{code: "ZA", callingCode: "27", nameKey: "region.za.name", status: 1, sortNo: 300},
		{code: "EG", callingCode: "20", nameKey: "region.eg.name", status: 1, sortNo: 310},
		{code: "NG", callingCode: "234", nameKey: "region.ng.name", status: 1, sortNo: 320},
		{code: "KE", callingCode: "254", nameKey: "region.ke.name", status: 1, sortNo: 330},
		{code: "MA", callingCode: "212", nameKey: "region.ma.name", status: 1, sortNo: 340},

		// 大洋洲
		{code: "AU", callingCode: "61", nameKey: "region.au.name", status: 1, sortNo: 350},
		{code: "NZ", callingCode: "64", nameKey: "region.nz.name", status: 1, sortNo: 360},
	}

	builders := make([]*ent.RegionCreate, 0, len(data))

	for _, item := range data {
		builders = append(builders,
			s.DB.Region.
				Create().
				SetCode(item.code).
				SetCallingCode(item.callingCode).
				SetNameKey(item.nameKey).
				SetStatus(item.status).
				SetSortNo(item.sortNo),
		)
	}

	// 只补充缺失的系统国家地区, 已存在的数据保持不变
	logx.Must(
		s.DB.Region.
			CreateBulk(builders...).
			OnConflictColumns(region.FieldCode).
			Ignore().
			Exec(ctx),
	)
}
