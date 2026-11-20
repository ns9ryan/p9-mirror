package password

import "testing"

// TestHash 验证生成的哈希不是明文
func TestHash(t *testing.T) {
	plainPassword := "P9@Member123"

	hash, err := Hash(plainPassword)
	if err != nil {
		t.Fatalf("生成密码哈希失败: %v", err)
	}
	if hash == plainPassword {
		t.Fatal("密码哈希不能等于明文")
	}
}

// TestCheck 验证正确密码, 错误密码和无效哈希
func TestCheck(t *testing.T) {
	plainPassword := "P9@Member123"

	hash, err := Hash(plainPassword)
	if err != nil {
		t.Fatalf("生成密码哈希失败: %v", err)
	}

	if !Check(hash, plainPassword) {
		t.Fatal("正确密码验证失败")
	}
	if Check(hash, "WrongPassword") {
		t.Fatal("错误密码不应通过验证")
	}
	if Check("invalid-hash", plainPassword) {
		t.Fatal("无效哈希不应通过验证")
	}
}
