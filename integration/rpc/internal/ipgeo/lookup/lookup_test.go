package lookup

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"oa.98ent.com/p9/integration/rpc/internal/ipgeo/cache"
	"oa.98ent.com/p9/integration/rpc/internal/ipgeo/geo"
	"oa.98ent.com/p9/integration/rpc/internal/ipgeo/source"
)

type stubSource struct {
	rec *geo.Record
	err error
	n   int
}

func (s *stubSource) Fetch(_ context.Context, _ string, _ string) (*geo.Record, error) {
	s.n++
	if s.rec != nil {
		cp := *s.rec
		return &cp, s.err
	}
	return nil, s.err
}

func (s *stubSource) Close() error { return nil }

func sampleRecord(ip, source, country string) *geo.Record {
	return &geo.Record{
		IP:           ip,
		CountryCode:  country,
		CountryNames: geo.LangNames{"en": "United States", "zh-CN": "美国"},
		Source:       source,
		UpdatedAt:    time.Now(),
	}
}

func fixedEnabled(list []geo.Provider) ListEnabledFunc {
	return func(_ context.Context, _ string) ([]geo.Provider, error) {
		out := make([]geo.Provider, 0, len(list))
		for _, p := range list {
			if p.IsEnabled {
				out = append(out, p)
			}
		}
		return out, nil
	}
}

func testLookup(list []geo.Provider, build func(geo.Provider) (geo.Source, error)) *Lookup {
	return &Lookup{
		listEnabled: fixedEnabled(list),
		cache:       cache.NewMemoryCache(),
		build:       build,
		byOp:        map[string]*opSources{},
	}
}

// 私有 IP
func TestLookup_PrivateIP(t *testing.T) {
	l := testLookup(nil, source.BuildSource)
	got := l.Search(context.Background(), "192.168.1.1", "zh-CN", "")
	if got.Country != "私有IP" || got.Source != geo.SourcePrivate {
		t.Fatalf("got %+v", got)
	}
}

// 缓存命中
func TestLookup_CacheHit(t *testing.T) {
	src := &stubSource{rec: sampleRecord("8.8.8.8", "maxmind", "US")}
	cch := cache.NewMemoryCache()
	_ = cch.Set(context.Background(), sampleRecord("8.8.8.8", "maxmind", "US"))
	l := &Lookup{
		listEnabled: fixedEnabled([]geo.Provider{{Name: "maxmind", Driver: geo.DriverHTTP, IsEnabled: true, Endpoint: "http://x/{ip}"}}),
		cache:       cch,
		build:       func(geo.Provider) (geo.Source, error) { return src, nil },
		byOp:        map[string]*opSources{},
	}
	got := l.Search(context.Background(), "8.8.8.8", "zh-CN", "")
	if got.Country != "美国" || src.n != 0 {
		t.Fatalf("got %+v n=%d", got, src.n)
	}
}

// 缓存过期会回源
func TestLookup_CacheExpiredFetchesSource(t *testing.T) {
	src := &stubSource{rec: sampleRecord("8.8.8.8", "maxmind", "US")}
	cch := cache.NewMemoryCache()
	stale := sampleRecord("8.8.8.8", "maxmind", "US")
	stale.ExpiresAt = time.Now().Add(-time.Minute)
	_ = cch.Set(context.Background(), stale)
	l := &Lookup{
		ttl:         time.Hour,
		listEnabled: fixedEnabled([]geo.Provider{{Name: "maxmind", Driver: geo.DriverHTTP, IsEnabled: true, Endpoint: "http://x/{ip}"}}),
		cache:       cch,
		build:       func(geo.Provider) (geo.Source, error) { return src, nil },
		byOp:        map[string]*opSources{},
	}
	got := l.Search(context.Background(), "8.8.8.8", "en", "")
	if got.Source != "maxmind" || src.n != 1 {
		t.Fatalf("got %+v n=%d", got, src.n)
	}
	hit, _ := cch.Get(context.Background(), cache.Key("maxmind", "8.8.8.8"))
	if hit == nil || !hit.ExpiresAt.After(time.Now().Add(50*time.Minute)) {
		t.Fatalf("expires_at not set: %+v", hit)
	}
}

