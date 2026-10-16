// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package main

import (
	"flag"
	"fmt"

	"oa.98ent.com/p9/agent-admin-api/internal/catalog"
	"oa.98ent.com/p9/agent-admin-api/internal/config"
	"oa.98ent.com/p9/agent-admin-api/internal/handler"
	"oa.98ent.com/p9/agent-admin-api/internal/svc"
	"oa.98ent.com/p9/common/i18n"
	"oa.98ent.com/p9/common/response"
	"oa.98ent.com/p9/common/validate"
	"oa.98ent.com/p9/core/common/middleware"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

var configFile = flag.String("f", "etc/agent_admin_api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 创建参数校验器
	v, err := validate.New(c.I18n.DefaultLanguage)
	logx.Must(err)
	// 注册全局参数校验器
	httpx.SetValidator(v)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()
	// 创建服务上下文
	ctx := svc.NewServiceContext(c)

	// 设置HTTP响应格式
	response.SetupHTTPX(ctx.Trans, i18n.CodePlatform, c.IsDebug())

	// 注册I18n语言中间件
	server.Use(i18n.NewI18nLangMiddleware(c.I18n.DefaultLanguage).Handle)
	// 注册客户端 IP 中间件
	server.Use(middleware.ClientIP)

	// 注册菜单、API目录、多语言数据
	logx.Must(catalog.Register(ctx))

	// 注册路由
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
