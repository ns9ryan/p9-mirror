# platform-operator API 文档

`platform-operator-api` 提供 P9 总网侧分站管理相关 HTTP API。

当前已实现并纳入本文档的功能：

- 分站管理
- 分站部署节点
- 分站档案
- 分站域名
- 分站管理员
- 基础资源分配汇总
- 语言分配
- 经营地区分配
- 代理子线路分配
- 游戏管理
- 游戏分类管理
- 游戏渠道管理
- 游戏提供商管理
- 游戏资源分配汇总
- 服务健康检查

---

## 一、接口约定

### 服务地址

默认开发端口：

```text
18002
```

实际 Host、域名和网关地址以部署环境为准。

例如直接访问开发服务器时：

```text
http://<host>:18002
```

### 鉴权

除 `GET /ping` 外，当前所有 `/admin/*` 接口均使用：

```text
Jwt, ActionLog, Authority
```

前端请求需要携带后台登录 Token：

```http
Authorization: Bearer <access_token>
```

Token 获取、刷新和过期处理规则参考同目录的 [core-api.md](./core-api.md)。

### 多语言

请求语言通过 `X-Lang` Header 指定：

```http
X-Lang: zh-CN
```

当前默认语言：

```text
zh-CN
```

错误文案等会根据当前请求语言返回。

### 请求格式

POST 接口统一使用 JSON：

```http
Content-Type: application/json
```

GET 接口参数通过 Query String 传递。

### 成功响应

成功响应统一格式：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {}
}
```

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `code` | int | 成功固定为 `0` |
| `msg` | string | 成功文案 |
| `data` | object | 接口业务数据 |

没有业务数据返回的接口，当前 `data` 为：

```json
{}
```

### 错误响应

失败时 `code` 使用对应 HTTP 状态码，例如：

```json
{
  "code": 400,
  "msg": "参数校验失败"
}
```

开发和测试环境可能额外返回 `debug`：

```json
{
  "code": 400,
  "msg": "参数校验失败",
  "debug": {
    "error": "...",
    "stack": "...",
    "cause": "..."
  }
}
```

`debug` 仅用于开发调试，前端业务逻辑不应依赖该字段。

### 分页

分页列表统一使用：

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `page` | query | 是 | int64 | 页码，从 `1` 开始 |
| `page_size` | query | 是 | int64 | 每页数量，范围 `1-100` |

### 时间字段

本文档中的时间字段统一为：

```text
Unix 毫秒时间戳
```

例如：

```text
1757836800000
```

### 公共枚举

#### 分站创建状态 `creation_status`

| 值 | 说明 |
| --- | --- |
| `1` | 草稿 |
| `2` | 已完成 |

#### 分站发布状态 `publish_status`

| 值 | 说明 |
| --- | --- |
| `1` | 未发布 |
| `2` | 发布中 |
| `3` | 已发布 |
| `4` | 发布失败 |

#### 分站状态 `status`

| 值 | 说明 |
| --- | --- |
| `1` | 正常 |
| `2` | 暂停 |
| `3` | 关闭 |

#### 部署节点状态 `node_status`

| 值 | 说明 |
| --- | --- |
| `1` | 启用 |
| `2` | 停用 |

节点状态和节点在线状态是两个不同概念：

- `node_status = 1` 表示节点允许使用。
- `node_online = true` 表示 node-agent 当前已经连接 node-dispatch。
- 保存分站部署节点时只要求节点启用。
- 正式发布分站时要求节点同时启用并在线。

#### 域名类型 `domain_type`

| 值 | 说明 |
| --- | --- |
| `1` | 分站后台 |
| `2` | 代理后台 |
| `3` | 会员 H5 |

#### 域名状态 `status`

| 值 | 说明 |
| --- | --- |
| `1` | 启用 |
| `2` | 停用 |

#### 管理员状态 `status`

| 值 | 说明 |
| --- | --- |
| `1` | 启用 |
| `2` | 停用 |

#### 代理子线路编码

| 编码 | 说明 |
| --- | --- |
| `CASH_PRODUCTION` | 现金正式线路 |
| `CASH_DEMO` | 现金试玩线路 |
| `CASH_TEST` | 现金测试线路 |
| `CREDIT_DEMO` | 信誉试玩线路 |
| `CREDIT_TEST` | 信誉测试线路 |

### 关联基础数据

创建或修改分站时使用的时区、结算币种，以及经营地区分配所使用的国家地区，都来自 `platform-base`。

前端可参考：

- [platform-base API 文档](./platform-base-api.md)

语言分配使用 Core 当前启用的语言，前端可参考：

- [core API 文档](./core-api.md)

部署节点来自 `node-dispatch`，前端可参考：

- [node-dispatch API 文档](./node-dispatch-api.md)

常用数据来源：

| 数据 | 建议接口 |
| --- | --- |
| 时区 | `GET /admin/timezone/list-all?status=1` |
| 结算币种 | `GET /admin/currency/list-all?status=1` |
| 国家地区 | `GET /admin/region/list-all?status=1` |
| 启用语言 | `GET /admin/i18n/lang/enabled` |
| 部署节点 | `GET /admin/node/list?page=1&page_size=100&status=1` |
| 代理子线路 | `GET /admin/operator/agent-line-allocation/list?operator_id=...` |

---

## 二、当前开发状态说明

当前前端对接时需要特别注意以下规则：

1. `POST /admin/operator/publish` 已经接入正式发布流程。调用成功表示发布任务已经提交并进入发布流程，不代表分站已经发布完成。
2. 发布开始后分站会进入 `publish_status = 2` 发布中状态，最终由调度任务执行结果更新为 `3` 已发布或 `4` 发布失败。
3. 发布结果以 node-dispatch 主动回调 platform-operator 为主；`POST /admin/operator/sync-publish-status` 用于主动查询调度任务状态并补偿同步发布结果。
4. `POST /admin/operator/complete` 完成创建前必须已经选择部署节点，并且节点必须处于启用状态；此阶段不要求节点在线。
5. `POST /admin/operator/complete` 当前不会统一检查档案、域名、语言、经营地区、代理子线路等配置是否全部完成，前端创建向导仍应按照页面流程控制调用时机。
6. 正式发布时，部署节点必须同时满足“启用”和“在线”。
7. 分站首次发布成功后会记录 `published_at`。一旦 `published_at` 已存在，部署节点、时区和结算币种均不能再修改。
8. 发布失败后，如果还没有首次发布成功，允许重新调整部署节点后再次发布。
9. `POST /admin/operator/delete` 当前只允许删除 `publish_status = 1` 未发布或 `publish_status = 4` 发布失败的分站。
10. 删除分站时，`platform-operator` 本地的语言、经营地区、代理子线路、域名、档案、管理员和分站数据会在同一个数据库事务中清理。
11. `platform-game` 游戏资源的跨服务清理当前仍保留 TODO。
12. `node-dispatch` 中已经保存的分站部署节点关系当前不在 `platform-operator` 删除事务范围内。

---

## 三、接口索引

- [服务健康检查](#get-ping)
    - [GET /ping](#get-ping)

- [分站管理](#operator)
    - [POST /admin/operator/create](#post-adminoperatorcreate)
    - [POST /admin/operator/update](#post-adminoperatorupdate)
    - [GET /admin/operator/get](#get-adminoperatorget)
    - [GET /admin/operator/list](#get-adminoperatorlist)
    - [POST /admin/operator/complete](#post-adminoperatorcomplete)
    - [POST /admin/operator/publish](#post-adminoperatorpublish)
    - [POST /admin/operator/sync-publish-status](#post-adminoperatorsync-publish-status)
    - [POST /admin/operator/delete](#post-adminoperatordelete)

- [分站部署节点](#operator-node)
    - [GET /admin/operator/node/get](#get-adminoperatornodeget)
    - [POST /admin/operator/node/save](#post-adminoperatornodesave)

- [分站档案](#operator-profile)
    - [POST /admin/operator/profile/create](#post-adminoperatorprofilecreate)
    - [POST /admin/operator/profile/update](#post-adminoperatorprofileupdate)
    - [GET /admin/operator/profile/get](#get-adminoperatorprofileget)

- [分站域名](#operator-domain)
    - [POST /admin/operator/domain/create](#post-adminoperatordomaincreate)
    - [POST /admin/operator/domain/update](#post-adminoperatordomainupdate)
    - [GET /admin/operator/domain/get](#get-adminoperatordomainget)
    - [GET /admin/operator/domain/list](#get-adminoperatordomainlist)
    - [POST /admin/operator/domain/delete](#post-adminoperatordomaindelete)

- [分站管理员](#operator-admin)
    - [POST /admin/operator/admin/create](#post-adminoperatoradmincreate)
    - [GET /admin/operator/admin/list](#get-adminoperatoradminlist)
    - [POST /admin/operator/admin/update](#post-adminoperatoradminupdate)
    - [POST /admin/operator/admin/resetPassword](#post-adminoperatoradminresetpassword)
    - [POST /admin/operator/admin/updateStatus](#post-adminoperatoradminupdatestatus)

- [基础资源分配](#basic-resource-allocation)
    - [GET /admin/operator/basic-resource-allocation/list](#get-adminoperatorbasic-resource-allocationlist)
    - [GET /admin/operator/language-allocation/list](#get-adminoperatorlanguage-allocationlist)
    - [POST /admin/operator/language-allocation/save](#post-adminoperatorlanguage-allocationsave)
    - [GET /admin/operator/region-allocation/list](#get-adminoperatorregion-allocationlist)
    - [POST /admin/operator/region-allocation/save](#post-adminoperatorregion-allocationsave)
    - [GET /admin/operator/agent-line-allocation/list](#get-adminoperatoragent-line-allocationlist)
    - [POST /admin/operator/agent-line-allocation/save](#post-adminoperatoragent-line-allocationsave)

- [游戏管理](#operator-game)
    - [POST /admin/operator/game/save-allocation](#post-adminoperatorgamesave-allocation)
    - [POST /admin/operator/game/batch-update-status](#post-adminoperatorgamebatch-update-status)
    - [GET /admin/operator/game/list](#get-adminoperatorgamelist)

- [游戏分类](#operator-game-category)
    - [POST /admin/operator/game-category/save-allocation](#post-adminoperatorgame-categorysave-allocation)
    - [POST /admin/operator/game-category/batch-update-status](#post-adminoperatorgame-categorybatch-update-status)
    - [GET /admin/operator/game-category/list](#get-adminoperatorgame-categorylist)

- [游戏渠道](#operator-game-channel)
    - [POST /admin/operator/game-channel/save-allocation](#post-adminoperatorgame-channelsave-allocation)
    - [POST /admin/operator/game-channel/batch-update-status](#post-adminoperatorgame-channelbatch-update-status)
    - [GET /admin/operator/game-channel/list](#get-adminoperatorgame-channellist)

- [游戏提供商](#operator-game-provider)
    - [POST /admin/operator/game-provider/save-allocation](#post-adminoperatorgame-providersave-allocation)
    - [POST /admin/operator/game-provider/batch-update-status](#post-adminoperatorgame-providerbatch-update-status)
    - [GET /admin/operator/game-provider/list](#get-adminoperatorgame-providerlist)

- [游戏资源分配](#operator-game-allocation)
    - [GET /admin/operator/game-allocation/list](#get-adminoperatorgame-allocationlist)

---

## 四、Ping

### GET /ping

服务健康检查，不需要 JWT。

请求：无。

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {}
}
```

