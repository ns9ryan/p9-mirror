package logic

import (
	"testing"

	"oa.98ent.com/p9/oss/common/oss"
	shared "oa.98ent.com/p9/oss/common/config"
	pb "oa.98ent.com/p9/oss/rpc/types/oss"
)

func TestProviderRoundTrip(t *testing.T) {
	in := &pb.Provider{
		Id: 9, Name: "minio", IsDefault: true, Driver: "s3",
		Endpoint: "http://127.0.0.1:9200", FileUrl: "http://cdn.example",
		AccessKey: "ak", SecretKey: "sk", PathStyle: true, Region: "default",
		Buckets: map[string]*pb.BucketSpec{
			"pub": {Mode: oss.ModePublic},
			"pri": {Mode: oss.ModePrivate, SignTtl: 600},
		},
	}
	spec := SpecFromProvider(in)
	if spec.OSS.FileURL != in.FileUrl || !spec.OSS.PathStyle {
		t.Fatalf("spec: %+v", spec)
	}
	if spec.OSS.Buckets["pri"].SignTTL != 600 || spec.OSS.Buckets["pub"].Mode != oss.ModePublic {
		t.Fatalf("buckets: %+v", spec.OSS.Buckets)
	}
}

func TestSpecFromNil(t *testing.T) {
	if spec := SpecFromProvider(nil); spec.OSS.Driver != "" || spec.OSS.Buckets != nil {
		t.Fatalf("nil: %+v", spec)
	}
	empty := SpecFromProvider(&pb.Provider{})
	if empty.OSS.Driver != "" || empty.OSS.Buckets != nil {
		t.Fatalf("empty: %+v", empty)
	}
}

func TestSpecFromSettings(t *testing.T) {
	spec := SpecFromSettings(shared.Settings{
		OSS: oss.Config{
			Driver: "s3", Endpoint: "http://e", AccessKey: "ak", SecretKey: "sk",
			Buckets: map[string]oss.Bucket{"images": {Mode: oss.ModePublic}},
		},
		Upload: shared.UploadSettings{
			DefaultScene: "public",
			Scenes:       map[string]shared.SceneSettings{"public": {Visibility: oss.ModePublic, Bucket: "images"}},
		},
	})
	if spec.OSS.Endpoint != "http://e" || spec.OSS.Buckets["images"].Mode != oss.ModePublic {
		t.Fatalf("spec: %+v", spec)
	}
}
