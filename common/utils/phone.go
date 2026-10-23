package utils

import (
	"strconv"
	"strings"
	"unicode"
)

// PhoneCallingCode 国际电话区号
type PhoneCallingCode struct {
	Region string // ISO 3166-1 alpha-2
	Code   string // 国际区号，不含加号
	Min    int    // 国内号码最小位数
	Max    int    // 国内号码最大位数
}

type phoneRegion struct {
	region string
	code   string
	min    int
	max    int
}

// 全球常用国家/地区国际区号与国内号码位数（不含区号、不含国内字冠 0）
var phoneRegions = []phoneRegion{
	// 东亚 / 东南亚 / 南亚
	{region: "CN", code: "86", min: 11, max: 11},
	{region: "HK", code: "852", min: 8, max: 8},
	{region: "MO", code: "853", min: 8, max: 8},
	{region: "TW", code: "886", min: 9, max: 9},
	{region: "JP", code: "81", min: 10, max: 10},
	{region: "KR", code: "82", min: 9, max: 10},
	{region: "MN", code: "976", min: 8, max: 8},
	{region: "KP", code: "850", min: 10, max: 10},
	{region: "VN", code: "84", min: 9, max: 10},
	{region: "TH", code: "66", min: 9, max: 9},
	{region: "MY", code: "60", min: 9, max: 10},
	{region: "SG", code: "65", min: 8, max: 8},
	{region: "ID", code: "62", min: 9, max: 12},
	{region: "PH", code: "63", min: 10, max: 10},
	{region: "KH", code: "855", min: 8, max: 9},
	{region: "LA", code: "856", min: 8, max: 10},
	{region: "MM", code: "95", min: 7, max: 10},
	{region: "BN", code: "673", min: 7, max: 7},
	{region: "TL", code: "670", min: 7, max: 8},
	{region: "IN", code: "91", min: 10, max: 10},
	{region: "PK", code: "92", min: 10, max: 10},
	{region: "BD", code: "880", min: 10, max: 10},
	{region: "LK", code: "94", min: 9, max: 9},
	{region: "NP", code: "977", min: 10, max: 10},
	{region: "BT", code: "975", min: 8, max: 8},
	{region: "MV", code: "960", min: 7, max: 7},
	{region: "AF", code: "93", min: 9, max: 9},
	// 中亚 / 西亚
	{region: "KZ", code: "7", min: 10, max: 10},
	{region: "UZ", code: "998", min: 9, max: 9},
	{region: "TM", code: "993", min: 8, max: 8},
	{region: "KG", code: "996", min: 9, max: 9},
	{region: "TJ", code: "992", min: 9, max: 9},
	{region: "AZ", code: "994", min: 9, max: 9},
	{region: "AM", code: "374", min: 8, max: 8},
	{region: "GE", code: "995", min: 9, max: 9},
	{region: "TR", code: "90", min: 10, max: 10},
	{region: "CY", code: "357", min: 8, max: 8},
	{region: "IL", code: "972", min: 8, max: 9},
	{region: "PS", code: "970", min: 9, max: 9},
	{region: "JO", code: "962", min: 9, max: 9},
	{region: "LB", code: "961", min: 7, max: 8},
	{region: "SY", code: "963", min: 9, max: 9},
	{region: "IQ", code: "964", min: 10, max: 10},
	{region: "IR", code: "98", min: 10, max: 10},
	{region: "SA", code: "966", min: 9, max: 9},
	{region: "AE", code: "971", min: 9, max: 9},
	{region: "QA", code: "974", min: 8, max: 8},
	{region: "KW", code: "965", min: 8, max: 8},
	{region: "BH", code: "973", min: 8, max: 8},
	{region: "OM", code: "968", min: 8, max: 8},
	{region: "YE", code: "967", min: 9, max: 9},
	// 欧洲
	{region: "RU", code: "7", min: 10, max: 10},
	{region: "UA", code: "380", min: 9, max: 9},
	{region: "BY", code: "375", min: 9, max: 9},
	{region: "MD", code: "373", min: 8, max: 8},
	{region: "GB", code: "44", min: 10, max: 10},
	{region: "IE", code: "353", min: 9, max: 9},
	{region: "FR", code: "33", min: 9, max: 9},
	{region: "DE", code: "49", min: 10, max: 11},
	{region: "IT", code: "39", min: 9, max: 10},
	{region: "ES", code: "34", min: 9, max: 9},
	{region: "PT", code: "351", min: 9, max: 9},
	{region: "NL", code: "31", min: 9, max: 9},
	{region: "BE", code: "32", min: 9, max: 9},
	{region: "LU", code: "352", min: 9, max: 9},
	{region: "CH", code: "41", min: 9, max: 9},
	{region: "AT", code: "43", min: 10, max: 13},
	{region: "PL", code: "48", min: 9, max: 9},
	{region: "CZ", code: "420", min: 9, max: 9},
	{region: "SK", code: "421", min: 9, max: 9},
	{region: "HU", code: "36", min: 9, max: 9},
	{region: "RO", code: "40", min: 9, max: 9},
	{region: "BG", code: "359", min: 8, max: 9},
	{region: "GR", code: "30", min: 10, max: 10},
	{region: "SE", code: "46", min: 9, max: 9},
	{region: "NO", code: "47", min: 8, max: 8},
	{region: "DK", code: "45", min: 8, max: 8},
	{region: "FI", code: "358", min: 9, max: 10},
	{region: "IS", code: "354", min: 7, max: 7},
	{region: "EE", code: "372", min: 7, max: 8},
	{region: "LV", code: "371", min: 8, max: 8},
	{region: "LT", code: "370", min: 8, max: 8},
	{region: "SI", code: "386", min: 8, max: 8},
	{region: "HR", code: "385", min: 8, max: 9},
	{region: "BA", code: "387", min: 8, max: 8},
	{region: "RS", code: "381", min: 8, max: 9},
	{region: "ME", code: "382", min: 8, max: 8},
	{region: "MK", code: "389", min: 8, max: 8},
	{region: "AL", code: "355", min: 9, max: 9},
	{region: "XK", code: "383", min: 8, max: 8},
	{region: "MT", code: "356", min: 8, max: 8},
	{region: "AD", code: "376", min: 6, max: 6},
	{region: "MC", code: "377", min: 8, max: 9},
	{region: "SM", code: "378", min: 6, max: 10},
	{region: "LI", code: "423", min: 7, max: 7},
	// 北美 / 中美 / 加勒比（NANP 共用 1）
	{region: "US", code: "1", min: 10, max: 10},
	{region: "CA", code: "1", min: 10, max: 10},
	{region: "MX", code: "52", min: 10, max: 10},
	{region: "GT", code: "502", min: 8, max: 8},
	{region: "BZ", code: "501", min: 7, max: 7},
	{region: "SV", code: "503", min: 8, max: 8},
	{region: "HN", code: "504", min: 8, max: 8},
	{region: "NI", code: "505", min: 8, max: 8},
	{region: "CR", code: "506", min: 8, max: 8},
	{region: "PA", code: "507", min: 8, max: 8},
	{region: "CU", code: "53", min: 8, max: 8},
	{region: "HT", code: "509", min: 8, max: 8},
	{region: "JM", code: "1", min: 10, max: 10},
	{region: "DO", code: "1", min: 10, max: 10},
	{region: "PR", code: "1", min: 10, max: 10},
	{region: "TT", code: "1", min: 10, max: 10},
	{region: "BS", code: "1", min: 10, max: 10},
	{region: "BB", code: "1", min: 10, max: 10},
	{region: "AG", code: "1", min: 10, max: 10},
	{region: "DM", code: "1", min: 10, max: 10},
	{region: "GD", code: "1", min: 10, max: 10},
	{region: "KN", code: "1", min: 10, max: 10},
	{region: "LC", code: "1", min: 10, max: 10},
	{region: "VC", code: "1", min: 10, max: 10},
	// 南美
	{region: "BR", code: "55", min: 10, max: 11},
	{region: "AR", code: "54", min: 10, max: 11},
	{region: "CO", code: "57", min: 10, max: 10},
	{region: "CL", code: "56", min: 9, max: 9},
	{region: "PE", code: "51", min: 9, max: 9},
	{region: "VE", code: "58", min: 10, max: 10},
	{region: "EC", code: "593", min: 9, max: 9},
	{region: "BO", code: "591", min: 8, max: 8},
	{region: "PY", code: "595", min: 9, max: 9},
	{region: "UY", code: "598", min: 8, max: 8},
	{region: "GY", code: "592", min: 7, max: 7},
	{region: "SR", code: "597", min: 6, max: 7},
	// 大洋洲
	{region: "AU", code: "61", min: 9, max: 9},
	{region: "NZ", code: "64", min: 8, max: 10},
	{region: "PG", code: "675", min: 8, max: 8},
	{region: "FJ", code: "679", min: 7, max: 7},
	{region: "SB", code: "677", min: 7, max: 7},
	{region: "VU", code: "678", min: 7, max: 7},
	{region: "WS", code: "685", min: 5, max: 7},
	{region: "TO", code: "676", min: 5, max: 7},
	{region: "KI", code: "686", min: 8, max: 8},
	{region: "FM", code: "691", min: 7, max: 7},
	{region: "MH", code: "692", min: 7, max: 7},
	{region: "PW", code: "680", min: 7, max: 7},
	{region: "NR", code: "674", min: 7, max: 7},
	{region: "TV", code: "688", min: 5, max: 6},
	// 非洲
	{region: "ZA", code: "27", min: 9, max: 9},
	{region: "EG", code: "20", min: 10, max: 10},
	{region: "NG", code: "234", min: 10, max: 10},
	{region: "KE", code: "254", min: 9, max: 9},
	{region: "MA", code: "212", min: 9, max: 9},
	{region: "DZ", code: "213", min: 9, max: 9},
	{region: "TN", code: "216", min: 8, max: 8},
	{region: "LY", code: "218", min: 9, max: 9},
	{region: "SD", code: "249", min: 9, max: 9},
	{region: "SS", code: "211", min: 9, max: 9},
	{region: "ET", code: "251", min: 9, max: 9},
	{region: "GH", code: "233", min: 9, max: 9},
	{region: "TZ", code: "255", min: 9, max: 9},
	{region: "UG", code: "256", min: 9, max: 9},
	{region: "AO", code: "244", min: 9, max: 9},
	{region: "MZ", code: "258", min: 9, max: 9},
	{region: "CM", code: "237", min: 8, max: 9},
	{region: "CI", code: "225", min: 10, max: 10},
	{region: "SN", code: "221", min: 9, max: 9},
	{region: "ZW", code: "263", min: 9, max: 9},
	{region: "ZM", code: "260", min: 9, max: 9},
	{region: "RW", code: "250", min: 9, max: 9},
	{region: "MW", code: "265", min: 9, max: 9},
	{region: "MG", code: "261", min: 9, max: 9},
	{region: "ML", code: "223", min: 8, max: 8},
	{region: "BF", code: "226", min: 8, max: 8},
	{region: "NE", code: "227", min: 8, max: 8},
	{region: "TD", code: "235", min: 8, max: 8},
	{region: "CD", code: "243", min: 9, max: 9},
	{region: "CG", code: "242", min: 9, max: 9},
	{region: "GA", code: "241", min: 7, max: 8},
	{region: "GN", code: "224", min: 8, max: 9},
	{region: "BJ", code: "229", min: 8, max: 8},
	{region: "TG", code: "228", min: 8, max: 8},
	{region: "SL", code: "232", min: 8, max: 8},
	{region: "LR", code: "231", min: 7, max: 9},
	{region: "MR", code: "222", min: 8, max: 8},
	{region: "NA", code: "264", min: 9, max: 9},
	{region: "BW", code: "267", min: 7, max: 8},
	{region: "MU", code: "230", min: 8, max: 8},
	{region: "SO", code: "252", min: 8, max: 9},
	{region: "DJ", code: "253", min: 8, max: 8},
	{region: "ER", code: "291", min: 7, max: 7},
	{region: "GM", code: "220", min: 7, max: 7},
	{region: "GW", code: "245", min: 7, max: 7},
	{region: "GQ", code: "240", min: 9, max: 9},
	{region: "CF", code: "236", min: 8, max: 8},
	{region: "BI", code: "257", min: 8, max: 8},
	{region: "LS", code: "266", min: 8, max: 8},
	{region: "SZ", code: "268", min: 8, max: 8},
	{region: "CV", code: "238", min: 7, max: 7},
	{region: "KM", code: "269", min: 7, max: 7},
	{region: "ST", code: "239", min: 7, max: 7},
	{region: "SC", code: "248", min: 7, max: 7},
}

