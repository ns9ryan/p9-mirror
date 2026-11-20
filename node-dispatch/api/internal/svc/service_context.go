// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package svc

import (
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"

	"oa.98ent.com/p9/common/i18n"
	"oa.98ent.com/p9/core/common/coreadapt"
	coremiddleware "oa.98ent.com/p9/core/common/middleware"
	"oa.98ent.com/p9/core/rpc/coreclient"
	"oa.98ent.com/p9/node-dispatch/api/internal/config"
	"oa.98ent.com/p9/node-dispatch/api/internal/locales"
	"oa.98ent.com/p9/node-dispatch/rpc/client/dispatchservice"
	"oa.98ent.com/p9/node-dispatch/rpc/client/nodeservice"
	"oa.98ent.com/p9/node-dispatch/rpc/client/pingservice"
)

// ServiceContext 服务上下文
type ServiceContext struct {
	// 服务配置
	Config config.Config // 服务配置

	// Core
	Core coreclient.Core // Core RPC客户端

	// Node Dispatch RPC
	PingRpc     pingservice.PingService         // Ping RPC
	NodeRpc     nodeservice.NodeService         // 节点RPC
	DispatchRpc dispatchservice.DispatchService // 调度任务RPC

	// 多语言
	Trans    *i18n.Translator // API翻译器
	I18nLang rest.Middleware  // Core多语言中间件

	// 认证权限
	Jwt       rest.Middleware // JWT认证中间件
	Authority rest.Middleware // 权限校验中间件

	// 日志
	ActionLog rest.Middleware // 操作日志中间件
	ErrorLog  rest.Middleware // 错误日志中间件
}

// NewServiceContext 创建服务上下文
func NewServiceContext(c config.Config) *ServiceContext {
	// 创建API翻译器
	trans, err := i18n.New(c.I18n, locales.FS)
	logx.Must(err)

	// ============================== Node Dispatch RPC ==============================

	// 创建Node Dispatch RPC客户端
	nodeDispatchClient := zrpc.MustNewClient(c.NodeDispatchRpc)

	// ============================== Core RPC ==============================

	// 创建Core RPC客户端
	coreClient := zrpc.MustNewClient(c.CoreRpc)
	coreCli := coreclient.NewCore(coreClient)

	// 设置Core多语言词典加载器
	coreadapt.SetDictLoader(coreCli)

	// 创建Core认证适配器
	auth := coreadapt.Auth(coreCli)

	jwt := coremiddleware.JWT(auth)
	authority := coremiddleware.Authority(auth)
	if c.IsDev() {
		// 如果是开发环境，跳过权限校验
		jwt = func(next http.HandlerFunc) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				next(w, r)
			}
		}
		authority = func(next http.HandlerFunc) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				next(w, r)
			}
		}
	}

	// ============================== Service Context ==============================

	return &ServiceContext{
		// 服务配置
		Config: c, // 服务配置

		// Core
		Core: coreCli, // Core RPC客户端

		// Node Dispatch RPC
		PingRpc:     pingservice.NewPingService(nodeDispatchClient),         // Ping RPC
		NodeRpc:     nodeservice.NewNodeService(nodeDispatchClient),         // 节点RPC
		DispatchRpc: dispatchservice.NewDispatchService(nodeDispatchClient), // 调度任务RPC

		// 多语言
		Trans:    trans,                                                     // API翻译器
		I18nLang: i18n.NewI18nLangMiddleware(c.I18n.DefaultLanguage).Handle, // Core多语言中间件

		// 认证权限
		Jwt:       jwt,       // JWT认证中间件
		Authority: authority, // 权限校验中间件

		// 日志
		ActionLog: coremiddleware.ActionLog(coreadapt.ActionRecorder(coreCli)),       // 操作日志中间件
		ErrorLog:  coremiddleware.ErrorLog(c.Name, coreadapt.ErrorRecorder(coreCli)), // 错误日志中间件
	}
}
