package ipgeo

import (
	"errors"

	geo "oa.98ent.com/p9/integration/rpc/internal/ipgeo"
	"oa.98ent.com/p9/integration/rpc/types/integration"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PB <-> 领域类型转换，以及领域错误到 gRPC status 的映射。

func toPBList(in []*geo.Info) []*integration.IpInfo {
	out := make([]*integration.IpInfo, 0, len(in))
	for _, item := range in {
		out = append(out, toPB(item))
	}
	return out
}

func toPB(in *geo.Info) *integration.IpInfo {
	if in == nil {
		return &integration.IpInfo{}
	}
	return &integration.IpInfo{
		Query:           in.Query,
		Continent:       in.Continent,
		ContinentCode:   in.ContinentCode,
		Country:         in.Country,
		CountryCode:     in.CountryCode,
		Province:        in.Province,
		ProvinceIsoCode: in.ProvinceIsoCode,
		City:            in.City,
		Zip:             in.Zip,
		Lat:             in.Lat,
		Lon:             in.Lon,
		Timezone:        in.Timezone,
		Isp:             in.Isp,
		Org:             in.Org,
		As:              in.As,
		Postal:          in.Postal,
		Source:          in.Source,
	}
}

func localToPB(in *geo.LocalIP) *integration.LocalIP {
	if in == nil {
		return &integration.LocalIP{}
	}
	return &integration.LocalIP{
		Id:              in.ID,
		Ip:              in.IP,
		Continent:       in.Continent,
		ContinentCode:   in.ContinentCode,
		Country:         in.Country,
		CountryCode:     in.CountryCode,
		Province:        in.Province,
		ProvinceIsoCode: in.ProvinceIsoCode,
		City:            in.City,
		Zip:             in.Postal,
		Lat:             in.Lat,
		Lon:             in.Lon,
		Timezone:        in.Timezone,
		Isp:             in.Isp,
		Org:             in.Org,
		As:              in.ASN,
		Postal:          in.Postal,
	}
}

func localFromPB(in *integration.LocalIP) geo.LocalIP {
	if in == nil {
		return geo.LocalIP{}
	}
	postal := in.GetPostal()
	if postal == "" {
		postal = in.GetZip()
	}
	return geo.LocalIP{
		ID:              in.GetId(),
		IP:              in.GetIp(),
		Continent:       in.GetContinent(),
		ContinentCode:   in.GetContinentCode(),
		Country:         in.GetCountry(),
		CountryCode:     in.GetCountryCode(),
		Province:        in.GetProvince(),
		ProvinceIsoCode: in.GetProvinceIsoCode(),
		City:            in.GetCity(),
		Postal:          postal,
		Lat:             in.GetLat(),
		Lon:             in.GetLon(),
		Timezone:        in.GetTimezone(),
		Isp:             in.GetIsp(),
		Org:             in.GetOrg(),
		ASN:             in.GetAs(),
	}
}

func policyItemsFromPB(in []*integration.PolicyItem) []geo.PolicyItem {
	out := make([]geo.PolicyItem, 0, len(in))
	for _, item := range in {
		if item == nil {
			continue
		}
		out = append(out, geo.PolicyItem{
			ProviderID: item.GetProviderId(),
			IsEnabled:  item.GetIsEnabled(),
			Priority:   int(item.GetPriority()),
		})
	}
	return out
}

func policiesToPB(list []geo.Provider) *integration.PolicyListResp {
	out := make([]*integration.PolicyView, 0, len(list))
	for i := range list {
		out = append(out, policyToPB(&list[i]))
	}
	return &integration.PolicyListResp{List: out, Total: int64(len(out))}
}

func policyToPB(p *geo.Provider) *integration.PolicyView {
	if p == nil {
		return &integration.PolicyView{}
	}
	return &integration.PolicyView{
		ProviderId: p.ID,
		Name:       p.Name,
		IsEnabled:  p.IsEnabled,
		Priority:   int32(p.Priority),
		Tag:        p.Tag,
	}
}

func geoNotReady() error {
	return status.Error(codes.Unavailable, "ip geo not ready")
}

func mapStoreErr(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, geo.ErrBadDriver),
		errors.Is(err, geo.ErrBadProvider),
		errors.Is(err, geo.ErrEmptyName),
		errors.Is(err, geo.ErrEmptyEndpoint),
		errors.Is(err, geo.ErrNameTaken),
		errors.Is(err, geo.ErrEmptyIP),
		errors.Is(err, geo.ErrBadIP),
		errors.Is(err, geo.ErrIPTaken),
		errors.Is(err, geo.ErrEmptyPolicy),
		errors.Is(err, geo.ErrTooManyPolicies),
		errors.Is(err, geo.ErrDuplicateProvider),
		errors.Is(err, geo.ErrBadTag):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, geo.ErrNotFound),
		errors.Is(err, geo.ErrLocalIPNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		return err
	}
}
