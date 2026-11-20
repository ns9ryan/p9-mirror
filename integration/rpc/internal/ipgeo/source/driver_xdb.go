package source

import (
	"context"
	"strings"
	"time"

	"oa.98ent.com/p9/integration/rpc/internal/ipgeo/geo"

	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
)

// XDB ip2region xdb 离线源。
type XDB struct {
	name     string        // 源名称，写入 Record.Source
	searcher *xdb.Searcher // 已打开的 xdb
}

// OpenXDB 打开 xdb 文件。
func OpenXDB(name, path string) (*XDB, error) {
	header, err := xdb.LoadHeaderFromFile(path)
	if err != nil {
		return nil, err
	}
	ver, err := xdb.VersionFromHeader(header)
	if err != nil {
		return nil, err
	}
	searcher, err := xdb.NewWithFileOnly(ver, path)
	if err != nil {
		return nil, err
	}
	return &XDB{name: name, searcher: searcher}, nil
}

func (x *XDB) Close() error {
	if x == nil || x.searcher == nil {
		return nil
	}
	x.searcher.Close()
	return nil
}

func (x *XDB) Fetch(_ context.Context, ip, lang string) (*geo.Record, error) {
	raw, err := x.searcher.Search(ip)
	if err != nil {
		return nil, err
	}
	return parseXDB(ip, raw, x.name, lang), nil
}

func parseXDB(ip, raw, source, lang string) *geo.Record {
	parts := strings.Split(raw, "|")
	if len(parts) < 2 {
		parts = strings.Split(raw, ",")
	}
	rec := &geo.Record{IP: ip, Source: source, UpdatedAt: time.Now()}
	if lang == "" {
		lang = geo.DefaultLang
	}
	setXDBPart(rec, parts, 0, func(s string) { rec.CountryNames = geo.LangNames{lang: s} })
	setXDBPart(rec, parts, 2, func(s string) { rec.ProvinceNames = geo.LangNames{lang: s} })
	setXDBPart(rec, parts, 3, func(s string) { rec.CityNames = geo.LangNames{lang: s} })
	setXDBPart(rec, parts, 4, func(s string) { rec.Isp = s })
	if len(parts) > 1 && looksISO(parts[1]) {
		rec.CountryCode = parts[1]
	}
	return rec
}

func setXDBPart(rec *geo.Record, parts []string, idx int, set func(string)) {
	if rec == nil || idx >= len(parts) {
		return
	}
	s := strings.TrimSpace(parts[idx])
	if s == "" || s == "0" {
		return
	}
	set(s)
}

func looksISO(s string) bool {
	s = strings.TrimSpace(s)
	return len(s) == 2
}
