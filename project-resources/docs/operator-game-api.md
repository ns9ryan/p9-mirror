# operator-game API 文档

`operator-game-api` 提供运营商游戏管理系统的 HTTP API。

当前包含：

- 游戏管理
- 游戏分类
- 游戏渠道
- 游戏供应商
- 游戏货币
- 发布数据同步

---

## 一、接口约定

### 服务地址

默认开发端口：

```text
18003
```

实际 Host 和域名以部署环境为准。

### 鉴权

除 `GET /ping` 外，所有接口均需要后台登录鉴权和接口权限校验。

请求头：

```http
Authorization: Bearer <access_token>
```

Token 获取、刷新和过期处理规则见同目录的 [core-api.md](./core-api.md)。

### 中间件

大部分接口使用以下中间件：

- `Jwt` - JWT 认证
- `ActionLog` - 操作日志记录
- `Authority` - 权限校验

### 请求格式

- GET 接口使用 Query 参数
- POST 接口使用 JSON Body

```http
Content-Type: application/json
```

### 响应格式

统一响应格式：

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

---

## 二、Ping 接口

### 1. 健康检查

**请求**

```http
GET /ping
```

**响应**

```json
{
  "message": "pong"
}
```

---

## 三、游戏管理接口

前缀：`/admin/game`

### 1. 游戏列表

**请求**

```http
GET /admin/game/list?page=1&page_size=15&game_code=&name=&category_code=&provider_code=&channel_code=&status=
```

**查询参数**

| 参数 | 类型 | 必需 | 说明 |
| --- | --- | --- | --- |
| `page` | int64 | 否 | 页码（默认1） |
| `page_size` | int64 | 否 | 每页大小（默认15） |
| `game_code` | string | 否 | 游戏编码 |
| `name` | string | 否 | 游戏名称（查询 source_name_i18n.default） |
| `category_code` | string | 否 | 分类编码筛选 |
| `provider_code` | string | 否 | 供应商编码筛选 |
| `channel_code` | string | 否 | 渠道编码筛选 |
| `status` | int16 | 否 | 状态筛选：1启用/2禁用 |

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "source_id": 1001,
        "game_code": "slot_001",
        "name": "炫彩老虎机",
        "status": 1,
        "category_id": "slots",
        "provider_id": "provider_1",
        "channel_id": "channel_1",
        "category_name": "老虎机",
        "provider_name": "供应商1",
        "channel_name": "渠道1",
        "currency": "CNY",
        "provider_key": "provider_key_1",
        "image_url": "https://example.com/image.jpg",
        "sort_no": 1,
        "supports_embed": true,
        "supports_redirect": true,
        "created_at": 1630000000,
        "updated_at": 1630000000
      }
    ],
    "total": 100
  }
}
```

### 2. 游戏详情

**请求**

```http
GET /admin/game/get?id=1
```

**查询参数**

| 参数 | 类型 | 必需 | 说明 |
| --- | --- | --- | --- |
| `id` | int64 | 是 | 游戏ID |

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "source_id": 1001,
    "game_code": "slot_001",
    "name": "炫彩老虎机",
    "status": 1,
    "category_id": "slots",
    "provider_id": "provider_1",
    "channel_id": "channel_1",
    "category_name": "老虎机",
    "provider_name": "供应商1",
    "channel_name": "渠道1",
    "currency": "CNY",
    "provider_key": "provider_key_1",
    "image_url": "https://example.com/image.jpg",
    "sort_no": 1,
    "supports_embed": true,
    "supports_redirect": true,
    "created_at": 1630000000,
    "updated_at": 1630000000
  }
}
```

### 3. 更新游戏

**请求**

```http
POST /admin/game/update
Content-Type: application/json

{
  "id": 1,
  "name": "新游戏名称",
  "sort_no": 2,
  "status": 1
}
```

**请求体**

| 字段 | 类型 | 必需 | 说明 |
| --- | --- | --- | --- |
| `id` | int64 | 是 | 游戏ID |
| `name` | string | 否 | 游戏名称 |
| `sort_no` | int32 | 否 | 排序号 |
| `status` | int16 | 否 | 状态：1启用/2禁用 |

**响应**

同游戏详情响应

---

## 四、游戏分类接口

前缀：`/admin/game-category`

### 1. 分类列表

**请求**

```http
GET /admin/game-category/list?page=1&page_size=15&category_code=&status=
```

**查询参数**