---

## 五、分站管理

<a id="operator"></a>

### OperatorInfo

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int64 | 分站 ID |
| `code` | string | 分站全局唯一业务编码，由后端创建时生成 |
| `name` | string | 分站名称 |
| `timezone_code` | string | IANA 时区编码 |
| `settlement_currency_code` | string | 结算币种编码 |
| `creation_status` | int64 | 创建状态：`1` 草稿，`2` 已完成 |
| `publish_status` | int64 | 发布状态：`1` 未发布，`2` 发布中，`3` 已发布，`4` 发布失败 |
| `status` | int64 | 分站状态：`1` 正常，`2` 暂停，`3` 关闭 |
| `remark` | string / null | 总网内部备注 |
| `published_at` | int64 / null | 首次发布成功时间，Unix 毫秒时间戳 |
| `created_at` | int64 | 创建时间，Unix 毫秒时间戳 |
| `updated_at` | int64 | 更新时间，Unix 毫秒时间戳 |

示例：

```json
{
  "id": 1,
  "code": "OP_8D7091378B244D89A51FB102251489F1",
  "name": "A01",
  "timezone_code": "Asia/Tokyo",
  "settlement_currency_code": "USD",
  "creation_status": 1,
  "publish_status": 1,
  "status": 1,
  "remark": "内部测试分站",
  "published_at": null,
  "created_at": 1757836800000,
  "updated_at": 1757836800000
}
```

### POST /admin/operator/create

创建分站基础信息。

新建分站默认：

```text
creation_status = 1
publish_status  = 1
status          = 1  // 未传 status 时
```

分站业务编码 `code` 由后端自动生成，前端不需要传。

#### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `name` | json | 是 | string | 分站名称，非空，最大 100 个字符 |
| `timezone_code` | json | 是 | string | 时区编码，最大 64 个字符 |
| `settlement_currency_code` | json | 是 | string | 结算币种编码，最大 16 个字符 |
| `status` | json | 否 | int64 | `1` 正常，`2` 暂停，`3` 关闭 |
| `remark` | json | 否 | string | 总网内部备注，最大 1000 个字符 |

请求示例：

```json
{
  "name": "A01",
  "timezone_code": "Asia/Tokyo",
  "settlement_currency_code": "USD",
  "status": 1,
  "remark": "内部测试分站"
}
```

#### 业务规则

- `timezone_code` 必须是 `platform-base` 中存在且已启用的时区。
- `settlement_currency_code` 必须是 `platform-base` 中存在且已启用的货币。
- 结算币种编码会自动去除首尾空格并转换为大写。
- 创建成功后返回系统生成的分站 ID 和业务编码。

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 1,
    "code": "OP_8D7091378B244D89A51FB102251489F1"
  }
}
```

### POST /admin/operator/update

修改分站基础信息。

#### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `id` | json | 是 | int64 | 分站 ID，大于 `0` |
| `name` | json | 否 | string | 分站名称，非空，最大 100 个字符 |
| `timezone_code` | json | 否 | string | 时区编码，最大 64 个字符 |
| `settlement_currency_code` | json | 否 | string | 结算币种编码，最大 16 个字符 |
| `status` | json | 否 | int64 | `1` 正常，`2` 暂停，`3` 关闭 |
| `remark` | json | 否 | string | 总网内部备注，最大 1000 个字符 |

除 `id` 外，至少需要传一个需要修改的字段。

请求示例：

```json
{
  "id": 1,
  "name": "A01 新名称",
  "status": 2,
  "remark": "临时暂停"
}
```

#### 业务规则

- 修改时区时，新时区必须存在且处于启用状态。
- 修改结算币种时，新币种必须存在且处于启用状态。
- `publish_status = 2` 发布中时不能修改时区或结算币种。
- `published_at` 已经存在时，表示该分站已经首次发布成功，之后不能再修改时区或结算币种。
- 其他允许修改的字段仍按照后端当前规则处理。

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {}
}
```