type phoneNSNRange struct {
	min int
	max int
}

var callingCodeNSN map[string]phoneNSNRange
var callingCodeIndex map[string]struct{}
var regionCallingCode map[string]string

func init() {
	callingCodeNSN = make(map[string]phoneNSNRange, len(phoneRegions))
	callingCodeIndex = make(map[string]struct{}, len(phoneRegions))
	regionCallingCode = make(map[string]string, len(phoneRegions))
	for _, item := range phoneRegions {
		regionCallingCode[item.region] = item.code
		callingCodeIndex[item.code] = struct{}{}
		if cur, ok := callingCodeNSN[item.code]; ok {
			if item.min < cur.min {
				cur.min = item.min
			}
			if item.max > cur.max {
				cur.max = item.max
			}
			callingCodeNSN[item.code] = cur
			continue
		}
		callingCodeNSN[item.code] = phoneNSNRange{min: item.min, max: item.max}
	}
}

// CallingCodes 返回支持的国际电话区号列表（不含加号）
func CallingCodes() []PhoneCallingCode {
	out := make([]PhoneCallingCode, 0, len(phoneRegions))
	for _, item := range phoneRegions {
		out = append(out, PhoneCallingCode{Region: item.region, Code: item.code, Min: item.min, Max: item.max})
	}
	return out
}

