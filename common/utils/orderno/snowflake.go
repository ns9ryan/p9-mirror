package orderno

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

const (
	// 自定义起始时间戳 (Epoch)，例如 2024-09-01 00:00:00 UTC 的毫秒值
	// 调整此时间可以控制生成的 ID 位数（当前差值约1~2年时正好为18位）
	epoch int64 = 1725148800000

	nodeBits     uint8 = 10 // 机器节点位数 (支持 0~1023)
	sequenceBits uint8 = 12 // 序列号位数 (每毫秒支持 0~4095)

	maxNode     int64 = -1 ^ (-1 << nodeBits)
	maxSequence int64 = -1 ^ (-1 << sequenceBits)

	nodeShift      uint8 = sequenceBits
	timestampShift uint8 = sequenceBits + nodeBits
)

type Snowflake struct {
	mu        sync.Mutex
	timestamp int64
	nodeID    int64
	sequence  int64
}

// NewSnowflake 创建新的雪花 ID 生成器
// nodeID 机器节点ID，必须在 0 到 1023 之间
// nodeID == 0 时，自动获取本机内网 IPv4 地址并计算 10 位 NodeID (0~1023)
func NewSnowflake(nodeID int64) *Snowflake {
	var err error
	if nodeID == 0 {
		nodeID, _, err = GetNodeIDFromLocalIP()
		if err != nil {
			panic(err)
		}
	}
	if nodeID < 0 || nodeID > maxNode {
		panic(fmt.Sprintf("nodeID 必须在 0 到 %d 之间", maxNode))
	}
	return &Snowflake{nodeID: nodeID}
}

// Generate 生成 64 位纯数字雪花 ID
func (s *Snowflake) Generate() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixMilli()
	if now == s.timestamp {
		s.sequence = (s.sequence + 1) & maxSequence
		if s.sequence == 0 {
			// 当前毫秒序列号用完，等待下一毫秒
			for now <= s.timestamp {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		s.sequence = 0
	}

	s.timestamp = now

	// 位运算拼接：时间戳差值 << 22 | 节点ID << 12 | 序列号
	id := ((now - epoch) << timestampShift) | (s.nodeID << nodeShift) | s.sequence
	return id
}

// GenerateWithPrefix 生成带前缀的 64 位纯数字雪花 ID
func (s *Snowflake) GenerateWithPrefix(prefix string) string {
	return fmt.Sprintf("%s%d", prefix, s.Generate())
}

// ParseID 反解你提供的 ID (230620057519058944) 内部结构
func (s *Snowflake) ParseID(id int64) {
	timeDiff := id >> timestampShift
	nodeID := (id >> nodeShift) & maxNode
	sequence := id & maxSequence
	fmt.Printf("ID: %d\n -> 时间戳差值(ms): %d\n -> 机器节点ID: %d\n -> 毫秒内序列号: %d\n",
		id, timeDiff, nodeID, sequence)
}

// GetNodeIDFromLocalIP 获取本机内网 IPv4 地址并计算 10 位 NodeID (0~1023)
func GetNodeIDFromLocalIP() (int64, string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return 0, "", fmt.Errorf("获取网卡地址失败: %w", err)
	}

	for _, addr := range addrs {
		// 检查是否为 IP 地址且非回环地址 (127.0.0.1)
		ipNet, ok := addr.(*net.IPNet)
		if !ok || ipNet.IP.IsLoopback() {
			continue
		}

		// 只处理 IPv4 地址
		ipv4 := ipNet.IP.To4()
		if ipv4 == nil {
			continue
		}

		// 优先匹配私有内网网段 (10.x.x.x, 172.16.x.x~172.31.x.x, 192.168.x.x)
		if ipv4.IsPrivate() {
			// 取 IP 第3段的低 2 位 + 第4段的全部 8 位，组成 10 位数字 (0~1023)
			// 可确保同一 /22 子网内最多 1024 个容器节点绝对不冲突
			nodeID := ((int64(ipv4[2]) & 0x03) << 8) | int64(ipv4[3])
			return nodeID % 1024, ipv4.String(), nil
		}
	}

	return 0, "", errors.New("未找到有效的本机内网 IPv4 地址")
}
