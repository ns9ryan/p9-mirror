package convert

import (
	"oa.98ent.com/p9/integration/api/internal/types"
	"oa.98ent.com/p9/integration/rpc/types/integration"
)

func SearchIpResp(in *integration.SearchIpResp) *types.SearchIpResp {
	resp := &types.SearchIpResp{List: make([]types.IpInfo, 0)}
	if in == nil {
		return resp
	}
	for _, item := range in.GetList() {
		resp.List = append(resp.List, IpInfo(item))
	}
	return resp
}

func IpInfo(in *integration.IpInfo) types.IpInfo {
	if in == nil {
		return types.IpInfo{}
	}
	return types.IpInfo{
		Query:           in.GetQuery(),
		Continent:       in.GetContinent(),
		ContinentCode:   in.GetContinentCode(),
		Country:         in.GetCountry(),
		CountryCode:     in.GetCountryCode(),
		Province:        in.GetProvince(),
		ProvinceIsoCode: in.GetProvinceIsoCode(),
		City:            in.GetCity(),
		Zip:             in.GetZip(),
		Lat:             in.GetLat(),
		Lon:             in.GetLon(),
		Timezone:        in.GetTimezone(),
		Isp:             in.GetIsp(),
		Org:             in.GetOrg(),
		As:              in.GetAs(),
		Postal:          in.GetPostal(),
		Source:          in.GetSource(),
	}
}

func LocalIPListData(in *integration.LocalIPListResp) *types.LocalIPListData {
	resp := &types.LocalIPListData{List: make([]types.LocalIPView, 0)}
	if in == nil {
		return resp
	}
	resp.Total = in.GetTotal()
	for _, item := range in.GetList() {
		resp.List = append(resp.List, LocalIPView(item))
	}
	return resp
}

func LocalIPView(in *integration.LocalIP) types.LocalIPView {
	if in == nil {
		return types.LocalIPView{}
	}
	return types.LocalIPView{
		Id:              in.GetId(),
		Ip:              in.GetIp(),
		Continent:       in.GetContinent(),
		ContinentCode:   in.GetContinentCode(),
		Country:         in.GetCountry(),
		CountryCode:     in.GetCountryCode(),
		Province:        in.GetProvince(),
		ProvinceIsoCode: in.GetProvinceIsoCode(),
		City:            in.GetCity(),
		Zip:             in.GetZip(),
		Lat:             in.GetLat(),
		Lon:             in.GetLon(),
		Timezone:        in.GetTimezone(),
		Isp:             in.GetIsp(),
		Org:             in.GetOrg(),
		As:              in.GetAs(),
		Postal:          in.GetPostal(),
	}
}

func LocalIP(in *types.LocalIPView) *integration.LocalIP {
	if in == nil {
		return &integration.LocalIP{}
	}
	return &integration.LocalIP{
		Id:              in.Id,
		Ip:              in.Ip,
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
	}
}

func PolicyListData(in *integration.PolicyListResp) *types.PolicyListData {
	resp := &types.PolicyListData{List: make([]types.PolicyView, 0)}
	if in == nil {
		return resp
	}
	resp.Total = in.GetTotal()
	for _, item := range in.GetList() {
		resp.List = append(resp.List, PolicyView(item))
	}
	return resp
}

func PolicyView(in *integration.PolicyView) types.PolicyView {
	if in == nil {
		return types.PolicyView{}
	}
	return types.PolicyView{
		ProviderId: in.GetProviderId(),
		Name:       in.GetName(),
		IsEnabled:  in.GetIsEnabled(),
		Priority:   int64(in.GetPriority()),
		Tag:        in.GetTag(),
	}
}

func PolicyItems(list []types.PolicyItem) []*integration.PolicyItem {
	out := make([]*integration.PolicyItem, 0, len(list))
	for _, item := range list {
		out = append(out, &integration.PolicyItem{
			ProviderId: item.ProviderId,
			IsEnabled:  item.IsEnabled,
			Priority:   int32(item.Priority),
		})
	}
	return out
}
