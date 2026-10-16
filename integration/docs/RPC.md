# 其它服务 gRPC

HTTP 管理后台见 [API.md](API.md)。业务进程请直连本服务 RPC，不要只走 HTTP 再绕一圈（可以，但多一跳）。

定义：改 `rpc/desc/*.proto`（`base.proto`、`ipgeo.proto`），goctls 合并为 `rpc/integration.proto`。

| | |
| --- | --- |
| package | `integration` |
| service | `Integration` |
| 客户端 | `oa.98ent.com/p9/integration/rpc/integrationclient` |
| 本机地址 | `127.0.0.1:19005`（`ListenOn`） |

本机先起依赖与 RPC（HTTP `go run ./api` 调不到这套 RPC）：

```powershell
cd integration
go run ./rpc
```

未配 `DB.DSN` 不能启动。本地 yaml 未设 `Mode` 时默认 `pro`，**不开** reflection。grpcurl 请带 proto，或在 `rpc/etc/integration.yaml` 设 `Mode: dev`。

租户：不要传 `operator_code`。取自 ctx claims / incoming metadata `x-operator-code`。`list` 不会是 `null`。查询顺序、缓存、私网行为与 HTTP 相同，见 [API.md](API.md) / [ip_search_pfd.md](ip_search_pfd.md)。

### 方法一览

proto 方法名（grpcurl 用 `integration.Integration/<Name>`，Go client 为 PascalCase）：

| RPC | 请求 | 返回 | 说明 |
| --- | --- | --- | --- |
| `ping` | `Empty` | `Empty` | 健康检查（`base`） |
| `searchIp` | `SearchIpReq` | `SearchIpResp` | 批量查 IP 地理信息 |
| `listPolicies` | `Empty` | `PolicyListResp` | 当前租户策略（未配则用默认） |
| `savePolicies` | `SavePolicyReq` | `PolicyListResp` | 保存当前租户策略覆盖 |
| `saveDefaultPolicies` | `SavePolicyReq` | `PolicyListResp` | 写服务商默认启用/优先级 |
| `listLocalIps` | `LocalIPListReq` | `LocalIPListResp` | 本地 IP 分页列表 |
| `getLocalIp` | `IDReq` | `LocalIP` | 本地 IP 详情 |
| `createLocalIp` | `LocalIP` | `LocalIP` | 新增本地 IP |
| `updateLocalIp` | `LocalIP` | `LocalIP` | 更新本地 IP（必填 `id`） |
| `deleteLocalIp` | `IDReq` | `Empty` | 删除本地 IP |

说明：本地 IP 相关方法名用 `LocalIp` 而非 `LocalIP`，避免 goctl 生成怪异文件名。

### 公共消息

| 消息 | 字段 | 说明 |
| --- | --- | --- |
| `Empty` | （无） | 空请求/空响应 |
| `IDReq` | `id` int64 | 主键 id |

### 失败（gRPC status）

| 情况 | code | msg |
| --- | --- | --- |
| 超过 100 个 IP / 不支持的 lang | `InvalidArgument` | `max ips is 100` / `unsupported lang` |
| 策略列表空 / 超过 100 条 / 重复 id / 非法项 | `InvalidArgument` | `policy list is required` / `max policies is 100` / `duplicate provider_id` / `invalid provider` |
| 服务商不存在 | `NotFound` | `provider not found` |
| 本地 IP 不存在 | `NotFound` | `local ip not found` |
| ip 非法 / 已存在 | `InvalidArgument` | `ip is required` / `invalid ip` / `ip already exists` |
| 查询引擎未就绪 | `Unavailable` | `ip geo not ready` |

---

## SearchIp

合法公网按该租户策略选出启用源（没有对应策略则用 Provider 默认），按 `priority` 从小到大；每个源先读缓存（KV，键 `name|IP`），miss 再查源。无效 IP 仍占一位，只回 `query`。私网 / 回环：`country` 为 `私有IP`，`source` 为 `private`。单 IP 某源失败继续下一条。

### 请求 `SearchIpReq`

| 字段 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- |
| `ips` | 是 | `repeated string` | 待查 IP，最多 100 |
| `lang` | 否 | `string` | 本地化语言，空则 `en`；允许：`en` `zh-CN` `pt-BR` `ru` `fr` `ja` `es` `de` |

```json
{ "ips": ["8.8.8.8", "1.1.1.1"], "lang": "zh-CN" }
```

### 返回 `SearchIpResp`

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `list` | `repeated IpInfo` | 与 `ips` 等长；不会是 `null` |

### `IpInfo`

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `query` | string | 请求中的原始 IP |
| `continent` | string | 大洲名（随 `lang`） |
| `continent_code` | string | 大洲代码，如 `NA` |
| `country` | string | 国家名（随 `lang`） |
| `country_code` | string | 国家代码，如 `US` |
| `province` | string | 省/州 |
| `province_iso_code` | string | 省/州 ISO，如 `CA` |
| `city` | string | 城市 |
| `zip` | string | 邮编（与 `postal` 同源） |
| `lat` | double | 纬度 |
| `lon` | double | 经度 |
| `timezone` | string | 时区 |
| `isp` | string | ISP |
| `org` | string | 组织 |
| `as` | string | ASN 文本 |
| `postal` | string | 邮编 |
| `source` | string | 命中服务商 `name`，或 `private` |

---

## 策略

服务商连接配置来自 yaml 种子，RPC 不管理。每个租户可覆盖 `is_enabled`、`priority`。该租户没有策略行时，用 `ip_geo_provider` 上的默认启用/优先级。

