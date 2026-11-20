package jwt

import (
	"errors"
	"strings"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
	"oa.98ent.com/p9/common/ctxdata"
)

type Claims struct {
	UserID       int64             `json:"user_id"`
	UserCode     string            `json:"user_code"`
	Username     string            `json:"username"`
	OperatorCode string            `json:"operator_code"`
	RoleCodes    []string          `json:"role_codes"`
	Salt         string            `json:"salt"`
	TokenType    ctxdata.TokenType `json:"token_type"`
	IsPlatform   bool              `json:"is_platform"`
	ClientIP     string            `json:"client_ip"`
	jwtv5.RegisteredClaims
}

// WithType 设置令牌类型
func (c Claims) WithType(tokenType ctxdata.TokenType) Claims {
	c.TokenType = tokenType
	return c
}

// Sign 签名令牌
func Sign(secret string, expireSec int64, c Claims) (string, int64, error) {
	if secret == "" {
		return "", 0, errors.New("jwt secret is empty")
	}
	if expireSec <= 0 {
		expireSec = 86400
	}
	if c.TokenType == "" {
		c.TokenType = ctxdata.TokenTypeAccess
	}
	now := time.Now()
	exp := now.Add(time.Duration(expireSec) * time.Second)
	c.RegisteredClaims = jwtv5.RegisteredClaims{
		IssuedAt:  jwtv5.NewNumericDate(now),
		ExpiresAt: jwtv5.NewNumericDate(exp),
	}
	t := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, c)
	s, err := t.SignedString([]byte(secret))
	return s, exp.Unix(), err
}

// Parse 解析令牌
func Parse(secret, token string) (*Claims, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errors.New("missing token")
	}
	parsed, err := jwtv5.ParseWithClaims(token, &Claims{}, func(t *jwtv5.Token) (any, error) {
		if _, ok := t.Method.(*jwtv5.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	c, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	return c, nil
}

// ParseTyped 解析指定类型的令牌
func ParseTyped(secret, token string, wantType ctxdata.TokenType) (*Claims, error) {
	c, err := Parse(secret, token)
	if err != nil {
		return nil, err
	}
	if !c.TokenType.IsWant(wantType) {
		return nil, errors.New("invalid token type")
	}
	return c, nil
}

// StripBearer 移除令牌前缀
func StripBearer(v string) string {
	v = strings.TrimSpace(v)
	if len(v) > 7 && strings.EqualFold(v[:7], "Bearer ") {
		return strings.TrimSpace(v[7:])
	}
	return v
}

// IsExpired 判断是否是过期错误
func IsExpired(err error) bool {
	return errors.Is(err, jwtv5.ErrTokenExpired)
}