### GET /admin/operator/get

按分站 ID 获取分站详情。

Query 示例：

```text
/admin/operator/get?id=1
```

#### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `id` | query | 是 | int64 | 分站 ID，大于 `0` |

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 1,
    "code": "OP_8D7091378B244D89A51FB102251489F1",
    "name": "A01",
    "timezone_code": "Asia/Tokyo",
    "settlement_currency_code": "USD",
    "creation_status": 1,
    "publish_status": 1,
    "status": 1,
    "remark": "内部测试分站",
    "published_at": null,
    "created_at": 1757836800000,
    "updated_at": 1757836800000
  }
}
```

### GET /admin/operator/list

获取分站管理列表。

Query 示例：

```text
/admin/operator/list?page=1&page_size=20&keyword=A01&creation_status=1&publish_status=1&status=1
```

#### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `page` | query | 是 | int64 | 页码，从 `1` 开始 |
| `page_size` | query | 是 | int64 | 每页数量，范围 `1-100` |
| `keyword` | query | 否 | string | 搜索关键字，匹配分站编码或名称，最大 100 个字符 |
| `creation_status` | query | 否 | int64 | `1` 草稿，`2` 已完成 |
| `publish_status` | query | 否 | int64 | `1` 未发布，`2` 发布中，`3` 已发布，`4` 发布失败 |
| `status` | query | 否 | int64 | `1` 正常，`2` 暂停，`3` 关闭 |

列表按创建时间倒序，同一创建时间下按 ID 倒序。

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "total": 1,
    "list": [
      {
        "id": 1,
        "code": "OP_8D7091378B244D89A51FB102251489F1",
        "name": "A01",
        "timezone_code": "Asia/Tokyo",
        "settlement_currency_code": "USD",
        "creation_status": 1,
        "publish_status": 1,
        "status": 1,
        "remark": "内部测试分站",
        "published_at": null,
        "created_at": 1757836800000,
        "updated_at": 1757836800000
      }
    ]
  }
}
```

### POST /admin/operator/complete

完成分站创建，将创建状态更新为：

```text
creation_status = 2
```

#### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `id` | json | 是 | int64 | 分站 ID，大于 `0` |

请求示例：

```json
{
  "id": 1
}
```

#### 业务规则

- 已经是 `creation_status = 2` 时再次调用会直接返回成功。
- 完成创建前必须已经选择部署节点。
- 当前部署节点必须处于启用状态。
- 完成创建阶段不要求节点当前在线。
- 当前后端不会在此接口统一校验档案、域名、语言、经营地区、代理子线路等配置是否全部完成。
- 前端应在创建向导各步骤完成后再调用该接口。

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {}
}
```

### POST /admin/operator/publish

发布分站。

发布采用异步调度流程。

调用成功仅表示：

```text
发布请求已接受
↓
分站进入 publish_status = 2
↓
调度任务开始执行
```

不代表已经发布完成。

#### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `id` | json | 是 | int64 | 分站 ID，大于 `0` |

请求示例：

```json
{
  "id": 1
}
```

#### 发布条件

只有满足以下条件才能发起发布：

- `creation_status = 2`
- `published_at = null`
- `publish_status = 1` 未发布，或 `publish_status = 4` 发布失败
- 已经选择部署节点
- 部署节点状态为启用
- 部署节点当前在线

以下情况不能发起发布：

```text
creation_status != 2
publish_status = 2
publish_status = 3
published_at != null
节点未选择
节点已停用
节点离线
```

#### 发布流程

```text
platform-operator
    ↓
node-dispatch SubmitTask
    ↓
WebSocket
    ↓
node-agent
    ↓
operator-base Initialize
    ↓
执行结果返回 node-dispatch
    ↓
node-dispatch 回调 platform-operator
    ↓
更新 publish_status
```

当前发布任务：

```text
target    = operator-base
task_type = CREATE_OPERATOR
```

任务参数当前主要包含：

```json
{
  "operator_code": "OP_8D7091378B244D89A51FB102251489F1"
}
```

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {}
}
```

发布成功响应后，前端应把当前分站视为：

```text
publish_status = 2
```

等待后续发布结果。

### POST /admin/operator/sync-publish-status

主动同步分站发布状态。

该接口用于发布结果的补偿查询，不建议作为独立人工操作按钮展示。

正常情况下发布结果由 node-dispatch 主动回调 platform-operator。

当前端发现分站长时间处于：

```text
publish_status = 2
```

时，可以调用该接口主动向 node-dispatch 查询实际调度任务状态。

#### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `id` | json | 是 | int64 | 分站 ID，大于 `0` |

请求示例：

```json
{
  "id": 1
}
```

#### 响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `operator_id` | int64 | 分站 ID |
| `operator_code` | string | 分站业务编码 |
| `publish_status` | int64 | 当前发布状态：`2` 发布中、`3` 已发布、`4` 发布失败 |

响应示例：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "operator_id": 1,
    "operator_code": "OP_8D7091378B244D89A51FB102251489F1",
    "publish_status": 3
  }
}
```

#### 状态处理

如果调度任务当前仍然是：

```text
1 待执行
2 执行中
```

则接口返回：

```text
publish_status = 2
```

如果调度任务已经：

```text
3 成功
```

则同步为：

```text
publish_status = 3
```

如果调度任务已经：

```text
4 失败
```

则同步为：

```text
publish_status = 4
```

如果分站本身已经是：

```text
publish_status = 3
```

或：

```text
publish_status = 4
```

则直接返回当前最终状态，不重复处理。

### POST /admin/operator/delete

删除分站。

#### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `id` | json | 是 | int64 | 分站 ID，大于 `0` |

请求示例：

```json
{
  "id": 1
}
```

#### 删除条件

只有下面两种发布状态允许删除：

| `publish_status` | 说明 | 是否允许删除 |
| --- | --- | --- |
| `1` | 未发布 | 是 |
| `2` | 发布中 | 否 |
| `3` | 已发布 | 否 |
| `4` | 发布失败 | 是 |

因此前端列表中的删除按钮建议只在：

```text
publish_status == 1 || publish_status == 4
```

时允许操作。

#### 删除范围

当前 `platform-operator` RPC 会在同一个本地数据库事务中依次删除：

```text
语言分配
→ 经营地区分配
→ 代理子线路分配
→ 分站域名
→ 分站档案
→ 分站管理员
→ 分站
```

任意一步失败时本地事务不会提交。

当前不属于这个本地事务的跨服务数据：

```text
platform-game 游戏资源分配
node-dispatch 分站部署节点关系
```

其中 `platform-game` 跨服务清理 RPC 当前仍保留 TODO。

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {}
}
```

---

## 六、分站部署节点

<a id="operator-node"></a>

分站部署节点用于指定该分站最终部署到哪个业务节点。

