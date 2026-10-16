package auth

import (
	"errors"
	"strings"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

const (
	tokenAccess  = "access"
	tokenRefresh = "refresh"
)

// Claims 是代理访问令牌和刷新令牌的载荷。
type Claims struct {
	UserID       int64  `json:"user_id"`
	Username     string `json:"username"`
	OperatorCode string `json:"operator_code"`
	Salt         string `json:"salt"`
	TokenType    string `json:"token_type"`
	ClientIP     string `json:"client_ip"`
	jwtv5.RegisteredClaims
}

// WithType 返回指定令牌类型的副本。
func (c Claims) WithType(tokenType string) Claims {
	c.TokenType = tokenType
	return c
}

// Sign 签发 HS256 令牌，并返回过期时间的 Unix 秒。
func Sign(secret string, expireSec int64, c Claims) (string, int64, error) {
	if secret == "" {
		return "", 0, errors.New("jwt secret is empty")
	}
	if expireSec <= 0 {
		expireSec = 86400
	}
	if c.TokenType == "" {
		c.TokenType = tokenAccess
	}
	now := time.Now()
	exp := now.Add(time.Duration(expireSec) * time.Second)
	c.RegisteredClaims = jwtv5.RegisteredClaims{
		IssuedAt:  jwtv5.NewNumericDate(now),
		ExpiresAt: jwtv5.NewNumericDate(exp),
	}
	t := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, c)
	signed, err := t.SignedString([]byte(secret))
	return signed, exp.Unix(), err
}

// Parse 校验并解析令牌。
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

// ParseTyped 校验令牌类型后解析。
func ParseTyped(secret, token, wantType string) (*Claims, error) {
	c, err := Parse(secret, token)
	if err != nil {
		return nil, err
	}
	if c.TokenType != wantType {
		return nil, errors.New("invalid token type")
	}
	return c, nil
}

// StripBearer 去掉 Bearer 前缀。
func StripBearer(v string) string {
	v = strings.TrimSpace(v)
	if len(v) > 7 && strings.EqualFold(v[:7], "Bearer ") {
		return strings.TrimSpace(v[7:])
	}
	return v
}

// IsExpired 判断错误是否为令牌过期。
func IsExpired(err error) bool {
	return errors.Is(err, jwtv5.ErrTokenExpired)
}
