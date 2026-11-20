package handler

import (
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
	"net/http"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"testing"
)

func TestMemberAssignmentRoutes(t *testing.T) {
	var cfg rest.RestConf
	if err := conf.LoadFromJsonBytes([]byte(`{"Name":"route-test","Host":"127.0.0.1","Port":0}`), &cfg); err != nil {
		t.Fatal(err)
	}
	server, err := rest.NewServer(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Stop()
	pass := func(next http.HandlerFunc) http.HandlerFunc { return next }
	RegisterHandlers(server, &svc.ServiceContext{Jwt: pass, ActionLog: pass, Authority: pass})
	seen := map[string]bool{}
	for _, r := range server.Routes() {
		key := r.Method + " " + r.Path
		if seen[key] {
			t.Fatalf("duplicate route %s", key)
		}
		seen[key] = true
	}
	for _, key := range []string{"POST /member/level/auto/update", "POST /member/level/fixed/update", "POST /member/fixed-level/update", "GET /member/detail", "POST /member/status/update", "POST /member/auth/errors/reset", "POST /member/remark/update", "POST /member/chat/update", "POST /member/betting/update", "POST /member/withdrawal/update"} {
		if !seen[key] {
			t.Errorf("missing route %s", key)
		}
	}
}
