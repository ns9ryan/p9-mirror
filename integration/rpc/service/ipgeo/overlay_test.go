package ipgeo

import (
	"errors"
	"testing"

	"oa.98ent.com/p9/integration/rpc/internal/ipgeo/geo"
)

func TestOverlayList_AppliesTenantPolicy(t *testing.T) {
	list := []geo.Provider{{
		ID: 1, Name: "maxmind", Driver: geo.DriverHTTP, IsEnabled: true, Priority: 1, Endpoint: "http://x/{ip}",
	}}
	out := overlayList(list, "op-a", map[int64]tenantPolicy{
		1: {IsEnabled: false, Priority: 9},
	})
	if len(out) != 1 || out[0].IsEnabled || out[0].Priority != 9 || out[0].OperatorCode != "op-a" {
		t.Fatalf("got %+v", out[0])
	}
	enabled := filterEnabled(out)
	if len(enabled) != 0 {
		t.Fatalf("enabled %+v", enabled)
	}
}

func TestOverlayList_NoPolicyKeepsDefault(t *testing.T) {
	list := []geo.Provider{{
		ID: 1, Name: "maxmind", Driver: geo.DriverHTTP, IsEnabled: true, Priority: 3, Endpoint: "http://x/{ip}",
	}}
	out := overlayList(list, "op-b", nil)
	if len(out) != 1 || !out[0].IsEnabled || out[0].Priority != 3 || out[0].OperatorCode != "op-b" {
		t.Fatalf("got %+v", out[0])
	}
}

func TestNormalizePolicyItems(t *testing.T) {
	if _, err := normalizePolicyItems(nil); !errors.Is(err, geo.ErrEmptyPolicy) {
		t.Fatalf("empty %v", err)
	}
	if _, err := normalizePolicyItems([]geo.PolicyItem{{ProviderID: 1}, {ProviderID: 1}}); !errors.Is(err, geo.ErrDuplicateProvider) {
		t.Fatalf("dup %v", err)
	}
	if _, err := normalizePolicyItems([]geo.PolicyItem{{ProviderID: 0}}); !errors.Is(err, geo.ErrBadProvider) {
		t.Fatalf("bad %v", err)
	}
}
