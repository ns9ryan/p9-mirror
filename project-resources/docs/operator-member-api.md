# 分站后台会员管理 API 对接文档

- **建议存放路径：** `project-resources/docs/operator-member-api.md`
- **适用服务：** `operator-admin-api`，业务由 `operator-crm` RPC 提供
- **对接范围：** 创建会员、会员列表
- **更新日期：** 2026-11-03
- **代码依据：** `p9-mirror/main`，Commit `5eea2ded699a508446c2a05cbe2730265490447d`；`desc/member/manage.api`、API Logic、CRM RPC Logic 与 DAO

本文记录当前源码已经定义的接口及行为，不代表已经通过部署环境的端到端联调。

## 一、基础约定

### 1. 请求地址

服务内部路由没有 `/admin` 前缀。直接访问开发服务时，例如：

```http
GET http://127.0.0.1:38010/member/list?page=1&page_size=20
```

现有会员层级、VIP 对接文档使用 `/admin/operator` 作为 Nginx 外部前缀，例如：

```http
GET https://ryan05.op.com/admin/operator/member/list?page=1&page_size=20
```

**注意：** 外部前缀以目标环境的 Nginx 实际配置为准。项目其他文档还存在 `/api` 前缀的约定，不能直接将上面外部 URL 当作所有环境的固定地址。

### 2. 认证与请求头

两个接口均经过 `Jwt + ActionLog + Authority`：需携带登录 Token，并由 Core 按 HTTP 方法和内部路径检查 API 权限。

```http
Authorization: Bearer <access_token>
X-Lang: zh-CN
```

`POST /member/create` 还需 `Content-Type: application/json`。`X-Lang` 可选，支持 `zh-CN`、`zh-HK`、`en-US`，默认 `zh-CN`。

API 权限登记为 `GET /member/list` 和 `POST /member/create`。当前厅身份由服务端可信上下文确定，前端**不传** `operator_code`。`agent_id` 只能作为查询筛选或创建会员归属参数，不能用来突破当前厅的访问范围。

### 3. 通用响应

