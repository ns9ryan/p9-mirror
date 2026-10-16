package convert

import (
	"strings"

	"oa.98ent.com/p9/oss/api/internal/types"
	"oa.98ent.com/p9/oss/common/oss"
	shared "oa.98ent.com/p9/oss/common/config"
	"oa.98ent.com/p9/oss/rpc/ossclient"
)

func ProviderFromOSS(cfg oss.Config, source string, maskSecret bool) *types.ProviderView {
	v := &types.ProviderView{
		Name:      "(runtime)",
		IsDefault: true,
		Driver:    cfg.Driver,
		Endpoint:  cfg.Endpoint,
		FileURL:   cfg.FileURL,
		AccessKey: cfg.AccessKey,
		PathStyle: cfg.PathStyle,
		Region:    cfg.Region,
		Buckets:   cfg.Buckets,
		Source:    source,
	}
	if !maskSecret {
		v.SecretKey = cfg.SecretKey
	}
	return v
}

func IsProviderSource(p *ossclient.Provider) bool {
	return p != nil && strings.HasPrefix(p.GetSource(), "provider:")
}

func Provider(in *ossclient.Provider) *types.ProviderView {
	if in == nil {
		return nil
	}
	return &types.ProviderView{
		ID:        in.GetId(),
		Name:      in.GetName(),
		IsDefault: in.GetIsDefault(),
		Driver:    in.GetDriver(),
		Endpoint:  in.GetEndpoint(),
		FileURL:   in.GetFileUrl(),
		AccessKey: in.GetAccessKey(),
		SecretKey: in.GetSecretKey(),
		PathStyle: in.GetPathStyle(),
		Region:    in.GetRegion(),
		Buckets:   bucketsFromPB(in.GetBuckets()),
		Source:    in.GetSource(),
	}
}

func ProviderReq(v *types.ProviderView) *ossclient.Provider {
	if v == nil {
		return &ossclient.Provider{}
	}
	return &ossclient.Provider{
		Id:        v.ID,
		Name:      v.Name,
		IsDefault: v.IsDefault,
		Driver:    v.Driver,
		Endpoint:  v.Endpoint,
		FileUrl:   v.FileURL,
		AccessKey: v.AccessKey,
		SecretKey: v.SecretKey,
		PathStyle: v.PathStyle,
		Region:    v.Region,
		Buckets:   bucketsToPB(v.Buckets),
	}
}

func SpecFromProvider(in *ossclient.Provider) shared.ProviderSpec {
	if in == nil {
		return shared.ProviderSpec{}
	}
	return shared.ProviderSpec{
		OSS: oss.Config{
			Driver: in.GetDriver(), Endpoint: in.GetEndpoint(), FileURL: in.GetFileUrl(),
			AccessKey: in.GetAccessKey(), SecretKey: in.GetSecretKey(),
			PathStyle: in.GetPathStyle(), Region: in.GetRegion(),
			Buckets: bucketsFromPB(in.GetBuckets()),
		},
	}
}

func File(in *ossclient.FileData) *types.FileData {
	if in == nil {
		return nil
	}
	return &types.FileData{
		URL: in.GetUrl(), Key: in.GetKey(), Scene: in.GetScene(),
		Visibility: in.GetVisibility(), URLExpiresIn: int(in.GetUrlExpiresIn()),
	}
}

func FileList(in *ossclient.FileList) *types.FileListData {
	out := &types.FileListData{List: []types.FileInfo{}}
	if in == nil {
		return out
	}
	out.Total = in.GetTotal()
	for _, row := range in.GetList() {
		if row == nil {
			continue
		}
		out.List = append(out.List, types.FileInfo{
			ID: row.GetId(), Key: row.GetKey(), Scene: row.GetScene(), Bucket: row.GetBucket(),
			Name: row.GetName(), Size: row.GetSize(), ContentType: row.GetContentType(),
			URL: row.GetUrl(), Visibility: row.GetVisibility(),
			URLExpiresIn: int(row.GetUrlExpiresIn()), CreatedAt: row.GetCreatedAt(),
		})
	}
	return out
}

func ProviderList(in *ossclient.ProviderList) *types.ProviderListData {
	out := &types.ProviderListData{List: []types.ProviderView{}}
	if in == nil {
		return out
	}
	out.Total = in.GetTotal()
	for _, row := range in.GetList() {
		if v := Provider(row); v != nil {
			out.List = append(out.List, *v)
		}
	}
	return out
}

func bucketsFromPB(in map[string]*ossclient.BucketSpec) map[string]oss.Bucket {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]oss.Bucket, len(in))
	for name, spec := range in {
		if spec == nil {
			continue
		}
		out[name] = oss.Bucket{Mode: spec.GetMode(), SignTTL: int(spec.GetSignTtl())}
	}
	return out
}

func bucketsToPB(in map[string]oss.Bucket) map[string]*ossclient.BucketSpec {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]*ossclient.BucketSpec, len(in))
	for name, b := range in {
		out[name] = &ossclient.BucketSpec{Mode: b.Mode, SignTtl: int32(b.SignTTL)}
	}
	return out
}
