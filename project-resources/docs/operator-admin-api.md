# operator-admin-api 说明文档

## 一、HTTP API

`operator-admin-api` 为分站后台 CRM HTTP 网关，转发 `operator-crm` 的 `AgentService` 等业务 RPC。  
默认开发地址 `http://127.0.0.1:38010`，**无**统一 `/admin` 前缀。JSON 字段以 `[operator-admin-api/desc](../../operator-admin-api/desc)` 为准。

登录鉴权、Token 刷新、过期处理见同目录 [core-api.md](./core-api.md)。本服务 JWT / 权限由分站 Core 校验；请求上下文中的 `operator_code` 会透传到 CRM。

## 接口索引

- [约定](#约定)
- [公开（无 JWT）](#公开无-jwt)
  - [GET /ping](#get-ping)
- [需登录（仅 Jwt）](#需登录仅-jwt)
  - [GET /phone-calling-codes](#get-phone-calling-codes)
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
    - [GET /agent/line/defaults](#get-agentlinedefaults)
    - [GET /agent/init/status](#get-agentinitstatus)
  - [代理树](#代理树)
    - [GET /agent/tree/node](#get-agenttreenode)
    - [GET /agent/tree/ancestors](#get-agenttreeancestors)
  - [代理登录日志](#代理登录日志)
    - [POST /agent/login-log/list](#post-agentlogin-loglist)
    - [POST /agent/login-log/associations](#post-agentlogin-logassociations)
  - [代理申请](#代理申请)
    - [POST /agent/apply/list](#post-agentapplylist)
    - [POST /agent/apply/review](#post-agentapplyreview)
  - [代理资料](#代理资料)
    - [GET /agent/profile](#get-agentprofile)
    - [POST /agent/password/reset](#post-agentpasswordreset)
    - [POST /agent/withdraw-password/reset](#post-agentwithdraw-passwordreset)
    - [POST /agent/share-rate/update](#post-agentshare-rateupdate)
    - [POST /agent/payout-restricted/update](#post-agentpayout-restrictedupdate)
    - [POST /agent/status/update](#post-agentstatusupdate)
    - [POST /agent/remark/update](#post-agentremarkupdate)
    - [POST /agent/profile/update](#post-agentprofileupdate)
    - [GET /agent/op-logs](#get-agentop-logs)

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

```http
Authorization: Bearer <access_token>
```

- **公开**：无 JWT，如 `GET /ping`
- **仅 Jwt**：需登录，无 ActionLog / Authority，如 `GET /common/phone-calling-codes`
- **Jwt + ActionLog + Authority**：业务受保护接口
  - **Jwt**：分站 Core `CheckToken`
  - **ActionLog**：记录操作日志
  - **Authority**：按 method + path 调 Core `Enforce`（Casbin）；依赖 Catalog 已注册的 `/agent/`* path，超管角色启动时会补授权

Token 获取见 [core-api.md](./core-api.md)。Dev 模式下 Jwt / Authority 透传，便于本地联调。

### 客户端元数据透传

全局中间件会把客户端信息写入上下文，并随 gRPC metadata 传给 CRM（用于资料变更操作日志等）：

| 来源             | 透传字段           | 说明             |
| -------------- | -------------- | -------------- |
| `ClientIP` 中间件 | 客户端 IP         | 由 core 解析请求 IP |
| `User-Agent`   | `login_device` | 标准请求头          |
| `X-Udid`       | `udid`         | 可选；客户端设备码      |

写代理资料相关接口（重置密码、改占成等）时建议带上 `X-Udid`，便于操作记录落库。

### 多语言

请求语言通过 Header 指定：

```http
X-Lang: zh-CN
```

| 语言编码    | 说明   |
| ------- | ---- |
| `zh-CN` | 简体中文 |
| `zh-HK` | 繁体中文 |
| `en-US` | 英文   |

默认 `zh-CN`。菜单 / API 描述词条由启动时 `RegisterCatalog` 写入 Core（i18n code=`operator`）。

### 公共类型

`AgentLevelItem`：

| 字段      | 类型     | 说明          |
| ------- | ------ | ----------- |
| `id`    | int64  | 等级主键；创建时可省略 |
| `level` | int32  | 等级 1~7      |
| `name`  | string | 等级名称        |
| `alias` | string | 等级别名        |

`AgentConfig`：

| 字段                         | 类型    | 说明                     |
| -------------------------- | ----- | ---------------------- |
| `id`                       | int64 | 配置主键；无配置时为 `0`         |
| `max_share_rate`           | int32 | 最高占成，百分比 0~100         |
| `apply_min_share_rate`     | int32 | 申请最低占成                 |
| `apply_max_share_rate`     | int32 | 申请最高占成；须 ≥ min 且 ≤ max |
| `cash_max_member_count`    | int64 | 现金代理最大会员数              |
| `cash_max_agent_per_ip`    | int64 | 现金 IP 最大可注册代理数         |
| `cash_production_agent_id` | int64 | 现金正式线指定代理，`0` 表示清空     |
| `cash_demo_agent_id`       | int64 | 现金试玩线指定代理              |
| `cash_test_agent_id`       | int64 | 现金测试线指定代理              |
| `credit_demo_agent_id`     | int64 | 信誉试玩线指定代理              |
| `credit_test_agent_id`     | int64 | 信誉测试线指定代理              |

`AgentCreateFields`：

| 字段                | 类型     | 说明                        |
| ----------------- | ------ | ------------------------- |
| `account`         | string | 登录账号                      |
| `code`            | string | 代理编码                      |
| `password`        | string | 登录密码明文                    |
| `agent_type`      | int32  | `1` 现金代理，`2` 信用代理         |
| `agent_line_code` | string | 子线路编码，如 `CASH_PRODUCTION` |
| `share_rate`      | int32  | 占成，百分比 0~100              |
| `promo_code`      | string | 对外推广码                     |
| `register_ip`     | string | 注册 IP，可选                  |

`AddAgentResp`：

| 字段           | 类型     | 说明             |
| ------------ | ------ | -------------- |
| `id`         | int64  | 代理 ID          |
| `account`    | string | 账号             |
| `code`       | string | 编码             |
| `parent_id`  | int64  | 上级 ID          |
| `top_id`     | int64  | 顶层 ID；顶层可为 `0` |
| `level`      | int32  | 层级             |
| `share_rate` | int32  | 占成             |

`AgentSearchItem`：

| 字段            | 类型     | 说明                |
| ------------- | ------ | ----------------- |
| `id`          | int64  | 代理 ID             |
| `account`     | string | 账号                |
| `level`       | int32  | 层级                |
| `level_alias` | string | 等级别名              |
| `line_code`   | string | `agent_line_code` |
| `share_rate`  | int32  | 当前占成，百分比 0~100    |

`AgentListItem`：

| 字段                    | 类型     | 说明                                 |
| --------------------- | ------ | ---------------------------------- |
| `id`                  | int64  | 代理 ID                              |
| `account`             | string | 账号                                 |
| `level`               | int32  | 层级                                 |
| `level_alias`         | string | 等级别名                               |
| `parent_id`           | int64  | 上级 ID                              |
| `parent_account`      | string | 上级账号                               |
| `top_id`              | int64  | 顶层 ID                              |
| `top_account`         | string | 顶层账号                               |
| `register_source`     | int32  | 注册来源：`1` PC，`2` app，`3` 管理员，`4` 代理 |
| `direct_member_count` | int64  | 直属会员数                              |
| `direct_agent_count`  | int64  | 直属代理数                              |
| `total_member_count`  | int64  | 全部会员数                              |
| `total_agent_count`   | int64  | 全部下级代理数                            |
| `share_rate`          | int32  | 占成                                 |
| `wallet_balance`      | int64  | 钱包余额（当前固定 `0`）                     |
| `total_commission`    | int64  | 累计佣金                               |
| `unpaid_commission`   | int64  | 未领取佣金                              |
| `status`              | int32  | `1` 正常，`2` 冻结，`3` 停用               |
| `payout_restricted`   | int32  | `1` 不限制，`2` 限制本人，`3` 限制全线          |
| `remark`              | string | 备注                                 |
| `promo_link`          | string | 推广链接（暂用 `promo_code`）              |
| `visit_uv`            | int64  | 访问 UV                              |
| `register_at`         | int64  | 注册时间 Unix 秒                        |
| `last_login_at`       | int64  | 最后登录 Unix 秒                        |

`TreeNodeAgentItem`：

| 字段                   | 类型     | 说明    |
| -------------------- | ------ | ----- |
| `id`                 | int64  | 代理 ID |
| `account`            | string | 账号    |
| `line_code`          | string | 线路编码  |
| `level`              | int32  | 层级    |
| `level_alias`        | string | 等级别名  |
| `direct_agent_count` | int64  | 直属代理数 |

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

## 需登录（仅 Jwt）

需 Header：`Authorization: Bearer <access_token>`，不校验 Casbin 权限、不写操作日志。

### GET /phone-calling-codes

返回 common 包支持的国际手机区号列表（`utils.CallingCodes`）。`code` 为国际区号，**不含**加号；`region` 为 ISO 3166-1 alpha-2。

**响应**

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "list": [
      { "region": "US", "code": "1" },
      { "region": "MY", "code": "60" },
      { "region": "SG", "code": "65" }
    ]
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

| 字段                          | 位置   | 必填  | 类型     | 说明                       |
| --------------------------- | ---- | --- | ------ | ------------------------ |
| `page`                      | json | 否   | int32  | 页码，从 1；默认按 CRM           |
| `page_size`                 | json | 否   | int32  | 每页条数，默认 20，上限 100        |
| `register_source`           | json | 否   | int32  | 注册来源，`0` 不限              |
| `level`                     | json | 否   | int32  | 层级，`0` 不限                |
| `status`                    | json | 否   | int32  | 状态，`0` 不限                |
| `payout_restricted`         | json | 否   | int32  | 出款限制，`0` 不限              |
| `time_filter_type`          | json | 否   | int32  | `1` 注册时间，`2` 最后登录，`0` 不筛 |
| `time_start` / `time_end`   | json | 否   | int64  | Unix 秒；成对有效              |
| `number_filter_type`        | json | 否   | int32  | `1`~`7` 数值筛选项，`0` 不筛     |
| `number_min` / `number_max` | json | 否   | int64  | `>0` 才生效                 |
| `agent_filter_type`         | json | 否   | int32  | `1` 本人，`2` 上级，`0` 不筛     |
| `agent_keyword`             | json | 否   | string | 纯数字按 ID，否则账号模糊           |

`number_filter_type`：`1` 直属会员数，`2` 直属代理数，`3` 所有下级总数，`4` 所有会员总数，`5` 占成比例，`6` 累计佣金，`7` 已领取佣金（累计−未领）。

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
    "list": [
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

| 字段          | 位置   | 必填  | 类型     | 说明                            |
| ----------- | ---- | --- | ------ | ----------------------------- |
| `agent`     | json | 是   | object | 见 `AgentCreateFields`         |
| `is_init`   | json | 否   | bool   | 是否初始化代理；`true` 时可走指定线/root 逻辑 |
| `parent_id` | json | 否   | int64  | 上级 ID；`0` 且非 init 时失败         |

```json
{
  "agent": {
    "account": "agent01",
    "password": "Passw0rd!",
    "agent_type": 1,
    "agent_line_code": "CASH_PRODUCTION",
    "share_rate": 50
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

| 字段        | 位置   | 必填  | 类型     | 说明                       |
| --------- | ---- | --- | ------ | ------------------------ |
| `keyword` | json | 是   | string | 纯数字：ID 或账号模糊；否则账号模糊      |
| `level`   | json | 否   | int32  | `<=0` 时默认按 CRM 规则（通常 ≥2） |
| `limit`   | json | 否   | int32  | 默认 20，上限 50              |

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
    "list": [
      {
        "id": 1002,
        "account": "agent01",
        "level": 2,
        "level_alias": "总代",
        "line_code": "CASH_PRODUCTION",
        "share_rate": 50
      }
    ]
  }
}
```

### GET /agent/operator/list

### PS: 用于设置分站代理线路默认代理

按层级查询代理，并按 `agent_line_code` 分组。对应 RPC `listOperatorAgents`。下发 operator-base `GetAgentLines` 的全部已开通线路，顺序与其一致。该层级没有代理时 `agents` 为 `[]`。未开通线路和未知线路不下发。

**请求**

| 字段      | 位置    | 必填  | 类型    | 说明                |
| ------- | ----- | --- | ----- | ----------------- |
| `level` | query | 是   | int32 | 层级 1~7（业务上常用 `2`） |

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
            "line_code": "CASH_PRODUCTION",
            "share_rate": 50
          }
        ]
      },
      {
        "line_code": "CASH_DEMO",
        "agents": []
      }
    ]
  }
}
```

分组顺序与已开通线路一致。没有代理的线路仍下发，`agents` 为 `[]`。未知 `line_code` 不下发。

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
    "list": [
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

| 字段      | 位置   | 必填  | 类型    | 说明                                        |
| ------- | ---- | --- | ----- | ----------------------------------------- |
| `list` | json | 是   | array | `AgentLevelItem`；`level` 1~7，同 level 不可重复 |

```json
{
  "list": [
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
    "list": [
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

### GET /agent/line/defaults

### PS: 用于新增代理的默认代理选择

返回当前分站已开通代理线在配置中的默认代理及占成。先取 operator-base `GetAgentLines`，再取 `getAgentConfig`。`agent_id` 大于 0 时用 `getAgentProfile` 取 `share_rate`（百分比 0~100）。未配置时 `agent_id` 与 `share_rate` 为 `0`。顺序与已开通线路一致。

**响应**

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "list": [
      {
        "line_code": "CASH_PRODUCTION",
        "agent_id": 1002,
        "share_rate": 50
      },
      {
        "line_code": "CASH_DEMO",
        "agent_id": 0,
        "share_rate": 0
      }
    ]
  }
}
```

### GET /agent/init/status

检测已开通代理线是否完成初始化，以及 `agent_level` 是否至少 3 层。线路以 `GetAgentLines` 为准；`listOperatorAgents` 固定查 `level=1`，该线路至少有一名代理则 `initialized` 为 `true`。`ready` 为等级已就绪且每条已开通线路都已初始化。没有开通线路时只看等级。

**响应**

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "ready": false,
    "levels_ready": true,
    "level_count": 3,
    "lines": [
      { "line_code": "CASH_PRODUCTION", "initialized": true },
      { "line_code": "CASH_DEMO", "initialized": false }
    ]
  }
}
```

---

## 代理树

### GET /agent/tree/node

按上级 ID 取直属下级；`parent_id` 为 `0` 或不传时返回顶层（`parent_id IS NULL`）。对应 RPC `treeNodeAgents`。

**请求**

| 字段          | 位置    | 必填  | 类型    | 说明              |
| ----------- | ----- | --- | ----- | --------------- |
| `parent_id` | query | 否   | int64 | 上级 ID；`0`/省略为顶层 |

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

### GET /agent/tree/ancestors

按代理 ID 或账号精确查找，返回当前代理及其全部上级。对应 RPC `treeNodeAncestorAgents`。当前代理在数组最前，其后为直接上级直到顶层。纯数字先按 ID 匹配，没有再按账号匹配；其他内容只按账号匹配。

**请求**

| 字段        | 位置    | 必填  | 类型     | 说明        |
| --------- | ----- | --- | ------ | --------- |
| `keyword` | query | 是   | string | 代理 ID 或账号 |

```http
GET /agent/tree/ancestors?keyword=agent01
```

**响应** `data` 为 `AgentSearchItem` 数组，含 `share_rate`（百分比 0~100）。

```json
{
  "code": 0,
  "msg": "ok",
  "data": [
    {
      "id": 1003,
      "account": "agent01",
      "level": 2,
      "level_alias": "总代",
      "line_code": "CASH_PRODUCTION",
      "share_rate": 50
    },
    {
      "id": 1001,
      "account": "parent01",
      "level": 1,
      "level_alias": "股东",
      "line_code": "CASH_PRODUCTION",
      "share_rate": 80
    }
  ]
}
```

---

## 代理登录日志

对应 RPC `listAgentLoginLogs`。登录类型、设备类型、登录结果数值码保留，并额外下发当前语言的 `*_label`（front 词典）。

### POST /agent/login-log/list

分页查询代理登录日志。

**请求**

| 字段             | 位置   | 必填  | 类型     | 说明                                               |
| -------------- | ---- | --- | ------ | ------------------------------------------------ |
| `page`         | body | 否   | int32  | 页码，默认 1                                          |
| `page_size`    | body | 否   | int32  | 每页条数，默认 20，上限 100                                |
| `time_start`   | body | 否   | int64  | 登录时间起始 unix 秒，`0` 不限                             |
| `time_end`     | body | 否   | int64  | 登录时间结束 unix 秒，`0` 不限                             |
| `login_type`   | body | 否   | int32  | `0` 不限，`1` 登录，`2` 注册                             |
| `device_type`  | body | 否   | int32  | `0` 不限，`1`~`5` PC网页 / H5 / Android / iOS / PC客户端 |
| `keyword_type` | body | 否   | int32  | `0` 不筛，`1` 代理 ID，`2` 登录 IP，`3` 设备码，`4` 登录地点      |
| `keyword`      | body | 否   | string | 与 `keyword_type` 配合                              |

```json
{
  "page": 1,
  "page_size": 20,
  "time_start": 1760256000,
  "time_end": 1760342399,
  "login_type": 0,
  "device_type": 0,
  "keyword_type": 1,
  "keyword": "61339567"
}
```

**响应**

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "total": 1,
    "list": [
      {
        "id": 1,
        "agent_id": 61339567,
        "account": "agent01",
        "login_type": 1,
        "login_type_label": "登录",
        "login_ip": "38.54.14.228",
        "login_location": "Vietnam",
        "udid": "4001297567",
        "login_device": "Chrome 149.0.0.0",
        "device_type": 1,
        "device_type_label": "PC网页",
        "login_at": 1760262518,
        "result": 1,
        "result_label": "成功"
      }
    ]
  }
}
```

### POST /agent/login-log/associations

关联检索登录日志。对应 RPC `listAgentLoginAssociations`。同 IP / 同设备代理数仅在时间范围与设备类型范围内统计（不计 `agent_id` / keyword）。设备类型数值码保留，并额外下发当前语言的 `device_type_label`（front 词典）。

**请求**

| 字段             | 位置   | 必填  | 类型     | 说明                                               |
| -------------- | ---- | --- | ------ | ------------------------------------------------ |
| `page`         | body | 否   | int32  | 页码，默认 1                                          |
| `page_size`    | body | 否   | int32  | 每页条数，默认 20，上限 100                                |
| `time_start`   | body | 否   | int64  | 登录时间起始 unix 秒，`0` 不限                             |
| `time_end`     | body | 否   | int64  | 登录时间结束 unix 秒，`0` 不限                             |
| `agent_id`     | body | 否   | int64  | 代理 ID，`0` 不限                                     |
| `device_type`  | body | 否   | int32  | `0` 不限，`1`~`5` PC网页 / H5 / Android / iOS / PC客户端 |
| `keyword_type` | body | 否   | int32  | `0` 不筛，`1` 登录 IP，`2` 设备码                         |
| `keyword`      | body | 否   | string | 与 `keyword_type` 配合                              |

```json
{
  "page": 1,
  "page_size": 20,
  "time_start": 1760256000,
  "time_end": 1760342399,
  "agent_id": 0,
  "device_type": 0,
  "keyword_type": 1,
  "keyword": "38.54.14.228"
}
```

**响应**

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "total": 1,
    "list": [
      {
        "id": 1,
        "agent_id": 61339567,
        "account": "agent01",
        "login_ip": "38.54.14.228",
        "login_location": "Vietnam",
        "udid": "4001297567",
        "login_device": "Chrome 149.0.0.0",
        "device_type": 1,
        "device_type_label": "PC网页",
        "login_at": 1760262518,
        "same_ip_agent_count": 3,
        "same_device_agent_count": 2
      }
    ]
  }
}
```

---

## 代理申请

对应 RPC `listAgentApplies` / `reviewAgentApply`。审核状态：`1` 待审核，`2` 已通过，`3` 人工拒绝，`4` 风控拒绝。占成为百分比 0~100。

### POST /agent/apply/list

分页查询代理申请。

**请求**

| 字段                        | 位置   | 必填  | 类型     | 说明                         |
| ------------------------- | ---- | --- | ------ | -------------------------- |
| `page`                    | json | 否   | int32  | 页码，从 1                     |
| `page_size`               | json | 否   | int32  | 默认 20，上限 100               |
| `audit_status`            | json | 否   | int32  | `0` 不限                     |
| `account_type`            | json | 否   | int32  | `1` 代理账号，`2` 会员账号，`0` 不限类型 |
| `account`                 | json | 否   | string | 账号模糊                       |
| `time_start` / `time_end` | json | 否   | int64  | 申请时间 Unix 秒，`0` 不限         |

```json
{
  "page": 1,
  "page_size": 20,
  "audit_status": 1
}
```

**响应**

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "list": [
      {
        "id": 1,
        "account": "apply01",
        "realname": "张三",
        "mobile": "13800000000",
        "register_source": 1,
        "apply_share_rate": 50,
        "member_account": "user01",
        "member_code": "100000000001",
        "audit_status": 1,
        "reject_reason": "",
        "apply_remark": "",
        "register_at": 1710000000,
        "audited_at": 0,
        "auditor": "",
        "agent_id": 0
      }
    ],
    "total": 1
  }
}
```

### POST /agent/apply/review

审批申请。仅待审核可审。通过时按现金正式线指定代理开户，响应中的 `password` 为明文且仅返回一次。拒绝时 `reject_reason` 必填。

**请求**

| 字段              | 位置   | 必填  | 类型     | 说明                       |
| --------------- | ---- | --- | ------ | ------------------------ |
| `id`            | json | 是   | int64  | 申请单 ID                   |
| `action`        | json | 是   | int32  | `1` 通过，`2` 人工拒绝，`3` 风控拒绝 |
| `reject_reason` | json | 拒绝时 | string | 拒绝理由                     |

```json
{
  "id": 1,
  "action": 1
}
```

**响应**

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "apply": {
      "id": 1,
      "account": "apply01",
      "audit_status": 2,
      "agent_id": 1002,
      "auditor": "admin",
      "audited_at": 1710001000
    },
    "password": "a1b2c3d4e5f6"
  }
}
```

---

## 代理资料

对应 RPC `getAgentProfile`、`resetLoginPassword`、`updateWithdrawPassword`、`updateAgentShareRate`、`updateAgentPayoutRestricted`、`updateAgentStatus`、`updateAgentRemark`、`updateAgentProfile`、`listAgentOpLogs`。

账号状态：`1` 正常，`2` 冻结，`3` 禁用。注册来源：`1` PC，`2` app，`3` 管理员，`4` 代理。占成为百分比 `0`–`100`。出款限制：`1` 不限制，`2` 限制本人，`3` 限制全线。代理类型：`1` 现金，`2` 信用。`invite_member_count` 为直属会员数。

### GET /agent/profile

按代理 ID 查询资料。

**请求**

| 字段         | 位置    | 必填  | 类型    | 说明    |
| ---------- | ----- | --- | ----- | ----- |
| `agent_id` | query | 是   | int64 | 代理 ID |

```http
GET /agent/profile?agent_id=1001
```

**响应** `data`

| 字段                                                                                                                                                  | 类型     | 说明      |
| --------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ------- |
| `agent_id`                                                                                                                                          | int64  | 代理 ID   |
| `remark`                                                                                                                                            | string | 主表备注    |
| `status`                                                                                                                                            | int32  | 账号状态    |
| `register_source`                                                                                                                                   | int32  | 注册来源    |
| `promo_code`                                                                                                                                        | string | 对外推广码   |
| `parent_account`                                                                                                                                    | string | 上级代理账号  |
| `register_at`                                                                                                                                       | int64  | 账号开通时间  |
| `last_login_at`                                                                                                                                     | int64  | 最后登录时间  |
| `last_login_ip`                                                                                                                                     | string | 最后登录 IP |
| `last_login_udid`                                                                                                                                   | string | 最后登录设备码 |
| `share_rate`                                                                                                                                        | int32  | 当前占成    |
| `login_count`                                                                                                                                       | int64  | 登录次数    |
| `invite_member_count`                                                                                                                               | int64  | 邀请用户数   |
| `payout_restricted`                                                                                                                                 | int32  | 出款限制    |
| `agent_type`                                                                                                                                        | int32  | 代理类型    |
| `avatar_url` / `realname` / `telegram` / `facebook` / `whatsapp` / `mobile` / `nickname` / `id_document_no` / `zalo` / `twitter` / `line` / `email` | string | 资料字段    |

### POST /agent/password/reset

服务端生成随机登录密码，只存哈希，不改令牌盐。响应中的 `password` 为明文且仅返回一次。

**请求**

| 字段         | 位置   | 必填  | 类型    | 说明    |
| ---------- | ---- | --- | ----- | ----- |
| `agent_id` | json | 是   | int64 | 代理 ID |

```json
{ "agent_id": 1001 }
```

**响应** `data`

```json
{ "password": "a1b2c3d4e5f67890" }
```

### POST /agent/withdraw-password/reset

调用方传入提现密码明文，只存哈希，响应不回传密码。

**请求**

| 字段         | 位置   | 必填  | 类型     | 说明     |
| ---------- | ---- | --- | ------ | ------ |
| `agent_id` | json | 是   | int64  | 代理 ID  |
| `password` | json | 是   | string | 提现密码明文 |

成功时 `data` 为空对象。

### POST /agent/share-rate/update

占成百分比 `0`–`100`，不能超过上级占成和配置上限，也不能低于直属下级里最高的占成。没有上级时只校验配置上限；没有直属下级时不校验下限。相等允许。

**请求**

| 字段           | 位置   | 必填  | 类型    | 说明    |
| ------------ | ---- | --- | ----- | ----- |
| `agent_id`   | json | 是   | int64 | 代理 ID |
| `share_rate` | json | 是   | int32 | 百分比   |

成功时 `data` 为空对象。

### POST /agent/payout-restricted/update

**请求**

| 字段                  | 位置   | 必填  | 类型    | 说明              |
| ------------------- | ---- | --- | ----- | --------------- |
| `agent_id`          | json | 是   | int64 | 代理 ID           |
| `payout_restricted` | json | 是   | int32 | `1` / `2` / `3` |

成功时 `data` 为空对象。

### POST /agent/status/update

对应 RPC `updateAgentStatus`。修改代理账号状态。

**请求**

| 字段         | 位置   | 必填  | 类型    | 说明                   |
| ---------- | ---- | --- | ----- | -------------------- |
| `agent_id` | json | 是   | int64 | 代理 ID                |
| `status`   | json | 是   | int32 | `1` 正常，`2` 冻结，`3` 禁用 |

成功时 `data` 为空对象。

### POST /agent/remark/update

只改主表备注。`remark` 可省略或传空字符串，表示清空。

**请求**

| 字段         | 位置   | 必填  | 类型     | 说明     |
| ---------- | ---- | --- | ------ | ------ |
| `agent_id` | json | 是   | int64  | 代理 ID  |
| `remark`   | json | 否   | string | 空字符串清空 |

成功时 `data` 为空对象。

### POST /agent/profile/update

按 `field_type` 更新资料，不含备注。被选中的字段省略或传空字符串表示清空。

| `field_type` | 字段               |
| ------------ | ---------------- |
| `0`          | 下列全部             |
| `1`          | `avatar_url`     |
| `2`          | `realname`       |
| `3`          | `telegram`       |
| `4`          | `facebook`       |
| `5`          | `whatsapp`       |
| `6`          | `mobile`         |
| `7`          | `nickname`       |
| `8`          | `id_document_no` |
| `9`          | `zalo`           |
| `10`         | `twitter`        |
| `11`         | `line`           |
| `12`         | `email`          |

**请求**

```json
{
  "agent_id": 1001,
  "field_type": 2,
  "realname": "张三"
}
```

**响应** `data` 与 `GET /agent/profile` 相同。

### GET /agent/op-logs

对应 RPC `listAgentOpLogs`。分页查询资料/账号变更操作记录（占成、密码、状态、备注、资料字段等）。`target_id` / `target_type` 均为可选过滤；不传则不限目标。

数值/code 字段保留，并额外下发当前语言的 label（front 词典）：

| 原字段 | label 字段 | 词典 key |
| --- | --- | --- |
| `op_target_type` | `op_target_type_label` | `agent.opTargetType.<n>` |
| `admin_type` | `admin_type_label` | `agent.adminType.<n>` |
| `action` | `action_label` | 与 `action` 相同（落库即为完整 i18n key） |
| `result` | `result_label` | `agent.opResult.<n>` |

`fail_reason` 若存的是错误 i18n key，会按 error 组翻译后返回。

权限：`agent:op-log:list`。

**请求**

| 字段           | 位置    | 必填  | 类型    | 说明                                      |
| ------------ | ----- | --- | ----- | --------------------------------------- |
| `target_id`  | query | 否   | int64 | 操作目标 ID，`0` 或不传表示不限                      |
| `target_type`| query | 否   | int32 | 操作目标类型：`0` 不限，`1` 代理，`2` 会员             |
| `page`       | query | 否   | int32 | 页码，默认 1                                 |
| `page_size`  | query | 否   | int32 | 每页条数，默认 20，上限 100                       |
| `time_start` | query | 否   | int64 | 操作时间起始 unix 秒，`0` 不限                    |
| `time_end`   | query | 否   | int64 | 操作时间结束 unix 秒，`0` 不限                    |
| `result`     | query | 否   | int32 | `0` 不限，`1` 成功，`2` 失败                    |

```http
GET /agent/op-logs?target_id=1001&target_type=1&page=1&page_size=20&result=0
```

**响应**

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "total": 1,
    "list": [
      {
        "id": 1,
        "op_target_id": 1001,
        "op_target_type": 1,
        "op_target_type_label": "代理账号",
        "op_target_account": "agent01",
        "admin_id": 2001,
        "admin_type": 1,
        "admin_type_label": "分站管理员",
        "admin_account": "admin01",
        "action": "agent.opAction.share_rate",
        "action_label": "更新占成",
        "before_value": "30",
        "after_value": "40",
        "client_ip": "10.0.0.1",
        "client_ip_location": "私有IP",
        "login_device": "Mozilla/5.0 ...",
        "udid": "device-abc",
        "operated_at": 1760262518,
        "result": 1,
        "result_label": "成功",
        "fail_reason": ""
      }
    ]
  }
}
```

`op_target_type`

| 值 | 说明 |
| -- | -- |
| `1` | 代理账号 |
| `2` | 会员账号 |

`admin_type`

| 值 | 说明 |
| -- | -- |
| `1` | 分站管理员 |
| `2` | 分站代理 |

`action`（落库为完整 front key，稳定可直接作词典查询）

| action | 说明 |
| --- | --- |
| `agent.opAction.create` | 创建代理 |
| `agent.opAction.share_rate` | 更新占成 |
| `agent.opAction.payout_restricted` | 更新出款限制 |
| `agent.opAction.status` | 更新状态 |
| `agent.opAction.remark` | 更新备注 |
| `agent.opAction.reset_login_password` | 重置登录密码 |
| `agent.opAction.update_login_password` | 修改登录密码 |
| `agent.opAction.update_withdraw_password` | 修改提款密码 |
| `agent.opAction.profile` | 整份资料更新（`field_type=0`，前后值为 JSON） |
| `agent.opAction.profile.avatar` / `…realname` / `…mobile` / … | 单字段资料更新 |

密码类变更的 `before_value` / `after_value` 为 `***`，不落明文。

---

## 二、Catalog 注册

服务启动时调用分站 Core `RegisterCatalog`（i18n code=`operator`），注册：

- 菜单：看板、代理管理（列表 / 树 / 配置 / 等级 / 申请）及按钮权限（含 `agent:op-log:list`）
- API：上表全部 path（`ServiceName=operator-admin-api`，含 `GET /agent/op-logs`）
- 多语言：`menu` / `api` / `front` / `error` 组词条（含 `agent.opAction.*`、`agent.opResult.*`）

菜单与 API 按 `name`、method+path upsert；超管角色自动补授权。详见 [core-api.md](./core-api.md) 「菜单 / API 自动迁移」。