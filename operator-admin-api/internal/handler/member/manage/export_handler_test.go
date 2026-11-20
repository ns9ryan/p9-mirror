package manage

import (
	"strings"
	"testing"
)

func TestStrictExportFilters(t *testing.T) {
	for _, body := range []string{`null`, `[]`, `{"unknown":1}`, `{"filter_ip":null}`, `{"range_today_deposit":null}`, `{"range_today_withdraw":[]}`, `{"agent_id":1,"agent_id":2}`, `{"keyword":"one"} {"keyword":"two"}`, `{"page":1}`, `{"sort_by":1}`, `{"phone":true}`, `{"agent_id":"1"}`} {
		if _, err := decodeExportRequest(strings.NewReader(body)); err == nil {
			t.Fatalf("silently accepted %s", body)
		}
	}
	request, err := decodeExportRequest(strings.NewReader(`{"keyword":"MiXeD","filter_ip":"192.0.2.1","filter_device":"device","agent_type":1,"columns":["username","status"]}`))
	if err != nil || request.Keyword != "MiXeD" || request.FilterIp != "192.0.2.1" || request.AgentType != 1 {
		t.Fatal("valid filters lost")
	}
}

func TestExportIncludeDemoSelection(t *testing.T) {
	for _, scenario := range []struct {
		body        string
		includeDemo bool
	}{
		{body: `{}`},
		{body: `{"include_demo":false}`},
		{body: `{"include_demo":true}`, includeDemo: true},
	} {
		request, err := decodeExportRequest(strings.NewReader(scenario.body))
		if err != nil || request.IncludeDemo != scenario.includeDemo {
			t.Fatalf("include_demo selection lost: %s, %v", scenario.body, err)
		}
	}
	for _, body := range []string{`{"include_demo":null}`, `{"include_demo":"true"}`, `{"include_demo":1}`} {
		if _, err := decodeExportRequest(strings.NewReader(body)); err == nil {
			t.Fatalf("invalid include_demo accepted: %s", body)
		}
	}
}