// 缓存命中续期
func TestLookup_CacheHitRenewsTTL(t *testing.T) {
	src := &stubSource{rec: sampleRecord("8.8.8.8", "maxmind", "US")}
	cch := cache.NewMemoryCache()
	rec := sampleRecord("8.8.8.8", "maxmind", "US")
	rec.ExpiresAt = time.Now().Add(time.Minute)
	_ = cch.Set(context.Background(), rec)
	l := &Lookup{
		ttl:         time.Hour,
		listEnabled: fixedEnabled([]geo.Provider{{Name: "maxmind", Driver: geo.DriverHTTP, IsEnabled: true, Endpoint: "http://x/{ip}"}}),
		cache:       cch,
		build:       func(geo.Provider) (geo.Source, error) { return src, nil },
		byOp:        map[string]*opSources{},
	}
	got := l.Search(context.Background(), "8.8.8.8", "zh-CN", "")
	if got.Country != "美国" || src.n != 0 {
		t.Fatalf("got %+v n=%d", got, src.n)
	}
	hit, _ := cch.Get(context.Background(), cache.Key("maxmind", "8.8.8.8"))
	if hit == nil || !hit.ExpiresAt.After(time.Now().Add(50*time.Minute)) {
		t.Fatalf("not renewed: %+v", hit)
	}
}

// 按优先级回退
func TestLookup_FallbackByPriority(t *testing.T) {
	primary := &stubSource{err: errors.New("timeout")}
	fallback := &stubSource{rec: sampleRecord("8.8.8.8", "mmdb", "US")}
	l := testLookup([]geo.Provider{
		{ID: 1, Name: "http-a", Driver: geo.DriverHTTP, IsEnabled: true, Priority: 1, Endpoint: "http://a/{ip}"},
		{ID: 2, Name: "mmdb", Driver: geo.DriverMMDB, IsEnabled: true, Priority: 2, Endpoint: "x.mmdb"},
	}, func(p geo.Provider) (geo.Source, error) {
		if p.Name == "http-a" {
			return primary, nil
		}
		return fallback, nil
	})
	got := l.Search(context.Background(), "8.8.8.8", "en", "")
	if got.Source != "mmdb" || primary.n != 1 || fallback.n != 1 {
		t.Fatalf("got %+v p=%d f=%d", got, primary.n, fallback.n)
	}
}

// 禁用的源不参与
func TestLookup_DisabledSourceSkipped(t *testing.T) {
	src := &stubSource{rec: sampleRecord("8.8.8.8", "live", "US")}
	dead := &stubSource{rec: sampleRecord("8.8.8.8", "dead", "US")}
	l := testLookup([]geo.Provider{
		{ID: 1, Name: "dead", Driver: geo.DriverHTTP, IsEnabled: false, Priority: 1, Endpoint: "http://d/{ip}"},
		{ID: 2, Name: "live", Driver: geo.DriverHTTP, IsEnabled: true, Priority: 2, Endpoint: "http://l/{ip}"},
	}, func(p geo.Provider) (geo.Source, error) {
		if p.Name == "dead" {
			return dead, nil
		}
		return src, nil
	})
	got := l.Search(context.Background(), "8.8.8.8", "en", "")
	if got.Source != "live" || dead.n != 0 {
		t.Fatalf("got %+v dead.n=%d", got, dead.n)
	}
}

// 批量与校验
func TestLookup_BatchAndValidation(t *testing.T) {
	src := &stubSource{rec: sampleRecord("8.8.8.8", "http", "US")}
	l := testLookup(
		[]geo.Provider{{Name: "http", Driver: geo.DriverHTTP, IsEnabled: true, Endpoint: "http://x/{ip}"}},
		func(geo.Provider) (geo.Source, error) { return src, nil },
	)
	list, err := l.SearchList(context.Background(), []string{"not-an-ip", "8.8.8.8"}, "en", "")
	if err != nil || list[1].Country != "United States" {
		t.Fatalf("err=%v list=%v", err, list)
	}
	if _, err := l.SearchList(context.Background(), make([]string, 101), "en", ""); !errors.Is(err, geo.ErrTooManyIPs) {
		t.Fatalf("want too many, %v", err)
	}
}

// 按字段映射解析
func TestHTTPSource_FieldMapParse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, pass, ok := r.BasicAuth(); !ok || pass != "key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		io.WriteString(w, `{"country":{"iso_code":"US","names":{"en":"United States"}},"lat":1.5}`)
	}))
	defer srv.Close()
	src := source.NewHTTPSource(geo.Provider{
		Name: "geo", Driver: geo.DriverHTTP, Endpoint: srv.URL + "/{ip}",
		AuthType: geo.AuthBasic, AccessKey: "id", SecretKey: "key", TimeoutSec: 2,
		FieldMap: map[string]string{
			"country_code":  "country.iso_code",
			"country_names": "country.names",
			"lat":           "lat",
		},
	})
	rec, err := src.Fetch(context.Background(), "8.8.8.8", "en")
	if err != nil || rec.CountryCode != "US" || rec.CountryNames["en"] != "United States" || rec.Latitude != 1.5 {
		t.Fatalf("rec=%+v err=%v", rec, err)
	}
}

