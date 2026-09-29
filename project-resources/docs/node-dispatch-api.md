# node-dispatch API 文档

`node-dispatch-api` 提供 P9 总网调度中心相关 HTTP API。

当前已实现并纳入本文档的功能：

- 节点管理
- 调度任务查询

当前调度任务后台仅提供查询能力，不提供手动创建、重试、取消等操作。

---

## 一、接口约定

### 服务地址

默认开发端口：

```text
28001
```

实际 Host、域名和网关地址以部署环境为准。

例如直接访问开发服务器时：

```text
http://<host>:28001
```

### 鉴权

当前所有 `/admin/*` 接口均使用：

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

当前支持：

| 语言编码 | 说明 |
| --- | --- |
| `zh-CN` | 简体中文 |
| `zh-HK` | 繁体中文 |
| `en-US` | 英文 |

未传 `X-Lang` 时使用默认语言。

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
  "msg": "成功",
  "data": {}
}
```

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `code` | int | 成功固定为 `0` |
| `msg` | string | 当前请求语言对应的成功文案 |
| `data` | object | 接口业务数据 |

没有业务数据返回时：

```json
{
  "code": 0,
  "msg": "成功",
  "data": {}
}
```

### 错误响应

失败时 `code` 使用对应 HTTP 状态码，例如：

```json
{
  "code": 400,
  "msg": "请求参数格式错误"
}
```

开发和测试环境可能额外返回 `debug`：

```json
{
  "code": 400,
  "msg": "请求参数格式错误",
  "debug": {
    "error": "...",
    "stack": "...",
    "cause": "..."
  }
}
```

`debug` 仅用于开发调试，前端业务逻辑不要依赖该字段。

### 分页

分页列表统一使用：

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `page` | query | 是 | int64 | 页码，从 `1` 开始 |
| `page_size` | query | 是 | int64 | 每页数量，范围 `1-100` |

### 时间字段

本文档中的时间字段统一返回：

```text
Unix 毫秒时间戳
```

例如：

```text
1757836800000
```

可空时间字段没有值时返回 `null` 或不返回该字段，前端应按可空字段处理。

---

## 二、公共枚举

### 节点状态 `status`

| 值 | 说明 |
| --- | --- |
| `1` | 启用 |
| `2` | 停用 |

节点未传 `status` 创建时，后端默认：

```text
status = 1
```

### 节点在线状态 `online`

| 值 | 说明 |
| --- | --- |
| `true` | 当前节点已连接 node-dispatch |
| `false` | 当前节点未连接 node-dispatch |

`online` 是 node-dispatch 根据当前 WebSocket 连接实时判断的运行状态，不是数据库中的固定状态字段。

因此：

```text
status = 1
```

只表示节点允许使用，不代表当前一定在线。

### 调度任务状态 `status`

| 值 | 说明 |
| --- | --- |
| `1` | 待执行 |
| `2` | 执行中 |
| `3` | 成功 |
| `4` | 失败 |

---

## 三、接口索引

### 节点管理

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/admin/node/create` | 创建节点 |
| POST | `/admin/node/update` | 修改节点 |
| GET | `/admin/node/get` | 获取节点详情 |
| GET | `/admin/node/list` | 获取节点列表 |
| POST | `/admin/node/reset-auth-secret` | 重置节点认证密钥 |

