package config

// CoreRPCConfig 连接 p9 core RPC，供 JWT CheckToken。endpoints 为空则不启用 JWT。
type CoreRPCConfig struct {
	Endpoints []string `mapstructure:"endpoints"`
	Timeout   int64    `mapstructure:"timeout"` // 毫秒；≤0 时中间件用 10000
}
