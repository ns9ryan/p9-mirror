package svc

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/operator-wallet/ent"
	"oa.98ent.com/p9/operator-wallet/ent/migrate"
	_ "oa.98ent.com/p9/operator-wallet/ent/runtime"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/config"
)

// ServiceContext 会员钱包RPC服务上下文
type ServiceContext struct {
	Config config.Config
	DB     *ent.Client // Ent数据库客户端
}

// NewServiceContext 创建会员钱包RPC服务上下文
func NewServiceContext(c config.Config) *ServiceContext {
	// 创建数据库驱动
	driver, err := c.DatabaseConf.NewDriver()
	logx.Must(err)

	// 创建Ent客户端配置
	entOpts := []ent.Option{
		ent.Log(logx.Info), // 使用go-zero日志输出SQL
		ent.Driver(driver), // 设置数据库驱动
	}

	// 开发和测试环境开启SQL调试
	if c.IsDebug() {
		entOpts = append(entOpts, ent.Debug())
	}

	// 创建Ent数据库客户端
	db := ent.NewClient(entOpts...)

	// 启动时同步数据库结构
	logx.Must(db.Schema.Create(
		context.Background(),
		migrate.WithDropColumn(false),  // 禁止自动删除字段
		migrate.WithDropIndex(false),   // 禁止自动删除索引
		migrate.WithForeignKeys(false), // 不创建数据库外键约束
	))

	return &ServiceContext{
		Config: c,
		DB:     db,
	}
}