// 缺少 Client
func TestNew_MissingDSN(t *testing.T) {
	_, err := New(Conf{})
	if !errors.Is(err, geo.ErrMissingDSN) {
		t.Fatalf("got %v", err)
	}
}

func TestValidateProvider(t *testing.T) {
	if err := geo.ValidateProvider(geo.Provider{Name: "a", Driver: "foo", Endpoint: "x"}); !errors.Is(err, geo.ErrBadDriver) {
		t.Fatalf("got %v", err)
	}
	if err := geo.ValidateProvider(geo.Provider{Name: "a", Driver: geo.DriverHTTP, Endpoint: ""}); !errors.Is(err, geo.ErrEmptyEndpoint) {
		t.Fatalf("got %v", err)
	}
	if err := geo.ValidateProvider(geo.Provider{Name: "local", Driver: geo.DriverLocal}); err != nil {
		t.Fatalf("local: %v", err)
	}
	if err := geo.ValidateProvider(geo.Provider{Name: "a", Driver: geo.DriverHTTP, Endpoint: "x", Tag: "foo"}); !errors.Is(err, geo.ErrBadTag) {
		t.Fatalf("tag: %v", err)
	}
	got := geo.NormalizeProvider(geo.Provider{Name: "a", Driver: geo.DriverHTTP, Endpoint: "http://x/{ip}"})
	if got.Tag != geo.TagOnline {
		t.Fatalf("http tag=%s", got.Tag)
	}
	if geo.TagFromDriver(geo.DriverLocal) != geo.TagLocal || geo.TagFromDriver(geo.DriverMMDB) != geo.TagOffline {
		t.Fatalf("driver tag")
	}
}

// 租户无策略用总配置
func TestLookup_NoPolicyUsesDefault(t *testing.T) {
	def := &stubSource{rec: sampleRecord("8.8.8.8", "geoip-city", "US")}
	l := testLookup([]geo.Provider{{
		Name: "geoip-city", Driver: geo.DriverMMDB, IsEnabled: true, Endpoint: "x.mmdb",
	}}, func(geo.Provider) (geo.Source, error) { return def, nil })
	got := l.Search(context.Background(), "8.8.8.8", "en", "op-a")
	if got.Source != "geoip-city" || def.n != 1 {
		t.Fatalf("got %+v n=%d", got, def.n)
	}
}

// 租户策略禁用主源
func TestLookup_TenantDisablesPrimary(t *testing.T) {
	primary := &stubSource{rec: sampleRecord("8.8.8.8", "primary", "US")}
	fallback := &stubSource{rec: sampleRecord("8.8.8.8", "fallback", "US")}
	l := &Lookup{
		listEnabled: func(_ context.Context, op string) ([]geo.Provider, error) {
			all := []geo.Provider{
				{ID: 1, Name: "primary", Driver: geo.DriverHTTP, IsEnabled: true, Priority: 1, Endpoint: "http://a/{ip}"},
				{ID: 2, Name: "fallback", Driver: geo.DriverHTTP, IsEnabled: true, Priority: 2, Endpoint: "http://b/{ip}"},
			}
			if op == "op-a" {
				all[0].IsEnabled = false
			}
			return fixedEnabled(all)(context.Background(), op)
		},
		cache: cache.NewMemoryCache(),
		build: func(p geo.Provider) (geo.Source, error) {
			if p.Name == "primary" {
				return primary, nil
			}
			return fallback, nil
		},
		byOp: map[string]*opSources{},
	}
	gotA := l.Search(context.Background(), "8.8.8.8", "en", "op-a")
	if gotA.Source != "fallback" || primary.n != 0 || fallback.n != 1 {
		t.Fatalf("tenant a got %+v primary=%d fallback=%d", gotA, primary.n, fallback.n)
	}
	gotB := l.Search(context.Background(), "8.8.8.8", "en", "op-b")
	if gotB.Source != "primary" || primary.n != 1 {
		t.Fatalf("tenant b got %+v primary=%d", gotB, primary.n)
	}
}

