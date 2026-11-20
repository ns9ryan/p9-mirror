package ipgeo

import (
	"context"
	"strings"

	"oa.98ent.com/p9/integration/rpc/ent"
	"oa.98ent.com/p9/integration/rpc/ent/localip"
	"oa.98ent.com/p9/integration/rpc/internal/ipgeo/geo"

	"github.com/zeromicro/go-zero/core/logx"
)

// --- 本地 IP 表 ip_geo_local（供 driver=local 与后台维护）---

// CreateLocalIP 写入本地 IP，并失效该 IP 的查询缓存。
func (s *Service) CreateLocalIP(ctx context.Context, in geo.LocalIP) (*geo.LocalIP, error) {
	if s == nil || s.Client == nil {
		return nil, geo.ErrMissingDSN
	}
	in, err := geo.NormalizeLocalIP(in)
	if err != nil {
		return nil, err
	}
	row, err := applyLocalCreate(s.Client.LocalIP.Create(), in).Save(ctx)
	if ent.IsConstraintError(err) {
		return nil, geo.ErrIPTaken
	}
	if err != nil {
		return nil, err
	}
	out := fromEntLocal(row)
	s.invalidateIP(ctx, out.IP)
	return &out, nil
}

// UpdateLocalIP 更新本地 IP，并失效新旧 IP 缓存。
func (s *Service) UpdateLocalIP(ctx context.Context, in geo.LocalIP) (*geo.LocalIP, error) {
	if s == nil || s.Client == nil {
		return nil, geo.ErrMissingDSN
	}
	old, err := s.GetLocalIP(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	in, err = geo.NormalizeLocalIP(in)
	if err != nil {
		return nil, err
	}
	row, err := applyLocalUpdate(s.Client.LocalIP.UpdateOneID(in.ID), in).Save(ctx)
	if ent.IsNotFound(err) {
		return nil, geo.ErrLocalIPNotFound
	}
	if ent.IsConstraintError(err) {
		return nil, geo.ErrIPTaken
	}
	if err != nil {
		return nil, err
	}
	out := fromEntLocal(row)
	s.invalidateIP(ctx, old.IP)
	if out.IP != old.IP {
		s.invalidateIP(ctx, out.IP)
	}
	return &out, nil
}

// DeleteLocalIP 删除本地 IP，并失效缓存。
func (s *Service) DeleteLocalIP(ctx context.Context, id int64) error {
	if s == nil || s.Client == nil {
		return geo.ErrMissingDSN
	}
	old, err := s.GetLocalIP(ctx, id)
	if err != nil {
		return err
	}
	if err := s.Client.LocalIP.DeleteOneID(id).Exec(ctx); err != nil {
		if ent.IsNotFound(err) {
			return geo.ErrLocalIPNotFound
		}
		return err
	}
	s.invalidateIP(ctx, old.IP)
	return nil
}

// GetLocalIP 按主键查本地 IP。
func (s *Service) GetLocalIP(ctx context.Context, id int64) (*geo.LocalIP, error) {
	if s == nil || s.Client == nil {
		return nil, geo.ErrMissingDSN
	}
	row, err := s.Client.LocalIP.Get(ctx, id)
	if ent.IsNotFound(err) {
		return nil, geo.ErrLocalIPNotFound
	}
	if err != nil {
		return nil, err
	}
	out := fromEntLocal(row)
	return &out, nil
}

// ListLocalIPs 分页列表；ip 非空时做包含匹配。
func (s *Service) ListLocalIPs(ctx context.Context, page, pageSize int, ip string) (*geo.LocalIPPage, error) {
	if s == nil || s.Client == nil {
		return nil, geo.ErrMissingDSN
	}
	page, pageSize = geo.NormalizePage(page, pageSize)
	q := s.Client.LocalIP.Query()
	if ip = strings.TrimSpace(ip); ip != "" {
		q = q.Where(localip.IPContains(ip))
	}
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := q.Order(ent.Asc(localip.FieldIP)).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, err
	}
	list := make([]geo.LocalIP, 0, len(rows))
	for _, row := range rows {
		list = append(list, fromEntLocal(row))
	}
	return &geo.LocalIPPage{List: list, Total: total}, nil
}

func (s *Service) invalidateIP(ctx context.Context, ip string) {
	if s == nil || s.Cache == nil || ip == "" {
		return
	}
	if err := s.Cache.DeleteByIP(ctx, ip); err != nil {
		logx.WithContext(ctx).Errorf("ipgeo cache delete %s: %v", ip, err)
	}
}

func applyLocalCreate(b *ent.LocalIPCreate, in geo.LocalIP) *ent.LocalIPCreate {
	return b.SetIP(in.IP).
		SetContinent(in.Continent).
		SetContinentCode(in.ContinentCode).
		SetCountry(in.Country).
		SetCountryCode(in.CountryCode).
		SetProvince(in.Province).
		SetProvinceIsoCode(in.ProvinceIsoCode).
		SetCity(in.City).
		SetPostal(in.Postal).
		SetLat(in.Lat).
		SetLon(in.Lon).
		SetTimezone(in.Timezone).
		SetIsp(in.Isp).
		SetOrg(in.Org).
		SetAsn(in.ASN)
}

func applyLocalUpdate(b *ent.LocalIPUpdateOne, in geo.LocalIP) *ent.LocalIPUpdateOne {
	return b.SetIP(in.IP).
		SetContinent(in.Continent).
		SetContinentCode(in.ContinentCode).
		SetCountry(in.Country).
		SetCountryCode(in.CountryCode).
		SetProvince(in.Province).
		SetProvinceIsoCode(in.ProvinceIsoCode).
		SetCity(in.City).
		SetPostal(in.Postal).
		SetLat(in.Lat).
		SetLon(in.Lon).
		SetTimezone(in.Timezone).
		SetIsp(in.Isp).
		SetOrg(in.Org).
		SetAsn(in.ASN)
}

func fromEntLocal(row *ent.LocalIP) geo.LocalIP {
	return geo.LocalIP{
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
}
