package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"oa.98ent.com/p9/oss/api/internal/catalog"
	"oa.98ent.com/p9/oss/api/internal/config"
	"oa.98ent.com/p9/oss/api/internal/handler"
	"oa.98ent.com/p9/oss/api/internal/middleware"
	"oa.98ent.com/p9/oss/api/internal/svc"
	shared "oa.98ent.com/p9/oss/common/config"

	"oa.98ent.com/p9/common/i18n"
	"oa.98ent.com/p9/common/response"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "api/etc/oss.yaml", "HTTP 配置；空或已设 ETCD_CONFIG_KEY 则读 etcd")

func main() {
	flag.Parse()
	path := strings.TrimSpace(*configFile)
	if os.Getenv("ETCD_CONFIG_KEY") != "" || path == "" {
		path = ""
	}
	proc.AddShutdownListener(shared.Close)

	c := config.MustLoad(path)
	server := rest.MustNewServer(c.RestConf, rest.WithCors(c.CROSConf.Address))
	defer server.Stop()
	// 国际化多语言中间件
	server.Use(i18n.NewI18nLangMiddleware(c.App.I18n.DefaultLanguage).Handle)
	// 客户端IP中间件
	server.Use(middleware.ClientIP)
	// 创建服务上下文
	ctx := svc.NewServiceContext(c)
	// 设置HTTP响应头
	response.SetupHTTPX(ctx.Trans, i18n.CodePlatform, c.IsDebug())
	// 注册Catalog
	if ctx.Core != nil {
		logx.Must(catalog.Register(ctx.Core))
	}
	handler.RegisterHandlers(server, ctx)
	fmt.Printf("Starting oss.api at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
