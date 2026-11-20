package svc

import (
	"oa.98ent.com/p9/core/rpc/coreclient"
	"oa.98ent.com/p9/operator-base/rpc/ent"
	_ "oa.98ent.com/p9/operator-base/rpc/ent/runtime"
	"oa.98ent.com/p9/operator-base/rpc/internal/config"
	operatorservice "oa.98ent.com/p9/operator-base/rpc/internal/operator"
	"oa.98ent.com/p9/platform-base/rpc/client/currencyservice"
	"oa.98ent.com/p9/platform-base/rpc/client/regionservice"
	"oa.98ent.com/p9/platform-base/rpc/client/timezoneservice"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
)

// ServiceContext 服务上下文
type ServiceContext struct {
	Config      config.Config
	DB          *ent.Client                     // Ent数据库客户端
	Operator    *operatorservice.Service        // operator业务服务
	CoreRpc     coreclient.Core                 // 本站语言资料RPC
	RegionRpc   regionservice.RegionService     // 总网国家地区RPC
	CurrencyRpc currencyservice.CurrencyService // 总网货币RPC
	TimezoneRpc timezoneservice.TimezoneService // 总网时区RPC
}

// NewServiceContext 创建服务上下文
func NewServiceContext(c config.Config) *ServiceContext {
	// 创建数据库驱动
	driver, err := c.DatabaseConf.NewDriver()
	logx.Must(err)

	// 创建Ent客户端配置
	entOpts := []ent.Option{
		ent.Log(logx.Info), // 使用go-zero日志输出SQL
		ent.Driver(driver), // 设置数据库驱动
	}

	// 开发和测试环境开启Ent调试模式
	if c.Mode == service.DevMode || c.Mode == service.TestMode {
		entOpts = append(entOpts, ent.Debug())
	}

	// 创建Ent数据库客户端
	db := ent.NewClient(entOpts...)

	// 创建operator业务服务
	operatorService := operatorservice.NewService(db, c)

	// 创建本站 Core RPC 客户端，业务通过 CoreRpc 查询语言资料。
	var coreRpc coreclient.Core
	if c.OperatorCoreRpc != nil {
		operatorCoreConf := *c.OperatorCoreRpc
		operatorCoreConf.NonBlock = true
		coreRpc = coreclient.NewCore(zrpc.MustNewClient(operatorCoreConf))
	}

	// 创建总网基础资料 RPC 客户端，各业务只检查实际使用的依赖。
	var regionRpc regionservice.RegionService
	var currencyRpc currencyservice.CurrencyService
	var timezoneRpc timezoneservice.TimezoneService
	if c.PlatformBaseRpc != nil {
		platformBaseConf := *c.PlatformBaseRpc
		platformBaseConf.NonBlock = true
		platformBase := zrpc.MustNewClient(platformBaseConf)
		regionRpc = regionservice.NewRegionService(platformBase)
		currencyRpc = currencyservice.NewCurrencyService(platformBase)
		timezoneRpc = timezoneservice.NewTimezoneService(platformBase)
	}

	// 返回服务上下文
	return &ServiceContext{
		Config:      c,
		DB:          db,
		Operator:    operatorService,
		CoreRpc:     coreRpc,
		RegionRpc:   regionRpc,
		CurrencyRpc: currencyRpc,
		TimezoneRpc: timezoneRpc,
	}
}
