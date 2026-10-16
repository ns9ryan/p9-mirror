package source

import (
	"context"
	"net"
	"time"

	"oa.98ent.com/p9/integration/rpc/internal/ipgeo/geo"

	"github.com/oschwald/geoip2-golang"
)

// MMDB MaxMind MMDB 离线源。
type MMDB struct {
	name   string         // 源名称，写入 Record.Source
	reader *geoip2.Reader // 已打开的 MMDB
}

// OpenNamedMMDB 打开 MMDB 文件。
func OpenNamedMMDB(name, path string) (*MMDB, error) {
	reader, err := geoip2.Open(path)
	if err != nil {
		return nil, err
	}
	return &MMDB{name: name, reader: reader}, nil
}

func (m *MMDB) Fetch(_ context.Context, ip, _ string) (*geo.Record, error) {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return nil, nil
	}
	city, err := m.reader.City(parsed)
	if err != nil {
		return nil, err
	}
	rec := recordFromCity(ip, city)
	rec.Source = m.name
	return rec, nil
}

func (m *MMDB) Close() error {
	if m == nil || m.reader == nil {
		return nil
	}
	return m.reader.Close()
}

func recordFromCity(ip string, city *geoip2.City) *geo.Record {
	rec := &geo.Record{
		IP:             ip,
		ContinentCode:  city.Continent.Code,
		ContinentNames: geo.LangNames(city.Continent.Names),
		CountryCode:    city.Country.IsoCode,
		CountryNames:   geo.LangNames(city.Country.Names),
		CityNames:      geo.LangNames(city.City.Names),
		PostalCode:     city.Postal.Code,
		Latitude:       city.Location.Latitude,
		Longitude:      city.Location.Longitude,
		Timezone:       city.Location.TimeZone,
		Source:         geo.DriverMMDB,
		UpdatedAt:      time.Now(),
	}
	if len(city.Subdivisions) > 0 {
		sub := city.Subdivisions[0]
		rec.ProvinceIsoCode = sub.IsoCode
		rec.ProvinceNames = geo.LangNames(sub.Names)
	}
	return rec
}
