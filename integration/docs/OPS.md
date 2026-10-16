# integration 运维与部署

HTTP 见 [API.md](API.md)，其它服务 gRPC 见 [RPC.md](RPC.md)。总览见 [integration.md](integration.md)。仓库索引见 [../README.md](../README.md)。

---

## 本地开发

- Go 1.22+
- PostgreSQL（`project-resources/docker/compose.yaml`：`root` / `root`，端口 `5432`）
- 本地直连，**不需要 etcd**

```powershell
cd integration
go run ./rpc
go run ./api
```

在 `rpc/` 或 `api/` 目录内启动时加 `-f etc/integration.yaml`。


| 地址                                                                                                                 | 说明                                                          |
| ------------------------------------------------------------------------------------------------------------------ | ----------------------------------------------------------- |
| [http://127.0.0.1:18005/p9/integration/ipgeo/ip/search](http://127.0.0.1:18005/p9/integration/ipgeo/ip/search)     | 查 IP                                                        |
| [http://127.0.0.1:18005/p9/integration/ipgeo/policy/list](http://127.0.0.1:18005/p9/integration/ipgeo/policy/list) | 当前租户策略                                                      |
| `127.0.0.1:19005`                                                                                                  | gRPC `Integration`                                          |
| `127.0.0.1:5432`                                                                                                   | PostgreSQL（库名 `integration`，不存在则 RPC 启动时 `CREATE DATABASE`） |


配置：`api/etc/integration.yaml`、`rpc/etc/integration.yaml`。

离线库路径相对 **RPC 工作目录**。在 `integration` 目录执行 `go run ./rpc` 时，yaml 用 `assets/GeoLite2-City.mmdb`；请把文件放到 `integration/assets/`。

---



## Docker 部署

`Dockerfile-*` + `deploy/docker-compose.yml`。

### 构建镜像

在 `integration/` 根目录（需能访问内网 Go module）：

```bash
make docker          # linux amd64 二进制 + 打镜像
# make publish-docker  # 推到 DOCKER_REPO
```

镜像名：`${DOCKER_REPO}/integration-rpc:${VERSION}`、`integration-api`。默认仓库见 `Makefile` / `deploy/.env`。

### 启动

```bash
cd deploy
docker compose --env-file .env up -d
```


| 端口      | 服务                                                 |
| ------- | -------------------------------------------------- |
| `18005` | HTTP API                                           |
| `19005` | gRPC RPC                                           |
| `15433` | PostgreSQL（库名 `integration`，避免与 core 的 `15432` 冲突） |


配置挂载：

- `deploy/etc/integration-rpc.yaml` → 容器 `etc/integration.yaml`
- `deploy/etc/integration-api.yaml` → 容器 `etc/integration.yaml`

API 的 `CoreRpc` 默认 `host.docker.internal:19000`（宿主机上的 core-rpc）。若 core 也在 Docker 且共用网络，可改成 `core-rpc:19000`，并把两个 compose 的 `p9-network` 设为同一 external network。

离线 mmdb/xdb：在 `docker-compose.yml` 取消注释 `../assets:/app/assets`，并在 rpc yaml 里 `IsEnabled: true`。

---



## 配置

HTTP 用 `IntegrationRpc.Endpoints`（本地 `127.0.0.1:19005`）和 `CoreRpc.Endpoints`**（JWT + Casbin，必填）**。启动时向 core `RegisterCatalog` 注册本服务 API。RPC 用 `DB` + `IpGeo`。

`DB.DSN` **必填**。`DB.Driver` 若填则必须是 `postgres`。库不存在会自动创建；表 `ip_geo_cache`、`ip_geo_provider`、`ip_geo_provider_policy`、`ip_geo_local` 由 Ent `Schema.Create` 建出（与 oss 相同）。

`ip_geo_provider` 为空时，把 yaml `IpGeo.Providers` 写入库作为**全局服务商**（默认 `IsEnabled`/`Priority`/`Tag`）。已有数据不覆盖；服务商无管理接口，改连接配置需清空该表再重启。种子里的 `local` 对应本地 IP 表，数据走 `POST /p9/integration/ipgeo/local/`*。

查询 IP：合法公网按租户策略选出启用源，每个源先查缓存（KV，键 `name|IP`）。默认 `IpGeo.CacheDriver: table` 写 `ip_geo_cache`（`payload` 一整包 JSON）；`memory` 不落库；`redis` 用 `IpGeo.CacheRedis`（`GET`/`SET`/`EXPIRE`，键前缀默认 `ipgeo`）。改策略不用清缓存，停用的源不会再被读到。`POST /policy/list` 按登录 `operator_code` 取策略；`POST /policy/save` 写租户策略；`POST /policy/default/save` 写 Provider 默认启用/优先级。其它服务可在 gRPC metadata `x-operator-code` 上传租户编码，不要放请求体。

`IpGeo.CacheTTL`：过期小时数；查询命中会按该时长续期（表驱动只 UPDATE `expires_at`；Redis 用 `EXPIRE`）；`0` = 不过期。默认 yaml 为 `168`。切 Redis：`CacheDriver: redis`，并保证 `CacheRedis.Host` 可连（与 core 一样默认 `127.0.0.1:6379`）。

生产若用 etcd 服务发现：RPC 配 `Etcd.Hosts` / `Key`，API 把 `IntegrationRpc` 改成同样的 `Etcd`（不要再写 `Endpoints`）。

---



## 排错


| 现象                                      | 处理                                                   |
| --------------------------------------- | ---------------------------------------------------- |
| `open etc/integration.yaml`             | 在仓库根执行 `go run ./rpc` / `go run ./api`               |
| `database "integration" does not exist` | 旧进程；当前版本会自动建库。确认 5432 账号是 yaml 里的 `root`/`root`      |
| `password authentication failed`        | 5432 被其它 Postgres 占用，密码不是 yaml 的                     |
| `dial tcp 127.0.0.1:12379`              | 本地不要配 Etcd；已改为直连                                     |
| `ip geo not ready` / RPC 连不上            | 先起 `go run ./rpc`，API yaml `Endpoints` 对上 `ListenOn` |
| Redis `dial tcp 127.0.0.1:6379`         | `CacheDriver: redis` 时先起 Redis；Host/Pass/DB 与实例一致   |
| mmdb / xdb 打开失败                         | 检查 `endpoint` 相对 RPC 工作目录是否存在                        |
| 查询 `source` 空、地理字段空                     | 无启用源，或三方都失败；看 RPC 日志 `ipgeo source`                  |


