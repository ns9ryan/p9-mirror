package config

import (
	"log"
	"strings"

	shared "oa.98ent.com/p9/oss/common/config"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/zrpc"
)

// Config RPC 进程配置。yaml/etcd 仍走 shared.Load（snake_case），再填 RpcServerConf。
type Config struct {
	zrpc.RpcServerConf
	App      *shared.AppConfig
	FromFile bool
}

func MustLoad(path string) Config {
	path = strings.TrimSpace(path)
	return FromApp(shared.Load(path), path != "")
}

func FromApp(app *shared.AppConfig, fromFile bool) Config {
	var c zrpc.RpcServerConf
	if err := conf.FillDefault(&c); err != nil {
		log.Fatalf("rpc conf: %v", err)
	}
	c.Name = "oss.rpc"
	c.ListenOn = strings.TrimSpace(app.RPC.ListenOn)
	c.Mode = "dev"
	c.Timeout = 300000
	c.DevServer.Enabled = false
	c.Log.ServiceName = "oss"
	c.Log.Mode = "console"
	c.Log.Encoding = "plain"
	c.Log.Level = "info"
	c.Log.Stat = false
	return Config{RpcServerConf: c, App: app, FromFile: fromFile}
}
