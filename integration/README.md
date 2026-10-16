# integration

P9 第三方接入：短信、邮件、客服、IP/GEO。当前已落地 IP/GEO。

## 文档


| 文件                                         | 说明            |
| ------------------------------------------ | ------------- |
| [docs/integration.md](docs/integration.md) | IP/GEO        |
| [docs/API.md](docs/API.md)                 | HTTP 参数与返回值   |
| [docs/RPC.md](docs/RPC.md)                 | 其它服务 gRPC     |
| [docs/OPS.md](docs/OPS.md)                 | 配置、本地启动、排错    |
| [docs/swagger.json](docs/swagger.json)     | OpenAPI（json） |
| [docs/swagger.yaml](docs/swagger.yaml)     | OpenAPI（yaml） |


## 目录

```text
integration/
├── README.md
├── Makefile                       # gen / build-linux / docker（对齐 core）
├── Dockerfile-api
├── Dockerfile-rpc
├── deploy/                        # compose 部署（对齐 core/deploy）
│   ├── docker-compose.yml
│   ├── .env
│   └── etc/{integration-api,integration-rpc}.yaml
├── docs/
├── api/                           # HTTP。仓库根：go run ./api
│   ├── integration.go
│   ├── desc/                      # main.api + base.api + ipgeo.api
│   ├── etc/integration.yaml
│   └── internal/{config,convert,handler,logic,svc,types}
└── rpc/                           # gRPC。仓库根：go run ./rpc
    ├── integration.go
    ├── integration.proto          # goctls 由 desc 合并生成（勿手改业务）
    ├── desc/*.proto               # 按资源拆分，改这里
    ├── types/integration/         # pb 生成物
    ├── integrationclient/         # zrpc 客户端
    ├── etc/integration.yaml
    ├── ent/
    ├── service/                   # 业务层（按域：ipgeo/…）
    └── internal/{config,db,ipgeo,logic,server,svc}
```