| 参数 | 类型 | 必需 | 说明 |
| --- | --- | --- | --- |
| `page` | int64 | 否 | 页码（默认1） |
| `page_size` | int64 | 否 | 每页大小（默认15） |
| `category_code` | string | 否 | 分类编码 |
| `status` | int16 | 否 | 状态筛选：1启用/2禁用 |

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "category_code": "slots",
        "name": "老虎机",
        "sort_no": 1,
        "status": 1,
        "created_at": 1630000000,
        "updated_at": 1630000000
      }
    ],
    "total": 50
  }
}
```

### 2. 分类详情

**请求**

```http
GET /admin/game-category/get?id=1
```

**查询参数**

| 参数 | 类型 | 必需 | 说明 |
| --- | --- | --- | --- |
| `id` | int64 | 是 | 分类ID |

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "category_code": "slots",
    "name": "老虎机",
    "sort_no": 1,
    "status": 1,
    "created_at": 1630000000,
    "updated_at": 1630000000
  }
}
```

### 3. 更新分类

**请求**

```http
POST /admin/game-category/update
Content-Type: application/json

{
  "id": 1,
  "sort_no": 2,
  "status": 1
}
```

**请求体**

| 字段 | 类型 | 必需 | 说明 |
| --- | --- | --- | --- |
| `id` | int64 | 是 | 分类ID |
| `sort_no` | int32 | 否 | 排序号 |
| `status` | int16 | 否 | 状态：1启用/2禁用 |

**响应**

同分类详情响应

---

## 五、游戏渠道接口

前缀：`/admin/game-channel`

### 1. 渠道列表

**请求**

```http
GET /admin/game-channel/list?page=1&page_size=15&channel_code=&name=&status=
```

**查询参数**

| 参数 | 类型 | 必需 | 说明 |
| --- | --- | --- | --- |
| `page` | int64 | 否 | 页码（默认1） |
| `page_size` | int64 | 否 | 每页大小（默认15） |
| `channel_code` | string | 否 | 渠道编码 |
| `name` | string | 否 | 渠道名称 |
| `status` | int16 | 否 | 状态筛选：1启用/2禁用 |

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "channel_code": "channel_1",
        "provider_count": 10,
        "game_count": 100,
        "name": "渠道1",
        "sort_no": 1,
        "load_type": 1,
        "status": 1,
        "created_at": 1630000000,
        "updated_at": 1630000000
      }
    ],
    "total": 20
  }
}
```

### 2. 渠道详情

**请求**

```http
GET /admin/game-channel/get?id=1
```

**查询参数**

| 参数 | 类型 | 必需 | 说明 |
| --- | --- | --- | --- |
| `id` | int64 | 是 | 渠道ID |

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "channel_code": "channel_1",
    "provider_count": 10,
    "game_count": 100,
    "name": "渠道1",
    "sort_no": 1,
    "load_type": 1,
    "status": 1,
    "created_at": 1630000000,
    "updated_at": 1630000000
  }
}
```

### 3. 更新渠道

**请求**

```http
POST /admin/game-channel/update
Content-Type: application/json

{
  "id": 1,
  "sort_no": 2,
  "status": 1
}
```

**请求体**

| 字段 | 类型 | 必需 | 说明 |
| --- | --- | --- | --- |
| `id` | int64 | 是 | 渠道ID |
| `sort_no` | int32 | 否 | 排序号 |
| `status` | int16 | 否 | 状态：1启用/2禁用 |

**响应**

同渠道详情响应

---

## 六、游戏供应商接口

前缀：`/admin/game-provider`

### 1. 供应商列表

**请求**

```http
GET /admin/game-provider/list?page=1&page_size=15&provider_code=&name=&status=
```

**查询参数**

| 参数 | 类型 | 必需 | 说明 |
| --- | --- | --- | --- |
| `page` | int64 | 否 | 页码（默认1） |
| `page_size` | int64 | 否 | 每页大小（默认15） |
| `provider_code` | string | 否 | 供应商编码 |
| `name` | string | 否 | 供应商名称 |
| `status` | int16 | 否 | 状态筛选：1启用/2禁用 |

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "provider_code": "provider_1",
        "channel_code": "channel_1",
        "name": "供应商1",
        "channel_name": "渠道1",
        "logo_url": "https://example.com/logo.jpg",
        "sort_no": 1,
        "status": 1,
        "created_at": 1630000000,
        "updated_at": 1630000000
      }
    ],
    "total": 30
  }
}
```

### 2. 供应商详情

**请求**

```http
GET /admin/game-provider/get?id=1
```

**查询参数**

| 参数 | 类型 | 必需 | 说明 |
| --- | --- | --- | --- |
| `id` | int64 | 是 | 供应商ID |

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "provider_code": "provider_1",
    "channel_code": "channel_1",
    "name": "供应商1",
    "channel_name": "渠道1",
    "logo_url": "https://example.com/logo.jpg",
    "sort_no": 1,
    "status": 1,
    "created_at": 1630000000,
    "updated_at": 1630000000
  }
}
```

