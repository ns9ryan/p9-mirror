# 任务管理 API 文档

## 概述

任务管理模块提供任务配置的列表查询、修改以及会员任务奖励订单的查询功能。

**Base URL:** `http://localhost:38010`

## API 端点

### 1. 任务配置列表

**请求方法:** `GET`

**请求路径:** `/task/config/list`

**请求参数:**

| 参数名 | 类型 | 必需 | 描述 |
|--------|------|------|------|
| page | int64 | 是 | 分页页码，从1开始，默认1 |
| page_size | int64 | 是 | 分页大小，默认20，最大100 |
| category | string | 否 | 任务分类过滤，可选值：ONCE\|DAILY\|WEEKLY\|MONTHLY |
| status | int32 | 否 | 任务状态过滤，1=启用，2=禁用 |

**请求示例:**

```bash
curl -X GET "http://localhost:38010/task/config/list?page=1&page_size=20&category=DAILY&status=1" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**响应示例:**

```json
{
  "total": 5,
  "page": 1,
  "page_size": 20,
  "list": [
    {
      "id": 1,
      "operator_code": "op001",
      "task_code": "REGISTER",
      "category": "ONCE",
      "name": "注册任务",
      "name_i18n": "{}",
      "description": "完成注册获得奖励",
      "description_i18n": "{}",
      "timezone": "Asia/Shanghai",
      "status": 1,
      "completion_amount": 0,
      "reward_amount": 100000,
      "payout_mode": "MANUAL",
      "sort": 1,
      "created_at": 1700000000,
      "updated_at": 1700000000
    }
  ]
}
```

**响应字段说明:**

| 字段名 | 类型 | 描述 |
|--------|------|------|
| id | int64 | 任务配置ID |
| operator_code | string | 分站代码 |
| task_code | string | 任务编码 |
| category | string | 任务分类（ONCE/DAILY/WEEKLY/MONTHLY） |
| name | string | 任务名称 |
| name_i18n | string | 任务名称国际化JSON |
| description | string | 任务描述 |
| description_i18n | string | 任务描述国际化JSON |
| timezone | string | 分站时区 |
| status | int32 | 任务状态（1=启用，2=禁用） |
| completion_amount | int32 | 完成条件USD金额 |
| reward_amount | int64 | 奖励金额（最小货币单位） |
| payout_mode | string | 派发方式（MANUAL/AUTO） |
| sort | int64 | 任务排序 |
| created_at | int64 | 创建时间（unix秒） |
| updated_at | int64 | 更新时间（unix秒） |

---

### 2. 更新任务配置

**请求方法:** `POST`

**请求路径:** `/task/config/update`

**请求体参数:**

| 参数名 | 类型 | 必需 | 描述 |
|--------|------|------|------|
| id | int64 | 是 | 任务配置ID |
| timezone | string | 否 | 分站时区 |
| status | int32 | 否 | 任务状态（1=启用，2=禁用） |
| completion_amount | int32 | 否 | 完成条件USD金额 |
| reward_amount | int64 | 否 | 奖励金额（最小货币单位） |
| payout_mode | string | 否 | 派发方式（MANUAL/AUTO） |
| sort | int64 | 否 | 任务排序 |

**说明:** 只允许修改 `timezone`、`status`、`completion_amount`、`reward_amount`、`payout_mode`、`sort` 字段

**请求示例:**

```bash
curl -X POST "http://192.168.10.220:38010/task/config/update" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "id": 1,
    "status": 1,
    "sort": 2,
    "reward_amount": 150000
  }'
```

**响应示例:**

```json
{
  "task_config": {
    "id": 1,
    "operator_code": "op001",
    "task_code": "REGISTER",
    "category": "ONCE",
    "name": "注册任务",
    "name_i18n": "{}",
    "description": "完成注册获得奖励",
    "description_i18n": "{}",
    "timezone": "Asia/Shanghai",
    "status": 1,
    "completion_amount": 0,
    "reward_amount": 150000,
    "payout_mode": "MANUAL",
    "sort": 2,
    "created_at": 1700000000,
    "updated_at": 1700000000
  }
}
```

---

### 3. 会员任务奖励订单查询

**请求方法:** `GET`

**请求路径:** `/task/reward/list`

**请求参数:**

| 参数名 | 类型 | 必需 | 描述 |
|--------|------|------|------|
| page | int64 | 是 | 分页页码，从1开始，默认1 |
| page_size | int64 | 是 | 分页大小，默认20，最大100 |
| member_id | int64 | 否 | 会员ID |
| task_code | string | 否 | 任务编码 |
| category | string | 否 | 任务分类 |
| status | int32 | 否 | 订单状态（1=待领取\|2=待派发\|11=任务已拒绝\|12=派发被拒绝\|99=已派发） |
| order_no | string | 否 | 订单号 |
| created_at_start | int64 | 否 | 创建时间范围开始，unix秒 |
| created_at_end | int64 | 否 | 创建时间范围结束，unix秒 |

**请求示例:**

```bash
curl -X GET "http://192.168.10.220:38010/task/reward/list?page=1&page_size=20&member_id=123&status=1" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**响应示例:**