部署关系实际由 `node-dispatch` 维护，`platform-operator` 通过 RPC 访问，不在 `platform-operator` 数据库中保存完整副本。

创建分站时建议作为第 6 步处理：

```text
1. 基础信息
2. 档案信息
3. 管理员
4. 域名
5. 资源分配
6. 部署节点
```

候选节点列表由 `node-dispatch-api` 提供：

```text
GET /admin/node/list?page=1&page_size=100&status=1
```

### OperatorNodeInfo

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `node_code` | string | 节点业务编码 |
| `node_name` | string | 节点名称 |
| `node_status` | int64 | 节点状态：`1` 启用，`2` 停用 |
| `node_online` | bool | 节点当前是否在线 |
| `node_last_seen_at` | int64 / null | 节点最近一次活动时间，Unix 毫秒时间戳 |
| `node_remark` | string / null | 节点运维备注 |

### GET /admin/operator/node/get

获取分站当前部署节点。

Query 示例：

```text
/admin/operator/node/get?operator_id=1
```

#### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `operator_id` | query | 是 | int64 | 分站 ID，大于 `0` |

已经选择部署节点时：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "operator_node": {
      "node_code": "NODE_1234567890ABCDEF1234567890ABCDEF",
      "node_name": "东京节点01",
      "node_status": 1,
      "node_online": true,
      "node_last_seen_at": 1759122000000,
      "node_remark": "东京生产节点"
    }
  }
}
```

草稿阶段尚未选择部署节点时：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "operator_node": null
  }
}
```

### POST /admin/operator/node/save

保存分站部署节点。

#### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `operator_id` | json | 是 | int64 | 分站 ID，大于 `0` |
| `node_code` | json | 是 | string | 节点业务编码，非空，最大 64 个字符 |

请求示例：

```json
{
  "operator_id": 1,
  "node_code": "NODE_1234567890ABCDEF1234567890ABCDEF"
}
```

#### 业务规则

- `operator_id` 对应的分站必须存在。
- `node_code` 对应的节点必须存在。
- 只有 `node_status = 1` 的启用节点可以保存为部署节点。
- 保存部署节点时不要求节点当前在线。
- `publish_status = 2` 发布中时不能修改部署节点。
- `published_at` 已经存在时不能修改部署节点。
- 首次发布成功前，未发布或发布失败状态可以重新选择节点。
- 保存采用覆盖语义，同一个分站只保留一个当前部署节点。
- 跨服务实际通过分站全局业务编码 `operator_code` 建立部署关系，不依赖 `platform-operator` 的本地自增 ID。

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {}
}
```

---

## 七、分站档案

<a id="operator-profile"></a>

一个分站最多只有一条档案记录。

前端建议流程：

```text
GET /admin/operator/profile/get
        ↓
profile == null
        → 调用 create
profile != null
        → 调用 update
```

### OperatorProfileInfo

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int64 | 档案 ID |
| `operator_id` | int64 | 分站 ID |
| `company_name` | string / null | 公司名称 |
| `contact_name` | string / null | 主要联系人名称 |
| `contact_email` | string / null | 主要联系人邮箱 |
| `remark` | string / null | 总网内部档案备注 |
| `created_at` | int64 | 创建时间，Unix 毫秒时间戳 |
| `updated_at` | int64 | 更新时间，Unix 毫秒时间戳 |

### POST /admin/operator/profile/create

创建分站档案。

#### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `operator_id` | json | 是 | int64 | 分站 ID，大于 `0` |
| `company_name` | json | 否 | string | 公司名称，最大 200 个字符 |
| `contact_name` | json | 否 | string | 主要联系人名称，最大 100 个字符 |
| `contact_email` | json | 否 | string | 主要联系人邮箱，邮箱格式，最大 255 个字符 |
| `remark` | json | 否 | string | 总网内部档案备注，最大 1000 个字符 |

请求示例：

```json
{
  "operator_id": 1,
  "company_name": "Example Company",
  "contact_name": "Ryan",
  "contact_email": "ryan@example.com",
  "remark": "内部档案备注"
}
```

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 1
  }
}
```

### POST /admin/operator/profile/update

修改分站档案。

#### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `operator_id` | json | 是 | int64 | 分站 ID，大于 `0` |
| `company_name` | json | 否 | string | 公司名称，最大 200 个字符 |
| `contact_name` | json | 否 | string | 主要联系人名称，最大 100 个字符 |
| `contact_email` | json | 否 | string | 主要联系人邮箱，邮箱格式，最大 255 个字符 |
| `remark` | json | 否 | string | 总网内部档案备注，最大 1000 个字符 |

除 `operator_id` 外，至少需要传一个需要修改的字段。

请求示例：

```json
{
  "operator_id": 1,
  "contact_name": "Ryan Chen",
  "contact_email": "ryan.chen@example.com"
}
```

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {}
}
```

### GET /admin/operator/profile/get

获取分站档案。

Query 示例：

```text
/admin/operator/profile/get?operator_id=1
```

#### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `operator_id` | query | 是 | int64 | 分站 ID，大于 `0` |

已创建档案时：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "profile": {
      "id": 1,
      "operator_id": 1,
      "company_name": "Example Company",
      "contact_name": "Ryan",
      "contact_email": "ryan@example.com",
      "remark": "内部档案备注",
      "created_at": 1757836800000,
      "updated_at": 1757836800000
    }
  }
}
```

分站存在但尚未创建档案时：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "profile": null
  }
}
```

---

## 八、分站域名

<a id="operator-domain"></a>

域名只传主机名，不包含协议、路径和端口。

正确示例：

```text
bo.example.com
agent.example.com
m.example.com
```

不要传：

```text
https://bo.example.com
bo.example.com:443
https://bo.example.com/login
```

后端会自动去除首尾空格并转换为小写。

### OperatorDomainInfo

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int64 | 域名 ID |
| `operator_id` | int64 | 分站 ID |
| `domain_name` | string | 域名，不包含协议和端口 |
| `domain_type` | int64 | `1` 分站后台，`2` 代理后台，`3` 会员 H5 |
| `status` | int64 | `1` 启用，`2` 停用 |
| `remark` | string / null | 总网内部备注 |
| `created_at` | int64 | 创建时间，Unix 毫秒时间戳 |
| `updated_at` | int64 | 更新时间，Unix 毫秒时间戳 |

### 域名约束

当前后端约束包括：

- 同一个分站不能重复保存同一个域名。
- 同一个启用中的域名不能同时绑定多个分站。
- 同一个分站的同一种域名类型只能存在一个启用中的域名。
- `domain_name` 必须是合法 FQDN。
- 创建时不传 `status`，默认按启用状态处理。

### POST /admin/operator/domain/create

创建分站域名。

#### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `operator_id` | json | 是 | int64 | 分站 ID，大于 `0` |
| `domain_name` | json | 是 | string | FQDN，最大 253 个字符，不含协议和端口 |
| `domain_type` | json | 是 | int64 | `1` 分站后台，`2` 代理后台，`3` 会员 H5 |
| `status` | json | 否 | int64 | `1` 启用，`2` 停用 |
| `remark` | json | 否 | string | 总网内部备注，最大 1000 个字符 |

请求示例：

```json
{
  "operator_id": 1,
  "domain_name": "bo.example.com",
  "domain_type": 1,
  "status": 1,
  "remark": "分站后台域名"
}
```

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 1
  }
}
```

### POST /admin/operator/domain/update

修改分站域名。

#### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `id` | json | 是 | int64 | 域名 ID，大于 `0` |
| `domain_name` | json | 否 | string | FQDN，最大 253 个字符 |
| `domain_type` | json | 否 | int64 | `1` 分站后台，`2` 代理后台，`3` 会员 H5 |
| `status` | json | 否 | int64 | `1` 启用，`2` 停用 |
| `remark` | json | 否 | string | 总网内部备注，最大 1000 个字符 |

请求示例：

```json
{
  "id": 1,
  "status": 2,
  "remark": "暂时停用"
}
```

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {}
}
```

### GET /admin/operator/domain/get

按域名 ID 获取详情。

Query 示例：

```text
/admin/operator/domain/get?id=1
```

#### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `id` | query | 是 | int64 | 域名 ID，大于 `0` |

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 1,
    "operator_id": 1,
    "domain_name": "bo.example.com",
    "domain_type": 1,
    "status": 1,
    "remark": "分站后台域名",
    "created_at": 1757836800000,
    "updated_at": 1757836800000
  }
}
```

### GET /admin/operator/domain/list

获取分站域名分页列表。

Query 示例：

```text
/admin/operator/domain/list?page=1&page_size=20&operator_id=1&domain_type=1&status=1&keyword=example
```

#### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `page` | query | 是 | int64 | 页码，从 `1` 开始 |
| `page_size` | query | 是 | int64 | 每页数量，范围 `1-100` |
| `operator_id` | query | 否 | int64 | 分站 ID，大于 `0` |
| `keyword` | query | 否 | string | 域名关键字，最大 253 个字符 |
| `domain_type` | query | 否 | int64 | `1` 分站后台，`2` 代理后台，`3` 会员 H5 |
| `status` | query | 否 | int64 | `1` 启用，`2` 停用 |

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "total": 1,
    "list": [
      {
        "id": 1,
        "operator_id": 1,
        "domain_name": "bo.example.com",
        "domain_type": 1,
        "status": 1,
        "remark": "分站后台域名",
        "created_at": 1757836800000,
        "updated_at": 1757836800000
      }
    ]
  }
}
```

### POST /admin/operator/domain/delete

删除分站域名。

#### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `id` | json | 是 | int64 | 域名 ID，大于 `0` |

请求示例：

```json
{
  "id": 1
}
```

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {}
}
```

---

## 九、分站管理员

<a id="operator-admin"></a>

管理员账号挂在具体分站下。同一个分站内 `username` 唯一。列表和详情均不返回密码。

明文密码长度 `6-32`，入库前由服务端哈希；表字段长度留给哈希值，前端不要按 64 位明文理解。

可通过更新接口修改账号、密码、显示名称和启停状态。当前不提供按 ID 删除单条账号、详情查询接口。删除分站时会按 `operator_id` 一并清理该分站下全部管理员账号，无需单独调用删除接口。

### OperatorAdminInfo

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int64 | 管理员 ID |
| `operator_id` | int64 | 分站 ID |
| `operator_name` | string | 分站名称 |
| `username` | string | 账号 |
| `display_name` | string | 显示名称 |
| `status` | int64 | `1` 启用，`2` 停用 |
| `created_at` | int64 | 创建时间，Unix 毫秒时间戳 |
| `updated_at` | int64 | 更新时间，Unix 毫秒时间戳 |

### 管理员约束

- 同一个分站下账号不可重复。
- 创建时不传 `status`，默认按启用状态处理。
- 列表、创建响应均不包含 `password`。

### POST /admin/operator/admin/create

创建分站管理员。

#### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `operator_id` | json | 是 | int64 | 分站 ID，大于 `0` |
| `username` | json | 是 | string | 账号，最大 64 个字符 |
| `password` | json | 是 | string | 明文密码，长度 `6-32` |
| `display_name` | json | 是 | string | 显示名称，最大 100 个字符 |
| `status` | json | 否 | int64 | `1` 启用，`2` 停用 |

请求示例：

```json
{
  "operator_id": 1,
  "username": "opadmin",
  "password": "Passw0rd",
  "display_name": "分站管理员",
  "status": 1
}
```

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 1
  }
}
```

### GET /admin/operator/admin/list

获取分站管理员分页列表。

Query 示例：

```text
/admin/operator/admin/list?page=1&page_size=20&operator_id=1&status=1&keyword=admin
```

#### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `page` | query | 是 | int64 | 页码，从 `1` 开始 |
| `page_size` | query | 是 | int64 | 每页数量，范围 `1-100` |
| `operator_id` | query | 否 | int64 | 分站 ID，大于 `0` |
| `keyword` | query | 否 | string | 匹配账号或显示名称，最大 100 个字符 |
| `status` | query | 否 | int64 | `1` 启用，`2` 停用 |

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "total": 1,
    "list": [
      {
        "id": 1,
        "operator_id": 1,
        "operator_name": "示例分站",
        "username": "opadmin",
        "display_name": "分站管理员",
        "status": 1,
        "created_at": 1757836800000,
        "updated_at": 1757836800000
      }
    ]
  }
}
```

### POST /admin/operator/admin/update

更新分站管理员。至少需要传入 `username`、`password`、`display_name`、`status` 中的一项。不修改所属分站。

#### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `id` | json | 是 | int64 | 管理员 ID，大于 `0` |
| `username` | json | 否 | string | 账号，最大 64 个字符 |
| `password` | json | 否 | string | 明文密码，长度 `6-32` |
| `display_name` | json | 否 | string | 显示名称，最大 100 个字符 |
| `status` | json | 否 | int64 | `1` 启用，`2` 停用 |

请求示例：

```json
{
  "id": 1,
  "username": "opadmin",
  "password": "Passw0rd",
  "display_name": "分站管理员",
  "status": 1
}
```

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {}
}
```

### POST /admin/operator/admin/resetPassword

重置分站管理员密码。

#### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `id` | json | 是 | int64 | 管理员 ID，大于 `0` |
| `password` | json | 是 | string | 新明文密码，长度 `6-32` |

请求示例：

```json
{
  "id": 1,
  "password": "NewPass1"
}
```

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {}
}
```

### POST /admin/operator/admin/updateStatus

更新分站管理员启停状态。

#### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `id` | json | 是 | int64 | 管理员 ID，大于 `0` |
| `status` | json | 是 | int64 | `1` 启用，`2` 停用 |

请求示例：

```json
{
  "id": 1,
  "status": 2
}
```

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {}
}
```

---

## 十、基础资源分配

<a id="basic-resource-allocation"></a>

基础资源分配页面当前由三类资源组成：

```text
基础资源分配
├─ 语言
├─ 经营地区
└─ 代理子线路
```

同时提供一个汇总列表，用于展示每个分站当前已经分配的资源数量。

### 保存接口说明

语言、经营地区、代理子线路三个 `/save` 接口都按“当前最终选择结果”保存。

例如分站原来分配：

```text
A, B, C
```

请求保存：

```text
A, C, D
```

保存后最终结果就是：

```text
A, C, D
```

也就是说：

- 不再传的旧数据会被取消分配。
- 新传的数据会新增分配。
- 重复编码会被后端去重。
- 传空数组 `[]` 表示清空当前这一类全部分配。

因此前端保存时应提交当前页面完整选中集合，不是只提交本次新增项。

### GET /admin/operator/basic-resource-allocation/list

获取各分站基础资源分配汇总列表。

#### BasicResourceAllocationInfo

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `operator_id` | int64 | 分站 ID |
| `operator_code` | string | 分站业务编码 |
| `operator_name` | string | 分站名称 |
| `language_count` | int64 | 已分配语言数量 |
| `region_count` | int64 | 已分配经营地区数量 |
| `agent_line_count` | int64 | 已分配代理子线路数量 |

Query 示例：

```text
/admin/operator/basic-resource-allocation/list?page=1&page_size=20&keyword=A01
```

#### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `page` | query | 是 | int64 | 页码，从 `1` 开始 |
| `page_size` | query | 是 | int64 | 每页数量，范围 `1-100` |
| `keyword` | query | 否 | string | 匹配分站编码或名称，最大 100 个字符 |

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "total": 1,
    "list": [
      {
        "operator_id": 1,
        "operator_code": "OP_8D7091378B244D89A51FB102251489F1",
        "operator_name": "A01",
        "language_count": 2,
        "region_count": 3,
        "agent_line_count": 2
      }
    ]
  }
}
```