### 调度任务

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/admin/dispatch-task/get` | 获取调度任务详情 |
| GET | `/admin/dispatch-task/list` | 获取调度任务列表 |

---

# 四、节点管理

## NodeInfo

节点信息结构：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int64 | 节点 ID |
| `code` | string | 节点全局唯一业务编码，由后端生成 |
| `name` | string | 节点名称 |
| `status` | int64 | 节点状态：`1` 启用，`2` 停用 |
| `online` | bool | 当前节点是否在线 |
| `last_seen_at` | int64 / null | 最近一次活动时间，Unix 毫秒时间戳 |
| `remark` | string / null | 运维备注 |
| `created_at` | int64 | 创建时间，Unix 毫秒时间戳 |
| `updated_at` | int64 | 更新时间，Unix 毫秒时间戳 |

示例：

```json
{
  "id": 1,
  "code": "NODE_1234567890ABCDEF1234567890ABCDEF",
  "name": "东京节点01",
  "status": 1,
  "online": true,
  "last_seen_at": 1759122000000,
  "remark": "东京生产节点",
  "created_at": 1759035600000,
  "updated_at": 1759122000000
}
```

---

## POST /admin/node/create

创建节点。

节点业务编码和节点认证密钥均由后端生成，前端不需要传。

### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `name` | json | 是 | string | 节点名称，非空，最大 100 个字符 |
| `status` | json | 否 | int64 | `1` 启用，`2` 停用；不传默认 `1` |
| `remark` | json | 否 | string | 运维备注，最大 1000 个字符 |

请求示例：

```json
{
  "name": "东京节点01",
  "status": 1,
  "remark": "东京生产节点"
}
```

### 响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int64 | 节点 ID |
| `code` | string | 后端生成的节点业务编码 |
| `auth_secret` | string | 节点认证密钥，仅本次创建返回 |

响应示例：

```json
{
  "code": 0,
  "msg": "成功",
  "data": {
    "id": 1,
    "code": "NODE_1234567890ABCDEF1234567890ABCDEF",
    "auth_secret": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
  }
}
```

### 前端特别注意

`auth_secret` **仅在创建成功时返回这一次**。

后端数据库保存的是认证密钥哈希，不保存可直接查询的明文认证密钥。

因此创建成功后，前端应明确提示管理员：

```text
请立即复制并保存节点认证密钥。
关闭后无法再次查看，如遗失只能重置。
```

不建议普通弹窗关闭后仍在页面中长期保存或展示该密钥。

---

## POST /admin/node/update

修改节点。

### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `id` | json | 是 | int64 | 节点 ID，大于 `0` |
| `name` | json | 否 | string | 节点名称，非空，最大 100 个字符 |
| `status` | json | 否 | int64 | `1` 启用，`2` 停用 |
| `remark` | json | 否 | string | 运维备注，最大 1000 个字符 |

除 `id` 外，至少需要传一个需要修改的字段。

请求示例：

```json
{
  "id": 1,
  "name": "东京节点01",
  "status": 2,
  "remark": "节点维护中"
}
```

响应：

```json
{
  "code": 0,
  "msg": "成功",
  "data": {}
}
```

### 业务规则

当节点状态修改为：

```text
status = 2
```

node-dispatch 会主动断开该节点当前的 WebSocket 连接。

因此节点停用后，其在线状态会变为：

```text
online = false
```

---

## GET /admin/node/get

获取节点详情。

### Query 示例

```text
/admin/node/get?id=1
```

### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `id` | query | 是 | int64 | 节点 ID，大于 `0` |

### 响应

响应中的 `data` 直接为 `NodeInfo`，不额外包一层 `node`。

```json
{
  "code": 0,
  "msg": "成功",
  "data": {
    "id": 1,
    "code": "NODE_1234567890ABCDEF1234567890ABCDEF",
    "name": "东京节点01",
    "status": 1,
    "online": true,
    "last_seen_at": 1759122000000,
    "remark": "东京生产节点",
    "created_at": 1759035600000,
    "updated_at": 1759122000000
  }
}
```

---

## GET /admin/node/list

获取节点分页列表。

### Query 示例

```text
/admin/node/list?page=1&page_size=20
```

按状态查询：

```text
/admin/node/list?page=1&page_size=20&status=1
```

按在线状态查询：

```text
/admin/node/list?page=1&page_size=20&online=true
```

关键字搜索：

```text
/admin/node/list?page=1&page_size=20&keyword=东京
```

### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `page` | query | 是 | int64 | 页码，从 `1` 开始 |
| `page_size` | query | 是 | int64 | 每页数量，范围 `1-100` |
| `keyword` | query | 否 | string | 搜索关键字，最大 100 个字符 |
| `status` | query | 否 | int64 | 节点状态：`1` 启用，`2` 停用 |
| `online` | query | 否 | bool | 是否在线：`true` / `false` |

### 搜索规则

`keyword` 当前匹配：

- 节点业务编码 `code`
- 节点名称 `name`

### 响应

```json
{
  "code": 0,
  "msg": "成功",
  "data": {
    "total": 1,
    "list": [
      {
        "id": 1,
        "code": "NODE_1234567890ABCDEF1234567890ABCDEF",
        "name": "东京节点01",
        "status": 1,
        "online": true,
        "last_seen_at": 1759122000000,
        "remark": "东京生产节点",
        "created_at": 1759035600000,
        "updated_at": 1759122000000
      }
    ]
  }
}
```

### 分站创建页面使用

分站创建第 6 步需要选择部署节点时，可以直接复用：

```text
GET /admin/node/list
```

建议至少传：

```text
status=1
```

例如：

```text
/admin/node/list?page=1&page_size=20&status=1
```

保存分站部署节点仍然调用 `platform-operator`：

```text
POST /admin/operator/node/save
```

也就是说：

```text
节点候选列表
    → node-dispatch API