// CallingCodeByRegion 按 ISO 3166-1 alpha-2 取国际区号
func CallingCodeByRegion(region string) (string, bool) {
	code, ok := regionCallingCode[strings.ToUpper(strings.TrimSpace(region))]
	return code, ok
}

// ParsePhone 解析「区号-手机号码」，区号不含加号。
// 合法示例：60-123456789、+60-123456789、86-13800138000。
func ParsePhone(raw string) (callingCode, nationalNumber string, ok bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", false
	}
	raw = strings.ReplaceAll(raw, " ", "")
	raw = strings.ReplaceAll(raw, "\t", "")
	parts := strings.Split(raw, "-")
	if len(parts) != 2 {
		return "", "", false
	}
	callingCode = strings.TrimPrefix(parts[0], "+")
	if _, err := strconv.Atoi(callingCode); err != nil || callingCode == "" {
		return "", "", false
	}
	if _, ok := callingCodeIndex[callingCode]; !ok {
		return "", "", false
	}
	nationalNumber = parts[1]
	if nationalNumber == "" || strings.ContainsFunc(nationalNumber, func(r rune) bool {
		return !unicode.IsDigit(r)
	}) {
		return "", "", false
	}
	nsn := callingCodeNSN[callingCode]
	n := len(nationalNumber)
	if n < nsn.min || n > nsn.max {
		return "", "", false
	}
	return callingCode, nationalNumber, true
}

// ValidPhone 校验手机号是否为「区号-手机号码」且区号、位数有效
func ValidPhone(raw string) bool {
	_, _, ok := ParsePhone(raw)
	return ok
}

// FormatPhone 格式化为「区号-手机号码」
func FormatPhone(callingCode, nationalNumber string) string {
	callingCode = strings.TrimPrefix(strings.TrimSpace(callingCode), "+")
	nationalNumber = strings.TrimSpace(nationalNumber)
	if callingCode == "" || nationalNumber == "" {
		return ""
	}
	return callingCode + "-" + nationalNumber
}