---

## 十一、语言分配

### LanguageAllocationInfo

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `language_code` | string | 语言编码 |
| `allocated_at` | int64 | 分配时间，Unix 毫秒时间戳 |

### GET /admin/operator/language-allocation/list

获取某个分站当前已经分配的语言。

Query 示例：

```text
/admin/operator/language-allocation/list?operator_id=1
```

#### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `operator_id` | query | 是 | int64 | 分站 ID，大于 `0` |

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "list": [
      {
        "language_code": "zh-CN",
        "allocated_at": 1757836800000
      },
      {
        "language_code": "en-US",
        "allocated_at": 1757836800000
      }
    ]
  }
}
```

说明：

- 此接口只返回当前已分配语言。
- 前端可使用 Core 的 `GET /admin/i18n/lang/enabled` 获取当前可选择的启用语言，再与本接口结果组合展示。

### POST /admin/operator/language-allocation/save

保存分站当前语言分配。

#### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `operator_id` | json | 是 | int64 | 分站 ID，大于 `0` |
| `language_codes` | json | 否 | string[] | 当前最终选中的语言编码，每项非空，单项最大 16 个字符 |

请求示例：

```json
{
  "operator_id": 1,
  "language_codes": [
    "zh-CN",
    "en-US"
  ]
}
```

清空全部语言分配：

```json
{
  "operator_id": 1,
  "language_codes": []
}
```

#### 业务规则

- 传入语言必须属于 Core 当前启用语言。
- 语言编码比较时不区分大小写，后端最终保存 Core 返回的标准语言编码。
- 保存采用全量覆盖语义。

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {}
}
```

---

## 十二、经营地区分配

### RegionAllocationInfo

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `region_code` | string | 国家地区编码 |
| `allocated_at` | int64 | 分配时间，Unix 毫秒时间戳 |

### GET /admin/operator/region-allocation/list

获取某个分站当前已经分配的经营地区。

Query 示例：

```text
/admin/operator/region-allocation/list?operator_id=1
```

#### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `operator_id` | query | 是 | int64 | 分站 ID，大于 `0` |

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "list": [
      {
        "region_code": "JP",
        "allocated_at": 1757836800000
      },
      {
        "region_code": "TH",
        "allocated_at": 1757836800000
      }
    ]
  }
}
```

说明：

- 此接口只返回当前已分配国家地区。
- 前端可使用 `platform-base` 的 `GET /admin/region/list-all?status=1` 获取当前可选择的启用国家地区。

### POST /admin/operator/region-allocation/save

保存分站当前经营地区分配。

#### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `operator_id` | json | 是 | int64 | 分站 ID，大于 `0` |
| `region_codes` | json | 否 | string[] | 当前最终选中的国家地区编码，每项固定 2 个字符 |

请求示例：

```json
{
  "operator_id": 1,
  "region_codes": [
    "JP",
    "TH"
  ]
}
```

清空全部经营地区分配：

```json
{
  "operator_id": 1,
  "region_codes": []
}
```

#### 业务规则

- 传入地区必须属于 `platform-base` 当前启用的国家地区。
- 国家地区编码会自动去除首尾空格并转换为大写。
- 保存采用全量覆盖语义。

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {}
}
```

---

## 十三、代理子线路分配

代理子线路目前不依赖基础数据表，后端使用固定编码。

当前支持：

| 编码 | 说明 |
| --- | --- |
| `CASH_PRODUCTION` | 现金正式线路 |
| `CASH_DEMO` | 现金试玩线路 |
| `CASH_TEST` | 现金测试线路 |
| `CREDIT_DEMO` | 信誉试玩线路 |
| `CREDIT_TEST` | 信誉测试线路 |

### AgentLineAllocationInfo

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `agent_line_code` | string | 代理子线路编码 |
| `allocated` | bool | 是否已经分配给当前分站 |
| `allocated_at` | int64 / null | 分配时间，未分配时为空 |

### GET /admin/operator/agent-line-allocation/list

获取某个分站的代理子线路分配状态。

与语言、经营地区接口不同，此接口会返回全部支持的代理子线路，并通过 `allocated` 标识当前是否已分配。

Query 示例：

```text
/admin/operator/agent-line-allocation/list?operator_id=1
```

#### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `operator_id` | query | 是 | int64 | 分站 ID，大于 `0` |

响应示例：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "list": [
      {
        "agent_line_code": "CASH_PRODUCTION",
        "allocated": true,
        "allocated_at": 1757836800000
      },
      {
        "agent_line_code": "CASH_DEMO",
        "allocated": true,
        "allocated_at": 1757836800000
      },
      {
        "agent_line_code": "CASH_TEST",
        "allocated": false,
        "allocated_at": null
      },
      {
        "agent_line_code": "CREDIT_DEMO",
        "allocated": false,
        "allocated_at": null
      },
      {
        "agent_line_code": "CREDIT_TEST",
        "allocated": false,
        "allocated_at": null
      }
    ]
  }
}
```

### POST /admin/operator/agent-line-allocation/save

保存分站当前代理子线路分配。

#### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `operator_id` | json | 是 | int64 | 分站 ID，大于 `0` |
| `agent_line_codes` | json | 否 | string[] | 当前最终选中的代理子线路编码，每项最大 32 个字符 |

请求示例：

```json
{
  "operator_id": 1,
  "agent_line_codes": [
    "CASH_PRODUCTION",
    "CASH_DEMO"
  ]
}
```

清空全部代理子线路分配：

```json
{
  "operator_id": 1,
  "agent_line_codes": []
}
```

#### 业务规则

- 只允许传当前后端支持的固定代理子线路编码。
- 编码必须使用上表中的标准值。
- 重复编码会自动去重。
- 保存采用全量覆盖语义。

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {}
}
```

---

## 十四、前端页面对接建议

当前接口和页面可以按下面方式对应：

