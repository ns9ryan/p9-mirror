package auth

import (
	"context"
	"errors"
	"testing"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/authpb"
)

func TestLoginRequiresFields(t *testing.T) {
	s := New(nil, nil, Config{})
	_, err := s.Login(context.Background(), &authpb.LoginReq{Password: "pw", OperatorCode: "op"})
	assertKey(t, err, i18n.AuthUsernamePasswordRequired)

	_, err = s.Login(context.Background(), &authpb.LoginReq{Account: "a", Password: "pw"})
	assertKey(t, err, i18n.AuthOperatorCodeRequired)
}

func TestLogoutRequiresRedis(t *testing.T) {
	s := New(nil, nil, Config{RefreshSecret: "refresh"})
	ctx := ctxdata.WithClaims(context.Background(), &ctxdata.Claims{UserID: 1, Salt: "salt", ExpiresAt: 1})
	ctx = ctxdata.WithRawToken(ctx, "access-token")
	err := s.Logout(ctx, &authpb.LogoutReq{RefreshToken: "refresh-token"})
	assertKey(t, err, i18n.InternalError)
}

func TestCheckClientIP(t *testing.T) {
	ctx := ctxdata.WithClientIP(context.Background(), "::ffff:10.0.0.1")
	if err := checkClientIP(ctx, &Claims{ClientIP: "10.0.0.1"}); err != nil {
		t.Fatal(err)
	}
	err := checkClientIP(ctx, &Claims{ClientIP: "10.0.0.2"})
	assertKey(t, err, i18n.AuthIPMismatch)
}

func assertKey(t *testing.T, err error, key string) {
	t.Helper()
	var e *xerr.Error
	if !errors.As(err, &e) || e.Message != key {
		t.Fatalf("err = %v, want %s", err, key)
	}
}