### 3. 更新供应商

**请求**

```http
POST /admin/game-provider/update
Content-Type: application/json

{
  "id": 1,
  "sort_no": 2,
  "status": 1
}
```

**请求体**

| 字段 | 类型 | 必需 | 说明 |
| --- | --- | --- | --- |
| `id` | int64 | 是 | 供应商ID |
| `sort_no` | int32 | 否 | 排序号 |
| `status` | int16 | 否 | 状态：1启用/2禁用 |

**响应**

同供应商详情响应

---

## 七、游戏货币接口

前缀：`/admin/game-currency`

### 1. 货币列表

**请求**

```http
GET /admin/game-currency/list?page=1&page_size=15&game_id=&currency_id=&status=
```

**查询参数**

| 参数 | 类型 | 必需 | 说明 |
| --- | --- | --- | --- |
| `page` | int64 | 否 | 页码（默认1） |
| `page_size` | int64 | 否 | 每页大小（默认15） |
| `game_id` | int64 | 否 | 游戏ID筛选 |
| `currency_id` | int64 | 否 | 货币ID筛选 |
| `status` | int16 | 否 | 状态筛选：1启用/2禁用 |

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "game_id": 1,
        "game_code": "slot_001",
        "game_name": "炫彩老虎机",
        "currency_id": 1,
        "currency_code": "CNY",
        "currency_name": "人民币",
        "status": 1,
        "created_at": 1630000000,
        "updated_at": 1630000000
      }
    ],
    "total": 200
  }
}
```

### 2. 货币详情

**请求**

```http
GET /admin/game-currency/get?id=1
```

**查询参数**

| 参数 | 类型 | 必需 | 说明 |
| --- | --- | --- | --- |
| `id` | int64 | 是 | 游戏货币ID |

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "game_id": 1,
    "game_code": "slot_001",
    "game_name": "炫彩老虎机",
    "currency_id": 1,
    "currency_code": "CNY",
    "currency_name": "人民币",
    "status": 1,
    "created_at": 1630000000,
    "updated_at": 1630000000
  }
}
```

### 3. 更新货币

**请求**

```http
POST /admin/game-currency/update
Content-Type: application/json

{
  "id": 1,
  "status": 1
}
```

**请求体**

| 字段 | 类型 | 必需 | 说明 |
| --- | --- | --- | --- |
| `id` | int64 | 是 | 游戏货币ID |
| `status` | int16 | 否 | 状态：1启用/2禁用 |

**响应**

同货币详情响应

---

## 八、发布数据同步接口

前缀：`/admin/publish-data`

### 1. 同步发布数据

**请求**

```http
POST /admin/publish-data/sync
Content-Type: application/json

{
  "op_code": "op_001"
}
```

**请求体**

| 字段 | 类型 | 必需 | 说明 |
| --- | --- | --- | --- |
| `op_code` | string | 是 | 分站代码 |

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "success": true,
    "message": "发布数据同步成功",
    "category_stat": {
      "total": 50,
      "success": 50,
      "failed": 0,
      "failed_reason": ""
    },
    "provider_stat": {
      "total": 30,
      "success": 30,
      "failed": 0,
      "failed_reason": ""
    },
    "channel_stat": {
      "total": 20,
      "success": 20,
      "failed": 0,
      "failed_reason": ""
    },
    "game_stat": {
      "total": 1000,
      "success": 1000,
      "failed": 0,
      "failed_reason": ""
    },
    "currency_stat": {
      "total": 100,
      "success": 100,
      "failed": 0,
      "failed_reason": ""
    },
    "created_at": 1630000000,
    "updated_at": 1630000000
  }
}
```

---

## 九、错误码

| 错误码 | 说明 |
| --- | --- |
| 0 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未授权（无Token或Token过期） |
| 403 | 禁止访问（无权限） |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

---

## 十、数据类型说明

### 状态值

| 值 | 说明 |
| --- | --- |
| 1 | 启用 |
| 2 | 禁用 |

### 时间戳

所有时间戳字段（`created_at`, `updated_at`）均为 Unix 时间戳（秒）。

---

## 十一、分页说明

分页接口使用 `page` 和 `page_size` 参数：

- `page`: 页码，从 1 开始（默认值 1）
- `page_size`: 每页大小（默认值 15）

响应数据包含 `list` 和 `total` 两个字段：

- `list`: 当前页的数据列表
- `total`: 总记录数

---
