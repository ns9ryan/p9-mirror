package store

import (
	"testing"

	"oa.98ent.com/p9/oss/rpc/ent"
)

func TestAttachTenant(t *testing.T) {
	c := ent.NewClient()
	attachTenant(c)
	if n := len(c.File.Interceptors()); n < 1 {
		t.Fatalf("File interceptors: %d", n)
	}
	if n := len(c.Provider.Interceptors()); n < 1 {
		t.Fatalf("Provider interceptors: %d", n)
	}
}
