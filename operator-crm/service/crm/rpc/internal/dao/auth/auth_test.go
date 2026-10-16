package auth

import (
	"testing"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

func TestSignAndParse(t *testing.T) {
	secret := "access-secret"
	token, exp, err := Sign(secret, 60, Claims{
		UserID:       7,
		Username:     "agent",
		OperatorCode: "op",
		Salt:         "0123456789abcdef",
		ClientIP:     "10.0.0.8",
	}.WithType(tokenAccess))
	if err != nil {
		t.Fatal(err)
	}
	if exp <= time.Now().Unix() {
		t.Fatalf("expire = %d", exp)
	}
	got, err := ParseTyped(secret, StripBearer("Bearer "+token), tokenAccess)
	if err != nil {
		t.Fatal(err)
	}
	if got.UserID != 7 || got.Username != "agent" || got.OperatorCode != "op" || got.ClientIP != "10.0.0.8" {
		t.Fatalf("claims = %+v", got)
	}
	if _, err := ParseTyped(secret, token, tokenRefresh); err == nil {
		t.Fatal("refresh type should be rejected")
	}
}

func TestParseExpired(t *testing.T) {
	secret := "access-secret"
	c := Claims{
		UserID:    1,
		TokenType: tokenAccess,
		RegisteredClaims: jwtv5.RegisteredClaims{
			ExpiresAt: jwtv5.NewNumericDate(time.Now().Add(-time.Minute)),
		},
	}
	token, err := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, c).SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	_, err = Parse(secret, token)
	if !IsExpired(err) {
		t.Fatalf("err = %v", err)
	}
}

func TestCheckPassword(t *testing.T) {
	hash, err := HashPassword("secret")
	if err != nil {
		t.Fatal(err)
	}
	if !CheckPassword(hash, "secret") {
		t.Fatal("password should match")
	}
	if CheckPassword(hash, "other") {
		t.Fatal("password should not match")
	}
}
