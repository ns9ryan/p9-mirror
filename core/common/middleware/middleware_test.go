package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"oa.98ent.com/p9/common/ctxdata"
)

func TestPreviewWriteDenied(t *testing.T) {
	allow := [][2]string{
		{http.MethodGet, "/core/user/info"},
		{http.MethodGet, "/core/user/perm"},
		{http.MethodGet, "/core/menu/role"},
		{http.MethodHead, "/core/user/detail"},
		{http.MethodOptions, "/core/role/detail"},
		{http.MethodPost, "/core/user/list"},
		{http.MethodPost, "/core/role/list"},
		{http.MethodPost, "/core/menu/list"},
		{http.MethodPost, "/core/api/list"},
		{http.MethodPost, "/core/authority/menu/role"},
		{http.MethodPost, "/core/authority/api/role"},
		{http.MethodPost, "/core/menu/update"},
		{http.MethodPost, "/core/authority/menu/update"},
	}
	deny := [][2]string{
		{http.MethodPost, "/core/user/create"},
		{http.MethodPost, "/core/user/update"},
		{http.MethodPost, "/core/user/delete"},
		{http.MethodPost, "/core/user/password"},
		{http.MethodPost, "/core/user/password/self"},
		{http.MethodPost, "/core/user/roles"},
		{http.MethodPost, "/core/user/ipWhitelist"},
		{http.MethodPost, "/core/logout"},
		{http.MethodPost, "/core/logout/all"},
		{http.MethodPost, "/core/role/create"},
		{http.MethodPost, "/core/authority/api/update"},
		{http.MethodPut, "/core/user/list"},
		{http.MethodDelete, "/core/user/info"},
	}
	for _, c := range allow {
		if previewWriteDenied(c[0], c[1]) {
			t.Fatalf("allow %s %s", c[0], c[1])
		}
	}
	for _, c := range deny {
		if !previewWriteDenied(c[0], c[1]) {
			t.Fatalf("deny %s %s", c[0], c[1])
		}
	}
}

func TestClientIPMiddleware(t *testing.T) {
	called := false
	h := ClientIP(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if got := ctxdata.ClientIPFromCtx(r.Context()); got != "10.0.0.1" {
			t.Fatalf("ip %q", got)
		}
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1")
	h(httptest.NewRecorder(), req)
	if !called {
		t.Fatal("next not called")
	}
}
