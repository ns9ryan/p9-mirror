package source

import (
	"context"

	"oa.98ent.com/p9/integration/rpc/ent"
	"oa.98ent.com/p9/integration/rpc/ent/localip"
	"oa.98ent.com/p9/integration/rpc/internal/ipgeo/geo"
)

// LocalLookup 按 IP 查本地表；未命中返回 (nil, nil)。
type LocalLookup func(ctx context.Context, ip string) (*geo.LocalIP, error)

// LocalSource 本地表源：查 LocalLookup，未命中返回空。
type LocalSource struct {
	name   string      // 源名称，写入 Record.Source
	lookup LocalLookup // 本地 IP 查询
}

// NewLocalSource 绑定本地 IP 查询函数。
func NewLocalSource(name string, lookup LocalLookup) *LocalSource {
	return &LocalSource{name: name, lookup: lookup}
}

// LookupByClient 用 ent 查 ip_geo_local。
func LookupByClient(db *ent.Client) LocalLookup {
	if db == nil {
		return nil
	}
	return func(ctx context.Context, ip string) (*geo.LocalIP, error) {
		row, err := db.LocalIP.Query().Where(localip.IPEQ(ip)).Only(ctx)
		if ent.IsNotFound(err) {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		out := geo.LocalIP{
			ID:              row.ID,
			IP:              row.IP,
			Continent:       row.Continent,
			ContinentCode:   row.ContinentCode,
			Country:         row.Country,
			CountryCode:     row.CountryCode,
			Province:        row.Province,
			ProvinceIsoCode: row.ProvinceIsoCode,
			City:            row.City,
			Postal:          row.Postal,
			Lat:             row.Lat,
			Lon:             row.Lon,
			Timezone:        row.Timezone,
			Isp:             row.Isp,
			Org:             row.Org,
			ASN:             row.Asn,
		}
		return &out, nil
	}
}

func (s *LocalSource) Close() error { return nil }

func (s *LocalSource) Fetch(ctx context.Context, ip, _ string) (*geo.Record, error) {
	if s == nil || s.lookup == nil {
		return nil, nil
	}
	row, err := s.lookup(ctx, ip)
	if err != nil || row == nil {
		return nil, err
	}
	return row.ToRecord(s.name), nil
}
