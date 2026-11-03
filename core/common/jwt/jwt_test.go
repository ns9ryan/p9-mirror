package jwt

import (
	"testing"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
	"oa.98ent.com/p9/common/ctxdata"
)

func TestSignParseRoundTrip(t *testing.T) {
	tok, exp, err := Sign("secret", 60, Claims{
		UserID:       1,
		UserCode:     "abc",
		Username:     "admin",
		OperatorCode: "demo",
		RoleCodes:    []string{"super_admin"},
		Salt:         "s1",
		TokenType:    ctxdata.TokenTypeAccess,
		ClientIP:     "10.0.0.1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if tok == "" || exp == 0 {
		t.Fatal("empty token")
	}
	c, err := Parse("secret", tok)
	if err != nil {
		t.Fatal(err)
	}
	if c.UserID != 1 || c.Salt != "s1" || c.OperatorCode != "demo" || !c.TokenType.IsAccess() || c.ClientIP != "10.0.0.1" {
		t.Fatalf("claims %+v", c)
	}
	if _, err := Parse("other", tok); err == nil {
		t.Fatal("expected bad secret")
	}
}

func TestParseTypedRejectsWrongType(t *testing.T) {
	access, _, err := Sign("access-secret", 60, Claims{UserID: 1, Salt: "s1", TokenType: ctxdata.TokenTypeAccess})
	if err != nil {
		t.Fatal(err)
	}
	refresh, _, err := Sign("refresh-secret", 120, Claims{UserID: 1, Salt: "s1", TokenType: ctxdata.TokenTypeRefresh})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseTyped("refresh-secret", access, ctxdata.TokenTypeRefresh); err == nil {
		t.Fatal("access must not parse as refresh")
	}
	if _, err := ParseTyped("access-secret", refresh, ctxdata.TokenTypeAccess); err == nil {
		t.Fatal("refresh must not parse as access")
	}
	c, err := ParseTyped("refresh-secret", refresh, ctxdata.TokenTypeRefresh)
	if err != nil || c.Salt != "s1" || !c.TokenType.IsRefresh() {
		t.Fatalf("refresh parse %+v %v", c, err)
	}
}

func TestSignParsePreview(t *testing.T) {
	tok, exp, err := Sign("secret", 60, Claims{
		UserID:       8,
		UserCode:     "u8",
		Username:     "admin",
		OperatorCode: "demo",
		RoleCodes:    []string{"super_admin"},
		Salt:         "s1",
		TokenType:    ctxdata.TokenTypePreview,
		IsPlatform:   true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if tok == "" || exp == 0 {
		t.Fatal("empty token")
	}
	c, err := Parse("secret", tok)
	if err != nil {
		t.Fatal(err)
	}
	if c.UserID != 8 || c.Username != "admin" || c.Salt != "s1" || c.OperatorCode != "demo" || !c.TokenType.IsPreview() || !c.IsPlatform {
		t.Fatalf("claims %+v", c)
	}
	if _, err := ParseTyped("secret", tok, ctxdata.TokenTypeRefresh); err == nil {
		t.Fatal("preview must not parse as refresh")
	}
	if _, err := ParseTyped("secret", tok, ctxdata.TokenTypeAccess); err == nil {
		t.Fatal("preview must not parse as access")
	}
	got, err := ParseTyped("secret", tok, ctxdata.TokenTypePreview)
	if err != nil || got.RoleCodes[0] != "super_admin" {
		t.Fatalf("preview parse %+v %v", got, err)
	}
}

func TestStripBearer(t *testing.T) {
	if got := StripBearer("Bearer abc"); got != "abc" {
		t.Fatalf("got %q", got)
	}
	if got := StripBearer("abc"); got != "abc" {
		t.Fatalf("got %q", got)
	}
}

func TestParseExpired(t *testing.T) {
	tok := mustSignExpired(t, "secret", Claims{UserID: 1, Salt: "s1", TokenType: ctxdata.TokenTypeAccess})
	_, err := Parse("secret", tok)
	if err == nil {
		t.Fatal("expected expired token")
	}
	if !IsExpired(err) {
		t.Fatalf("IsExpired=false err=%v", err)
	}
	_, badSecret := Parse("other", tok)
	if IsExpired(nil) || IsExpired(badSecret) {
		t.Fatal("IsExpired must be only for expiry")
	}
}

func mustSignExpired(t *testing.T, secret string, c Claims) string {
	t.Helper()
	c.RegisteredClaims = jwtv5.RegisteredClaims{
		IssuedAt:  jwtv5.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		ExpiresAt: jwtv5.NewNumericDate(time.Now().Add(-time.Hour)),
	}
	s, err := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, c).SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	return s
}
