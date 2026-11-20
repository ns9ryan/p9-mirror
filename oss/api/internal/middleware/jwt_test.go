package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestJWTFromRPC_无地址则放行(t *testing.T) {
	mw := JWTFromRPC(nil, 0)
	called := false
	h := mw(func(http.ResponseWriter, *http.Request) { called = true })
	h(nil, nil)
	if !called {
		t.Fatal("empty endpoints should pass through")
	}
}

func TestAuthFromRPC_无地址则放行(t *testing.T) {
	cli, jwt, auth := AuthFromRPC(nil, 0)
	if cli != nil {
		t.Fatal("expected nil core client")
	}
	called := 0
	jwt(func(http.ResponseWriter, *http.Request) { called++ })(nil, nil)
	auth(func(http.ResponseWriter, *http.Request) { called++ })(nil, nil)
	if called != 2 {
		t.Fatalf("called=%d", called)
	}
}

func TestActionLogFromCore_multipart不截断body(t *testing.T) {
	payload := strings.Repeat("x", 20*1024)
	mw := ActionLogFromCore(nil)
	h := mw(func(_ http.ResponseWriter, r *http.Request) {
		got, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != payload {
			t.Fatalf("got %d", len(got))
		}
	})
	req := httptest.NewRequest(http.MethodPost, "/p9/oss/file/upload", strings.NewReader(payload))
	req.Header.Set("Content-Type", "multipart/form-data; boundary=----x")
	h(httptest.NewRecorder(), req)
}

func TestActionLogFromCore_无客户端则放行(t *testing.T) {
	mw := ActionLogFromCore(nil)
	called := false
	h := mw(func(http.ResponseWriter, *http.Request) { called = true })
	h(nil, nil)
	if !called {
		t.Fatal("nil core should pass through")
	}
}