| 页面 | 主要接口 |
| --- | --- |
| 分站列表 | `GET /admin/operator/list`、`GET /admin/operator/get`、`POST /admin/operator/publish`、`POST /admin/operator/sync-publish-status`、`POST /admin/operator/delete` |
| 创建分站-基础信息 | `POST /admin/operator/create`、`POST /admin/operator/update` |
| 创建分站-档案信息 | `GET /admin/operator/profile/get`、`POST /admin/operator/profile/create`、`POST /admin/operator/profile/update` |
| 创建分站-域名信息 | `/admin/operator/domain/*` |
| 创建分站-基础资源 | `/admin/operator/language-allocation/*`、`/admin/operator/region-allocation/*`、`/admin/operator/agent-line-allocation/*` |
| 创建分站-部署节点 | `GET /admin/node/list`、`GET /admin/operator/node/get`、`POST /admin/operator/node/save` |
| 创建完成 | `POST /admin/operator/complete` |
| 分站详情 | `GET /admin/operator/get` + 档案/域名/资源分配/部署节点接口 |
| 域名管理 | `/admin/operator/domain/*` |
| 管理员账号 | `/admin/operator/admin/*` |
| 基础资源分配列表 | `GET /admin/operator/basic-resource-allocation/list` |
| 基础资源分配-语言 | `/admin/operator/language-allocation/*` |
| 基础资源分配-经营地区 | `/admin/operator/region-allocation/*` |
| 基础资源分配-代理子线路 | `/admin/operator/agent-line-allocation/*` |

### 分站创建向导

当前建议创建流程：

```text
步骤1 基础信息
    ↓
POST /admin/operator/create

步骤2 档案信息
    ↓
profile create / update

步骤3 管理员
    ↓
admin create / update

步骤4 域名
    ↓
domain create / update

步骤5 资源分配
    ↓
language / region / agent-line

步骤6 部署节点
    ↓
GET  /admin/node/list
POST /admin/operator/node/save

完成创建
    ↓
POST /admin/operator/complete
```

保存部署节点时节点可以暂时离线，只需要处于启用状态。

真正点击“发布”时，节点必须启用并在线。

### 发布状态处理

建议前端发布流程：

```text
点击发布
    ↓
POST /admin/operator/publish
    ↓
publish_status = 2
    ↓
刷新分站数据
    ↓
正常等待回调更新状态
```

如果长时间仍为：

```text
publish_status = 2
```

可以调用：

```text
POST /admin/operator/sync-publish-status
```

主动补偿同步。

`sync-publish-status` 属于发布流程内部状态同步能力，当前不建议单独做“同步状态”按钮。

### 分站列表按钮权限

当前分站列表按钮权限：

| Permission | 说明 |
| --- | --- |
| `operator:detail` | 分站详情 |
| `operator:create` | 创建分站 |
| `operator:update` | 编辑分站 |
| `operator:publish` | 发布分站 |
| `operator:delete` | 删除分站 |

`sync-publish-status` 不单独配置按钮权限。

即使前端通过点击分站名称、图标或“查看”文字进入详情页，也应使用：

```text
operator:detail
```

控制详情访问入口。

### 基础资源分配

基础资源分配页面的三个 Tab：

```text
基础资源分配
├─ 语言
├─ 经营地区
└─ 代理子线路
```

不需要拆成三个独立左侧菜单页面。

---

## 十五、游戏管理

<a id="operator-game"></a>

### OperatorGameInfo

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int64 | 游戏 ID |
| `op_code` | string | 分站业务编码 |
| `game_code` | string | 游戏业务编码 |
| `name` | string | 游戏名称 |
| `status` | int32 | 状态：`1` 启用，`2` 停用 |
| `check_status` | int32 | 分配校验状态，按后端枚举返回 |
| `created_at` | int64 | 创建时间，Unix 毫秒时间戳 |
| `updated_at` | int64 | 更新时间，Unix 毫秒时间戳 |

### POST /admin/operator/game/save-allocation

保存分站游戏分配。

#### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `op_code` | json | 是 | string | 分站业务编码，最大 64 个字符 |
| `items` | json | 是 | SaveOperatorGameAllocationItem[] | 分配项列表 |

#### SaveOperatorGameAllocationItem

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `code` | string | 游戏业务编码 |
| `check_status` | int32 | 分配校验状态，按后端枚举返回 |

请求示例：

```json
{
  "op_code": "OP_8D7091378B244D89A51FB102251489F1",
  "items": [
    {
      "code": "GAME_001",
      "check_status": 1
    },
    {
      "code": "GAME_002",
      "check_status": 2
    }
  ]
}
```

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "total": 2,
    "created": 1,
    "deleted": 0,
    "exist": 1,
    "failed": 0
  }
}
```

### POST /admin/operator/game/batch-update-status

批量修改分站游戏状态。

#### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `ids` | json | 是 | int64[] | 游戏 ID 列表 |
| `status` | json | 是 | int32 | 新状态，`1` 启用，`2` 停用 |

请求示例：

```json
{
  "ids": [1, 2, 3],
  "status": 2
}
```

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "total": 3,
    "success": 3,
    "failed": 0
  }
}
```

### GET /admin/operator/game/list

获取分站游戏列表。

Query 示例：

```text
/admin/operator/game/list?page=1&page_size=20&op_code=OP_8D7091378B244D89A51FB102251489F1&status=1
```

#### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `page` | query | 是 | int32 | 页码，从 `1` 开始 |
| `page_size` | query | 是 | int32 | 每页数量，范围 `1-100` |
| `op_code` | query | 否 | string | 分站业务编码，最大 64 个字符 |
| `game_code` | query | 否 | string | 游戏业务编码，最大 64 个字符 |
| `status` | query | 否 | int32 | 状态，`1` 启用，`2` 停用 |

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "items": [
      {
        "id": 1,
        "op_code": "OP_8D7091378B244D89A51FB102251489F1",
        "game_code": "GAME_001",
        "name": "示例游戏1",
        "status": 1,
        "check_status": 1,
        "created_at": 1757836800000,
        "updated_at": 1757836800000
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 20
  }
}
```

---

## 十六、游戏分类

<a id="operator-game-category"></a>

### OperatorGameCategoryInfo

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int64 | 分类 ID |
| `op_code` | string | 分站业务编码 |
| `category_code` | string | 分类业务编码 |
| `status` | int32 | 状态：`1` 启用，`2` 停用 |
| `check_status` | int32 | 分配校验状态，按后端枚举返回 |
| `created_at` | int64 | 创建时间 |
| `updated_at` | int64 | 更新时间 |

### POST /admin/operator/game-category/save-allocation

保存分站游戏分类分配。

#### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `op_code` | json | 是 | string | 分站业务编码，最大 64 个字符 |
| `items` | json | 是 | SaveOperatorGameCategoryAllocationItem[] | 分配项列表 |

#### SaveOperatorGameCategoryAllocationItem

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `code` | string | 分类业务编码 |
| `check_status` | int32 | 分配校验状态，按后端枚举返回 |

请求示例：

```json
{
  "op_code": "OP_8D7091378B244D89A51FB102251489F1",
  "items": [
    {
      "code": "CATEGORY_001",
      "check_status": 1
    }
  ]
}
```

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "total": 1,
    "created": 1,
    "deleted": 0,
    "exist": 0,
    "failed": 0
  }
}
```

### POST /admin/operator/game-category/batch-update-status

批量修改分站游戏分类状态。

#### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `ids` | json | 是 | int64[] | 分类 ID 列表 |
| `status` | json | 是 | int32 | 新状态，`1` 启用，`2` 停用 |

请求示例：