`tag`：`在线`（`driver=http`）、`本地`（`driver=local`）、`离线`（`mmdb`、`ip2region`）。列表按有效 `priority`、`id` 升序，一次返回全部服务商。

### `PolicyView`（列表回显）

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `provider_id` | int64 | 全局服务商 id，保存时回传 |
| `name` | string | 服务商名称 |
| `is_enabled` | bool | 当前租户是否启用；未配策略时为默认 |
| `priority` | int32 | 当前租户优先级，越小越先 |
| `tag` | string | `在线` / `本地` / `离线` |

### `PolicyItem`（保存入参，不含 `name`/`tag`）

| 字段 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- |
| `provider_id` | 是 | int64 | 全局服务商 id |
| `is_enabled` | 是 | bool | 是否启用 |
| `priority` | 是 | int32 | 优先级，越小越先 |

### `SavePolicyReq` / `PolicyListResp`

| 消息 | 字段 | 说明 |
| --- | --- | --- |
| `SavePolicyReq` | `list` `repeated PolicyItem` | 不能为空；最多 100；同一 `provider_id` 不能重复 |
| `PolicyListResp` | `list` `repeated PolicyView`、`total` int64 | `list` 禁止 null |

| RPC | 行为 |
| --- | --- |
| `listPolicies` | 当前租户策略；未配则用默认 |
| `savePolicies` | 租户覆盖：不存在则创建，存在则更新；不改连接配置、不改其它租户 |
| `saveDefaultPolicies` | 写入 Provider 默认；已有租户策略的行不受影响；返回的是**默认**启用/优先级 |

```json
{
  "list": [
    { "provider_id": 1, "is_enabled": true, "priority": 0 },
    { "provider_id": 2, "is_enabled": false, "priority": 10 }
  ]
}
```

---

## 本地 IP

`driver=local` 的数据源。全局一份，不按租户隔离。是否启用、优先级看当前租户策略。表中没有该 IP 时，查询继续走后面的源。增删改会清掉该 IP 的查询缓存。

`page` 从 1；`page_size` 默认 20、最大 100；`page` / `page_size` ≤0 用默认，不报错。

### `LocalIP`

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int64 | 主键；创建时忽略，更新时必填 |
| `ip` | string | IPv4/IPv6，全局唯一；创建必填 |
| `continent` / `continent_code` | string | 大洲名 / 代码 |
| `country` / `country_code` | string | 国家名 / 代码 |
| `province` / `province_iso_code` | string | 省/州及 ISO |
| `city` | string | 城市 |
| `zip` / `postal` | string | 邮编，写入同一列；填其一即可 |
| `lat` / `lon` | double | 经纬度 |
| `timezone` / `isp` / `org` / `as` | string | 其它 |

### `LocalIPListReq` / `LocalIPListResp`

| 消息 | 字段 | 说明 |
| --- | --- | --- |
| `LocalIPListReq` | `page` int32、`page_size` int32、`ip` string | `ip` 模糊筛选，可空 |
| `LocalIPListResp` | `list` `repeated LocalIP`、`total` int64 | `list` 禁止 null |

| RPC | 请求示例 | 返回 |
| --- | --- | --- |
| `listLocalIps` | `{ "page": 1, "page_size": 20, "ip": "8.8" }` | `{ list, total }` |
| `getLocalIp` | `{ "id": 1 }` | `LocalIP` |
| `createLocalIp` | `{ "ip": "8.8.8.8", "country": "中国", "city": "深圳" }` | `LocalIP` |
| `updateLocalIp` | `LocalIP`（必填 `id`） | `LocalIP` |
| `deleteLocalIp` | `{ "id": 1 }` | `Empty` |

---

## Demo（grpcurl）

```powershell
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

PowerShell 会拆坏 JSON 引号，从 stdin 传入：

```powershell
cd integration
'{"ips":["8.8.8.8"],"lang":"zh-CN"}' | grpcurl -plaintext -d '@' `
  -import-path rpc -proto integration.proto `
  127.0.0.1:19005 integration.Integration/SearchIp
```

```powershell
'{"id":1}' | grpcurl -plaintext -d '@' `
  -import-path rpc -proto integration.proto `
  127.0.0.1:19005 integration.Integration/GetLocalIp
```

```powershell
'{}' | grpcurl -plaintext -d '@' `
  -import-path rpc -proto integration.proto `
  127.0.0.1:19005 integration.Integration/ListPolicies
```

---

## Demo（Go）

其它服务在 `go.mod` 里：

```text
require oa.98ent.com/p9/integration v0.0.0
```

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/integration/rpc/integrationclient"

	"github.com/zeromicro/go-zero/zrpc"
)

func main() {
	cli := integrationclient.NewIntegration(zrpc.MustNewClient(zrpc.RpcClientConf{
		Endpoints: []string{"127.0.0.1:19005"},
		NonBlock:  true,
		Timeout:   30000,
	}))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ctx = ctxdata.WithClaims(ctx, &ctxdata.Claims{OperatorCode: "op-xxx"})

	out, err := cli.SearchIp(ctx, &integrationclient.SearchIpReq{
		Ips:  []string{"8.8.8.8"},
		Lang: "zh-CN",
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, item := range out.GetList() {
		fmt.Printf("query=%s country=%s source=%s\n",
			item.GetQuery(), item.GetCountry(), item.GetSource())
	}
}
```
