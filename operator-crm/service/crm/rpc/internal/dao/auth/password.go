package auth

import "golang.org/x/crypto/bcrypt"

// HashPassword 生成登录密码哈希。
func HashPassword(pw string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(h), err
}

// CheckPassword 比对登录密码和哈希。
func CheckPassword(hash, pw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw)) == nil
}
