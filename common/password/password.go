// ============================================================================
// Package password 提供基于 bcrypt 的密码哈希与校验
// ============================================================================

package password

import "golang.org/x/crypto/bcrypt"

// Hash 生成 bcrypt 密码哈希
func Hash(plainPassword string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	return string(hash), err
}

// Check 检查明文密码是否与哈希匹配
func Check(hash, plainPassword string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plainPassword)) == nil
}