成功：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {}
}
```

失败示例：

```json
{
  "code": 403,
  "msg": "无权访问"
}
```

常见错误状态为 `400`（参数错误）、`401`（未授权）、`403`（无权限）、`498`（Access Token 过期）和 `500`（服务错误）；以实际返回状态与 `msg` 为准。

### 4. 数据约定

- 列表的 `page`、`page_size` 必填，分别要求 `page >= 1`、`1 <= page_size <= 100`。
- 本文所有时间参数与会员时间字段均为 **Unix 毫秒时间戳**，与旧代理接口文档中的 Unix 秒不同。
- 金额使用当前厅结算币种的**最小单位**，接口不返回浮点金额。
- 列表统计是当前已有数据的累计/结算周期统计；尚未接入的钱包余额、优惠等字段不在接口中。
- 可选联系方式与登录信息需允许空值；当前列表 RPC **暂不填充** `phone` 和 `email`，前端不可假定会收到脱敏后的号码或邮箱。

## 二、接口一览

| 功能 | 方法 | 服务内部路径 | 请求方式 |
|---|---|---|---|
| 创建会员 | POST | `/member/create` | JSON Body |
| 会员列表 | GET | `/member/list` | Query 参数 |

## 三、创建会员

### POST /member/create

由分站后台为指定分层代理创建会员。

**请求示例**

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

**请求字段**

| 字段 | 类型 | 必填 | 规则与说明 |
|---|---|---|---|
| `account` | string | 是 | 登录账号，4～20 位；仅允许英文字母、数字、`-`、`_`，当前厅内不能重复 |
| `password` | string | 是 | 初始密码，8～32 位；仅允许 ASCII 可见字符，不允许空格 |
| `nickname` | string | 否 | 会员昵称，最多 100 个 Unicode 字符；未填或去除首尾空白后为空时使用账号 |
| `agent_id` | int64 | 是 | 所属分层代理 ID，大于 0；对应代理须存在、状态正常且层级有效 |
| `remark` | string | 否 | 后台备注，最多 200 个 Unicode 字符；空白备注不保存 |

**成功响应示例**

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

| 响应字段 | 类型 | 说明 |
|---|---|---|
| `data.id` | int64 | 新创建会员的数据库 ID |
| `data.code` | string | 系统生成的 12 位会员对外编号，注意是字符串 |

**业务说明**

- 创建成功后，注册方式为 `account`，注册渠道为 `admin`（这两个值在当前 RPC 中仍标记为临时常量）。
- 系统同时生成会员邀请码、创建会员初始记录，并更新所属代理及其上级的会员统计；这些内容不在本接口响应中返回。
- 密码由服务端哈希保存，前端不可把明文密码用于日志或二次展示。

## 四、会员列表

### GET /member/list

查询当前厅会员，支持关键字、时间、等级、注册信息、状态、代理归属、联系方式、充值条件及统计排序。

**请求示例**

```http
GET /member/list?page=1&page_size=20&keyword=member&status=1&sort_by=1&sort_order=2
Authorization: Bearer <access_token>
```

**Query 参数**

| 字段 | 类型 | 必填 | 规则与说明 |
|---|---|---|---|
| `page` | int64 | 是 | 页码，从 1 开始 |
| `page_size` | int64 | 是 | 每页数量，1～100 |
| `keyword` | string | 否 | 模糊搜索会员账号、12 位会员编号或昵称，首尾空白会去除 |
| `time_type` | int32 | 否 | `0` 不筛选、`1` 注册时间、`2` 最后登录时间 |
| `start_at` | int64 | 否 | 起始 Unix 毫秒时间，包含该时刻 |
| `end_at` | int64 | 否 | 结束 Unix 毫秒时间，不包含该时刻 |
| `auto_level_id` | int64 | 否 | 当前自动层级 ID，大于 0 |
| `fixed_level_id` | int64 | 否 | 已分配的固定层级 ID，大于 0 |
| `register_method` | string | 否 | 注册方式，精确匹配；提供时去空白后不能为空，最长 32 字节 |
| `register_channel` | string | 否 | 注册渠道，精确匹配；提供时去空白后不能为空，最长 32 字节 |
| `status` | int64 | 否 | 会员状态：`1` 正常、`2` 冻结、`3` 禁用 |
| `agent_id` | int64 | 否 | 所属分层代理 ID，精确筛选，大于 0 |
| `betting_disabled` | bool | 否 | `true` 只查询被禁止投注的会员 |
| `no_phone` | bool | 否 | `true` 只查询未绑定手机号的会员 |
| `no_email` | bool | 否 | `true` 只查询未绑定邮箱的会员 |
| `deposited` | bool | 否 | `true` 只查询有成功充值记录的会员 |
| `not_deposited` | bool | 否 | `true` 只查询没有成功充值记录的会员 |
| `sort_by` | int32 | 否 | `0` 注册时间、`1` 累计充值金额、`2` 充值次数、`3` 累计提现金额、`4` 提现次数 |
| `sort_order` | int32 | 否 | `0` 默认降序、`1` 升序、`2` 降序；`sort_by=0` 时注册时间始终倒序 |

**筛选与排序规则**

- `time_type=0` 时不可传 `start_at`、`end_at`；`time_type=1/2` 时至少传入一个时间边界。两个边界都传时必须满足 `start_at < end_at`。时间区间为**左闭右开**。
- `deposited=true` 和 `not_deposited=true` 不可同时出现；充值判断依据为成功充值次数 `total_deposit_count > 0`。没有统计记录视为未充值。
- `no_phone`、`no_email` 根据后台联系方式检索摘要是否存在判断，与列表是否返回手机号或邮箱是两回事。
- `sort_by=0` 或未传时按注册时间倒序，再按会员 ID 倒序；此时 `sort_order=1` 也不会改为升序。
- 统计字段排序按 `sort_order` 控制方向，会员 ID 为次级倒序；统计记录不存在的会员在升序时靠前、降序时靠后。
- 传入不支持的筛选枚举、无效 ID 或冲突的筛选条件会被拒绝，不能假定服务会自动忽略。

**成功响应示例**

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
            { "id": 3, "name": "重点会员" }
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
        "last_login_at": null,
        "last_login_ip": null,
        "phone": null,
        "email": null,
        "inviter_member_id": null,
        "remark": "后台创建",
        "stats": {
          "total_deposit": 0,
          "total_deposit_count": 0,
          "total_withdrawal": 0,
          "total_withdrawal_count": 0,
          "total_deposit_withdrawal_diff": 0,
          "total_bet": 0,
          "total_valid_bet": 0,
          "total_water": 0,
          "total_win_loss": 0,
          "current_period_deposit": 0,
          "current_period_withdrawal": 0,
          "current_period_bet": 0,
          "current_period_valid_bet": 0,
          "current_period_water": 0,
          "current_period_win_loss": 0,
          "current_period_started_at": null
        }
      }
    ]
  }
}
```

上面是**字段结构示例**，数值和记录仅用于说明，并非真实业务数据。可选字段未设置时前端应兼容 `null`；列表无记录时 `data.list` 返回 `[]`、`total` 返回 `0`。

**响应字段：基础与状态**

