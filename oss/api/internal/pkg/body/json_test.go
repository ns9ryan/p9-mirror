package body

import (
	"net/http"
	"strings"
	"testing"

	"oa.98ent.com/p9/oss/common/oss"
)

func TestJSONProviderCreate(t *testing.T) {
	raw := `{
		"name": "local-minio",
		"is_default": false,
		"driver": "s3",
		"endpoint": "http://127.0.0.1:9200",
		"file_url": "http://127.0.0.1:9200",
		"access_key": "minioadmin",
		"secret_key": "minioadmin",
		"path_style": true,
		"region": "",
		"buckets": {
			"oss-bucket": { "mode": "public" },
			"oss-bucket-private": { "mode": "private", "sign_ttl": 600 }
		}
	}`
	r := httptestRequest(raw)
	var v struct {
		Name      string                `json:"name"`
		AccessKey string                `json:"access_key"`
		Buckets   map[string]oss.Bucket `json:"buckets"`
	}
	if err := JSON(r, &v); err != nil {
		t.Fatal(err)
	}
	if v.Name != "local-minio" || v.AccessKey != "minioadmin" {
		t.Fatalf("got %+v", v)
	}
	if v.Buckets["oss-bucket"].Mode != "public" {
		t.Fatalf("buckets=%v", v.Buckets)
	}
}

func httptestRequest(raw string) *http.Request {
	r, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(raw))
	if err != nil {
		panic(err)
	}
	return r
}
