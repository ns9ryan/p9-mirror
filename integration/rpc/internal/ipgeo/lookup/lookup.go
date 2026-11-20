package lookup

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"oa.98ent.com/p9/integration/rpc/ent"
	"oa.98ent.com/p9/integration/rpc/internal/ipgeo/cache"
	"oa.98ent.com/p9/integration/rpc/internal/ipgeo/geo"
	"oa.98ent.com/p9/integration/rpc/internal/ipgeo/source"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/singleflight"
)

const (
	maxIPs      = 100
	maxParallel = 8
)

// ListEnabledFunc 返回某租户当前启用的定位源（已按优先级排序）。
// 生产由 service/ipgeo.Service.ListEnabledProviders 注入。
type ListEnabledFunc func(ctx context.Context, operatorCode string) ([]geo.Provider, error)

// Conf Lookup 运行配置（由 svc 注入 Client / Cache / 选源函数）。
type Conf struct {
	Client      *ent.Client     // 必填；local 驱动查 ip_geo_local
	Cache       cache.Cache     // 必填；查询结果缓存（svc 打开后注入）
	CacheTTL    time.Duration   // 写入缓存的 TTL；<=0 表示不过期
	ListEnabled ListEnabledFunc // 必填；通常为 service/ipgeo.ListEnabledProviders
}

// Lookup IP/GEO 查询编排器：只负责 Search，不写管理 CRUD。
// 管理端 LocalIP/策略走 rpc/service.Deps。
type Lookup struct {
	listEnabled ListEnabledFunc                        // 按租户取启用源
	cache       cache.Cache                            // 按源缓存查询结果
	ttl         time.Duration                          // 写入缓存时的 TTL
	sf          singleflight.Group                     // 同键回源合并
	build       func(geo.Provider) (geo.Source, error) // 构造 Source；可测注入

	mu   sync.Mutex
	byOp map[string]*opSources // 租户 -> 已打开的有序源列表
}

// opSource 租户下一条已打开的查询源。
type opSource struct {
	id   int64      // Provider 主键
	name string     // Provider.Name，作缓存键前缀
	src  geo.Source // 运行时 Source
}

// opSources 某租户当前启用源快照。
type opSources struct {
	fp  string     // 配置指纹，变则重建 src
	src []opSource // 按优先级排好的源
}

// New 构造 Lookup；不关闭传入的 Client / Cache（生命周期归 svc）。
func New(cfg Conf) (*Lookup, error) {
	if cfg.Client == nil {
		return nil, geo.ErrMissingDSN
	}
	if cfg.ListEnabled == nil {
		return nil, geo.ErrBadProvider
	}
	if cfg.Cache == nil {
		return nil, geo.ErrBadCacheDriver
	}
	lookupLocal := source.LookupByClient(cfg.Client)
	return &Lookup{
		ttl:         cfg.CacheTTL,
		cache:       cfg.Cache,
		listEnabled: cfg.ListEnabled,
		build: func(p geo.Provider) (geo.Source, error) {
			return source.Build(p, lookupLocal)
		},
		byOp: map[string]*opSources{},
	}, nil
}

// Close 关闭已打开的查询源（不关 Client/Cache）。
func (l *Lookup) Close() error {
	if l == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, cur := range l.byOp {
		closeSources(cur.src)
	}
	l.byOp = nil
	return nil
}

// SearchList 并行批量查询，最多 100 个 IP。
func (l *Lookup) SearchList(ctx context.Context, ips []string, lang, operatorCode string) ([]*geo.Info, error) {
	if len(ips) > maxIPs {
		return nil, geo.ErrTooManyIPs
	}
	if lang == "" {
		lang = geo.DefaultLang
	}
	if !geo.SupportedLang(lang) {
		return nil, geo.ErrBadLang
	}
	operatorCode = geo.NormalizeOperatorCode(operatorCode)
	out := make([]*geo.Info, len(ips))
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(maxParallel)
	for i, ip := range ips {
		i, ip := i, ip
		g.Go(func() error {
			out[i] = l.Search(ctx, ip, lang, operatorCode)
			return nil
		})
	}
	_ = g.Wait()
	return out, nil
}

