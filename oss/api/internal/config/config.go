package config

import (
	"log"
	"strconv"
	"strings"

	shared "oa.98ent.com/p9/oss/common/config"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

const defaultOssRPCTimeoutMS int64 = 10000

// Config HTTP 进程配置。yaml/etcd 仍走 shared.Load（snake_case），再填 RestConf。
type Config struct {
	rest.RestConf
	CROSConf struct {
		Address string
	}
	CoreRpc zrpc.RpcClientConf
	OssRpc  zrpc.RpcClientConf
	Upload  shared.UploadSettings
	App     *shared.AppConfig
}

// IsDebug 是否为调试模式
func (c *Config) IsDebug() bool {
	return c.Mode == service.DevMode || c.Mode == service.TestMode
}

func MustLoad(path string) Config {
	return FromApp(shared.Load(strings.TrimSpace(path)))
}

func FromApp(app *shared.AppConfig) Config {
	settings := app.Settings()
	c := Config{
		RestConf: restConf(settings),
		CoreRpc:  zrpc.RpcClientConf{Endpoints: app.CoreRPC.Endpoints, NonBlock: true, Timeout: app.CoreRPC.Timeout},
		OssRpc:   zrpc.RpcClientConf{Endpoints: app.OssRPC.Endpoints, NonBlock: true, Timeout: ossTimeout(app)},
		Upload:   settings.Upload,
		App:      app,
	}
	c.CROSConf.Address = corsAddress(app)
	return c
}

func corsAddress(app *shared.AppConfig) string {
	if app != nil {
		if a := strings.TrimSpace(app.CROSConf.Address); a != "" {
			return a
		}
	}
	return "*"
}

func ossTimeout(app *shared.AppConfig) int64 {
	if app.OssRPC.Timeout > 0 {
		return app.OssRPC.Timeout
	}
	return defaultOssRPCTimeoutMS
}

func restConf(settings shared.Settings) rest.RestConf {
	port, err := strconv.Atoi(strings.TrimSpace(settings.Port))
	if err != nil || port <= 0 {
		port = 80
	}
	maxBytes := settings.Upload.PresignMaxSize + shared.MaxRequestOverhead
	if n := settings.Upload.MaxSize + shared.MaxRequestOverhead; n > maxBytes {
		maxBytes = n
	}
	var c rest.RestConf
	if err := conf.FillDefault(&c); err != nil {
		log.Fatalf("rest conf: %v", err)
	}
	c.Host = "0.0.0.0"
	c.Port = port
	c.Timeout = 300000
	c.MaxBytes = maxBytes
	c.MaxConns = 10000
	c.Name = "oss.api"
	c.Mode = "dev"
	c.DevServer.Enabled = false
	c.Log.ServiceName = "oss"
	c.Log.Mode = "console"
	c.Log.Encoding = "plain"
	c.Log.Level = "info"
	c.Log.Stat = false
	return c
}
