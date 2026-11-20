package logic

import (
	"oa.98ent.com/p9/oss/common/oss"
	shared "oa.98ent.com/p9/oss/common/config"
	"oa.98ent.com/p9/oss/rpc/ent"
	"oa.98ent.com/p9/oss/rpc/internal/store"
	pb "oa.98ent.com/p9/oss/rpc/types/oss"
)

func SpecFromSettings(s shared.Settings) shared.ProviderSpec {
	c := shared.CloneSettings(s)
	return shared.ProviderSpec{OSS: c.OSS}
}

func SpecFromProvider(v *pb.Provider) shared.ProviderSpec {
	if v == nil {
		return shared.ProviderSpec{}
	}
	return shared.ProviderSpec{
		OSS: oss.Config{
			Driver:    v.GetDriver(),
			Endpoint:  v.GetEndpoint(),
			FileURL:   v.GetFileUrl(),
			AccessKey: v.GetAccessKey(),
			SecretKey: v.GetSecretKey(),
			PathStyle: v.GetPathStyle(),
			Region:    v.GetRegion(),
			Buckets:   BucketsFromPB(v.GetBuckets()),
		},
	}
}

func ProviderFromEnt(p *ent.Provider, maskSecret bool) *pb.Provider {
	if p == nil {
		return &pb.Provider{}
	}
	spec, _ := store.SpecOf(p)
	out := &pb.Provider{
		Id:        p.ID,
		Name:      p.Name,
		IsDefault: p.IsDefault,
		Driver:    spec.OSS.Driver,
		Endpoint:  spec.OSS.Endpoint,
		FileUrl:   spec.OSS.FileURL,
		AccessKey: spec.OSS.AccessKey,
		PathStyle: spec.OSS.PathStyle,
		Region:    spec.OSS.Region,
		Buckets:   BucketsToPB(spec.OSS.Buckets),
	}
	if !maskSecret {
		out.SecretKey = spec.OSS.SecretKey
	}
	return out
}

func BucketsFromPB(in map[string]*pb.BucketSpec) map[string]oss.Bucket {
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

func BucketsToPB(in map[string]oss.Bucket) map[string]*pb.BucketSpec {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]*pb.BucketSpec, len(in))
	for name, b := range in {
		out[name] = &pb.BucketSpec{Mode: b.Mode, SignTtl: int32(b.SignTTL)}
	}
	return out
}