```json
{
  "total": 10,
  "page": 1,
  "page_size": 20,
  "list": [
    {
      "id": 1,
      "operator_code": "op001",
      "member_id": 123,
      "task_code": "REGISTER",
      "kafka_id": "kafka_123",
      "category": "ONCE",
      "period_key": "ALL",
      "payout_mode": "MANUAL",
      "amount": 100000,
      "status": 1,
      "order_no": "order_20231113_001",
      "wallet_transaction_id": "txn_123",
      "device_id": "device_123",
      "wallet_id": "wallet_123",
      "claim_ip": "127.0.0.1",
      "payout_at": 0,
      "created_at": 1700000000,
      "updated_at": 1700000000
    }
  ]
}
```

**响应字段说明:**

| 字段名 | 类型 | 描述 |
|--------|------|------|
| id | int64 | 订单ID |
| operator_code | string | 分站代码 |
| member_id | int64 | 会员ID |
| task_code | string | 任务编码 |
| kafka_id | string | Kafka消息ID |
| category | string | 任务分类 |
| period_key | string | 周期键（ALL/日期/周/月） |
| payout_mode | string | 派发方式（MANUAL/AUTO） |
| amount | int64 | 奖励金额（最小货币单位） |
| status | int32 | 订单状态（1=待领取\|2=待派发\|11=任务已拒绝\|12=派发被拒绝\|99=已派发） |
| order_no | string | 订单号 |
| wallet_transaction_id | string | 钱包交易ID |
| device_id | string | 领取设备ID |
| wallet_id | string | 领取钱包ID |
| claim_ip | string | 领取IP |
| payout_at | int64 | 派发时间（unix秒） |
| created_at | int64 | 创建时间（unix秒） |
| updated_at | int64 | 更新时间（unix秒） |

---

### 4. 更新会员任务奖励状态

**请求方法:** `POST`

**请求路径:** `/task/reward/update-status`

**请求体参数:**

| 参数名 | 类型 | 必需 | 描述 |
|--------|------|------|------|
| operator_code | string | 是 | 分站代码 |
| id | int64 | 是 | 奖励订单ID |
| status | int32 | 是 | 新状态，只允许为 2（待派发） |

**说明:**
- 只允许将订单状态从 `1`（待领取）改为 `2`（待派发）
- 其他状态转换不被允许
- 操作需要有对应分站的权限

**请求示例:**

```bash
curl -X POST "http://192.168.10.220:38010/task/reward/update-status" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "operator_code": "op001",
    "id": 1,
    "status": 2
  }'
```

**响应示例:**

```json
{
  "member_task_reward": {
    "id": 1,
    "operator_code": "op001",
    "member_id": 123,
    "task_code": "REGISTER",
    "kafka_id": "kafka_123",
    "category": "ONCE",
    "period_key": "ALL",
    "payout_mode": "MANUAL",
    "amount": 100000,
    "status": 2,
    "order_no": "order_20231113_001",
    "wallet_transaction_id": "txn_123",
    "device_id": "device_123",
    "wallet_id": "wallet_123",
    "claim_ip": "127.0.0.1",
    "payout_at": 0,
    "created_at": 1700000000,
    "updated_at": 1700000000
  }
}
```

**响应字段说明:**

返回更新后的完整 `MemberTaskReward` 对象，字段说明同上一个接口。

**错误情况:**

| 错误情况 | 错误码 | 说明 |
|---------|--------|------|
| 订单不存在 | 404 | 指定ID的订单不存在 |
| 订单状态不是1 | 400 | only_allow_from_status_1 - 只有状态为1的订单才能更新 |
| 新状态不是2 | 400 | only_allow_status_2 - 只允许更新为状态2 |
| 权限不足 | 403 | 分站代码不匹配或无权限操作 |
| 无效的Token | 401 | 认证失败 |

---

## 错误响应

所有 API 出错时返回统一错误格式：

```json
{
  "code": 400,
  "msg": "错误信息描述"
}
```

