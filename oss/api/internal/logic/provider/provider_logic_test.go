package provider

import (
	"context"
	"testing"

	"oa.98ent.com/p9/oss/api/internal/svc"
	"oa.98ent.com/p9/oss/common/oss"
	shared "oa.98ent.com/p9/oss/common/config"
	"oa.98ent.com/p9/oss/rpc/ossclient"

	"google.golang.org/grpc"
)

type activeStub struct {
	ossclient.Oss
	p *ossclient.Provider
}

func (s activeStub) GetActiveProvider(
	_ context.Context, _ *ossclient.Empty, _ ...grpc.CallOption,
) (*ossclient.Provider, error) {
	return s.p, nil
}

func TestActive_无服务商返回进程oss(t *testing.T) {
	l := NewLogic(context.Background(), &svc.ServiceContext{
		Oss:          activeStub{p: &ossclient.Provider{Name: "(none)", Source: "yaml"}},
		ConfigSource: "yaml",
		Settings: shared.Settings{OSS: oss.Config{
			Driver: "s3", Endpoint: "http://127.0.0.1:9200", AccessKey: "minioadmin",
			SecretKey: "secret", PathStyle: true,
			Buckets: map[string]oss.Bucket{"oss-bucket": {Mode: oss.ModePublic}},
		}},
	})
	got, err := l.Active()
	if err != nil {
		t.Fatal(err)
	}
	if got.Endpoint != "http://127.0.0.1:9200" || got.Source != "yaml" || got.SecretKey != "" {
		t.Fatalf("%+v", got)
	}
	if got.Buckets["oss-bucket"].Mode != oss.ModePublic {
		t.Fatalf("buckets: %+v", got.Buckets)
	}
}

func TestActive_有服务商不回退(t *testing.T) {
	l := NewLogic(context.Background(), &svc.ServiceContext{
		Oss: activeStub{p: &ossclient.Provider{
			Name: "prod", Endpoint: "http://s3", Source: "provider:prod",
		}},
		Settings: shared.Settings{OSS: oss.Config{Endpoint: "http://yaml"}},
	})
	got, err := l.Active()
	if err != nil || got.Endpoint != "http://s3" || got.Name != "prod" {
		t.Fatalf("%+v %v", got, err)
	}
}
