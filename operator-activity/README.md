# operator-activity

P9 厅侧活动业务服务（纯 RPC）。

负责运营活动、活动规则、活动任务、会员参与记录及活动奖励管理等相关业务。

不负责会员钱包、资金账务、全民代理推广奖励及游戏注单等独立业务。

## 目录

```
operator-activity/
├── proto/        # gRPC proto
├── client/       # zrpc 客户端（供其它服务引用）
├── internal/     # 服务实现
├── pb/           # protoc 生成代码
├── ent/          # Ent Schema 与生成代码
├── pkg/database/ # PostgreSQL / Ent 驱动
└── etc/          # 配置（默认 ListenOn 39006，Etcd Key: operator-activity-rpc）
```

## 生成代码

```bash
make rpc              # 根据 proto 生成 RPC
make ent-new name=Foo # 新建 Ent Schema
make ent              # 生成 Ent 代码
make gen              # rpc + ent
```

## 本地启动

```bash
go run . -f etc/operator_activity.yaml
```

## 活动引擎

详细流程（Kafka → 插件 → Engine 流水线 → 审核/派发 / Temporal）见：

- [docs/activity-engine.md](docs/activity-engine.md)
