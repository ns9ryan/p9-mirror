package svc

import (
	"github.com/redis/go-redis/v9"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"

	"oa.98ent.com/p9/platform-base/rpc/ent"
	_ "oa.98ent.com/p9/platform-base/rpc/ent/runtime"
	"oa.98ent.com/p9/platform-base/rpc/internal/cache/basedata"
	nicknamecache "oa.98ent.com/p9/platform-base/rpc/internal/cache/nickname"
	"oa.98ent.com/p9/platform-base/rpc/internal/config"
)

type ServiceContext struct {
	BaseData  *basedata.Store
	Nicknames *nicknamecache.Store
	Config    config.Config
	DB        *ent.Client // Ent数据库客户端
}

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

	var rdb redis.UniversalClient
	if c.DataRedis.Host != "" {
		rdb = redis.NewClient(&redis.Options{Addr: c.DataRedis.Host, Password: c.DataRedis.Pass, DB: c.DataRedis.DB})
	}
	return &ServiceContext{
		BaseData:  basedata.New(db, rdb, c.ListCacheTTL),
		Nicknames: nicknamecache.New(db, rdb, c.ListCacheTTL),
		Config:    c,
		DB:        db,
	}
}
