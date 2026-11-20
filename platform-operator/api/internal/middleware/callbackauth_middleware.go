// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package middleware

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"net/http"
	"strings"
)

// CallbackAuthMiddleware 调度回调认证中间件
type CallbackAuthMiddleware struct {
	secretHash [sha256.Size]byte // 回调认证密钥哈希
}

// NewCallbackAuthMiddleware 创建调度回调认证中间件
func NewCallbackAuthMiddleware(secret string) (*CallbackAuthMiddleware, error) {
	// 整理认证密钥
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return nil, fmt.Errorf("dispatch callback secret is required")
	}

	// 保存认证密钥哈希
	return &CallbackAuthMiddleware{
		secretHash: sha256.Sum256([]byte(secret)), // 回调认证密钥哈希
	}, nil
}

// Handle 处理调度回调认证
func (m *CallbackAuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 解析Bearer认证密钥
		token, ok := parseBearerToken(r.Header.Get("Authorization"))
		if !ok {
			writeUnauthorized(w)
			return
		}

		// 比较回调认证密钥哈希
		tokenHash := sha256.Sum256([]byte(token))
		if subtle.ConstantTimeCompare(tokenHash[:], m.secretHash[:]) != 1 {
			writeUnauthorized(w)
			return
		}

		// 继续处理请求
		next(w, r)
	}
}

// parseBearerToken 解析Bearer认证密钥
func parseBearerToken(value string) (string, bool) {
	// 拆分认证类型和认证密钥
	parts := strings.SplitN(strings.TrimSpace(value), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}

	// 整理认证密钥
	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", false
	}

	return token, true
}

// writeUnauthorized 返回未认证响应
func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
}
