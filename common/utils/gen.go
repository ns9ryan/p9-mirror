package utils

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"math/big"
	"os"
	"sync/atomic"
	"time"
)

const (
	ShortIDLength = 10 // 短 ID 长度，10 位 Base62 字符串
	LongIDLength  = 20 // 长 ID 长度，20 位 Base62 字符串

	minPasswordLen  = 8                                                               // 密码长度至少为8位
	passwordLower   = "abcdefghijkmnopqrstuvwxyz"                                     // 小写字母
	passwordUpper   = "ABCDEFGHJKLMNPQRSTUVWXYZ"                                      // 大写字母
	passwordDigit   = "23456789"                                                      // 数字
	passwordSpecial = "!@#$%^&*"                                                      // 特殊字符：!@#$%^&*
	passwordAll     = passwordLower + passwordUpper + passwordDigit + passwordSpecial // 所有字符
)

var (
	machineTag uint32        // 机器标签
	seq        atomic.Uint32 // 序号
)

func init() {
	h := sha1.New()
	// 主机名或环境变量 MACHINE_ID
	h.Write([]byte(hostnameOrEnv()))
	var pid [4]byte
	binary.BigEndian.PutUint32(pid[:], uint32(os.Getpid()))
	h.Write(pid[:])
	// 随机8字节，作为额外混淆因子
	var extra [8]byte
	_, _ = rand.Read(extra[:])
	h.Write(extra[:])
	sum := h.Sum(nil)
	// 取前4字节，作为机器标签
	machineTag = binary.BigEndian.Uint32(sum[:4])

	// 随机4字节，作为序号种子
	var seed [4]byte
	if _, err := rand.Read(seed[:]); err == nil {
		seq.Store(binary.BigEndian.Uint32(seed[:])) // 存储序号种子，用于生成唯一的序号
	}
}

// hostnameOrEnv 获取主机名或环境变量 MACHINE_ID
func hostnameOrEnv() string {
	if v := os.Getenv("MACHINE_ID"); v != "" {
		return v
	}
	h, err := os.Hostname()
	if err != nil || h == "" {
		return "unknown"
	}
	return h
}

// encodeBase62FixedBig 将大整数转换为 Base62 定长字符串
func encodeBase62FixedBig(v *big.Int, width int) string {
	const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	base := big.NewInt(62)
	mod := new(big.Int)
	out := make([]byte, width)
	for i := width - 1; i >= 0; i-- {
		v.DivMod(v, base, mod)
		out[i] = alphabet[mod.Int64()]
	}
	return string(out)
}

// mixEntropy 混入时间 / 机器 / 序号，不覆盖随机字节
func mixEntropy(buf []byte) {
	var mix [8]byte
	binary.BigEndian.PutUint32(mix[0:4], uint32(time.Now().Unix()))
	binary.BigEndian.PutUint32(mix[4:8], seq.Add(1)^machineTag)
	for i := range buf {
		buf[i] ^= mix[i%len(mix)]
	}
}

// Gen 生成一个 Base62 定长字符串
func Gen(n int) (string, error) {
	if n < 1 {
		return "", fmt.Errorf("promo: invalid length %d", n)
	}
	// 约需的 bit；用字节数向上取整
	needBits := n * 6 // 略大于 log2(62)≈5.95，够用
	needBytes := (needBits + 7) / 8
	buf := make([]byte, needBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	// XOR 混入时间 / 机器 / 序号，不覆盖随机字节
	mixEntropy(buf)
	v := new(big.Int).SetBytes(buf)
	return encodeBase62FixedBig(v, n), nil
}

// GenShortID 生成一个 10 位 Base62 字符串
func GenShortID() (string, error) {
	return Gen(ShortIDLength)
}

// GenLongID 生成一个 20 位 Base62 字符串
func GenLongID() (string, error) {
	return Gen(LongIDLength)
}

// randCharset 随机选择一个字符
func randCharset(charset string) (byte, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
	if err != nil {
		return 0, err
	}
	return charset[n.Int64()], nil
}

// shuffleBytes 随机打乱字节数组
func shuffleBytes(buf []byte) error {
	for i := len(buf) - 1; i > 0; i-- {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return err
		}
		j := int(n.Int64())
		buf[i], buf[j] = buf[j], buf[i]
	}
	return nil
}

// GenRandomPassword 生成一个随机密码，长度至少 8 位，含大小写、数字和符号
func GenRandomPassword(length int) (string, error) {
	if length < minPasswordLen {
		return "", fmt.Errorf("password: invalid length %d, must be at least %d", length, minPasswordLen)
	}
	buf := make([]byte, length)
	sets := []string{passwordLower, passwordUpper, passwordDigit, passwordSpecial}
	for i, set := range sets {
		b, err := randCharset(set)
		if err != nil {
			return "", err
		}
		buf[i] = b
	}
	for i := len(sets); i < length; i++ {
		b, err := randCharset(passwordAll)
		if err != nil {
			return "", err
		}
		buf[i] = b
	}
	if err := shuffleBytes(buf); err != nil {
		return "", err
	}
	return string(buf), nil
}