| 字段 | 类型 | 说明 |
|---|---|---|
| `data.total` | int64 | 符合条件的会员总数，不是当前页记录数 |
| `data.list` | array | 当前页会员列表 |
| `list[].id` | int64 | 会员数据库 ID |
| `list[].code` | string | 12 位会员编号 |
| `list[].account` | string | 会员登录账号 |
| `list[].nickname` | string | 会员昵称 |
| `list[].agent_id` | int64 | 所属分层代理 ID |
| `list[].agent_account` | string | 所属分层代理账号 |
| `list[].status` | int64 | `1` 正常、`2` 冻结、`3` 禁用 |
| `list[].chat_disabled` | bool | 是否禁止发言 |
| `list[].betting_disabled` | bool | 是否禁止投注 |
| `list[].withdrawal_disabled` | bool | 是否禁止提款 |
| `list[].register_method` | string | 注册方式 |
| `list[].register_channel` | string | 注册渠道 |
| `list[].created_at` | int64 | 注册时间，Unix 毫秒 |
| `list[].last_login_at` | int64 / null | 最后登录时间，Unix 毫秒 |
| `list[].last_login_ip` | string / null | 最后登录 IP |
| `list[].phone` | string / null | 预留的脱敏手机号；当前列表 RPC 暂不填充 |
| `list[].email` | string / null | 预留的脱敏邮箱；当前列表 RPC 暂不填充 |
| `list[].inviter_member_id` | int64 / null | 邀请人会员 ID |
| `list[].remark` | string / null | 后台备注 |

**响应字段：会员等级 `list[].levels`**

| 字段 | 类型 | 说明 |
|---|---|---|
| `auto_level_id` | int64 | 当前自动层级 ID |
| `auto_level_name` | string | 当前自动层级名称 |
| `fixed_levels` | array | 当前分配的固定层级列表，按 ID 升序；无记录为 `[]` |
| `fixed_levels[].id` | int64 | 固定层级 ID |
| `fixed_levels[].name` | string | 固定层级名称 |
| `vip_level_id` | int64 | 当前 VIP 等级 ID |
| `vip_level` | int64 | 当前 VIP 等级编号 |

**响应字段：会员统计 `list[].stats`**

| 字段 | 类型 | 说明 |
|---|---|---|
| `total_deposit` | int64 | 累计充值金额 |
| `total_deposit_count` | int64 | 累计成功充值次数 |
| `total_withdrawal` | int64 | 累计提现金额 |
| `total_withdrawal_count` | int64 | 累计成功提现次数 |
| `total_deposit_withdrawal_diff` | int64 | 累计存提差额，充值减提现 |
| `total_bet` | int64 | 累计投注金额 |
| `total_valid_bet` | int64 | 累计有效投注金额 |
| `total_water` | int64 | 累计洗码量 |
| `total_win_loss` | int64 | 累计会员净输赢 |
| `current_period_deposit` | int64 | 当前结算周期充值金额 |
| `current_period_withdrawal` | int64 | 当前结算周期提现金额 |
| `current_period_bet` | int64 | 当前结算周期投注金额 |
| `current_period_valid_bet` | int64 | 当前结算周期有效投注金额 |
| `current_period_water` | int64 | 当前结算周期洗码量 |
| `current_period_win_loss` | int64 | 当前结算周期会员净输赢 |
| `current_period_started_at` | int64 / null | 当前结算周期开始时间，Unix 毫秒；无有效周期开始时间时为空 |

统计金额和计数无对应记录时返回 `0`。输赢字段按当前 Proto 说明，正数表示会员盈利，负数表示会员亏损。当前周期字段由现有会员统计数据映射，前端不要自行推算周期起点。

## 五、权限与联调注意事项

1. 权限目录已在 `operator-admin-api/internal/catalog/catalog.go` 中登记两个 API，多语言描述已在 `catalog_i18n.go` 中登记。需启动注册流程使 Core 写入目录，并确保测试运营角色被授予相应 API 权限。
2. 当前 CRM 会员列表 RPC 检查 `claims.UserType == UserTypeOperator`，但镜像中的 Core `CheckTokenResp` 与 `coreadapt.Auth` 未传递 `UserType`。**正式联调前必须修复可信身份传递链路**，不能通过直接信任请求参数来绕过校验。
3. 会员列表 SQL 依赖当前厅数据隔离以及服务端可信上下文。对 `agent_id` 的筛选不能成为跨厅查询入口。
4. 建议依次验证分页、空列表、关键字、时间边界、等级过滤、状态过滤、充值互斥、统计排序、不同运营角色的权限结果。
5. 本文按最新镜像源码整理，尚未在运行环境验证实际响应；若 API 网关或 JSON 序列化行为与示例不同，以实测为准并同步更新文档。