**常见错误码:**

| 错误码 | 说明 |
|--------|------|
| 400 | 请求参数错误 |
| 401 | 未授权（无效的Token） |
| 403 | 无权限访问 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

---

## 使用示例

### Python 示例

```python
import requests

BASE_URL = "http://localhost:38010"
TOKEN = "your_access_token"

headers = {
    "Authorization": f"Bearer {TOKEN}",
    "Content-Type": "application/json"
}

# 获取任务配置列表
response = requests.get(
    f"{BASE_URL}/task/config/list",
    params={
        "page": 1,
        "page_size": 20,
        "status": 1
    },
    headers=headers
)
print(response.json())

# 修改任务配置
response = requests.post(
    f"{BASE_URL}/task/config/update",
    json={
        "id": 1,
        "status": 1,
        "sort": 2,
        "reward_amount": 150000
    },
    headers=headers
)
print(response.json())

# 查询会员任务奖励订单
response = requests.get(
    f"{BASE_URL}/task/reward/list",
    params={
        "page": 1,
        "page_size": 20,
        "member_id": 123
    },
    headers=headers
)
print(response.json())

# 更新会员任务奖励状态（从1改为2）
response = requests.post(
    f"{BASE_URL}/task/reward/update-status",
    json={
        "operator_code": "op001",
        "id": 1,
        "status": 2
    },
    headers=headers
)
print(response.json())
```

### JavaScript/Node.js 示例

```javascript
const BASE_URL = "http://192.168.10.220:38010";
const TOKEN = "your_access_token";

const headers = {
  "Authorization": `Bearer ${TOKEN}`,
  "Content-Type": "application/json"
};

// 获取任务配置列表
fetch(`${BASE_URL}/task/config/list?page=1&page_size=20&status=1`, {
  method: "GET",
  headers: headers
})
  .then(res => res.json())
  .then(data => console.log(data));

// 修改任务配置
fetch(`${BASE_URL}/task/config/update`, {
  method: "POST",
  headers: headers,
  body: JSON.stringify({
    id: 1,
    status: 1,
    sort: 2,
    reward_amount: 150000
  })
})
  .then(res => res.json())
  .then(data => console.log(data));

// 更新会员任务奖励状态（从1改为2）
fetch(`${BASE_URL}/task/reward/update-status`, {
  method: "POST",
  headers: headers,
  body: JSON.stringify({
    operator_code: "op001",
    id: 1,
    status: 2
  })
})
  .then(res => res.json())
  .then(data => console.log(data));
```

---

## 数据字典

### 会员任务奖励订单状态（status）

| 状态值 | 状态名称 | 描述 |
|--------|---------|------|
| 1 | 待领取 | TaskStatusPendingClaim - 奖励订单已生成，会员可以领取 |
| 2 | 待派发 | TaskStatusPendingDispatch - 会员已领取，等待派发到钱包 |
| 11 | 任务已拒绝 | TaskStatusRejected - 任务被拒绝 |
| 12 | 派发被拒绝 | TaskStatusDispatchRejected - 派发时被拒绝 |
| 99 | 已派发 | TaskStatusDispatched - 奖励已成功派发到钱包 |

### 任务分类（category）

| 分类值 | 分类名称 | 描述 |
|--------|---------|------|
| ONCE | 一次性任务 | 只需完成一次 |
| DAILY | 每日任务 | 每天可以完成一次 |
| WEEKLY | 每周任务 | 每周可以完成一次 |
| MONTHLY | 每月任务 | 每月可以完成一次 |

### 派发方式（payout_mode）

| 派发方式 | 描述 |
|---------|------|
| MANUAL | 手动派发 - 由管理员手动触发派发 |
| AUTO | 自动派发 - 系统自动派发奖励 |

### 任务配置状态（status - TaskConfig）

| 状态值 | 状态名称 | 描述 |
|--------|---------|------|
| 1 | 启用 | 任务配置已启用，可以被会员完成 |
| 2 | 禁用 | 任务配置已禁用，会员无法完成 |

---

## 注意事项

1. **认证:** 所有API请求需要在 `Authorization` 请求头中提供有效的JWT Token
2. **分页:** 默认分页大小为20，最大为100
3. **时间戳:** 所有时间字段都是 Unix 秒时间戳
4. **金额单位:** 所有金额都是以最小货币单位表示（如美分、分等）
5. **排序:** 默认按 `sort` 字段升序排列
6. **时区:** timezone 字段遵循 IANA 时区标准（如 Asia/Shanghai）
