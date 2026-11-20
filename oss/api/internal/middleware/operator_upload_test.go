package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zeromicro/go-zero/rest"
	"oa.98ent.com/p9/common/ctxdata"
)

func TestSelectUploadAuth(t *testing.T) {
	for _, scenario := range []struct {
		name, scope, path         string
		wantDefault, wantOperator bool
	}{
		{name: "existing platform upload", path: operatorUploadPath, wantDefault: true},
		{name: "operator upload", scope: operatorAuthScope, path: operatorUploadPath, wantOperator: true},
		{name: "provider access rejected", scope: operatorAuthScope, path: "/p9/oss/provider/list"},
		{name: "unknown scope rejected", scope: "unknown", path: operatorUploadPath},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			var defaultCalled, operatorCalled bool
			defaultAuth := func(next http.HandlerFunc) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) { defaultCalled = true; next(w, r) }
			}
			operatorAuth := func(next http.HandlerFunc) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) { operatorCalled = true; next(w, r) }
			}
			request := httptest.NewRequest(http.MethodPost, scenario.path, nil)
			request.Header.Set(operatorAuthScopeHeader, scenario.scope)
			SelectUploadAuth(defaultAuth, operatorAuth)(func(http.ResponseWriter, *http.Request) {})(httptest.NewRecorder(), request)
			if defaultCalled != scenario.wantDefault || operatorCalled != scenario.wantOperator {
				t.Fatalf("default=%v operator=%v", defaultCalled, operatorCalled)
			}
		})
	}
}

func TestOperatorUploadAuthRequiresConfiguredCoreAndOperatorIdentity(t *testing.T) {
	for _, scenario := range []struct {
		name                          string
		configured, platform, preview bool
		wantCalled                    bool
	}{
		{name: "missing core", configured: false},
		{name: "platform token", configured: true, platform: true},
		{name: "preview token", configured: true, preview: true},
		{name: "operator access", configured: true, wantCalled: true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			jwt := func(next http.HandlerFunc) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					tokenType := ctxdata.TokenTypeAccess
					if scenario.preview {
						tokenType = ctxdata.TokenTypePreview
					}
					ctx := ctxdata.WithClaims(r.Context(), &ctxdata.Claims{UserID: 1, UserType: ctxdata.UserTypeOperator, OperatorCode: "OP_test", IsPlatform: scenario.platform, TokenType: tokenType})
					next(w, r.WithContext(ctx))
				}
			}
			identity := rest.Middleware(func(next http.HandlerFunc) http.HandlerFunc { return next })
			called := false
			OperatorUploadAuth(scenario.configured, jwt, identity, identity)(func(http.ResponseWriter, *http.Request) { called = true })(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, operatorUploadPath, nil))
			if called != scenario.wantCalled {
				t.Fatalf("called=%v", called)
			}
		})
	}
}
