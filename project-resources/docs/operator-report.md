# Operator Report Service API 文档

## 概述

Operator Report Service 提供会员分析相关的API接口，包括会员增长趋势统计和会员概览等功能。

---

## API 列表

### 1. 会员增长趋势

**接口**: `POST /operator-report/dashboard/get-member-growth-trend`

**描述**: 获取会员增长趋势数据，支持按日期或小时维度统计注册数、登录数、留存率等。

#### 请求参数

```json
{
  "start_date": "2026-10-15",           // 开始日期 (YYYY-MM-DD)，必须
  "end_date": "2026-10-22",             // 结束日期 (YYYY-MM-DD)，必须
  "granularity": "day",                 // 统计粒度: day(日)、hour(小时)，必须
  "agent_id": "AGENT_001",              // 代理ID，可选，不传则统计整个分站
  "source": "NA",                       // 来源过滤，可选 (北美、非北美)
  "type": "register"                    // 统计类型: register(注册数)、login(登录数)、retention1d~retention60d(留存数)，必须
}
```

**type 参数说明**:
- `register`: 按时间段统计新增注册用户数
- `login`: 按时间段统计活跃登录用户数

**留存人数统计** (计数类型):
- `retention1d`: 注册后第1天至少登录一次的用户数
- `retention2d`: 注册后第1～2天至少登录一次的用户数
- `retention3d`: 注册后第1～3天至少登录一次的用户数
- `retention7d`: 注册后第1～7天至少登录一次的用户数
- `retention15d`: 注册后第1～15天至少登录一次的用户数
- `retention30d`: 注册后第1～30天至少登录一次的用户数
- `retention60d`: 注册后第1～60天至少登录一次的用户数

**留存率统计** (百分比类型，保留整数):
- `retention1dRate`: 1日留存率 = (retention_1d_users / total_new_users) * 100，返回整数
- `retention2dRate`: 2日留存率 = (retention_2d_users / total_new_users) * 100，返回整数
- `retention3dRate`: 3日留存率 = (retention_3d_users / total_new_users) * 100，返回整数
- `retention7dRate`: 7日留存率 = (retention_7d_users / total_new_users) * 100，返回整数
- `retention15dRate`: 15日留存率 = (retention_15d_users / total_new_users) * 100，返回整数
- `retention30dRate`: 30日留存率 = (retention_30d_users / total_new_users) * 100，返回整数
- `retention60dRate`: 60日留存率 = (retention_60d_users / total_new_users) * 100，返回整数

#### 响应参数

```json
{
  "code": 0,                            // 0=成功，非0=失败
  "message": "success",                 // 错误信息
  "list": [                             // 统计结果列表
    {
      "date": "2026-10-15",             // 日期 (YYYY-MM-DD 或 YYYY-MM-DD HH:00)
      "count": 150                      // 统计数量说明：
                                        // - register/login: 新增数/登录数
                                        // - retention*: 留存人数
                                        // - retention*Rate: 留存率（百分比整数，如123表示12.3%）
    },
    {
      "date": "2026-10-16",
      "count": 180
    }
  ]
}
```

#### 留存率计算说明

留存率统计基于注册日期维度的 `mv_user_retention_daily` 物化视图，计算公式为：

- **留存率 (%) = (留存人数 / 当日新增用户数) × 100**

例如：
- 2026-10-15 注册新增用户 100 人
- 其中有 12 人在注册后第 1 天至少登录一次 → retention_1d_rate = (12 / 100) × 100 = 12（返回整数）
- 其中有 10 人在注册后第 1～7 天至少登录一次 → retention_7d_rate = (10 / 100) × 100 = 10（返回整数）

**注意**：
- 留存率返回值为百分比整数（0-100 之间）
- 例如 12.34% 会返回为 1234（即 count = 1234）等等
- 如果当日新增用户数为 0，则留存率返回 0

#### 示例

**注册趋势请求**:
```bash
curl -X POST http://localhost:8888/operator-report/dashboard/get-member-growth-trend \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {token}" \
  -d '{
    "start_date": "2026-10-15",
    "end_date": "2026-10-22",
    "granularity": "day",
    "type": "register"
  }'
```

**1日留存率请求**:
```bash
curl -X POST http://localhost:8888/operator-report/dashboard/get-member-growth-trend \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {token}" \
  -d '{
    "start_date": "2026-10-15",
    "end_date": "2026-10-22",
    "granularity": "day",
    "type": "retention1dRate"
  }'
```

**7日留存人数请求**:
```bash
curl -X POST http://localhost:8888/operator-report/dashboard/get-member-growth-trend \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {token}" \
  -d '{
    "start_date": "2026-10-15",
    "end_date": "2026-10-22",
    "granularity": "day",
    "type": "retention7d"
  }'
```

**注册趋势响应**:
```json
{
  "code": 0,
  "message": "success",
  "list": [
    {"date": "2026-10-15", "count": 150},
    {"date": "2026-10-16", "count": 180},
    {"date": "2026-10-17", "count": 165},
    {"date": "2026-10-18", "count": 200},
    {"date": "2026-10-19", "count": 190},
    {"date": "2026-10-20", "count": 210},
    {"date": "2026-10-21", "count": 220},
    {"date": "2026-10-22", "count": 1286}
  ]
}
```

