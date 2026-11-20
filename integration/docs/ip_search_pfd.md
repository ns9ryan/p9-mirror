# IP 查询流程图

接口：`POST /p9/integration/ipgeo/ip/search`

HTTP 对接见 [API.md](API.md#post-p9integrationipgeoipsearch)。其它服务 gRPC 见 [RPC.md](RPC.md#searchip)。

`/ip/search` **不按请求体选源**。租户来自登录 JWT 的 `operator_code`（其它服务走 gRPC metadata `x-operator-code`），再用该租户的策略决定查哪些源、先查谁。

缓存是 **KV**：键 `name|规范化IP`（例如 `maxmind|8.8.8.8`），值是整包 `Record` JSON。默认 `ip_geo_cache`（列 `ip` + `payload`）；yaml `IpGeo.CacheDriver` 可换成 `memory` 或 `redis`（`CacheRedis`）。同一 IP 不同源各一条。策略换源后去读新源的键，没有再 Fetch 并写回。停用某源就不会再碰它的缓存。

## 选源规则

1. 拉全部服务商（yaml 种子写入的 `ip_geo_provider`）
2. 有该租户在 `ip_geo_provider_policy` 的行，就覆盖 `is_enabled` / `priority`；没有就用服务商默认
3. 丢掉未启用的
4. 按有效 `priority` 从小到大（相同再按 `id`）依次：先读该源缓存（键 `name|IP`），miss 再 Fetch；**第一个有结果的就停**

默认种子一般是：`local`(0) → `maxmind`(1) → `geoip-city`(2) → `ip2region`(3)。改 `/policy/save` 只影响当前租户的启用和顺序，不必清全表缓存。

单条失败不影响整批：其它 IP 仍会 200 返回。

---

## 1. 调用入口

```mermaid
flowchart LR
  FE["后台 / 前端"] -->|"POST /p9/integration/ipgeo/ip/search<br/>JWT + Casbin"| API["integration-api"]
  API -->|"gRPC SearchIp"| RPC["integration-rpc"]
  SVC["其它 Go 服务"] -->|"gRPC SearchIp<br/>metadata x-operator-code"| RPC
  RPC --> LOOKUP["Lookup.SearchList"]
```

HTTP 路径：`Jwt` → `ActionLog` → `Authority` → `SearchIpHandler` → API Logic 转调 RPC。API 不直连业务库。

RPC 路径：`SearchIpLogic` 用 `AccountOperatorCode(ctx)` 取租户，再调 `Lookup.SearchList`。

---

## 2. 批量查询主流程

```mermaid
flowchart TD
  A["入口：ips + lang"] --> B["取 operator_code<br/>JWT / x-operator-code / default"]
  B --> C{"ips 超过 100<br/>或 lang 不支持?"}
  C -->|是| Z["整批失败<br/>HTTP 400 / gRPC InvalidArgument"]
  C -->|否| D["lang 空则 en"]
  D --> E["每个 IP 并行 Search<br/>最多 8 路 errgroup"]
  E --> F["等全部完成"]
  F --> G["按请求顺序组装 list<br/>单条失败不拖垮整批"]
  G --> H["返回 SearchIpResp"]
```

约束：

| 项 | 规则 |
| --- | --- |
| 批量上限 | 最多 100 个 IP，超限整批失败 |
| 并发 | `errgroup` 上限 8 |
| 语言 | 空则 `en`；仅 `en` `zh-CN` `pt-BR` `ru` `fr` `ja` `es` `de` |
| 租户 | 忽略请求体；空则 `default` |
| 回源合并 | 同租户同规范化 IP 走 `singleflight` |

---

## 3. 单 IP 查询

```mermaid
flowchart TD
  A["Search(ip, lang, operator_code)"] --> B{"规范化 IP 合法?"}
  B -->|否| B1["只回 query<br/>地理字段空"]
  B -->|是| C{"私网 / 回环 / 链路本地?"}
  C -->|是| C1["country=私有IP<br/>source=private"]
  C -->|否| D["singleflight<br/>键: operator_code|规范化IP"]
  D --> E["resolve：按租户策略选源"]
  E --> F{"有 Record?"}
  F -->|否| F1["只回 query"]
  F -->|是| G["Record.ToInfo(lang)<br/>按语言挑地名"]
```

---

## 4. 选源 + 缓存 + Fetch

```mermaid
flowchart TD
  A["ListEnabledProviders(operator_code)"] --> B["读 ip_geo_provider 全部服务商"]
  B --> C["读该租户 ip_geo_provider_policy"]
  C --> D{"该服务商有策略行?"}
  D -->|有| E["覆盖 is_enabled / priority"]
  D -->|无| F["用服务商默认 is_enabled / priority"]
  E --> G["丢掉 is_enabled=false"]
  F --> G
  G --> H["按 priority、id 升序"]
  H --> I["sourcesFor：按配置指纹复用已打开的 Source"]

  I --> N["按顺序取下一个源"]
  N --> L{"该源缓存命中?<br/>键: name|IP"}
  L -->|是且未过期| M["Touch 续期 TTL<br/>返回该源 Record"]
  L -->|否 / 过期 / 读失败当 miss| O["src.Fetch(ip, lang)"]
  O --> P{"成功且非空?"}
  P -->|失败或空| Q{"还有下一个源?"}
  Q -->|是| N
  Q -->|否| R["无结果"]
  P -->|成功| S["写入缓存 KV<br/>key=name|IP  value=Record JSON"]
  S --> T["返回该源 Record<br/>第一个有结果即停"]
```

缓存驱动（yaml `IpGeo.CacheDriver`）：

| 驱动 | 存储 |
| --- | --- |
| `table`（默认） | PostgreSQL `ip_geo_cache`，主键即缓存键 |
| `memory` | 进程内 map |
| `redis` | `CacheRedis`，Get 用 PTTL 回填过期 |

`CacheTTL` 单位小时；`0` 表示不过期。命中且配置了 TTL 时 `Touch` 续期。缓存 miss 或读失败都当未命中，继续 Fetch。

---

## 5. 各源 Fetch

```mermaid
flowchart TD
  A["Source.Fetch"] --> B{"Driver?"}
  B -->|local| C["查 ip_geo_local<br/>精确匹配规范化 IP"]
  C --> C1{"有行?"}
  C1 -->|否| C2["空结果，试下一源"]
  C1 -->|是| C3["转 Record"]

  B -->|http| D["展开 Endpoint 模板<br/>{ip} {lang}"]
  D --> D1["鉴权 none/basic/bearer/header"]
  D1 --> D2["GET JSON，FieldMap 用 gjson 填 Record"]
  D2 --> D3{"HTTP 200?"}
  D3 -->|否| D4["记日志，试下一源"]
  D3 -->|是| D5["Record"]

  B -->|mmdb| E["读 GeoLite2 .mmdb"]
  E --> E1["Record"]

  B -->|ip2region| F["读 ip2region .xdb"]
  F --> F1["Record"]
```

默认种子顺序（表空时由 yaml `IpGeo.Providers` 写入，已有数据不覆盖）：

| 优先级 | name | driver | 说明 |
| --- | --- | --- | --- |
| 0 | `local` | `local` | 后台 `/local/*` 维护的本地表 |
| 1 | `maxmind` | `http` | MaxMind Country Web API |
| 2 | `geoip-city` | `mmdb` | 默认未启用 |
| 3 | `ip2region` | `ip2region` | 离线省市库 |

---

## 6. 时序（HTTP 查询）

```mermaid
sequenceDiagram
  participant U as 调用方
  participant API as integration-api
  participant Core as core RPC
  participant RPC as integration-rpc
  participant PG as PostgreSQL
  participant Src as 定位源

  U->>API: POST /ipgeo/ip/search
  API->>Core: CheckToken + Casbin
  API->>RPC: gRPC SearchIp(ips, lang)
  Note over RPC: operator_code 来自 JWT / metadata
  RPC->>PG: 服务商 + 租户策略
  loop 每个启用源，直到命中
    RPC->>PG: 缓存 Get name|IP
    alt 命中且未过期
      RPC->>PG: Touch 续期
    else miss
      RPC->>Src: Fetch
      Src-->>RPC: Record 或失败
      opt 成功
        RPC->>PG: 缓存 Set
      end
    end
  end
  RPC-->>API: list
  API-->>U: {code, msg, data.list}
```

其它服务直连 gRPC 时跳过 API / Casbin，租户走 `x-operator-code`。

---

## 代码位置

| 环节 | 文件 |
| --- | --- |
| HTTP 路由 / 鉴权 | `api/internal/handler/routes.go` |
| HTTP 转 RPC | `api/internal/logic/ipgeo/search_ip_logic.go` |
| RPC 入口 | `rpc/internal/logic/ipgeo/search_ip_logic.go` |
| 批量 / 单 IP / 缓存编排 | `rpc/internal/ipgeo/lookup/lookup.go` |
| 租户策略叠加 | `rpc/service/ipgeo/provider.go`、`overlay.go` |
| 驱动构造 | `rpc/internal/ipgeo/source/build.go` |
| 缓存 KV | `rpc/internal/ipgeo/cache/` |
