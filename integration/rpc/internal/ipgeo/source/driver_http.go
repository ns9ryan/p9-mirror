package source

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"oa.98ent.com/p9/integration/rpc/internal/ipgeo/geo"

	"github.com/tidwall/gjson"
)

// HTTPSource HTTP 在线源：模板 URL、鉴权与 FieldMap。
type HTTPSource struct {
	name      string            // 源名称，写入 Record.Source
	endpoint  string            // URL 模板，支持 {ip}/{lang} 等占位
	authType  string            // none / basic / bearer / header
	accessKey string            // 访问密钥或 basic 用户名
	secretKey string            // 密钥或 basic 密码
	headers   map[string]string // 固定请求头
	fieldMap  map[string]string // gjson 路径 → Record 字段名
	client    *http.Client      // 带超时的 HTTP 客户端
}

// NewHTTPSource 按 Provider 建 HTTP 源。
func NewHTTPSource(p geo.Provider) *HTTPSource {
	return &HTTPSource{
		name:      p.Name,
		endpoint:  p.Endpoint,
		authType:  p.AuthType,
		accessKey: p.AccessKey,
		secretKey: p.SecretKey,
		headers:   p.Headers,
		fieldMap:  p.FieldMap,
		client:    &http.Client{Timeout: time.Duration(p.TimeoutSec) * time.Second},
	}
}

func (h *HTTPSource) Close() error { return nil }

func (h *HTTPSource) Fetch(ctx context.Context, ip, lang string) (*geo.Record, error) {
	url := expandTemplate(h.endpoint, ip, lang)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	applyAuth(req, h.authType, h.accessKey, h.secretKey)
	for k, v := range h.headers {
		req.Header.Set(k, v)
	}
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "application/json")
	}
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status %d: %s", resp.StatusCode, geo.Truncate(string(body), 256))
	}
	rec := &geo.Record{IP: ip, Source: h.name, UpdatedAt: time.Now()}
	applyFieldMap(rec, body, h.fieldMap, lang)
	return rec, nil
}

func expandTemplate(endpoint, ip, lang string) string {
	replacer := strings.NewReplacer("{ip}", ip, "{lang}", lang)
	return replacer.Replace(endpoint)
}

func applyAuth(req *http.Request, authType, accessKey, secretKey string) {
	switch authType {
	case geo.AuthBasic:
		req.SetBasicAuth(accessKey, secretKey)
	case geo.AuthBearer:
		req.Header.Set("Authorization", "Bearer "+secretKey)
	case geo.AuthHeader:
		if accessKey != "" {
			req.Header.Set(accessKey, secretKey)
		}
	}
}

func applyFieldMap(rec *geo.Record, body []byte, fieldMap map[string]string, lang string) {
	for field, path := range fieldMap {
		if path == "" {
			continue
		}
		path = expandTemplate(path, rec.IP, lang)
		setField(rec, field, gjson.GetBytes(body, path), lang)
	}
}

func setField(rec *geo.Record, field string, val gjson.Result, lang string) {
	if !val.Exists() {
		return
	}
	switch field {
	case "country_code":
		rec.CountryCode = val.String()
	case "continent_code":
		rec.ContinentCode = val.String()
	case "province_iso_code":
		rec.ProvinceIsoCode = val.String()
	case "postal", "zip":
		rec.PostalCode = val.String()
	case "timezone":
		rec.Timezone = val.String()
	case "isp":
		rec.Isp = val.String()
	case "org":
		rec.Org = val.String()
	case "as":
		rec.ASN = val.String()
	case "lat":
		rec.Latitude = toFloat(val)
	case "lon":
		rec.Longitude = toFloat(val)
	default:
		setNameField(rec, field, val, lang)
	}
}

func setNameField(rec *geo.Record, field string, val gjson.Result, lang string) {
	names := jsonToNames(val, lang)
	switch field {
	case "country", "country_names":
		rec.CountryNames = mergeNames(rec.CountryNames, names)
	case "continent", "continent_names":
		rec.ContinentNames = mergeNames(rec.ContinentNames, names)
	case "province", "province_names":
		rec.ProvinceNames = mergeNames(rec.ProvinceNames, names)
	case "city", "city_names":
		rec.CityNames = mergeNames(rec.CityNames, names)
	}
}

func jsonToNames(val gjson.Result, lang string) geo.LangNames {
	if val.IsObject() {
		out := geo.LangNames{}
		val.ForEach(func(k, v gjson.Result) bool {
			out[k.String()] = v.String()
			return true
		})
		return out
	}
	if s := val.String(); s != "" {
		return geo.LangNames{lang: s}
	}
	return nil
}

func mergeNames(dst, src geo.LangNames) geo.LangNames {
	if len(src) == 0 {
		return dst
	}
	if dst == nil {
		return src
	}
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func toFloat(val gjson.Result) float64 {
	if val.Type == gjson.Number {
		return val.Float()
	}
	f, _ := strconv.ParseFloat(val.String(), 64)
	return f
}
