package fieldcrypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
)

var (
	ErrEmptyPlaintext    = errors.New("fieldcrypto: empty plaintext")
	ErrEmptyNormalized   = errors.New("fieldcrypto: empty normalized input")
	ErrInvalidEncKey     = errors.New("fieldcrypto: enc key must be 16, 24, or 32 bytes")
	ErrEmptyMacKey       = errors.New("fieldcrypto: mac key must not be empty")
	ErrInvalidCiphertext = errors.New("fieldcrypto: invalid ciphertext")
)

// Cryptor AES-GCM 可逆加密 + HMAC-SHA256 检索摘要。
type Cryptor struct {
	gcm    cipher.AEAD
	macKey []byte
}

// New 创建 Cryptor。encKey 长度须为 16/24/32；macKey 非空。
func New(encKey, macKey []byte) (*Cryptor, error) {
	switch len(encKey) {
	case 16, 24, 32:
	default:
		return nil, ErrInvalidEncKey
	}
	if len(macKey) == 0 {
		return nil, ErrEmptyMacKey
	}
	block, err := aes.NewCipher(encKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	mac := make([]byte, len(macKey))
	copy(mac, macKey)
	return &Cryptor{gcm: gcm, macKey: mac}, nil
}

// Seal 加密明文，返回 base64(nonce||sealed)。明文不可为空。
func (c *Cryptor) Seal(plain string) (string, error) {
	if c == nil || c.gcm == nil {
		return "", fmt.Errorf("fieldcrypto: nil cryptor")
	}
	if plain == "" {
		return "", ErrEmptyPlaintext
	}
	nonce := make([]byte, c.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := c.gcm.Seal(nil, nonce, []byte(plain), nil)
	out := make([]byte, 0, len(nonce)+len(sealed))
	out = append(out, nonce...)
	out = append(out, sealed...)
	return base64.StdEncoding.EncodeToString(out), nil
}

// Open 解密 Seal 产出的明文。
func (c *Cryptor) Open(ciphertext string) (string, error) {
	if c == nil || c.gcm == nil {
		return "", fmt.Errorf("fieldcrypto: nil cryptor")
	}
	raw, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", ErrInvalidCiphertext
	}
	ns := c.gcm.NonceSize()
	if len(raw) < ns+1 {
		return "", ErrInvalidCiphertext
	}
	nonce, sealed := raw[:ns], raw[ns:]
	plain, err := c.gcm.Open(nil, nonce, sealed, nil)
	if err != nil {
		return "", ErrInvalidCiphertext
	}
	return string(plain), nil
}

// LookupHash 对已规范化字符串做 HMAC-SHA256，返回 64 字符 hex。
func (c *Cryptor) LookupHash(normalized string) (string, error) {
	if c == nil {
		return "", fmt.Errorf("fieldcrypto: nil cryptor")
	}
	if normalized == "" {
		return "", ErrEmptyNormalized
	}
	mac := hmac.New(sha256.New, c.macKey)
	_, _ = mac.Write([]byte(normalized))
	return hex.EncodeToString(mac.Sum(nil)), nil
}

// NormalizeEmail trim + 小写。
func NormalizeEmail(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// NormalizePhone trim、去空格、去掉前导 '+'（兼容「区号-号码」）。
func NormalizePhone(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "")
	return strings.TrimPrefix(s, "+")
}
