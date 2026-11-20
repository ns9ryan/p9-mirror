# 分站后台活动管理 API 对接文档

- **适用服务：** `operator-admin-api`
- **对接范围：** 创建活动、编辑活动、活动详情、活动列表、关闭活动、发布活动、活动枚举参数、活动订单列表、活动订单审核列表、活动订单审核
- **更新日期：** 2026-11-19
- **代码依据：** `operator-admin-api/desc/activity/manage.api`、`operator-admin-api/desc/activity/order.api`，RPC 透传 `operator-activity` ActivityService

## 目录索引

- [一、基础约定](#一基础约定)
  - [1. 请求地址与认证](#1-请求地址与认证)
  - [2. 通用响应](#2-通用响应)
  - [3. 数据格式](#3-数据格式)
- [二、接口一览](#二接口一览)
- [三、创建活动](#三创建活动)
  - [1. 创建活动请求参数](#1-创建活动请求参数)
  - [2. 创建活动响应](#2-创建活动响应)
- [四、编辑活动](#四编辑活动)
- [五、活动详情](#五活动详情)
  - [1. 活动详情请求参数](#1-活动详情请求参数)
  - [2. 活动详情响应](#2-活动详情响应)
- [六、活动列表](#六活动列表)
  - [1. 活动列表查询参数](#1-活动列表查询参数)
  - [2. 活动列表响应](#2-活动列表响应)
- [七、关闭活动](#七关闭活动)
- [八、发布活动](#八发布活动)
- [九、活动枚举参数](#九活动枚举参数)
- [十、活动订单列表](#十活动订单列表)
  - [1. 活动订单列表查询参数](#1-活动订单列表查询参数)
  - [2. 活动订单列表响应](#2-活动订单列表响应)
- [十一、活动订单审核列表](#十一活动订单审核列表)
  - [1. 活动订单审核列表查询参数](#1-活动订单审核列表查询参数)
  - [2. 活动订单审核列表响应](#2-活动订单审核列表响应)
- [十二、活动订单审核](#十二活动订单审核)
- [十三、公共数据结构](#十三公共数据结构)
  - [1. ActivitySettings](#1-activitysettings)
  - [2. ActivityRewardSchema](#2-activityrewardschema)
  - [3. ActivityReferConfig](#3-activityreferconfig)
  - [4. ActivityLuckWheelConfig](#4-activityluckwheelconfig)
  - [5. ActivityI18nItem](#5-activityi18nitem)
- [十四、枚举说明](#十四枚举说明)
- [十五、业务错误](#十五业务错误)
- [十六、前端对接注意事项](#十六前端对接注意事项)

---

## 一、基础约定

### 1. 请求地址与认证

**接口内部路径：**

| 功能 | 方法 | 路径 |
|---|---|---|
| 创建活动 | POST | `/activity/create` |
| 编辑活动 | POST | `/activity/update` |
| 活动详情 | GET | `/activity/detail` |
| 活动列表 | GET | `/activity/list` |
| 关闭活动 | POST | `/activity/close` |
| 发布活动 | POST | `/activity/publish` |
| 活动枚举参数 | GET | `/activity/enum-params` |
| 活动订单列表 | GET | `/activity/order/list` |
| 活动订单审核列表 | GET | `/activity/order/audit/list` |
| 活动订单审核 | POST | `/activity/order/audit` |

直接访问开发服务时，例如：

```http
GET http://127.0.0.1:38010/activity/list?page=1&page_size=20
```

部分部署环境通过 Nginx 增加 `/admin/operator` 前缀：

```http
GET https://ryan05.op.com/admin/operator/activity/list?page=1&page_size=20
```

实际请求地址以当前环境配置的 API Base URL 为准，前端不要在业务接口中重复拼接固定前缀。无论外部前缀如何调整，本文列出的服务内部路由不变。

**请求头：**

```http
Authorization: Bearer <access_token>
X-Lang: zh-CN
Content-Type: application/json
```

| 请求头 | 说明 |
|---|---|
| `Authorization` | 登录 Token，全部活动接口必需 |
| `X-Lang` | 响应语言，可选，支持 `zh-CN`、`zh-HK`、`en-US`，默认 `zh-CN` |
| `Content-Type` | POST JSON 请求使用 `application/json` |

全部接口经 `Jwt`、`ActionLog`、`Authority` 中间件，需登录且具备对应 API 权限。

### 2. 通用响应

**成功响应：**

```json
{
  "code": 0,
  "msg": "ok",
  "data": {}
}
```

**失败响应：**

```json
{
  "code": 403,
  "msg": "无权访问"
}
```

常见 HTTP 状态码：

| 状态码 | 说明 |
|---|---|
| `400` | 请求参数错误，或业务校验未通过 |
| `401` | 未登录或认证失败 |
| `403` | 没有接口访问权限 |
| `498` | Access Token 已过期 |
| `500` | 服务内部错误 |

### 3. 数据格式

| 数据类型 | 格式 | 说明 |
|---|---|---|
| 时间点 | Unix **秒**时间戳 | 字段名以 `_at_sec` / `_sec` 结尾，如 `open_at_sec`、`created_at_sec` |
| 金额 | int64 | 当前厅结算币种的最小单位，不是浮点金额 |
| ID | int64 | 数据库主键或业务关联 ID |
| 可选字段 | 可能为空或不返回 | 前端需要兼容空值 / `null` |
| 列表 | array | 无记录时返回空数组 `[]` |

---

## 二、接口一览

| 功能 | 方法 | 路径 | 说明 |
|---|---|---|---|
| 创建活动 | POST | `/activity/create` | 新建活动（默认可为草稿） |
| 编辑活动 | POST | `/activity/update` | 仅草稿可编辑，编辑后仍为草稿 |
| 活动详情 | GET | `/activity/detail` | 按 ID 查询完整活动 |
| 活动列表 | GET | `/activity/list` | 分页筛选活动 |
| 关闭活动 | POST | `/activity/close` | 将活动置为已结束 |
| 发布活动 | POST | `/activity/publish` | 仅草稿可发布为进行中 |
| 活动枚举参数 | GET | `/activity/enum-params` | 表单枚举选项，`label` 随 `X-Lang` |
| 活动订单列表 | GET | `/activity/order/list` | 分页筛选活动订单，不下发订单自增 ID |
| 活动订单审核列表 | GET | `/activity/order/audit/list` | 只查待审核订单 |
| 活动订单审核 | POST | `/activity/order/audit` | 按订单号通过或拒绝 |

---

## 三、创建活动

### 1. 创建活动请求参数

**POST /activity/create**

**请求示例：**

```http
POST /activity/create
Content-Type: application/json
Authorization: Bearer <access_token>
```

```json
{
  "type": 1,
  "status": 1,
  "apply_mode": 1,
  "section": 1,
  "cycle_mode": 1,
  "allowed_fixed_level": [1, 2],
  "allowed_auto_level": [1],
  "allowed_clients": [1, 2],
  "allowed_game_ids": [],
  "allowed_agent_ids": [],
  "apply_limit": [],
  "claim_limit": [],
  "is_permanent": false,
  "open_at_sec": 1731283200,
  "close_at_sec": 1733961600,
  "release_at_sec": 1731283200,
  "auto_audit": true,
  "auto_audit_max_amt": 100000,
  "payout_mode": 3,
  "payout_timing": 1,
  "claim_expire_sec": 86400,
  "settings": {
    "deposit_reward_type": 1,
    "bonus_mode": 1,
    "reward_schemas": [
      {
        "id": 1,
        "threshold": 10000,
        "amount": 1000
      }
    ]
  },
  "wagering_multiplier": 1,
  "wagering_base": 2,
  "show_in_list": true,
  "show_in_sidebar": true,
  "sort_weight": 100,
  "i18n": [
    {
      "lang": "zh-CN",
      "name": "首充送彩金",
      "rule": "首充满额赠送",
      "description": "新会员首次充值即可领取彩金",
      "image": "https://cdn.example.com/banner-zh.png"
    }
  ]
}
```

**请求字段：**

| 字段 | 类型 | 必填 | 规则与说明 |
|---|---|---|---|
| `type` | int32 | 是 | 活动类型：`1` 充值、`2` 洗码、`3` 推荐会员、`4` 幸运转盘、`5`～`14` 见枚举说明 |
| `status` | int32 | 否 | 活动状态；`0`/未传默认草稿；`1` 草稿、`2` 进行中、`3` 已结束 |
| `apply_mode` | int32 | 是 | 参与模式：`1` 自动、`2` 手动 |
| `section` | int32 | 是 | 活动所属：`1` 大厅、`2` 电子、`3` 体育 |
| `cycle_mode` | int32 | 否 | 循环模式，见[枚举说明](#九枚举说明) |
| `allowed_fixed_level` | int64[] | 否 | 允许的固定层级 ID |
| `allowed_auto_level` | int64[] | 否 | 允许的自动层级 ID |
| `allowed_clients` | int64[] | 否 | 允许的客户端类型，见枚举 |
| `allowed_game_ids` | int64[] | 否 | 允许的游戏 ID |
| `allowed_agent_ids` | int64[] | 否 | 允许的代理 ID |
| `apply_limit` | int64[] | 否 | 参与限制类型列表 |
| `claim_limit` | int64[] | 否 | 领取限制类型列表 |
| `is_permanent` | bool | 否 | 是否永久活动 |
| `open_at_sec` | int64 | 否 | 开始时间，Unix 秒 |
| `close_at_sec` | int64 | 否 | 结束时间，Unix 秒 |
| `release_at_sec` | int64 | 否 | 发布时间，Unix 秒 |
| `auto_audit` | bool | 否 | 是否自动审核 |
| `auto_audit_max_amt` | int64 | 否 | 自动审核金额上限（最小单位） |
| `payout_mode` | int32 | 否 | 派发模式，见枚举 |
| `payout_timing` | int32 | 否 | 派发时机，见枚举 |
| `claim_expire_sec` | int64 | 否 | 手动领取过期秒数 |
| `settings` | object | 否 | 活动配置，见 [ActivitySettings](#1-activitysettings)；类型为幸运转盘时需提供 `luck_wheel` |
| `wagering_multiplier` | int32 | 否 | 洗码倍数 |
| `wagering_base` | int32 | 否 | 洗码基数，见枚举 |
| `show_in_list` | bool | 否 | 是否在列表显示 |
| `show_in_sidebar` | bool | 否 | 是否在侧边栏显示 |
| `sort_weight` | int32 | 否 | 排序权重，数值越大越靠前（以服务端排序为准） |
| `create_uid` | int64 | 否 | 创建人 UID；通常可由服务端从登录态写入 |
| `i18n` | object[] | 否 | 多语言名称/规则/说明/图片，见 [ActivityI18nItem](#5-activityi18nitem) |

### 2. 创建活动响应

**成功响应示例：**

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 10001
  }
}
```

**响应字段：**

| 字段 | 类型 | 说明 |
|---|---|---|
| `data.id` | int64 | 新创建活动 ID |

---

## 四、编辑活动

**POST /activity/update**

仅 **草稿**（`status = 1`）可编辑；编辑后仍保持草稿。非草稿返回 `activity.notDraft`。`status` 入参忽略；`i18n` 按语言整批替换。

请求字段与创建活动基本一致，额外必填 `id`；不含 `create_uid` / `status`。

**请求示例：**

```http
POST /activity/update
Content-Type: application/json
Authorization: Bearer <access_token>
```

```json
{
  "id": 10001,
  "type": 1,
  "apply_mode": 1,
  "section": 1,
  "cycle_mode": 1,
  "is_permanent": false,
  "open_at_sec": 1731283200,
  "close_at_sec": 1733961600,
  "auto_audit": true,
  "auto_audit_max_amt": 100000,
  "payout_mode": 3,
  "payout_timing": 1,
  "settings": {
    "deposit_reward_type": 1,
    "bonus_mode": 1,
    "reward_schemas": [
      {
        "id": 1,
        "threshold": 10000,
        "amount": 1000
      }
    ]
  },
  "show_in_list": true,
  "sort_weight": 100,
  "i18n": [
    {
      "lang": "zh-CN",
      "name": "首充送彩金（改）",
      "rule": "首充满额赠送",
      "description": "新会员首次充值即可领取彩金",
      "image": "https://cdn.example.com/banner-zh.png"
    }
  ]
}
```

**成功响应示例：**

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 10001
  }
}
```

| 字段 | 类型 | 说明 |
|---|---|---|
| `data.id` | int64 | 已编辑的活动 ID |

---

## 五、活动详情

### 1. 活动详情请求参数

**GET /activity/detail**

**请求示例：**

```http
GET /activity/detail?id=10001
Authorization: Bearer <access_token>
```

**Query 参数：**

| 字段 | 类型 | 必填 | 规则与说明 |
|---|---|---|---|
| `id` | int64 | 是 | 活动 ID，必须大于 0 |

### 2. 活动详情响应

**成功响应示例：**

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 10001,
    "type": 1,
    "status": 1,
    "apply_mode": 1,
    "section": 1,
    "cycle_mode": 1,
    "allowed_fixed_level": [1, 2],
    "allowed_auto_level": [1],
    "allowed_clients": [1, 2],
    "allowed_game_ids": [],
    "allowed_agent_ids": [],
    "apply_limit": [],
    "claim_limit": [],
    "is_permanent": false,
    "open_at_sec": 1731283200,
    "close_at_sec": 1733961600,
    "release_at_sec": 1731283200,
    "auto_audit": true,
    "auto_audit_max_amt": 100000,
    "payout_mode": 3,
    "payout_timing": 1,
    "claim_expire_sec": 86400,
    "settings": {
      "deposit_reward_type": 1,
      "bonus_mode": 1,
      "reward_schemas": [
        {
          "id": 1,
          "threshold": 10000,
          "amount": 1000
        }
      ]
    },
    "wagering_multiplier": 1,
    "wagering_base": 2,
    "show_in_list": true,
    "show_in_sidebar": true,
    "sort_weight": 100,
    "create_uid": 1,
    "update_uid": 1,
    "version": 1,
    "created_at_sec": 1731280000,
    "updated_at_sec": 1731280000,
    "i18n": [
      {
        "lang": "zh-CN",
        "name": "首充送彩金",
        "rule": "首充满额赠送",
        "description": "新会员首次充值即可领取彩金",
        "image": "https://cdn.example.com/banner-zh.png"
      }
    ]
  }
}
```

**响应字段（**`data` **即 ActivityDetail）：**

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | int64 | 活动 ID |
| `type` | int32 | 活动类型 |
| `status` | int32 | 活动状态 |
| `apply_mode` | int32 | 参与模式 |
| `section` | int32 | 活动所属：`1` 大厅、`2` 电子、`3` 体育 |
| `cycle_mode` | int32 | 循环模式 |
| `allowed_fixed_level` | int64[] | 允许的固定层级 |
| `allowed_auto_level` | int64[] | 允许的自动层级 |
| `allowed_clients` | int64[] | 允许的客户端 |
| `allowed_game_ids` | int64[] | 允许的游戏 ID |
| `allowed_agent_ids` | int64[] | 允许的代理 ID |
| `apply_limit` | int64[] | 参与限制 |
| `claim_limit` | int64[] | 领取限制 |
| `is_permanent` | bool | 是否永久 |
| `open_at_sec` | int64 / null | 开始时间（秒） |
| `close_at_sec` | int64 / null | 结束时间（秒） |
| `release_at_sec` | int64 / null | 发布时间（秒） |
| `auto_audit` | bool | 是否自动审核 |
| `auto_audit_max_amt` | int64 | 自动审核金额上限 |
| `payout_mode` | int32 | 派发模式 |
| `payout_timing` | int32 | 派发时机 |
| `claim_expire_sec` | int64 | 领取过期秒数 |
| `settings` | object / null | 活动配置 |
| `wagering_multiplier` | int32 | 洗码倍数 |
| `wagering_base` | int32 | 洗码基数 |
| `show_in_list` | bool | 列表展示 |
| `show_in_sidebar` | bool | 侧边栏展示 |
| `sort_weight` | int32 | 排序权重 |
| `create_uid` | int64 | 创建人 |
| `update_uid` | int64 | 更新人 |
| `version` | int32 | 版本号 |
| `created_at_sec` | int64 | 创建时间（秒） |
| `updated_at_sec` | int64 | 更新时间（秒） |
| `i18n` | object[] | 多语言条目（含各语言图片与说明） |

---

## 六、活动列表

### 1. 活动列表查询参数

**GET /activity/list**

**请求示例：**

```http
GET /activity/list?page=1&page_size=20&name=首充&type=1&status=2&lang=zh-CN
Authorization: Bearer <access_token>
```

**Query 参数：**

| 字段 | 类型 | 必填 | 规则与说明 |
|---|---|---|---|
| `page` | int64 | 是 | 当前页码，从 1 开始 |
| `page_size` | int64 | 是 | 每页数量，1～100 |
| `name` | string | 否 | 活动名称模糊匹配（按 `lang` 对应 i18n.name） |
| `type` | int32 | 否 | 活动类型筛选 |
| `client` | int64 | 否 | 参与客户端筛选 |
| `status` | int32 | 否 | 活动状态筛选 |
| `lang` | string | 否 | 名称检索/展示语言，默认 `zh-CN` |

### 2. 活动列表响应

**成功响应示例：**

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "total": 1,
    "list": [
      {
        "id": 10001,
        "type": 1,
        "status": 2,
        "name": "首充送彩金",
        "image": "https://cdn.example.com/banner-zh.png",
        "allowed_clients": [1, 2],
        "open_at_sec": 1731283200,
        "close_at_sec": 1733961600,
        "release_at_sec": 1731283200,
        "sort_weight": 100,
        "show_in_list": true,
        "updated_at_sec": 1731280000,
        "auto_audit": true,
        "auto_audit_max_amt": 100000,
        "allowed_fixed_level": [1, 2],
        "allowed_auto_level": [1],
        "wagering_multiplier": 1,
        "allowed_game_ids": [],
        "is_permanent": false,
        "created_at_sec": 1731280000,
        "create_uid": 1,
        "update_uid": 1,
        "section": 1
      }
    ]
  }
}
```

**响应字段：**

| 字段 | 类型 | 说明 |
|---|---|---|
| `data.total` | int64 | 符合条件的总数 |
| `data.list` | object[] | 列表项，见下表 |

**列表项字段：**

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | int64 | 活动 ID |
| `type` | int32 | 活动类型 |
| `status` | int32 | 活动状态 |
| `name` | string | 按请求语言解析出的活动名称 |
| `image` | string | 按请求语言解析出的活动图片 URL |
| `allowed_clients` | int64[] | 允许的客户端 |
| `open_at_sec` | int64 / null | 开始时间（秒） |
| `close_at_sec` | int64 / null | 结束时间（秒） |
| `release_at_sec` | int64 / null | 发布时间（秒） |
| `sort_weight` | int32 | 排序权重 |
| `show_in_list` | bool | 是否列表展示 |
| `updated_at_sec` | int64 | 更新时间（秒） |
| `auto_audit` | bool | 是否自动审核 |
| `auto_audit_max_amt` | int64 | 自动审核金额上限 |
| `allowed_fixed_level` | int64[] | 固定层级 |
| `allowed_auto_level` | int64[] | 自动层级 |
| `wagering_multiplier` | int32 | 洗码倍数 |
| `allowed_game_ids` | int64[] | 游戏 ID |
| `is_permanent` | bool | 是否永久 |
| `created_at_sec` | int64 | 创建时间（秒） |
| `create_uid` | int64 | 创建人 |
| `update_uid` | int64 | 更新人 |
| `section` | int32 | 活动所属：`1` 大厅、`2` 电子、`3` 体育 |

---

## 七、关闭活动

**POST /activity/close**

将活动置为已结束（`status = 3`）。

**请求示例：**

```http
POST /activity/close
Content-Type: application/json
Authorization: Bearer <access_token>
```

```json
{
  "id": 10001
}
```

**请求字段：**

| 字段 | 类型 | 必填 | 规则与说明 |
|---|---|---|---|
| `id` | int64 | 是 | 活动 ID，必须大于 0 |

**成功响应示例：**

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 10001
  }
}
```

| 字段 | 类型 | 说明 |
|---|---|---|
| `data.id` | int64 | 已关闭的活动 ID |

---

## 八、发布活动

**POST /activity/publish**

仅 **草稿**（`status = 1`）活动可发布；发布成功后变为进行中（`status = 2`）。非草稿会返回 `activity.notDraft`。

**请求示例：**

```http
POST /activity/publish
Content-Type: application/json
Authorization: Bearer <access_token>
```

```json
{
  "id": 10001
}
```

**请求字段：**

| 字段 | 类型 | 必填 | 规则与说明 |
|---|---|---|---|
| `id` | int64 | 是 | 活动 ID，必须大于 0 |

**成功响应示例：**

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 10001
  }
}
```

| 字段 | 类型 | 说明 |
|---|---|---|
| `data.id` | int64 | 已发布的活动 ID |

---

## 九、活动枚举参数

**GET /activity/enum-params**

返回活动创建/编辑表单可用的枚举选项。`label` 来自 Core **front** 词典（catalog 种子 `activity.type.1`、`activity.applyMode.1` 等），按请求头 `X-Lang`（`zh-CN` / `zh-HK` / `en-US`，默认 `zh-CN`）翻译。不含 `OrderStatus`、`EventType`，且排除值为 `0` 的 Unspecified 项。

**请求示例：**

```http
GET /activity/enum-params
Authorization: Bearer <access_token>
X-Lang: zh-CN
```

无查询参数。

**成功响应示例：**

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "type": [
      { "value": 1, "label": "充值" },
      { "value": 2, "label": "洗码" },
      { "value": 3, "label": "推荐会员" },
      { "value": 4, "label": "幸运转盘" },
      { "value": 5, "label": "助力" },
      { "value": 14, "label": "签到" }
    ],
    "status": [
      { "value": 1, "label": "草稿" },
      { "value": 2, "label": "进行中" },
      { "value": 3, "label": "已结束" }
    ],
    "apply_mode": [
      { "value": 1, "label": "自动参与" },
      { "value": 2, "label": "手动参与" }
    ],
    "section": [
      { "value": 1, "label": "大厅" },
      { "value": 2, "label": "电子" },
      { "value": 3, "label": "体育" }
    ],
    "cycle_mode": [],
    "payout_mode": [],
    "payout_timing": [],
    "bonus_mode": [],
    "bet_claim_mode": [],
    "deposit_reward_type": [],
    "wheel_trigger_mode": [],
    "wagering_base": [],
    "client_type": [],
    "apply_limit_type": [],
    "claim_limit_type": [],
    "currency": "USDT",
    "timezone": "Asia/Tokyo"
  }
}
```

| 字段 | 类型 | 说明 |
|---|---|---|
| `type` | object[] | 活动类型 |
| `status` | object[] | 活动状态 |
| `apply_mode` | object[] | 参与模式 |
| `section` | object[] | 活动所属 |
| `cycle_mode` | object[] | 循环模式 |
| `payout_mode` | object[] | 奖励派发模式 |
| `payout_timing` | object[] | 奖励派发时间 |
| `bonus_mode` | object[] | 奖金模式 |
| `bet_claim_mode` | object[] | 投注奖励领取模式 |
| `deposit_reward_type` | object[] | 充值奖励类型 |
| `wheel_trigger_mode` | object[] | 幸运转盘触发维度 |
| `wagering_base` | object[] | 洗码计算基数 |
| `client_type` | object[] | 客户端类型（对应 `allowed_clients`） |
| `apply_limit_type` | object[] | 参与限制类型（对应 `apply_limit`） |
| `claim_limit_type` | object[] | 领取限制类型（对应 `claim_limit`） |
| `currency` | string | 货币 |
| `timezone` | string | 时区  |

每个数组元素：

| 字段 | 类型 | 说明 |
|---|---|---|
| `value` | int32 | 枚举值 |
| `label` | string | 当前语言文案（`i18n.TG` 翻译 front 词条） |

完整取值见 [十四、枚举说明](#十四枚举说明)。

---

## 十、活动订单列表

活动名称按请求头 `X-Lang` 查询，没有对应语言时为空字符串。列表不下发订单自增 ID，用 `order_no` 标识订单。时间为 Unix 秒，空时间不返回。

### 1. 活动订单列表查询参数

**GET /activity/order/list**

**请求示例：**

```http
GET /activity/order/list?page=1&page_size=20&order_no=AO123&activity_id=10001&member_id=20001&type=1&status=1&time_field=2&start_sec=1731283200&end_sec=1733961600
Authorization: Bearer <access_token>
X-Lang: zh-CN
```

**Query 参数：**

| 字段 | 类型 | 必填 | 规则与说明 |
|---|---|---|---|
| `page` | int64 | 是 | 当前页码，从 1 开始 |
| `page_size` | int64 | 是 | 每页数量，1～100 |
| `order_no` | string | 否 | 订单号，精确匹配 |
| `activity_id` | int64 | 否 | 活动 ID |
| `member_id` | int64 | 否 | 会员 ID |
| `type` | int32 | 否 | 活动类型 |
| `status` | int32 | 否 | 订单状态，见枚举 |
| `time_field` | int32 | 条件 | `1` 领取时间（`payout_at`），`2` 申请时间（`applied_at`） |
| `start_sec` | int64 | 否 | 时间范围开始，Unix 秒，含本秒 |
| `end_sec` | int64 | 否 | 时间范围结束，Unix 秒，含本秒 |

传了 `start_sec` 或 `end_sec` 时，`time_field` 必须是 `1` 或 `2`。只传 `time_field`、没有起止时间时不加时间条件。只传一端时只按这一端过滤。开始时间大于结束时间返回 `common.invalidParam`。

### 2. 活动订单列表响应

**成功响应示例：**

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "total": 1,
    "list": [
      {
        "order_no": "AO123",
        "activity_id": 10001,
        "activity_name": "首充送彩金",
        "type": 1,
        "member_id": 20001,
        "rewards_total": 100000,
        "channel_source": 1,
        "payout_mode": 3,
        "status": 1,
        "claim_expire_at_sec": 1731369600,
        "payout_at_sec": 1731283200,
        "applied_at_sec": 1731280000,
        "approved_at_sec": 1731281000,
        "approver_uid": 9
      }
    ]
  }
}
```

**响应字段：**

| 字段 | 类型 | 说明 |
|---|---|---|
| `data.total` | int64 | 符合条件的总数 |
| `data.list` | object[] | 列表项，见下表 |

**列表项字段：**

| 字段 | 类型 | 说明 |
|---|---|---|
| `order_no` | string | 订单号 |
| `activity_id` | int64 | 活动 ID |
| `activity_name` | string | 当前语言的活动名称，没有则为空 |
| `type` | int32 | 活动类型 |
| `member_id` | int64 | 会员 ID |
| `rewards_total` | int64 | 赠送金额，最小单位 |
| `channel_source` | int32 | 渠道来源，同客户端类型 |
| `payout_mode` | int32 | 领取方式 |
| `status` | int32 | 订单状态 |
| `claim_expire_at_sec` | int64 / 不返回 | 奖励过期时间 |
| `payout_at_sec` | int64 / 不返回 | 领取时间，对应计划或实际发放时间 |
| `applied_at_sec` | int64 / 不返回 | 申请时间 |
| `approved_at_sec` | int64 / 不返回 | 审批时间 |
| `approver_uid` | int64 | 审批人后台用户 ID，没有审批人为 `0` |

---

## 十一、活动订单审核列表

只返回 `status = 1`（待审核）的订单。活动名称同样按 `X-Lang` 查询。不下发订单自增 ID。

### 1. 活动订单审核列表查询参数

**GET /activity/order/audit/list**

**请求示例：**

```http
GET /activity/order/audit/list?page=1&page_size=20&activity_id=10001&member_id=20001&type=1&audit_status=1&time_field=2&start_sec=1731283200&end_sec=1733961600
Authorization: Bearer <access_token>
X-Lang: zh-CN
```

**Query 参数：**

| 字段 | 类型 | 必填 | 规则与说明 |
|---|---|---|---|
| `page` | int64 | 是 | 当前页码，从 1 开始 |
| `page_size` | int64 | 是 | 每页数量，1～100 |
| `order_no` | string | 否 | 订单号，精确匹配 |
| `activity_id` | int64 | 否 | 活动 ID |
| `member_id` | int64 | 否 | 会员 ID |
| `type` | int32 | 否 | 活动类型 |
| `audit_status` | int32 | 否 | `1` 人工审核，`2` 风控拦截；不传不限 |
| `time_field` | int32 | 条件 | `1` 领取时间，`2` 申请时间。规则与订单列表相同 |
| `start_sec` | int64 | 否 | 时间范围开始，Unix 秒，含本秒 |
| `end_sec` | int64 | 否 | 时间范围结束，Unix 秒，含本秒 |

订单状态固定为待审核，不能通过查询参数改成其他状态。

### 2. 活动订单审核列表响应

**成功响应示例：**

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "total": 1,
    "list": [
      {
        "order_no": "AO123",
        "activity_id": 10001,
        "activity_name": "首充送彩金",
        "type": 1,
        "member_id": 20001,
        "rewards_total": 100000,
        "channel_source": 1,
        "payout_mode": 1,
        "status": 1,
        "applied_at_sec": 1731280000
      }
    ]
  }
}
```

**列表项字段：**

| 字段 | 类型 | 说明 |
|---|---|---|
| `order_no` | string | 订单号，审核接口使用该字段 |
| `activity_id` | int64 | 活动 ID |
| `activity_name` | string | 当前语言的活动名称，没有则为空 |
| `type` | int32 | 活动类型 |
| `member_id` | int64 | 会员 ID |
| `rewards_total` | int64 | 赠送金额，最小单位 |
| `channel_source` | int32 | 渠道来源 |
| `payout_mode` | int32 | 领取方式 |
| `status` | int32 | 固定为 `1` 待审核 |
| `applied_at_sec` | int64 / 不返回 | 申请时间 |

---

## 十二、活动订单审核

**POST /activity/order/audit**

按订单号审核。`action = 1` 通过并启动领取或派发；`action = 2` 拒绝，不启动派发。只有待审核订单可以审。活动引擎未启动时，通过返回 `activity.engineDisabled`，拒绝仍可提交。

**请求示例：**

```http
POST /activity/order/audit
Content-Type: application/json
Authorization: Bearer <access_token>
```

```json
{
  "order_no": "AO123",
  "action": 1,
  "remark": "人工审核通过"
}
```

**请求参数：**

| 字段 | 类型 | 必填 | 规则与说明 |
|---|---|---|---|
| `order_no` | string | 是 | 订单号 |
| `action` | int32 | 是 | `1` 通过，`2` 拒绝 |
| `remark` | string | 否 | 审核备注。为空时通过默认「人工审核通过」，拒绝默认「人工审核拒绝」 |

**成功响应示例：**

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "order_no": "AO123"
  }
}
```

| 字段 | 类型 | 说明 |
|---|---|---|
| `data.order_no` | string | 已审核的订单号 |

---

## 十三、公共数据结构

### 1. ActivitySettings

| 字段 | 类型 | 说明 |
|---|---|---|
| `deposit_reward_type` | int32 | 充值奖励类型（可组合位掩码），见枚举 |
| `bet_claim_mode` | int32 | 投注奖励领取模式，见枚举 |
| `bonus_mode` | int32 | 奖金模式，见枚举 |
| `reward_schemas` | object[] | 通用奖励阶梯，见 [ActivityRewardSchema](#2-activityrewardschema) |
| `refer_config` | object | 推荐活动配置，见 [ActivityReferConfig](#3-activityreferconfig) |
| `luck_wheel` | object | 幸运转盘配置，见 [ActivityLuckWheelConfig](#4-activityluckwheelconfig)；`type=4` 时必填 |

### 2. ActivityRewardSchema

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | int64 | 档位 ID |
| `threshold` | int64 | 达标阈值（金额或次数等，按活动类型解释） |
| `amount` | int64 | 固定奖励金额（最小单位） |
| `rate` | int32 | 比例（按奖金模式使用） |
| `max_bonus` | int64 | 奖励上限 |
| `min_bonus` | int64 | 奖励下限 |
| `count` | int32 | 次数类配置（如转盘次数） |

### 3. ActivityReferConfig

| 字段 | 类型 | 说明 |
|---|---|---|
| `min_deposit_amount` | int64 | 被推荐人最低充值 |
| `min_bet_amount` | int64 | 被推荐人最低投注 |
| `ip_limit` | int32 | IP 限制 |
| `device_limit` | int32 | 设备限制 |

### 4. ActivityLuckWheelConfig

| 字段 | 类型 | 说明 |
|---|---|---|
| `wheel_trigger_mode` | int32 | `1` 仅投注、`2` 仅充值、`3` 投注或充值 |
| `bet_schemas` | object[] | 投注达标阶梯 |
| `deposit_schemas` | object[] | 充值达标阶梯 |

### 5. ActivityI18nItem

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `lang` | string | 是 | 语言码，如 `zh-CN`、`zh-HK`、`en-US` |
| `name` | string | 否 | 活动名称 |
| `rule` | string | 否 | 活动规则文案 |
| `description` | string | 否 | 活动说明，最长 1024 个字符 |
| `image` | string | 否 | 该语言对应的活动图片 URL |

---

## 十四、枚举说明

| 字段                            | 取值                    | 说明                                   |
| ----------------------------- | --------------------- | ------------------------------------ |
| `type(活动类型)`                  | 1～14                 | 充值 / 洗码 / 推荐会员 / 幸运转盘 / 助力 / 盲盒 / 新人彩金 / 全民代理 / 救援金 / 幸运注单 / VIP奖金 / 红包雨 / 图文(广告) / 签到 |
| `status(活动状态)`                | 1 / 2 / 3             | 草稿 / 进行中 / 已结束                       |
| `apply_mode(参与模式)`            | 1 / 2                 | 自动参与 / 手动参与                          |
| `section(活动所属)`               | 1 / 2 / 3             | 大厅 / 电子 / 体育                         |
| `cycle_mode(循环模式)`            | 0 / 1 / 2 / 3 / 4 / 5 | 未指定 / 不循环 / 每日 / 每周 / 每月 / 无限        |
| `payout_mode(奖励派发模式)`         | 1 / 2 / 3             | 手动领取过期作废 / 手动领取过期自动领 / 自动派发          |
| `payout_timing(奖励派发时间)`       | 1 / 2                 | 实时 / 次日                              |
| `bonus_mode(奖金模式)`            | 1～5                   | 固定多档 / 比例多档 / 固定阶梯最高 / 比例阶梯最高 / 单独奖励 |
| `bet_claim_mode(洗码`领取模式`)`    | 1 / 2 / 3             | 逐条 / 最高档 / 累计                        |
| `deposit_reward_type(`充值类型`)` | 1 / 2/ 3 / 4          | `1` 首充、`2` 次存、`3` 累计、`4` 单笔          |
| `wheel_trigger_mode`          | 1 / 2 / 3             | 仅投注 / 仅充值 / 投注或充值                    |
| `wagering_base`               | 0 / 1 / 2             | 未指定 / 本金+赠送 / 仅赠送                    |
| `allowed_clients` / `channel_source` | 1 / 2 / 3 / 4    | PC Web / 移动 Web / Android / iOS      |
| `status`（订单）                    | 1～8                  | 待审核 / 审核通过 / 审核拒绝 / 待领取 / 已领取 / 已过期 / 待派发 / 已派发 |
| `audit_status`                  | 1 / 2                 | 人工审核 / 风控拦截                        |
| `time_field`                    | 1 / 2                 | 领取时间 / 申请时间                        |
| `action`（订单审核）                 | 1 / 2                 | 通过 / 拒绝                             |
| `apply_limit` 元素              | 1 / 2                 | 同设备号不限次数 / 同设备 IP 限一次                |
| `claim_limit` 元素              | 1～5                   | 充值 / 绑卡 / 短信 / 邮箱 / 实名后才可领取          |

---

## 十五、业务错误

错误文案由 `X-Lang` 翻译，以下为 i18n key：

| key | 说明 |
|---|---|
| `activity.invalidType` | 活动类型无效 |
| `activity.invalidApplyMode` | 参与模式无效 |
| `activity.invalidSection` | 活动所属无效 |
| `activity.luckWheelSettingsRequired` | 幸运转盘配置不能为空 |
| `activity.notDraft` | 仅草稿活动可发布 |
| `activity.notFound` | 活动或订单不存在 |
| `activity.invalidId` | 活动 ID 无效 |
| `activity.orderNotPending` | 订单不在待审核状态 |
| `activity.engineDisabled` | 活动引擎未启动，无法通过审核 |

另可能返回通用错误如 `common.invalidParam`、`common.unauthorized` 等。

---

## 十六、前端对接注意事项

1. 全部活动接口需要登录 Token，并具备 catalog 中对应 API 权限。活动管理在「活动管理」分组，订单三个接口在「活动订单」分组：`/activity/order/list`、`/activity/order/audit/list`、`/activity/order/audit`。
2. 菜单路径为「优惠中心 → 活动中心」；前端路由组件约定为 `activity/index`。
3. 列表 `lang` 同时影响名称模糊检索与列表项 `name` 展示语言；未传默认 `zh-CN`。
4. 时间字段均为 **Unix 秒**，不要按会员模块的毫秒时间戳处理。
5. `type = 4`（幸运转盘）创建时必须提供 `settings.luck_wheel`，否则返回 `activity.luckWheelSettingsRequired`。
6. 发布接口仅对草稿生效；已进行中或已结束活动调用会失败。
7. 金额、阈值类字段使用 int64 最小单位，前端展示时再按币种精度换算。
8. 表单下拉选项请调用 `GET /activity/enum-params`，不要前端硬编码枚举文案；`label` 依赖 Core front 词典与 catalog 注册，切换语言时带上 `X-Lang` 重新拉取。
9. 订单列表和审核列表不下发订单自增 ID，审核请传 `order_no`。活动名称跟随 `X-Lang`，这两个列表没有 `lang` 查询参数。