**1日留存率响应**:
```json
{
  "code": 0,
  "message": "success",
  "list": [
    {"date": "2026-10-15", "count": 1234},
    {"date": "2026-10-16", "count": 1089},
    {"date": "2026-10-17", "count": 1515},
    {"date": "2026-10-18", "count": 1200},
    {"date": "2026-10-19", "count": 1368},
    {"date": "2026-10-20", "count": 1381},
    {"date": "2026-10-21", "count": 1409},
    {"date": "2026-10-22", "count": 1342}
  ]
}
```

**7日留存人数响应**:
```json
{
  "code": 0,
  "message": "success",
  "list": [
    {"date": "2026-10-15", "count": 45},
    {"date": "2026-10-16", "count": 52},
    {"date": "2026-10-17", "count": 48},
    {"date": "2026-10-18", "count": 58},
    {"date": "2026-10-19", "count": 62},
    {"date": "2026-10-20", "count": 71},
    {"date": "2026-10-21", "count": 75},
    {"date": "2026-10-22", "count": 89}
  ]
}
```

---

### 2. 会员概览

**接口**: `GET /operator-report/dashboard/get-member-summary`

**描述**: 获取会员概览数据，包括会员总数、新增、活跃等关键指标及其增长比例。

#### 请求参数

```json
{
  "agent_id": "AGENT_001"               // 代理ID，可选，不传则统计整个分站
}
```

#### 响应参数

```json
{
  "code": 0,                            // 0=成功，非0=失败
  "message": "success",                 // 错误信息
  "data": {
    "total_members": 128640,            // 会员总数
    "members_growth_rate": 12.8,        // 较昨日会员总数上涨比例 (%)
    "today_new_members": 1286,          // 今日新增（今日注册）
    "last_24h_new_members": 5000,       // 过去24小时新增
    "last_24h_new_growth_rate": 8.2,    // 过去24小时较前24小时新增增长比例 (%)
    "today_deposit_amount": 286420,     // 今日充值金额（单位：元）
    "deposit_growth_rate": 6.4,         // 充值增长比例 (%)
    "active_members": 18562,            // 活跃会员总数
    "today_active_members": 8562,       // 今日活跃会员数
    "last_24h_active_members": 12500,   // 过去24小时活跃会员数
    "last_24h_active_growth_rate": 5.6  // 过去24小时较前24小时活跃增长比例 (%)
  }
}
```

#### 数据说明

| 字段 | 说明 | 备注 |
|------|------|------|
| total_members | 会员总数 | 历史注册总数（去重） |
| members_growth_rate | 较昨日会员增长比例 | (今日总数 - 昨日总数) / 昨日总数 * 100 |
| today_new_members | 今日新增 | register_time >= 今天00:00 且 < 明天00:00 的用户计数 |
| last_24h_new_members | 过去24小时新增 | 现在时刻前24小时内注册的用户数 |
| last_24h_new_growth_rate | 过去24小时新增增长比例 | (过去24h新增 - 前一个24h新增) / 前一个24h新增 * 100 |
| today_deposit_amount | 今日充值金额 | 单位：元，需要从deposit相关表查询 |
| deposit_growth_rate | 充值增长比例 | (今日充值 - 昨日充值) / 昨日充值 * 100，默认值 |
| active_members | 活跃会员总数 | 历史登录去重总数 |
| today_active_members | 今日活跃会员数 | login_time >= 今天00:00 且 < 明天00:00 的用户计数（去重） |
| last_24h_active_members | 过去24小时活跃会员数 | 现在时刻前24小时内登录的用户数（去重） |
| last_24h_active_growth_rate | 过去24小时活跃增长比例 | (过去24h活跃 - 前一个24h活跃) / 前一个24h活跃 * 100 |

#### 示例

**请求**:
```bash
curl -X GET 'http://localhost:8888/operator-report/dashboard/get-member-summary' \
  -H "Authorization: Bearer {token}"
```

或指定代理ID:
```bash
curl -X GET 'http://localhost:8888/operator-report/dashboard/get-member-summary?agent_id=AGENT_001' \
  -H "Authorization: Bearer {token}"
```

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total_members": 128640,
    "members_growth_rate": 12.8,
    "today_new_members": 1286,
    "last_24h_new_members": 5000,
    "last_24h_new_growth_rate": 8.2,
    "today_deposit_amount": 286420,
    "deposit_growth_rate": 6.4,
    "active_members": 18562,
    "today_active_members": 8562,
    "last_24h_active_members": 12500,
    "last_24h_active_growth_rate": 5.6
  }
}
```

---

## 错误处理

### 常见错误码

| 错误码 | 说明 |
|-------|------|
| 0 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未授权（token无效或过期） |
| 403 | 禁止访问 |
| 500 | 服务器内部错误 |

### 错误响应示例

```json
{
  "code": 400,
  "message": "invalid date format",
  "data": null
}
```

---

## 认证与授权

- 所有接口都需要通过 JWT Token 认证
- Token 通过 `Authorization: Bearer {token}` 请求头传入
- 分站编码通过 Token 的上下文数据自动获取

---

## 数据来源

所有数据来自以下Doris数据表：

- `dwd_member_register_event`: 会员注册事件表
- `dwd_member_login_event`: 登录事件明细表
- `mv_user_retention_daily`: 用户留存统计物化视图

---

## 更新历史

| 日期 | 版本 | 变更说明 |
|------|------|--------|
| 2026-10-22 | 1.1 | 增加留存率统计类型支持（retention*Rate），返回百分比整数，支持1日、2日、3日、7日、15日、30日、60日留存率统计 |
| 2026-10-22 | 1.0 | 初始版本，支持会员增长趋势和会员概览接口 |