分站选择了哪个节点
    → platform-operator API
```

两者职责不同。

---

## POST /admin/node/reset-auth-secret

重新生成节点认证密钥。

### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `id` | json | 是 | int64 | 节点 ID，大于 `0` |

请求示例：

```json
{
  "id": 1
}
```

### 响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `auth_secret` | string | 新认证密钥，仅本次返回 |

响应示例：

```json
{
  "code": 0,
  "msg": "成功",
  "data": {
    "auth_secret": "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"
  }
}
```

### 业务规则

认证密钥重置成功后：

1. 原认证密钥立即失效。
2. node-dispatch 会主动断开该节点当前连接。
3. node-agent 必须更新为新的 `auth_secret` 后才能重新认证连接。

与创建节点一样，新 `auth_secret` 只返回一次。

前端应在重置操作前进行确认，并在成功后提供复制操作。

---

# 五、调度任务

调度任务页面当前为**只读查询页面**。

当前后台不提供：

- 手动创建任务
- 重试任务
- 取消任务
- 修改任务

业务服务通过内部 RPC 创建调度任务，不通过总网后台 HTTP API 创建。

---

## DispatchTaskInfo

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `task_no` | string | 调度中心生成的全局任务编号 |
| `request_no` | string | 调用方生成的请求编号 |
| `target` | string | 任务目标服务 |
| `task_type` | string | 任务类型 |
| `params` | string | 任务参数，内容为 JSON 字符串 |
| `node_code` | string | 执行节点业务编码 |
| `status` | int64 | 任务状态：`1` 待执行、`2` 执行中、`3` 成功、`4` 失败 |
| `result` | string / null | 任务执行结果，内容为 JSON 字符串 |
| `error_message` | string / null | 任务执行失败原因 |
| `started_at` | int64 / null | 开始执行时间，Unix 毫秒时间戳 |
| `finished_at` | int64 / null | 执行结束时间，Unix 毫秒时间戳 |
| `created_at` | int64 | 创建时间，Unix 毫秒时间戳 |
| `updated_at` | int64 | 更新时间，Unix 毫秒时间戳 |

### JSON 字符串字段

当前：

```text
params
result
```

接口类型都是 `string`，字符串内部保存 JSON。

例如：

```json
{
  "params": "{\"operator_code\":\"OP_001\"}",
  "result": "{\"success\":true}"
}
```

前端如果需要结构化展示，可以自行解析 JSON。

建议详情页面同时支持：

- 格式化 JSON 展示
- 原始内容展示

如果内容无法解析为 JSON，则直接按普通字符串展示，避免页面报错。

---

## GET /admin/dispatch-task/get

按调度任务编号获取任务详情。

### Query 示例

```text
/admin/dispatch-task/get?task_no=TASK_001
```

### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `task_no` | query | 是 | string | 调度任务编号，非空，最大 64 个字符 |

后台 HTTP API 当前只使用：

```text
task_no
```

查询详情。

虽然内部 RPC 还支持通过 `request_no` 查询，但该能力当前没有暴露给前端。

### 响应

```json
{
  "code": 0,
  "msg": "成功",
  "data": {
    "task_no": "TASK_001",
    "request_no": "REQ_001",
    "target": "operator-base",
    "task_type": "CREATE_OPERATOR",
    "params": "{\"operator_code\":\"OP_001\"}",
    "node_code": "NODE_1234567890ABCDEF1234567890ABCDEF",
    "status": 3,
    "result": "{\"success\":true}",
    "error_message": null,
    "started_at": 1759122000000,
    "finished_at": 1759122001000,
    "created_at": 1759121999000,
    "updated_at": 1759122001000
  }
}
```

响应中的 `data` 直接为 `DispatchTaskInfo`，不额外包一层 `task`。

---

## GET /admin/dispatch-task/list

获取调度任务分页列表。

### Query 示例

```text
/admin/dispatch-task/list?page=1&page_size=20
```

按状态筛选：

```text
/admin/dispatch-task/list?page=1&page_size=20&status=4
```

按节点筛选：

```text
/admin/dispatch-task/list?page=1&page_size=20&node_code=NODE_1234567890ABCDEF1234567890ABCDEF
```

按任务类型筛选：

```text
/admin/dispatch-task/list?page=1&page_size=20&task_type=CREATE_OPERATOR
```

关键字搜索：

```text
/admin/dispatch-task/list?page=1&page_size=20&keyword=TASK_001
```

### 请求参数

| 字段 | 位置 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- | --- |
| `page` | query | 是 | int64 | 页码，从 `1` 开始 |
| `page_size` | query | 是 | int64 | 每页数量，范围 `1-100` |
| `keyword` | query | 否 | string | 搜索关键字，最大 64 个字符 |
| `task_type` | query | 否 | string | 任务类型，最大 64 个字符 |
| `status` | query | 否 | int64 | `1` 待执行、`2` 执行中、`3` 成功、`4` 失败 |
| `node_code` | query | 否 | string | 执行节点业务编码，最大 64 个字符 |

### 搜索规则

`keyword` 当前匹配：

- `task_no`
- `request_no`

### 响应

```json
{
  "code": 0,
  "msg": "成功",
  "data": {
    "total": 1,
    "list": [
      {
        "task_no": "TASK_001",
        "request_no": "REQ_001",
        "target": "operator-base",
        "task_type": "CREATE_OPERATOR",
        "params": "{\"operator_code\":\"OP_001\"}",
        "node_code": "NODE_1234567890ABCDEF1234567890ABCDEF",
        "status": 3,
        "result": "{\"success\":true}",
        "error_message": null,
        "started_at": 1759122000000,
        "finished_at": 1759122001000,
        "created_at": 1759121999000,
        "updated_at": 1759122001000
      }
    ]
  }
}
```

### 节点详情页面使用

节点详情页面如果需要展示该节点的调度任务记录，可以直接复用：

```text
GET /admin/dispatch-task/list
```

并传：

```text
node_code=<当前节点编码>
```

例如：

```text
/admin/dispatch-task/list?page=1&page_size=20&node_code=NODE_1234567890ABCDEF1234567890ABCDEF
```

不需要新增节点专属的任务列表接口。

---

# 六、菜单与按钮权限

当前总网后台菜单结构：

```text
调度中心
├── 节点管理
└── 调度任务
```

## 节点管理

路由：

```text
/dispatch/node
```

前端组件：

```text
dispatch/node/index
```

当前按钮权限：

| 权限 | 说明 |
| --- | --- |
| `node:detail` | 节点详情 |
| `node:create` | 新增节点 |
| `node:update` | 编辑节点 |
| `node:resetAuthSecret` | 重置认证密钥 |

## 调度任务

路由：

```text
/dispatch/task
```

前端组件：

```text
dispatch/task/index
```

当前按钮权限：

| 权限 | 说明 |
| --- | --- |
| `dispatchTask:detail` | 调度任务详情 |

调度任务当前为只读页面，因此没有新增、编辑、删除等按钮权限。

---

# 七、前端对接重点

1. 节点 `status` 和 `online` 是两个不同概念：
    - `status`：节点是否启用。
    - `online`：node-agent 当前是否通过 WebSocket 在线。

2. 创建节点时 `code` 和 `auth_secret` 均由后端生成，前端不要提供输入框。

3. 创建节点返回的 `auth_secret` 只显示一次，应提供明显的复制和保存提示。

4. 重置认证密钥后旧密钥立即失效，当前节点连接也会被断开。

5. 节点列表支持：
    - 关键字
    - 启停状态
    - 在线状态
    - 分页

6. 分站创建第 6 步选择节点时，可以直接复用节点列表接口，并使用 `status=1` 筛选启用节点。

7. 分站最终保存哪个部署节点，调用的是 `platform-operator` 的：
   ```text
   POST /admin/operator/node/save
   ```

8. 调度任务后台当前只有列表和详情，不提供人工创建、重试、取消和修改任务。

9. 调度任务的 `params`、`result` 是 JSON 内容的字符串，不是直接嵌套的 JSON object，前端展示时需要自行解析。

10. `result`、`error_message`、`started_at`、`finished_at` 都可能为空，前端需要按可空字段处理。

11. 调度任务列表的 `keyword` 同时搜索：
    - `task_no`
    - `request_no`

12. 节点详情中的任务记录直接复用调度任务列表，并通过 `node_code` 筛选，不需要单独接口。

13. 当前完整调度链路的集中联调测试暂未在本阶段执行，现阶段先用于前端接口对接；后续会统一验证节点创建、node-agent 连接、在线状态、停用断线、密钥重置以及任务状态流转。