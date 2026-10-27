package svc

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/ent/migrate"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/config"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/dao"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/dao/agent"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/dao/auth"
)

type ServiceContext struct {
	Config config.Config
	DB     *ent.Client           // Ent数据库客户端
	Redis  redis.UniversalClient // 令牌黑名单
	DAO    *dao.Group            // 数据访问
	Auth   *auth.Service         // 代理鉴权
	Agent  *agent.Service        // 代理管理
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

	// 调试环境开启Ent调试模式
	if c.IsDebug() {
		entOpts = append(entOpts, ent.Debug())
	}

	// 创建Ent数据库客户端
	db := ent.NewClient(entOpts...)

	// 创建数据库表
	logx.Must(db.Schema.Create(
		context.Background(),
		migrate.WithDropIndex(true),    // 允许删除旧索引
		migrate.WithDropColumn(true),   // 允许删除旧字段
		migrate.WithForeignKeys(false), // 不创建数据库外键约束
	))

	var rdb redis.UniversalClient
	if c.DataRedis.Host != "" {
		rdb = redis.NewClient(&redis.Options{
			Addr:     c.DataRedis.Host,
			Password: c.DataRedis.Pass,
			DB:       c.DataRedis.DB,
		})
	}

	return &ServiceContext{
		Config: c,
		DB:     db,
		Redis:  rdb,
		DAO:    dao.New(db),
		Auth: auth.New(db, rdb, auth.Config{
			AccessSecret:  c.AgentAdminJwt.AccessSecret,
			AccessExpire:  c.AgentAdminJwt.AccessExpire,
			RefreshSecret: c.AgentAdminJwt.RefreshSecret,
			RefreshExpire: c.AgentAdminJwt.RefreshExpire,
		}),
		Agent: agent.New(db),
	}
}
