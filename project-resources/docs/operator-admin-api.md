# operator-admin-api 说明文档

## 一、HTTP API

`operator-admin-api` 为分站后台 CRM HTTP 网关，转发 `operator-crm` 的 `AgentService` 等业务 RPC。  
默认开发地址 `http://127.0.0.1:38010`，**无**统一 `/admin` 前缀。JSON 字段以 [`operator-admin-api/desc`](../../operator-admin-api/desc) 为准。

登录鉴权、Token 刷新、过期处理见同目录 [core-api.md](./core-api.md)。本服务 JWT / 权限由分站 Core 校验；请求上下文中的 `operator_code` 会透传到 CRM。

## 接口索引

- [约定](#约定)
- [公开（无 JWT）](#公开无-jwt)
  - [GET /ping](#get-ping)
- [受保护（Jwt + ActionLog + Authority）](#受保护jwt--actionlog--authority)
  - [代理列表 / 开户](#代理列表--开户)
    - [POST /agent/list](#post-agentlist)
    - [POST /agent/create](#post-agentcreate)
    - [POST /agent/search](#post-agentsearch)
    - [GET /agent/operator/list](#get-agentoperatorlist)
  - [代理等级](#代理等级)
    - [GET /agent/level/list](#get-agentlevellist)
    - [POST /agent/level/update](#post-agentlevelupdate)
  - [代理配置](#代理配置)
    - [GET /agent/config](#get-agentconfig)
    - [POST /agent/config/update](#post-agentconfigupdate)
  - [代理树](#代理树)
    - [GET /agent/tree/node](#get-agenttreenode)

---

## 约定

### 响应结构体

成功：

```json
{ "code": 0, "msg": "ok", "data": {} }
```

失败时 HTTP 状态码与 `code` 相同，无 `data`：

```json
{ "code": 401, "msg": "未登录" }
```

access 过期（前端用 refresh 后续请求）：

```json
{ "code": 498, "msg": "登录已过期" }
```

时间字段为 Unix 秒。下文示例均为完整 HTTP body 中的 `data` 形态说明时，仍给出完整信封。

### 鉴权

除 `GET /ping` 外，受保护接口中间件为 **Jwt + ActionLog + Authority**：

```http
Authorization: Bearer <access_token>
```

- **Jwt**：分站 Core `CheckToken`
- **ActionLog**：记录操作日志
- **Authority**：按 method + path 调 Core `Enforce`（Casbin）；依赖 Catalog 已注册的 `/agent/*` path，超管角色启动时会补授权

Token 获取见 [core-api.md](./core-api.md)。Dev 模式下 Jwt / Authority 透传，便于本地联调。

### 多语言

请求语言通过 Header 指定：

```http
X-Lang: zh-CN
```

| 语言编码 | 说明 |
| --- | --- |
| `zh-CN` | 简体中文 |
| `zh-HK` | 繁体中文 |
| `en-US` | 英文 |

默认 `zh-CN`。菜单 / API 描述词条由启动时 `RegisterCatalog` 写入 Core（i18n code=`operator`）。

### 公共类型

`AgentLevelItem`：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int64 | 等级主键；创建时可省略 |
| `level` | int32 | 等级 1~7 |
| `name` | string | 等级名称 |
| `alias` | string | 等级别名 |

`AgentConfig`：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int64 | 配置主键；无配置时为 `0` |
| `max_share_rate` | int32 | 最高占成，百分比 0~100 |
| `apply_min_share_rate` | int32 | 申请最低占成 |
| `apply_max_share_rate` | int32 | 申请最高占成；须 ≥ min 且 ≤ max |
| `cash_max_member_count` | int64 | 现金代理最大会员数 |
| `cash_max_agent_per_ip` | int64 | 现金 IP 最大可注册代理数 |
| `cash_production_agent_id` | int64 | 现金正式线指定代理，`0` 表示清空 |
| `cash_demo_agent_id` | int64 | 现金试玩线指定代理 |
| `cash_test_agent_id` | int64 | 现金测试线指定代理 |
| `credit_demo_agent_id` | int64 | 信誉试玩线指定代理 |
| `credit_test_agent_id` | int64 | 信誉测试线指定代理 |

`AgentCreateFields`：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `account` | string | 登录账号 |
| `code` | string | 代理编码 |
| `password` | string | 登录密码明文 |
| `agent_type` | int32 | `1` 现金代理，`2` 信用代理 |
| `agent_line_code` | string | 子线路编码，如 `CASH_PRODUCTION` |
| `share_rate` | int32 | 占成，百分比 0~100 |
| `promo_code` | string | 对外推广码 |
| `register_ip` | string | 注册 IP，可选 |

`AddAgentResp`：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int64 | 代理 ID |
| `account` | string | 账号 |
| `code` | string | 编码 |
| `parent_id` | int64 | 上级 ID |
| `top_id` | int64 | 顶层 ID；顶层可为 `0` |
| `level` | int32 | 层级 |
| `share_rate` | int32 | 占成 |

`AgentSearchItem`：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int64 | 代理 ID |
| `account` | string | 账号 |
| `level` | int32 | 层级 |
| `level_alias` | string | 等级别名 |
| `line_code` | string | `agent_line_code` |

`AgentListItem`：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int64 | 代理 ID |
| `account` | string | 账号 |
| `level` | int32 | 层级 |
| `level_alias` | string | 等级别名 |
| `parent_id` | int64 | 上级 ID |
| `parent_account` | string | 上级账号 |
| `top_id` | int64 | 顶层 ID |
| `top_account` | string | 顶层账号 |
| `register_source` | int32 | 注册来源：`1` PC，`2` app，`3` 管理员，`4` 代理 |
| `direct_member_count` | int64 | 直属会员数 |
| `direct_agent_count` | int64 | 直属代理数 |
| `total_member_count` | int64 | 全部会员数 |
| `total_agent_count` | int64 | 全部下级代理数 |
| `share_rate` | int32 | 占成 |
| `wallet_balance` | int64 | 钱包余额（当前固定 `0`） |
| `total_commission` | int64 | 累计佣金 |
| `unpaid_commission` | int64 | 未领取佣金 |
| `status` | int32 | `1` 正常，`2` 冻结，`3` 停用 |
| `payout_restricted` | int32 | `1` 不限制，`2` 限制本人，`3` 限制全线 |
| `remark` | string | 备注 |
| `promo_link` | string | 推广链接（暂用 `promo_code`） |
| `visit_uv` | int64 | 访问 UV |
| `register_at` | int64 | 注册时间 Unix 秒 |
| `last_login_at` | int64 | 最后登录 Unix 秒 |

`TreeNodeAgentItem`：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int64 | 代理 ID |
| `account` | string | 账号 |
| `line_code` | string | 线路编码 |
| `level` | int32 | 层级 |
| `level_alias` | string | 等级别名 |
| `direct_agent_count` | int64 | 直属代理数 |

---

## 公开（无 JWT）

### GET /ping

健康检查。

**响应**

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "message": "pong"
  }
}
```

---

## 受保护（Jwt + ActionLog + Authority）

需 Header：`Authorization: Bearer <access_token>`，且角色需具备对应 API 权限。

## 代理列表 / 开户

### POST /agent/list

分页代理列表，支持筛选。对应 RPC `listAgents`。

**请求**

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `page` | json | 否 | int32 | 页码，从 1；默认按 CRM |
| `page_size` | json | 否 | int32 | 每页条数，默认 20，上限 100 |
| `register_source` | json | 否 | int32 | 注册来源，`0` 不限 |
| `level` | json | 否 | int32 | 层级，`0` 不限 |
| `status` | json | 否 | int32 | 状态，`0` 不限 |
| `payout_restricted` | json | 否 | int32 | 出款限制，`0` 不限 |
| `time_filter_type` | json | 否 | int32 | `1` 注册时间，`2` 最后登录，`0` 不筛 |
| `time_start` / `time_end` | json | 否 | int64 | Unix 秒；成对有效 |
| `number_filter_type` | json | 否 | int32 | `1`~`7` 数值筛选项，`0` 不筛 |
| `number_min` / `number_max` | json | 否 | int64 | `>0` 才生效 |
| `agent_filter_type` | json | 否 | int32 | `1` 本人，`2` 上级，`0` 不筛 |
| `agent_keyword` | json | 否 | string | 纯数字按 ID，否则账号模糊 |

`number_filter_type`：`1` 直属会员，`2` 直属代理，`3` 全部下级代理，`4` 全部会员，`5` 占成，`6` 累计佣金，`7` 已领取佣金（累计−未领）。

```json
{
  "page": 1,
  "page_size": 20,
  "level": 2,
  "status": 1
}
```

**响应**

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "items": [
      {
        "id": 1002,
        "account": "agent01",
        "level": 2,
        "level_alias": "总代",
        "parent_id": 1001,
        "parent_account": "root",
        "top_id": 1001,
        "top_account": "root",
        "register_source": 3,
        "direct_member_count": 10,
        "direct_agent_count": 2,
        "total_member_count": 100,
        "total_agent_count": 5,
        "share_rate": 50,
        "wallet_balance": 0,
        "total_commission": 0,
        "unpaid_commission": 0,
        "status": 1,
        "payout_restricted": 1,
        "remark": "",
        "promo_link": "ABC123",
        "visit_uv": 0,
        "register_at": 1710000000,
        "last_login_at": 0
      }
    ],
    "total": 1
  }
}
```

### POST /agent/create

分站添加代理。对应 RPC `addOperatorAgent`。

**请求**

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `agent` | json | 是 | object | 见 `AgentCreateFields` |
| `is_init` | json | 否 | bool | 是否初始化代理；`true` 时可走指定线/root 逻辑 |
| `parent_id` | json | 否 | int64 | 上级 ID；`0` 且非 init 时失败 |

```json
{
  "agent": {
    "account": "agent01",
    "code": "A001",
    "password": "Passw0rd!",
    "agent_type": 1,
    "agent_line_code": "CASH_PRODUCTION",
    "share_rate": 50,
    "promo_code": "ABC123"
  },
  "is_init": false,
  "parent_id": 1001
}
```

**响应**

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 1002,
    "account": "agent01",
    "code": "A001",
    "parent_id": 1001,
    "top_id": 1001,
    "level": 2,
    "share_rate": 50
  }
}
```

### POST /agent/search

按关键字搜索代理（指定代理选择）。对应 RPC `searchAgents`。

**请求**

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `keyword` | json | 是 | string | 纯数字：ID 或账号模糊；否则账号模糊 |
| `level` | json | 否 | int32 | `<=0` 时默认按 CRM 规则（通常 ≥2） |
| `limit` | json | 否 | int32 | 默认 20，上限 50 |

```json
{
  "keyword": "agent",
  "level": 2,
  "limit": 20
}
```

**响应**

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "items": [
      {
        "id": 1002,
        "account": "agent01",
        "level": 2,
        "level_alias": "总代",
        "line_code": "CASH_PRODUCTION"
      }
    ]
  }
}
```

### GET /agent/operator/list

按层级查询代理，并按 `agent_line_code` 分组。对应 RPC `listOperatorAgents`。

**请求**

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `level` | query | 是 | int32 | 层级 1~7（业务上常用 `2`） |

```http
GET /agent/operator/list?level=2
```

**响应**

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "groups": [
      {
        "line_code": "CASH_PRODUCTION",
        "agents": [
          {
            "id": 1002,
            "account": "agent01",
            "level": 2,
            "level_alias": "总代",
            "line_code": "CASH_PRODUCTION"
          }
        ]
      }
    ]
  }
}
```

分组顺序按已知线路常量（`CASH_PRODUCTION`、`CASH_DEMO`、`CASH_TEST`、`CREDIT_DEMO`、`CREDIT_TEST`）；未知 `line_code` 排在末尾。仅返回有代理的组。

---

## 代理等级

### GET /agent/level/list

当前租户代理等级列表。对应 RPC `listAgentLevels`。无请求参数。

**响应**

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "items": [
      {
        "id": 1,
        "level": 1,
        "name": "系统",
        "alias": "系统"
      },
      {
        "id": 2,
        "level": 2,
        "name": "总代",
        "alias": "总代"
      }
    ]
  }
}
```

### POST /agent/level/update

批量 upsert 代理等级。对应 RPC `updateAgentLevels`。

**请求**

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `items` | json | 是 | array | `AgentLevelItem`；`level` 1~7，同 level 不可重复 |

```json
{
  "items": [
    { "level": 2, "name": "总代", "alias": "总代" },
    { "level": 3, "name": "一级", "alias": "一级" }
  ]
}
```

**响应**

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "items": [
      { "id": 2, "level": 2, "name": "总代", "alias": "总代" },
      { "id": 3, "level": 3, "name": "一级", "alias": "一级" }
    ]
  }
}
```

---

## 代理配置

### GET /agent/config

获取当前租户代理配置。无配置时字段为 `0`（`id=0`）。对应 RPC `getAgentConfig`。

**响应**

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 1,
    "max_share_rate": 100,
    "apply_min_share_rate": 0,
    "apply_max_share_rate": 80,
    "cash_max_member_count": 10000,
    "cash_max_agent_per_ip": 5,
    "cash_production_agent_id": 1002,
    "cash_demo_agent_id": 0,
    "cash_test_agent_id": 0,
    "credit_demo_agent_id": 0,
    "credit_test_agent_id": 0
  }
}
```

### POST /agent/config/update

更新（upsert）代理配置。对应 RPC `updateAgentConfig`。指定代理 ID 为 `0` 表示清空。

**请求**

字段同 `AgentConfig`（不含 `id`）。

```json
{
  "max_share_rate": 100,
  "apply_min_share_rate": 0,
  "apply_max_share_rate": 80,
  "cash_max_member_count": 10000,
  "cash_max_agent_per_ip": 5,
  "cash_production_agent_id": 1002,
  "cash_demo_agent_id": 0,
  "cash_test_agent_id": 0,
  "credit_demo_agent_id": 0,
  "credit_test_agent_id": 0
}
```

**响应**

同 [GET /agent/config](#get-agentconfig) 的 `data`。

---

## 代理树

### GET /agent/tree/node

按上级 ID 取直属下级；`parent_id` 为 `0` 或不传时返回顶层（`parent_id IS NULL`）。对应 RPC `treeNodeAgents`。

**请求**

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `parent_id` | query | 否 | int64 | 上级 ID；`0`/省略为顶层 |

```http
GET /agent/tree/node?parent_id=1001
```

**响应**

```json
{
  "code": 0,
  "msg": "ok",
  "data": [
    {
      "id": 1002,
      "account": "agent01",
      "line_code": "CASH_PRODUCTION",
      "level": 2,
      "level_alias": "总代",
      "direct_agent_count": 3
    }
  ]
}
```

---

## 二、Catalog 注册

服务启动时调用分站 Core `RegisterCatalog`（i18n code=`operator`），注册：

- 菜单：看板、代理管理（列表 / 树 / 配置 / 等级）及按钮权限
- API：上表全部 path（`ServiceName=operator-admin-api`）
- 多语言：`menu` / `api` / `front` / `error` 组词条

菜单与 API 按 `name`、method+path upsert；超管角色自动补授权。详见 [core-api.md](./core-api.md) 「菜单 / API 自动迁移」。
