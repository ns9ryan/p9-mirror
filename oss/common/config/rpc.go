package config

// RPCConfig 本服务对外 gRPC。listen_on 为空则不启动。
type RPCConfig struct {
	ListenOn string `mapstructure:"listen_on"`
}

// OssRPCConfig API 连接档案 RPC。endpoints 空则不连，上传只走 common/oss。
type OssRPCConfig struct {
	Endpoints []string `mapstructure:"endpoints"`
	Timeout   int64    `mapstructure:"timeout"` // 毫秒；≤0 时用 10000
}