```json
{
  "ids": [1, 2],
  "status": 2
}
```

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "total": 2,
    "success": 2,
    "failed": 0
  }
}
```

### GET /admin/operator/game-category/list

获取分站游戏分类列表。

Query 示例：

```text
/admin/operator/game-category/list?page=1&page_size=20&op_code=OP_8D7091378B244D89A51FB102251489F1&status=1
```

#### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `page` | query | 是 | int32 | 页码，从 `1` 开始 |
| `page_size` | query | 是 | int32 | 每页数量，范围 `1-100` |
| `op_code` | query | 否 | string | 分站业务编码，最大 64 个字符 |
| `category_code` | query | 否 | string | 分类业务编码，最大 64 个字符 |
| `status` | query | 否 | int32 | 状态，`1` 启用，`2` 停用 |

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "items": [
      {
        "id": 1,
        "op_code": "OP_8D7091378B244D89A51FB102251489F1",
        "category_code": "CATEGORY_001",
        "status": 1,
        "check_status": 1,
        "created_at": 1757836800000,
        "updated_at": 1757836800000
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 20
  }
}
```

---

## 十七、游戏渠道

<a id="operator-game-channel"></a>

### OperatorGameChannelInfo

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int64 | 渠道 ID |
| `op_code` | string | 分站业务编码 |
| `channel_code` | string | 渠道业务编码 |
| `status` | int32 | 状态：`1` 启用，`2` 停用 |
| `check_status` | int32 | 分配校验状态，按后端枚举返回 |
| `created_at` | int64 | 创建时间 |
| `updated_at` | int64 | 更新时间 |

### POST /admin/operator/game-channel/save-allocation

保存分站游戏渠道分配。

#### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `op_code` | json | 是 | string | 分站业务编码，最大 64 个字符 |
| `items` | json | 是 | SaveOperatorGameChannelAllocationItem[] | 分配项列表 |

#### SaveOperatorGameChannelAllocationItem

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `code` | string | 渠道业务编码 |
| `check_status` | int32 | 分配校验状态，按后端枚举返回 |

请求示例：

```json
{
  "op_code": "OP_8D7091378B244D89A51FB102251489F1",
  "items": [
    {
      "code": "CHANNEL_001",
      "check_status": 1
    }
  ]
}
```

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "total": 1,
    "created": 1,
    "deleted": 0,
    "exist": 0,
    "failed": 0
  }
}
```

### POST /admin/operator/game-channel/batch-update-status

批量修改分站游戏渠道状态。

#### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `ids` | json | 是 | int64[] | 渠道 ID 列表 |
| `status` | json | 是 | int32 | 新状态，`1` 启用，`2` 停用 |

请求示例：

```json
{
  "ids": [1, 2],
  "status": 2
}
```

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "total": 2,
    "success": 2,
    "failed": 0
  }
}
```

### GET /admin/operator/game-channel/list

获取分站游戏渠道列表。

Query 示例：

```text
/admin/operator/game-channel/list?page=1&page_size=20&op_code=OP_8D7091378B244D89A51FB102251489F1&status=1
```

#### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `page` | query | 是 | int32 | 页码，从 `1` 开始 |
| `page_size` | query | 是 | int32 | 每页数量，范围 `1-100` |
| `op_code` | query | 否 | string | 分站业务编码，最大 64 个字符 |
| `channel_code` | query | 否 | string | 渠道业务编码，最大 64 个字符 |
| `status` | query | 否 | int32 | 状态，`1` 启用，`2` 停用 |

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "items": [
      {
        "id": 1,
        "op_code": "OP_8D7091378B244D89A51FB102251489F1",
        "channel_code": "CHANNEL_001",
        "status": 1,
        "check_status": 1,
        "created_at": 1757836800000,
        "updated_at": 1757836800000
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 20
  }
}
```

---

## 十八、游戏提供商

<a id="operator-game-provider"></a>

### OperatorGameProviderInfo

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int64 | 提供商 ID |
| `op_code` | string | 分站业务编码 |
| `provider_code` | string | 提供商业务编码 |
| `status` | int32 | 状态：`1` 启用，`2` 停用 |
| `check_status` | int32 | 分配校验状态，按后端枚举返回 |
| `created_at` | int64 | 创建时间 |
| `updated_at` | int64 | 更新时间 |

### POST /admin/operator/game-provider/save-allocation

保存分站游戏提供商分配。

#### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `op_code` | json | 是 | string | 分站业务编码，最大 64 个字符 |
| `items` | json | 是 | SaveOperatorGameProviderAllocationItem[] | 分配项列表 |

#### SaveOperatorGameProviderAllocationItem

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `code` | string | 提供商业务编码 |
| `check_status` | int32 | 分配校验状态，按后端枚举返回 |

请求示例：

```json
{
  "op_code": "OP_8D7091378B244D89A51FB102251489F1",
  "items": [
    {
      "code": "PROVIDER_001",
      "check_status": 1
    }
  ]
}
```

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "total": 1,
    "created": 1,
    "deleted": 0,
    "exist": 0,
    "failed": 0
  }
}
```

### POST /admin/operator/game-provider/batch-update-status

批量修改分站游戏提供商状态。

#### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `ids` | json | 是 | int64[] | 提供商 ID 列表 |
| `status` | json | 是 | int32 | 新状态，`1` 启用，`2` 停用 |

请求示例：

```json
{
  "ids": [1, 2],
  "status": 2
}
```

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "total": 2,
    "success": 2,
    "failed": 0
  }
}
```

### GET /admin/operator/game-provider/list

获取分站游戏提供商列表。

Query 示例：

```text
/admin/operator/game-provider/list?page=1&page_size=20&op_code=OP_8D7091378B244D89A51FB102251489F1&status=1
```

#### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `page` | query | 是 | int32 | 页码，从 `1` 开始 |
| `page_size` | query | 是 | int32 | 每页数量，范围 `1-100` |
| `op_code` | query | 否 | string | 分站业务编码，最大 64 个字符 |
| `provider_code` | query | 否 | string | 提供商业务编码，最大 64 个字符 |
| `status` | query | 否 | int32 | 状态，`1` 启用，`2` 停用 |

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "items": [
      {
        "id": 1,
        "op_code": "OP_8D7091378B244D89A51FB102251489F1",
        "provider_code": "PROVIDER_001",
        "status": 1,
        "check_status": 1,
        "created_at": 1757836800000,
        "updated_at": 1757836800000
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 20
  }
}
```

---

## 十九、游戏资源分配

<a id="operator-game-allocation"></a>

### OperatorGameAllocationInfo

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int64 | 分配记录 ID |
| `op_code` | string | 分站业务编码 |
| `op_name` | string | 分站名称 |
| `game_count` | int32 | 分配的游戏数量 |
| `game_category_count` | int32 | 分配的游戏分类数量 |
| `game_provider_count` | int32 | 分配的游戏提供商数量 |
| `game_channel_count` | int32 | 分配的游戏渠道数量 |
| `updated_at` | int64 | 更新时间 |

### GET /admin/operator/game-allocation/list

获取游戏资源分配汇总列表。

Query 示例：

```text
/admin/operator/game-allocation/list?page=1&page_size=20&op_code=OP_8D7091378B244D89A51FB102251489F1&status=1
```

#### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `page` | query | 是 | int32 | 页码，从 `1` 开始 |
| `page_size` | query | 是 | int32 | 每页数量，范围 `1-100` |
| `op_code` | query | 否 | string | 分站业务编码，最大 64 个字符 |
| `status` | query | 否 | int32 | 状态过滤 |

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "code": 0,
    "message": "success",
    "items": [
      {
        "id": 1,
        "op_code": "OP_8D7091378B244D89A51FB102251489F1",
        "op_name": "示例分站",
        "game_count": 10,
        "game_category_count": 5,
        "game_provider_count": 3,
        "game_channel_count": 4,
        "updated_at": 1757836800000
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 20
  }
}
```