package svc

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"

	"oa.98ent.com/p9/operator-activity/ent"
	"oa.98ent.com/p9/operator-activity/ent/migrate"
	_ "oa.98ent.com/p9/operator-activity/ent/runtime"
	"oa.98ent.com/p9/operator-activity/internal/config"
	"oa.98ent.com/p9/operator-activity/modules/activity/plugin"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"
	taskqueue "oa.98ent.com/p9/operator-activity/task/queue"
	"oa.98ent.com/p9/operator-base/rpc/client/operatorservice"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/client/member"
)

// ActivityRuntime 活动引擎运行时。未启动引擎时为空。
type ActivityRuntime interface {
	PipelineFactory(activityType types.ActivityType) plugin.PipelineFactory
}

type ServiceContext struct {
	ServiceName     string
	Config          config.Config
	DB              *ent.Client                     // Ent数据库客户端
	MemberRpc       member.Member                   // 会员管理RPC
	OperatorBaseRpc operatorservice.OperatorService // 分站基础RPC
	RedisQueue      *taskqueue.RedisQueue           // Redis队列
	ActivityRuntime ActivityRuntime                 // 活动引擎，未启用时为 nil
}

func NewServiceContext(c config.Config) *ServiceContext {
	driver, err := c.DatabaseConf.NewDriver()
	logx.Must(err)

	entOpts := []ent.Option{
		ent.Log(logx.Info),
		ent.Driver(driver),
	}
	if c.IsDebug() {
		entOpts = append(entOpts, ent.Debug())
	}

	db := ent.NewClient(entOpts...)
	// 创建业务RPC客户端
	crmClient := zrpc.MustNewClient(c.CrmRpc)
	// 创建分站基础RPC客户端
	baseClient := zrpc.MustNewClient(c.OperatorBaseRpc)

	logx.Must(db.Schema.Create(
		context.Background(),
		migrate.WithDropIndex(true),
		migrate.WithDropColumn(true),
		migrate.WithForeignKeys(false),
	))

	return &ServiceContext{
		Config:          c,
		DB:              db,
		ServiceName:     c.Name,
		MemberRpc:       member.NewMember(crmClient),
		OperatorBaseRpc: operatorservice.NewOperatorService(baseClient),
	}
}
