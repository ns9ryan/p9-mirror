# 总站会员昵称管理 API

## 业务规则

总站 `platform-base` 保存共用昵称库，支持新增、修改、删除、详情和分页列表。昵称去除首尾空白后不能为空，最多 100 个 Unicode 字符；昵称库中的名称唯一（按数据库原值区分大小写）。删除为物理删除，不影响已创建会员的昵称。

`operator-crm` 新增会员时，已填写的昵称去除首尾空白后直接使用；未填、空串或全空白时调用 `NicknameService.Random`，从库中等概率选一条并保存到会员资料。不同会员允许使用同一昵称，不消耗昵称库记录。库为空时 RPC 返回 `NotFound`，会员沿用账号兜底；RPC 不可用、未配置或数据库异常时创建失败，不把故障当成空库。

## HTTP 约定

服务为 `platform-base-api`，直连服务无需添加其他网关前缀。所有接口使用现有 `Jwt,ActionLog,Authority` 中间件，需总站后台 Token 和对应 API 权限。写操作的 RPC 额外校验有效总站后台身份，拒绝分站身份和预览令牌。

请求头：

```http
Authorization: Bearer <access_token>
Content-Type: application/json
```

GET 参数放在 URL 查询字符串，POST 参数放在 JSON Body。响应使用现有统一包装：

```json
{"code":0,"data":{},"msg":"ok"}
```

下表及示例中的响应表示 `data` 部分。错误文案随语言变化，不应依赖固定文案。

| 方法 | 路径 | 用途 |
| --- | --- | --- |
| POST | `/admin/nickname/create` | 新增 |
| POST | `/admin/nickname/update` | 修改 |
| POST | `/admin/nickname/delete` | 删除 |
| GET | `/admin/nickname/get` | 详情 |
| GET | `/admin/nickname/list` | 分页列表 |

### 新增

`POST /admin/nickname/create`

```json
{"name":"快乐小熊"}
```

`name` 必填，去除首尾空白后 1～100 个字符。成功响应：

```json
{"nickname":{"id":1,"name":"快乐小熊"}}
```

### 修改

`POST /admin/nickname/update`

```json
{"id":1,"name":"阳光小熊"}
```

`id` 为已有记录的正整数主键；`name` 规则与新增相同。成功响应 `{}`，不存在的主键返回 NotFound，不会新增记录。

### 删除

`POST /admin/nickname/delete`

```json
{"id":1}
```

成功响应 `{}`。不存在的主键返回 NotFound。允许删除最后一条；此后新会员未指定昵称时使用账号兜底。

### 详情

`GET /admin/nickname/get?id=1`

成功响应：

```json
{"nickname":{"id":1,"name":"阳光小熊"}}
```

`id` 必须为正整数，不存在返回 NotFound。

### 分页列表

`GET /admin/nickname/list?page=1&page_size=20&keyword=小熊`

| 参数 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- |
| page | int64 | 1 | 页码，至少 1；与 page_size 计算的偏移量须在支持范围内 |
| page_size | int64 | 20 | 每页 1～100 条 |
| keyword | string | 空串 | 最多 100 字符，去除首尾空白后按名称不区分大小写包含查询 |

按 ID 升序返回，空列表为 `[]`：

```json
{"total":1,"list":[{"id":1,"name":"阳光小熊"}]}
```

管理列表与详情直接查数据库，可立即查看保存结果。

## RPC

定义：`platform-base/rpc/proto/types/platform_base_nickname.proto`；服务：`platform_base.NicknameService`。

| 方法 | 请求 | 响应 |
| --- | --- | --- |
| Create | name | nickname: {id, name} |
| Update | id, name | 空消息 |
| Delete | id | 空消息 |
| Get | id | nickname: {id, name} |
| List | page, page_size, keyword | total, list |
| Random | 空消息 | nickname: {id, name} |

直接调用 List RPC 必须显式传合法分页参数，默认值由 HTTP 层提供。Random 仅内部使用，无管理端 HTTP 路由，不进行额外写入。

参数错误及名称重复为 `InvalidArgument`；记录不存在或随机池为空为 `NotFound`；数据库故障为 `Internal`；写入身份不符为 `PermissionDenied`。HTTP 由现有 RPC 错误转换器处理。

## Redis 缓存设计

参照提供的 `default.go`、`sentinel_cache.go`，采用 MessagePack 序列化和 cache-aside。Redis 初始化沿用 core 的 `redis.UniversalClient`、`DataRedis.Host/Pass/DB` 模式；`RotateSalt` 所在的 core service 是查阅入口。

- 只缓存随机抽取使用的完整小型昵称池，管理分页与详情直接读数据库。
- `platform-base:nickname:revision` 保存版本号，初始缺失时视为 0；版本键不设置过期时间。
- 数据键为 `platform-base:nickname:pool:<版本号>`，内容为 `{ID, Name}` 数组的 MessagePack 二进制，正缓存 TTL 为 1 小时加 0～299 秒随机抖动。
- 空库使用 `NF` 哨兵，TTL 为 30～39 秒；数据库查询失败不写空库哨兵。
- 命中时直接从数组等概率随机选择，不每次查询数据库，也不缓存上次抽中的单个昵称。
- 未命中使用进程内 singleflight 合并同版本的回源；共享加载有独立的 5 秒超时，等待者可按自身上下文取消。
- 新增、修改、删除成功后增加版本号。并发旧加载只写旧版本的数据键，后续请求读取新版本；旧键按 TTL 自动回收。
- 版本读取失败、Redis 故障或缓存损坏时回源数据库并记录日志；损坏内容会删除。
- 修改后的失效操作使用独立的 3 秒超时，不因 HTTP 客户端断开而取消。失效失败会记录日志，数据库保存结果仍成功；此时旧缓存最多保留其剩余 TTL（正缓存上限约 65 分钟）。Redis 与数据库不提供跨系统原子事务，监控时应关注失效失败日志。

## 部署与联调

1. 先部署 platform-base RPC。启动时通过现有 `MustMigrate` / Ent Schema 创建 `nickname` 表，包含 id、name（唯一）、created_at、updated_at；不自动填充业务昵称。
2. 在 RPC 实际使用的配置文件中设置 DataRedis。未配置 Host 时允许无缓存运行：

```yaml
DataRedis:
  Host: 127.0.0.1:16379
  Pass: ""
  DB: 0
```

3. 部署 platform-base API；启动时按现有机制登记昵称页面菜单、增删改按钮权限及五个 HTTP API 目录和翻译。菜单路径 `/base-data/nickname`，组件 `base-data/nickname/index`，前端页面需要由前端对接实现。
4. 部署更新后的 operator-crm；沿用已有 PlatformBaseRpc 连接配置，未填写昵称时依赖新增 Random RPC。所有 platform-base RPC 实例需使用同一 Redis 数据库，以共享缓存版本。
5. 依次新增、查询、修改、删除昵称并验证列表；新增会员分别验证自填昵称、未填随机选择和昵称池为空时账号兜底。

本次未向任何实际环境写入昵称或执行数据库迁移。变更涉及 platform-base、operator-crm、project-resources 三个独立仓库，发布时需一并纳入。
