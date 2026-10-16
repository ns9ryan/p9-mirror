package svc

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/common/entmixin"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/ent/intercept"
	"oa.98ent.com/p9/operator-crm/ent/migrate"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/config"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/dao/agent"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/dao/auth"
)

type ServiceContext struct {
	Config config.Config
	DB     *ent.Client           // Ent数据库客户端
	Redis  redis.UniversalClient // 令牌黑名单
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

	// 附加租户过滤器
	attachTenant(db)

	// 创建数据库表
	logx.Must(db.Schema.Create(context.Background(), migrate.WithDropIndex(true), migrate.WithDropColumn(true)))

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
		Auth: auth.New(db, rdb, auth.Config{
			AccessSecret:  c.AgentAdminJwt.AccessSecret,
			AccessExpire:  c.AgentAdminJwt.AccessExpire,
			RefreshSecret: c.AgentAdminJwt.RefreshSecret,
			RefreshExpire: c.AgentAdminJwt.RefreshExpire,
		}),
		Agent: agent.New(db),
	}
}

// attachTenant 附加租户过滤器
func attachTenant(c *ent.Client) {
	fOperatorCode := intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
		entmixin.FilterOperatorCode(ctx, q)
		return nil
	})
	c.AgentUser.Intercept(fOperatorCode)
	c.AgentLevel.Intercept(fOperatorCode)
	c.AgentConfig.Intercept(fOperatorCode)
}
