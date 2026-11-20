# 分站后台会员管理 API 对接文档

- **适用服务：** `operator-admin-api`
- **对接范围：** 创建会员、批量创建会员、会员列表、会员详情
- **更新日期：** 2026-11-17

## 目录索引

- [一、基础约定](#一基础约定)
    - [1. 请求地址与认证](#1-请求地址与认证)
    - [2. 通用响应](#2-通用响应)
    - [3. 数据格式](#3-数据格式)
- [二、接口一览](#二接口一览)
- [三、创建会员](#三创建会员)
    - [1. 创建会员请求参数](#1-创建会员请求参数)
    - [2. 创建会员响应](#2-创建会员响应)
    - [3. 批量创建会员](#3-批量创建会员)
- [四、会员列表](#四会员列表)
    - [1. 会员列表查询参数](#1-会员列表查询参数)
    - [2. 筛选与排序规则](#2-筛选与排序规则)
    - [3. 会员列表响应](#3-会员列表响应)
- [五、会员详情](#五会员详情)
    - [1. 会员详情请求参数](#1-会员详情请求参数)
    - [2. 会员详情响应](#2-会员详情响应)
- [六、公共数据结构](#六公共数据结构)
    - [1. 会员等级信息](#1-会员等级信息)
    - [2. 会员统计信息](#2-会员统计信息)
- [七、前端对接注意事项](#七前端对接注意事项)

---

## 一、基础约定

### 1. 请求地址与认证

**接口内部路径：**

| 功能 | 方法 | 路径 |
|---|---|---|
| 创建会员 | POST | `/member/create` |
| 批量创建会员 | POST | `/member/batch/create` |
| 会员列表 | GET | `/member/list` |
| 会员详情 | GET | `/member/detail` |

直接访问开发服务时，例如：

```http
GET http://127.0.0.1:38010/member/list?page=1&page_size=20
```

部分部署环境通过 Nginx 增加 `/admin/operator` 前缀：

```http
GET https://ryan05.op.com/admin/operator/member/list?page=1&page_size=20
```

实际请求地址以当前环境配置的 API Base URL 为准，前端不要在业务接口中重复拼接固定前缀。

**请求头：**

```http
Authorization: Bearer <access_token>
X-Lang: zh-CN
```

| 请求头 | 说明 |
|---|---|
| `Authorization` | 登录 Token |
| `X-Lang` | 响应语言，可选，支持 `zh-CN`、`zh-HK`、`en-US`，默认 `zh-CN` |
| `Content-Type` | 创建会员时使用 `application/json` |

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
| `400` | 请求参数错误 |
| `401` | 未登录或认证失败 |
| `403` | 没有接口访问权限 |
| `498` | Access Token 已过期 |
| `500` | 服务内部错误 |

### 3. 数据格式

| 数据类型 | 格式 | 说明 |
|---|---|---|
| 普通时间点 | Unix 毫秒时间戳 | 注册时间、最后登录时间等 |
| 日期 | `YYYY-MM-DD` | 会员生日，不涉及时区转换 |
| 金额 | int64 | 当前厅结算币种的最小单位，不是浮点金额 |
| ID | int64 | 数据库主键或业务关联 ID |
| 可选字段 | 可能为空或不返回 | 前端需要兼容空值 |
| 列表 | array | 无记录时返回空数组 `[]` |

---

## 二、接口一览

| 功能 | 方法 | 路径 | 说明 |
|---|---|---|---|
| 创建会员 | POST | `/member/create` | 创建会员账号 |
| 批量创建会员 | POST | `/member/batch/create` | 按账号前缀或后缀生成 1 至 2000 个会员账号 |
| 会员列表 | GET | `/member/list` | 分页查询会员 |
| 会员详情 | GET | `/member/detail` | 查询会员完整资料 |

---

## 三、创建会员

### 1. 创建会员请求参数

**POST /member/create**

**请求示例：**

```http
POST /member/create
Content-Type: application/json
Authorization: Bearer <access_token>
```

```json
{
  "account": "member001",
  "password": "Abcd1234!",
  "nickname": "测试会员",
  "agent_id": 1002,
  "remark": "后台创建"
}
```

**请求字段：**

| 字段 | 类型 | 必填 | 规则与说明 |
|---|---|---|---|
| `account` | string | 是 | 登录账号，4～20 位，仅允许英文字母、数字、`-`、`_`，当前厅内不能重复 |
| `password` | string | 是 | 初始密码，8～32 位，仅允许 ASCII 可见字符，不允许空格 |
| `nickname` | string | 否 | 会员昵称，最多 100 个 Unicode 字符；未填写或去除首尾空白后为空时使用账号 |
| `agent_id` | int64 | 是 | 所属分层代理 ID，大于 0，必须选择有效且状态正常的代理 |
| `remark` | string | 否 | 后台备注，最多 200 个 Unicode 字符，空白备注不保存 |

### 2. 创建会员响应

**成功响应示例：**

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 10001,
    "code": "123456789012"
  }
}
```

**响应字段：**

| 字段 | 类型 | 说明 |
|---|---|---|
| `data.id` | int64 | 新创建会员的数据库 ID |
| `data.code` | string | 系统生成的 12 位会员对外编号，注意是字符串 |

---

### 3. 批量创建会员

`POST /member/batch/create`，使用当前分站的 API Base URL，携带 `Authorization: Bearer <access_token>` 和 `Content-Type: application/json`。调用账号需拥有该接口权限。

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `account_prefix` | string | 与后缀至少填一项 | 账号前缀 |
| `account_suffix` | string | 与前缀至少填一项 | 账号后缀 |
| `count` | int32 | 是 | 创建数量，整数，范围 1 至 2000 |
| `agent_id` | int64 | 是 | 所选的正常分层代理 ID，可通过 `POST /agent/search` 搜索代理 |

前缀和后缀合计最多 16 个字符，仅支持英文字母、数字、下划线和短横线。后端在中间拼接 4 至 8 位随机小写字母或数字，检查同一分站内的账号重复。最终账号不超过 20 个字符。

```json
{
  "account_prefix": "test_",
  "account_suffix": "_vip",
  "count": 2,
  "agent_id": 123
}
```

成功响应示例：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "members": [
      { "id": 501, "code": "618274930516", "account": "test_9b72ca3d_vip" },
      { "id": 502, "code": "719405632851", "account": "test_ax2fc691_vip" }
    ]
  }
}
```

`data.members` 的长度等于请求的 `count`。每项返回会员 ID（`id`）、会员编号（`code`，字符串）和登录账号（`account`）。初始登录密码固定为 `12345678`。昵称由总站随机分配。会员初始为 VIP0、普通会员、状态正常，禁止投注关闭，手机号、邮箱、备注留空。

接口同步返回，数据库按小批写入并在同一事务提交。任何写入失败时整批回滚。前端提交后应禁用提交按钮直至请求结束。接口服务允许最长 10 分钟，请结合网关设置配置客户端超时。请求超时或断线后先查询会员列表再决定是否重试，重复提交可能创建另一批账号。

---

## 四、会员列表

### 1. 会员列表查询参数

**GET /member/list**

**请求示例：**

```http
GET /member/list?page=1&page_size=20&keyword=member&status=1
Authorization: Bearer <access_token>
```

**Query 参数：**

| 字段 | 类型 | 必填 | 规则与说明 |
|---|---|---|---|
| `page` | int64 | 是 | 当前页码，从 1 开始 |
| `page_size` | int64 | 是 | 每页数量，1～100 |
| `keyword` | string | 否 | 模糊搜索会员账号、12 位会员编号或昵称 |
| `time_type` | int32 | 否 | 时间筛选类型：`0` 不筛选、`1` 注册时间、`2` 最后登录时间 |
| `start_at` | int64 | 否 | 起始时间，Unix 毫秒时间戳，包含该时刻 |
| `end_at` | int64 | 否 | 结束时间，Unix 毫秒时间戳，不包含该时刻 |
| `auto_level_id` | int64 | 否 | 当前自动层级 ID，大于 0 |
| `fixed_level_id` | int64 | 否 | 已分配的固定层级 ID，大于 0 |
| `register_method` | string | 否 | 注册方式，精确匹配，最长 32 字节 |
| `register_channel` | string | 否 | 注册渠道，精确匹配，最长 32 字节 |
| `status` | int64 | 否 | 会员状态：`1` 正常、`2` 冻结、`3` 禁用 |
| `agent_id` | int64 | 否 | 所属分层代理 ID，大于 0 |
| `betting_disabled` | bool | 否 | `true` 只查询禁止投注的会员 |
| `no_phone` | bool | 否 | `true` 只查询未绑定手机号的会员 |
| `no_email` | bool | 否 | `true` 只查询未绑定邮箱的会员 |
| `deposited` | bool | 否 | `true` 只查询有成功充值记录的会员 |
| `not_deposited` | bool | 否 | `true` 只查询没有成功充值记录的会员 |
| `sort_by` | int32 | 否 | 排序字段，具体见下方枚举 |
| `sort_order` | int32 | 否 | 排序方向，具体见下方枚举 |

**时间筛选类型：**

| 值 | 说明 |
|---|---|
| `0` | 不筛选时间 |
| `1` | 注册时间 |
| `2` | 最后登录时间 |

**排序字段 `sort_by`：**

| 值 | 说明 |
|---|---|
| `0` | 注册时间 |
| `1` | 累计充值金额 |
| `2` | 累计充值成功次数 |
| `3` | 累计提现金额 |
| `4` | 累计提现成功次数 |

**排序方向 `sort_order`：**

| 值 | 说明 |
|---|---|
| `0` | 默认降序 |
| `1` | 升序 |
| `2` | 降序 |

### 2. 筛选与排序规则

- `time_type=0` 时不能传入 `start_at`、`end_at`
- `time_type=1` 或 `2` 时，至少需要传入一个时间边界
- 同时传入起止时间时，必须满足 `start_at < end_at`
- 时间范围采用左闭右开，即包含开始时间，不包含结束时间
- `deposited=true` 和 `not_deposited=true` 不能同时使用
- `sort_by=0` 或未传时，按照注册时间倒序排列，此时 `sort_order` 不生效
- `sort_by=1～4` 时，按照对应统计字段排序
- 排序方向未指定时默认降序
- 统计字段排序时，没有统计记录的会员升序排在前面，降序排在后面

### 3. 会员列表响应

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
        "code": "123456789012",
        "account": "member001",
        "nickname": "测试会员",
        "agent_id": 1002,
        "agent_account": "agent001",
        "levels": {
          "auto_level_id": 1,
          "auto_level_name": "普通会员",
          "fixed_levels": [
            {
              "id": 3,
              "name": "重点会员"
            }
          ],
          "vip_level_id": 1,
          "vip_level": 1
        },
        "status": 1,
        "chat_disabled": false,
        "betting_disabled": false,
        "withdrawal_disabled": false,
        "register_method": "account",
        "register_channel": "admin",
        "created_at": 1793664000000,
        "last_login_at": 1793667600000,
        "last_login_ip": "192.0.2.11",
        "inviter_member_id": null,
        "remark": "后台创建",
        "stats": {
          "total_deposit": 100000,
          "total_deposit_count": 2,
          "total_withdrawal": 20000,
          "total_withdrawal_count": 1,
          "total_deposit_withdrawal_diff": 80000,
          "total_bet": 50000,
          "total_valid_bet": 40000,
          "total_water": 1000,
          "total_win_loss": -5000,
          "current_period_deposit": 30000,
          "current_period_withdrawal": 10000,
          "current_period_bet": 20000,
          "current_period_valid_bet": 18000,
          "current_period_water": 500,
          "current_period_win_loss": -2000,
          "current_period_started_at": 1793577600000
        }
      }
    ]
  }
}
```

以上数据仅用于展示字段结构。

**响应字段：**

| 字段 | 类型 | 说明 |
|---|---|---|
| `data.total` | int64 | 符合条件的会员总数 |
| `data.list` | array | 当前页会员列表 |
| `list[].id` | int64 | 会员数据库 ID |
| `list[].code` | string | 12 位会员编号 |
| `list[].account` | string | 会员登录账号 |
| `list[].nickname` | string | 会员昵称 |
| `list[].agent_id` | int64 | 所属分层代理 ID |
| `list[].agent_account` | string | 所属分层代理账号 |
| `list[].levels` | object | 会员等级信息，见[会员等级信息](#1-会员等级信息) |
| `list[].status` | int64 | 会员状态 |
| `list[].chat_disabled` | bool | 是否禁止发言 |
| `list[].betting_disabled` | bool | 是否禁止投注 |
| `list[].withdrawal_disabled` | bool | 是否禁止提款 |
| `list[].register_method` | string | 注册方式 |
| `list[].register_channel` | string | 注册渠道 |
| `list[].created_at` | int64 | 注册时间，Unix 毫秒时间戳 |
| `list[].last_login_at` | int64 / null | 最后登录时间，Unix 毫秒时间戳 |
| `list[].last_login_ip` | string / null | 最后登录 IP |
| `list[].phone` | string / null | 手机号字段，当前接口暂不返回有效值 |
| `list[].email` | string / null | 邮箱字段，当前接口暂不返回有效值 |
| `list[].inviter_member_id` | int64 / null | 邀请人会员 ID |
| `list[].remark` | string / null | 后台会员备注 |
| `list[].stats` | object | 会员统计信息，见[会员统计信息](#2-会员统计信息) |

没有符合条件的会员时，返回：

```json
{
  "total": 0,
  "list": []
}
```

---

## 五、会员详情

### 1. 会员详情请求参数

**GET /member/detail**

根据会员数据库主键查询会员详情。

**请求示例：**

```http
GET /member/detail?member_id=10001
Authorization: Bearer <access_token>
```

**Query 参数：**

| 字段 | 类型 | 必填 | 规则与说明 |
|---|---|---|---|
| `member_id` | int64 | 是 | 会员数据库主键，必须大于 0 |

### 2. 会员详情响应

**成功响应示例：**

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 10001,
    "code": "123456789012",
    "account": "member001",
    "nickname": "测试会员",
    "avatar_url": "/todo/default-member-avatar.png",
    "avatar_frame_url": "/todo/default-member-avatar-frame.png",
    "birthday": "1995-08-20",
    "gender": 1,
    "region_code": "CN",
    "agent_id": 1002,
    "agent_account": "agent001",
    "agent_line_code": "line01",
    "levels": {
      "auto_level_id": 1,
      "auto_level_name": "普通会员",
      "fixed_levels": [
        {
          "id": 3,
          "name": "重点会员"
        }
      ],
      "vip_level_id": 1,
      "vip_level": 1
    },
    "status": 1,
    "chat_disabled": false,
    "betting_disabled": false,
    "withdrawal_disabled": false,
    "register_method": "account",
    "register_channel": "admin",
    "created_at": 1793664000000,
    "register_ip": "192.0.2.10",
    "register_device_code": "device-example",
    "register_domain": "example.com",
    "last_login_at": 1793667600000,
    "last_login_ip": "192.0.2.11",
    "inviter_member_id": null,
    "remark": "后台创建",
    "stats": {
      "total_deposit": 100000,
      "total_deposit_count": 2,
      "total_withdrawal": 20000,
      "total_withdrawal_count": 1,
      "total_deposit_withdrawal_diff": 80000,
      "total_bet": 50000,
      "total_valid_bet": 40000,
      "total_water": 1000,
      "total_win_loss": -5000,
      "current_period_deposit": 30000,
      "current_period_withdrawal": 10000,
      "current_period_bet": 20000,
      "current_period_valid_bet": 18000,
      "current_period_water": 500,
      "current_period_win_loss": -2000,
      "current_period_started_at": 1793577600000
    }
  }
}
```

以上数据仅用于展示字段结构，具体内容以接口实际返回为准。

**会员基本信息：**

| 字段 | 类型 | 说明 |
|---|---|---|
| `data.id` | int64 | 会员数据库主键 |
| `data.code` | string | 12 位会员对外编号 |
| `data.account` | string / null | 会员登录账号 |
| `data.nickname` | string | 会员昵称 |
| `data.avatar_url` | string | 会员头像 URL |
| `data.avatar_frame_url` | string | 会员头像框 URL |
| `data.birthday` | string / null | 会员生日，格式 `YYYY-MM-DD` |
| `data.gender` | int64 | 性别：`0` 未设置、`1` 男、`2` 女、`3` 保密 |
| `data.region_code` | string / null | 国家或地区编码 |

**代理归属：**

| 字段 | 类型 | 说明 |
|---|---|---|
| `data.agent_id` | int64 | 所属分层代理 ID |
| `data.agent_account` | string | 所属分层代理账号 |
| `data.agent_line_code` | string | 所属代理线路编码 |

**会员等级：**

| 字段 | 类型 | 说明 |
|---|---|---|
| `data.levels` | object | 会员等级信息，见[会员等级信息](#1-会员等级信息) |

**状态与业务限制：**

| 字段 | 类型 | 说明 |
|---|---|---|
| `data.status` | int64 | `1` 正常、`2` 冻结、`3` 禁用 |
| `data.chat_disabled` | bool | 是否禁止发言 |
| `data.betting_disabled` | bool | 是否禁止投注 |
| `data.withdrawal_disabled` | bool | 是否禁止提款 |

**注册信息：**

| 字段 | 类型 | 说明 |
|---|---|---|
| `data.register_method` | string | 注册方式 |
| `data.register_channel` | string | 注册渠道 |
| `data.created_at` | int64 | 注册时间，Unix 毫秒时间戳 |
| `data.register_ip` | string / null | 注册 IP |
| `data.register_device_code` | string / null | 注册设备标识 |
| `data.register_domain` | string / null | 注册域名 |

**最后登录信息：**

| 字段 | 类型 | 说明 |
|---|---|---|
| `data.last_login_at` | int64 / null | 最后登录时间，Unix 毫秒时间戳 |
| `data.last_login_ip` | string / null | 最后登录 IP |

**运营信息：**

| 字段 | 类型 | 说明 |
|---|---|---|
| `data.inviter_member_id` | int64 / null | 邀请人会员 ID |
| `data.remark` | string / null | 后台会员备注 |

**统计信息：**

| 字段 | 类型 | 说明 |
|---|---|---|
| `data.stats` | object | 会员累计及当前周期统计，见[会员统计信息](#2-会员统计信息) |

当前会员详情接口不包含手机号、邮箱、钱包余额及钱包流水。

---

## 六、公共数据结构

会员列表和会员详情共用以下结构，字段统一在本节说明。

### 1. 会员等级信息

对应字段：

- 会员列表：`list[].levels`
- 会员详情：`data.levels`

| 字段 | 类型 | 说明 |
|---|---|---|
| `auto_level_id` | int64 | 当前自动层级 ID |
| `auto_level_name` | string | 当前自动层级名称 |
| `fixed_levels` | array | 已分配的固定层级列表，无记录时返回空数组 |
| `fixed_levels[].id` | int64 | 固定层级 ID |
| `fixed_levels[].name` | string | 固定层级名称 |
| `vip_level_id` | int64 | 当前 VIP 等级 ID |
| `vip_level` | int64 | 当前 VIP 等级编号 |

### 2. 会员统计信息

对应字段：

- 会员列表：`list[].stats`
- 会员详情：`data.stats`

**累计资金统计：**

| 字段 | 类型 | 说明 |
|---|---|---|
| `total_deposit` | int64 | 累计充值金额 |
| `total_deposit_count` | int64 | 累计充值成功次数 |
| `total_withdrawal` | int64 | 累计提现金额 |
| `total_withdrawal_count` | int64 | 累计提现成功次数 |
| `total_deposit_withdrawal_diff` | int64 | 累计存提差额，充值金额减去提现金额 |

**累计游戏统计：**

| 字段 | 类型 | 说明 |
|---|---|---|
| `total_bet` | int64 | 累计投注金额 |
| `total_valid_bet` | int64 | 累计有效投注金额 |
| `total_water` | int64 | 累计洗码量 |
| `total_win_loss` | int64 | 累计会员净输赢，正数表示盈利，负数表示亏损 |

**当前结算周期统计：**

| 字段 | 类型 | 说明 |
|---|---|---|
| `current_period_deposit` | int64 | 当前周期充值金额 |
| `current_period_withdrawal` | int64 | 当前周期提现金额 |
| `current_period_bet` | int64 | 当前周期投注金额 |
| `current_period_valid_bet` | int64 | 当前周期有效投注金额 |
| `current_period_water` | int64 | 当前周期洗码量 |
| `current_period_win_loss` | int64 | 当前周期会员净输赢，正数表示盈利，负数表示亏损 |
| `current_period_started_at` | int64 / null | 当前结算周期开始时间，Unix 毫秒时间戳 |

统计说明：

- 所有金额均使用当前厅结算币种的最小单位
- 没有统计记录时，统计金额和次数返回 `0`
- 当前周期开始时间没有有效值时为空

---

## 七、前端对接注意事项

1. 会员详情通过 `member_id` 查询，不要使用会员对外编号 `code` 代替数据库主键
2. 会员生日使用日期字符串展示，不要转换成 Unix 时间戳，避免时区导致日期偏移
3. 注册时间、最后登录时间和列表时间筛选均使用 Unix 毫秒时间戳
4. 金额需要根据当前厅结算币种的精度转换展示，不要直接按浮点数处理接口数据
5. `fixed_levels` 是数组，可能包含多个固定层级，不能按单个对象处理
6. 可选字段需要兼容空值或不返回的情况
7. 当前列表接口的 `phone`、`email` 字段暂不提供有效数据，不要依赖这两个字段展示联系方式
8. 当前会员详情接口尚未包含钱包数据，钱包信息需要后续接入对应接口
9. 会员列表和详情均只能访问当前厅的数据，前端不需要也不应该自行传递厅归属参数
---

## 八、会员联系方式绑定验证码记录

- 分站 API：`GET /member/contact/binding-codes/list`，使用会员管理的 JWT 和权限校验。
- 分站 RPC：`Member.ListMemberContactBindingCodes`。
- 部署后重启分站 API 以注册接口目录，并给操作账号的角色分配此 GET 接口权限。请求携带 `Authorization: Bearer <token>`。
- 不传筛选条件即查询列表；筛选条件按 AND 组合。厅归属从鉴权上下文读取。
- 按 `created_at DESC, id DESC` 排序。接口仅查询记录，不负责发送、校验或绑定。

| 查询参数 | 类型 | 说明 |
| --- | --- | --- |
| `page` | int64 | 页码，默认 1 |
| `page_size` | int64 | 每页数量，默认 20，最大 100 |
| `id` | int64 | 验证码记录 ID，精确匹配 |
| `member_id` | int64 | 会员主键 ID，精确匹配 |
| `channel` | int32 | 1 手机短信，2 邮件 |
| `status` | int32 | 0 待发送，1 已发送，2 已使用，3 已作废，4 发送失败；不传则查全部 |
| `created_from` | int64 | 创建时间起点，包含，Unix 毫秒 |
| `created_before` | int64 | 创建时间终点，不包含，Unix 毫秒 |

响应数据：`total` 为当前筛选条件下的总记录数，`list` 为当前页记录。每条记录包含 `id`、`member_id`、`channel`、`status`、`failed_attempts`、`created_at`、`sent_at`、`expires_at`、`used_at`。时间统一用 Unix 毫秒，未发送或未核销时对应时间为 `0`。过期由 `expires_at` 判断，不自动改变发送状态。`verification_code` 以字符串返回已失效的验证码，保留前导零。仍可用于验证的验证码以及只有摘要的历史记录返回空字符串。

在分站 CRM 的 PostgreSQL 数据库执行以下 SQL，再调用该接口。

```sql
CREATE TABLE member_contact_binding_code (
    id               bigint GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    operator_code    varchar(64) NOT NULL,
    member_id        bigint NOT NULL,
    channel          smallint NOT NULL CHECK (channel IN (1, 2)),
    target_hash      bytea NOT NULL CHECK (octet_length(target_hash) = 32),
    verification_code text,
    code_hash        bytea NOT NULL CHECK (octet_length(code_hash) = 32),
    status           smallint NOT NULL DEFAULT 0 CHECK (status IN (0, 1, 2, 3, 4)),
    failed_attempts  integer NOT NULL DEFAULT 0 CHECK (failed_attempts >= 0),
    created_at       timestamptz(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    sent_at          timestamptz(3),
    expires_at       timestamptz(3) NOT NULL,
    used_at          timestamptz(3),

    CHECK (expires_at > created_at),
    CHECK (status NOT IN (1, 2) OR sent_at IS NOT NULL),
    CHECK ((status = 2) = (used_at IS NOT NULL))
);

COMMENT ON TABLE member_contact_binding_code IS '会员手机号、邮箱绑定验证码记录';
COMMENT ON COLUMN member_contact_binding_code.id IS '主键 ID，也是本次验证码申请的标识';
COMMENT ON COLUMN member_contact_binding_code.operator_code IS '所属分站编码';
COMMENT ON COLUMN member_contact_binding_code.member_id IS '会员 ID';
COMMENT ON COLUMN member_contact_binding_code.channel IS '发送渠道：1 手机短信，2 邮件';
COMMENT ON COLUMN member_contact_binding_code.target_hash IS '规范化手机号或邮箱的 HMAC-SHA256 摘要';
COMMENT ON COLUMN member_contact_binding_code.verification_code IS '验证码明文，历史记录可为空';
COMMENT ON COLUMN member_contact_binding_code.code_hash IS '与本次申请绑定的验证码 HMAC-SHA256 摘要';
COMMENT ON COLUMN member_contact_binding_code.status IS '0 待发送，1 已发送，2 已使用，3 已作废，4 发送失败';
COMMENT ON COLUMN member_contact_binding_code.failed_attempts IS '验证码校验失败次数';
COMMENT ON COLUMN member_contact_binding_code.created_at IS '验证码记录创建时间';
COMMENT ON COLUMN member_contact_binding_code.sent_at IS '发送 RPC 返回成功的时间';
COMMENT ON COLUMN member_contact_binding_code.expires_at IS '验证码过期时间';
COMMENT ON COLUMN member_contact_binding_code.used_at IS '验证码核销时间';

CREATE INDEX idx_member_contact_binding_code_lookup
    ON member_contact_binding_code
    (operator_code, member_id, channel, target_hash, created_at DESC, id DESC);

CREATE INDEX idx_member_contact_binding_code_operator_created_at
    ON member_contact_binding_code (operator_code, created_at DESC, id DESC);

CREATE INDEX idx_member_contact_binding_code_expires_at
    ON member_contact_binding_code (expires_at);
```

已执行旧版建表 SQL 的环境需要执行以下增量 SQL。发送流程写入记录时需同时保存 `verification_code` 和 `code_hash`。

```sql
ALTER TABLE member_contact_binding_code
    ADD COLUMN IF NOT EXISTS verification_code text;
COMMENT ON COLUMN member_contact_binding_code.verification_code IS '验证码明文，历史记录可为空';
```


### 发送验证码与确认修改

`POST /member/contact/binding-codes/send` 调用分站 RPC `Member.SendMemberContactBindingCode`。请求包含 `member_id`、`channel` 和新手机号或新邮箱 `target`。渠道 `1` 为短信，手机号需含国际区号。渠道 `2` 为邮件。

```json
{"member_id": 123, "channel": 2, "target": "member@example.com"}
```

发送服务受理后返回 `record_id` 和 Unix 毫秒时间戳 `expires_at`。验证码为六位数字，有效期五分钟，同一厅、会员和渠道每 60 秒允许申请一次。新的发送请求会作废该渠道旧的待发送或已发送记录。连续输错五次后该记录失效。

`POST /member/contact/update` 修改邮箱或手机号时，分别携带 `email_verification` 或 `phone_verification`，两者结构相同。单次修改两项时须提交两份验证码。现有联系方式不能通过传空字符串清除。仅修改昵称等其他资料时无需验证码。

```json
{
  "member_id": 123,
  "email": "member@example.com",
  "email_verification": {
    "record_id": 456,
    "verification_code": "012345"
  }
}


{
  "member_id": 123,
  "phone": "5561981737725",
  "phone_verification": {
    "record_id": 456,
    "verification_code": "012345"
  }
}
```

分站 CRM 的 `IntegrationRpc` 必须指向已配置邮件和短信通道的 integration RPC，CRM 的 `FieldCrypto` 需要配置加密和摘要密钥。部署时重新注册并分配列表及发送路由的 API 权限，保留联系方式更新路由的原有权限。

发送验证码路由的 HTTP 超时为 45 秒。内网配置中 API 到 CRM 为 40 秒、CRM 服务为 38 秒、CRM 到 integration 为 30 秒，给发送结果落库预留时间。SMTP 连接或发送超时仍需根据 integration 日志检查供应商网络和端口。
