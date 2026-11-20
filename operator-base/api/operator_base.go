// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package main

import (
	"flag"
	"fmt"

	"oa.98ent.com/p9/operator-base/api/internal/catalog"
	"oa.98ent.com/p9/operator-base/api/internal/config"
	"oa.98ent.com/p9/operator-base/api/internal/handler"
	"oa.98ent.com/p9/operator-base/api/internal/httpresponse"
	"oa.98ent.com/p9/operator-base/api/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/operator_base.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	// 本站 Core 是启动必需依赖；未配置时直接报错，不接受 HTTP 请求。
	if ctx.Core == nil {
		logx.Must(fmt.Errorf("CoreRpc 未配置，无法启动 operator-base API"))
	}
	// 调用本站 Core 登记接口目录，沿用现有重试；登记成功后才启动 HTTP，失败则退出。
	logx.Must(catalog.Register(ctx.Core))
	httpresponse.SetupHTTPX()
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
