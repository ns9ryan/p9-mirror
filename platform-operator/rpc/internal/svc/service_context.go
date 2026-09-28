package svc

import (
	nodedispatchoperatornodeservice "oa.98ent.com/p9/node-dispatch/rpc/client/operatornodeservice"
	"oa.98ent.com/p9/platform-operator/rpc/ent"
	_ "oa.98ent.com/p9/platform-operator/rpc/ent/runtime"
	"oa.98ent.com/p9/platform-operator/rpc/internal/config"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config
	DB     *ent.Client // Ent数据库客户端

	// Node Dispatch RPC
	NodeDispatchOperatorNodeRpc nodedispatchoperatornodeservice.OperatorNodeService // 分站部署节点RPC
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

	// ============================== Node Dispatch RPC ==============================

	// 创建Node Dispatch RPC客户端
	nodeDispatchClient := zrpc.MustNewClient(c.NodeDispatchRpc)

	// ============================== Service Context ==============================

	return &ServiceContext{
		Config: c,
		DB:     db,

		// Node Dispatch RPC
		NodeDispatchOperatorNodeRpc: nodedispatchoperatornodeservice.NewOperatorNodeService(nodeDispatchClient), // 分站部署节点RPC
	}
}