// 停用源不读该源缓存
func TestLookup_DisabledSourceIgnoresCache(t *testing.T) {
	primary := &stubSource{rec: sampleRecord("8.8.8.8", "primary", "US")}
	fallback := &stubSource{rec: sampleRecord("8.8.8.8", "fallback", "CN")}
	cch := cache.NewMemoryCache()
	_ = cch.Set(context.Background(), sampleRecord("8.8.8.8", "primary", "US"))
	l := &Lookup{
		listEnabled: func(_ context.Context, _ string) ([]geo.Provider, error) {
			return []geo.Provider{
				{ID: 2, Name: "fallback", Driver: geo.DriverHTTP, IsEnabled: true, Priority: 2, Endpoint: "http://b/{ip}"},
			}, nil
		},
		cache: cch,
		build: func(p geo.Provider) (geo.Source, error) {
			if p.Name == "primary" {
				return primary, nil
			}
			return fallback, nil
		},
		byOp: map[string]*opSources{},
	}
	got := l.Search(context.Background(), "8.8.8.8", "en", "op-a")
	if got.Source != "fallback" || got.CountryCode != "CN" || primary.n != 0 || fallback.n != 1 {
		t.Fatalf("got %+v primary=%d fallback=%d", got, primary.n, fallback.n)
	}
}

// 本地表命中
func TestLookup_LocalTableHit(t *testing.T) {
	fallback := &stubSource{rec: sampleRecord("8.8.8.8", "http", "US")}
	lookupLocal := memLocalLookup(map[string]geo.LocalIP{
		"8.8.8.8": {IP: "8.8.8.8", Country: "中国", CountryCode: "CN", City: "深圳"},
	})
	l := &Lookup{
		listEnabled: fixedEnabled([]geo.Provider{
			{Name: "local", Driver: geo.DriverLocal, IsEnabled: true, Priority: 0},
			{Name: "http", Driver: geo.DriverHTTP, IsEnabled: true, Priority: 1, Endpoint: "http://x/{ip}"},
		}),
		cache: cache.NewMemoryCache(),
		build: func(p geo.Provider) (geo.Source, error) { return source.Build(p, lookupLocal) },
		byOp:  map[string]*opSources{},
	}
	got := l.Search(context.Background(), "8.8.8.8", "zh-CN", "")
	if got.Country != "中国" || got.City != "深圳" || got.Source != "local" || fallback.n != 0 {
		t.Fatalf("got %+v fallback=%d", got, fallback.n)
	}
}

// 本地表未命中回落
func TestLookup_LocalMissFallback(t *testing.T) {
	fallback := &stubSource{rec: sampleRecord("1.1.1.1", "http", "US")}
	l := &Lookup{
		listEnabled: fixedEnabled([]geo.Provider{
			{Name: "local", Driver: geo.DriverLocal, IsEnabled: true, Priority: 0},
			{Name: "http", Driver: geo.DriverHTTP, IsEnabled: true, Priority: 1, Endpoint: "http://x/{ip}"},
		}),
		cache: cache.NewMemoryCache(),
		build: func(p geo.Provider) (geo.Source, error) {
			if p.Driver == geo.DriverLocal {
				return source.NewLocalSource(p.Name, memLocalLookup(nil)), nil
			}
			return fallback, nil
		},
		byOp: map[string]*opSources{},
	}
	got := l.Search(context.Background(), "1.1.1.1", "en", "")
	if got.Source != "http" || fallback.n != 1 {
		t.Fatalf("got %+v n=%d", got, fallback.n)
	}
}

// 改本地表并清缓存后查询拿到新值
func TestLookup_UpdateLocalClearsCache(t *testing.T) {
	rows := map[string]geo.LocalIP{
		"8.8.8.8": {IP: "8.8.8.8", Country: "旧", CountryCode: "CN"},
	}
	lookupLocal := memLocalLookup(rows)
	cch := cache.NewMemoryCache()
	l := &Lookup{
		listEnabled: fixedEnabled([]geo.Provider{{Name: "local", Driver: geo.DriverLocal, IsEnabled: true, Priority: 0}}),
		cache:       cch,
		build:       func(p geo.Provider) (geo.Source, error) { return source.Build(p, lookupLocal) },
		byOp:        map[string]*opSources{},
	}
	ctx := context.Background()
	if got := l.Search(ctx, "8.8.8.8", "zh-CN", ""); got.Country != "旧" {
		t.Fatalf("first %+v", got)
	}
	rows["8.8.8.8"] = geo.LocalIP{IP: "8.8.8.8", Country: "新", CountryCode: "CN"}
	if err := cch.DeleteByIP(ctx, "8.8.8.8"); err != nil {
		t.Fatal(err)
	}
	if got := l.Search(ctx, "8.8.8.8", "zh-CN", ""); got.Country != "新" {
		t.Fatalf("after update %+v", got)
	}
}

func memLocalLookup(rows map[string]geo.LocalIP) source.LocalLookup {
	var mu sync.RWMutex
	return func(_ context.Context, ip string) (*geo.LocalIP, error) {
		mu.RLock()
		defer mu.RUnlock()
		if rows == nil {
			return nil, nil
		}
		row, ok := rows[ip]
		if !ok {
			return nil, nil
		}
		cp := row
		return &cp, nil
	}
}