// Search 单 IP 查询：私网直返，否则按租户策略选源（缓存优先，miss 再 Fetch）。
func (l *Lookup) Search(ctx context.Context, ip, lang, operatorCode string) *geo.Info {
	operatorCode = geo.NormalizeOperatorCode(operatorCode)
	normalized, err := geo.NormalizeIP(ip)
	if err != nil {
		return &geo.Info{Query: ip}
	}
	if geo.IsPrivate(net.ParseIP(normalized)) {
		return geo.PrivateInfo(ip)
	}
	v, err, _ := l.sf.Do(operatorCode+"|"+normalized, func() (any, error) {
		return l.resolve(ctx, normalized, lang, operatorCode)
	})
	if err != nil || v == nil {
		return &geo.Info{Query: ip}
	}
	rec, _ := v.(*geo.Record)
	return rec.ToInfo(lang, ip)
}

func (l *Lookup) resolve(ctx context.Context, ip, lang, operatorCode string) (*geo.Record, error) {
	list, err := l.listEnabled(ctx, operatorCode)
	if err != nil {
		logx.WithContext(ctx).Errorf("ipgeo list providers: %v", err)
		return nil, nil
	}
	for _, item := range l.sourcesFor(operatorCode, list) {
		if rec := l.cacheGet(ctx, cache.Key(item.name, ip)); rec != nil {
			return rec, nil
		}
		rec, err := item.src.Fetch(ctx, ip, lang)
		if err != nil {
			logx.WithContext(ctx).Errorf("ipgeo source: %v", err)
			continue
		}
		if rec == nil {
			continue
		}
		rec.Source = item.name
		rec.OperatorCode = operatorCode
		l.cacheSet(ctx, rec)
		return rec, nil
	}
	return nil, nil
}

func (l *Lookup) sourcesFor(operatorCode string, list []geo.Provider) []opSource {
	fp := providerFP(list)
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.byOp == nil {
		l.byOp = map[string]*opSources{}
	}
	cur := l.byOp[operatorCode]
	if cur != nil && cur.fp == fp {
		return cur.src
	}
	built := l.buildList(list)
	if cur != nil {
		closeSources(cur.src)
	}
	l.byOp[operatorCode] = &opSources{fp: fp, src: built}
	return built
}

func (l *Lookup) buildList(list []geo.Provider) []opSource {
	built := make([]opSource, 0, len(list))
	for _, p := range list {
		src, err := l.build(p)
		if err != nil {
			logx.Errorf("ipgeo build provider %s: %v", p.Name, err)
			continue
		}
		built = append(built, opSource{id: p.ID, name: p.Name, src: src})
	}
	return built
}

func providerFP(list []geo.Provider) string {
	fp := ""
	for _, p := range list {
		fp += fmt.Sprintf("%d:%s:%s:%d:%s:%v;", p.ID, p.OperatorCode, p.Driver, p.Priority, p.Endpoint, p.IsEnabled)
	}
	return fp
}

func closeSources(list []opSource) {
	for _, item := range list {
		if item.src != nil {
			_ = item.src.Close()
		}
	}
}

func (l *Lookup) cacheGet(ctx context.Context, key string) *geo.Record {
	if l.cache == nil {
		return nil
	}
	rec, err := l.cache.Get(ctx, key)
	if err != nil {
		logx.WithContext(ctx).Errorf("ipgeo cache get %s: %v", key, err)
		return nil
	}
	if rec == nil || l.cacheExpired(rec) {
		return nil
	}
	l.cacheTouch(ctx, key)
	return rec
}

func (l *Lookup) cacheExpired(rec *geo.Record) bool {
	if l.ttl <= 0 {
		return false
	}
	return rec.ExpiresAt.IsZero() || !time.Now().Before(rec.ExpiresAt)
}

func (l *Lookup) cacheTouch(ctx context.Context, key string) {
	if l.cache == nil || l.ttl <= 0 {
		return
	}
	if err := l.cache.Touch(ctx, key, time.Now().Add(l.ttl)); err != nil {
		logx.WithContext(ctx).Errorf("ipgeo cache touch %s: %v", key, err)
	}
}

func (l *Lookup) cacheSet(ctx context.Context, rec *geo.Record) {
	if l.cache == nil || rec == nil {
		return
	}
	if l.ttl > 0 {
		rec.ExpiresAt = time.Now().Add(l.ttl)
	} else {
		rec.ExpiresAt = time.Time{}
	}
	if err := l.cache.Set(ctx, rec); err != nil {
		logx.WithContext(ctx).Errorf("ipgeo cache set %s: %v", rec.IP, err)
	}
}
