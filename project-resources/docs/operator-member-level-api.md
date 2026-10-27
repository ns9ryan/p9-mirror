# 分站后台会员层级 API 对接文档

-   建议存放路径：`project-resources/docs/operator-member-level-api.md`
-   适用服务：`operator-admin-api`
-   对接范围：会员自动层级、会员固定层级
-   依据：`operator-admin-api/desc/member/auto_level.api`、`operator-admin-api/desc/member/fixed_level.api`
    及当前 RPC Logic / DAO 实现
-   更新日期：2026-10-27

## 一、基础约定

### 1. 请求地址

分站后台通过域名访问 Nginx，统一使用 `/admin/operator` 前缀。

``` text
{分站域名}/admin/operator/{接口路径}
```

例如：

``` text
GET https://ryan05.op.com/admin/operator/member/fixed-level/list?page=1&page_size=20
```

Nginx 将外部 `/admin/operator/` 前缀去掉，再转发到
`operator-admin-api`。服务内部路由不包含 `/admin/operator` 前缀。

### 2. 请求头

业务接口均需要登录 Token，并经过 API 权限校验。

``` http
Authorization: Bearer <access_token>
Content-Type: application/json
X-Lang: zh-CN
```

-   `Authorization`：使用登录接口获得的 access token。
-   `Content-Type`：POST JSON 请求使用 `application/json`。
-   `X-Lang`：可选，指定语言；当前支持 `zh-CN`、`zh-HK`、`en-US`，默认
    `zh-CN`。

### 3. 通用响应结构

成功响应：

``` json
{
  "code": 0,
  "msg": "ok",
  "data": {}
}
```

失败响应示例：

``` json
{
  "code": 403,
  "msg": "无权访问"
}
```

常见状态：

  HTTP 状态码   含义
  ------------- -------------------------------------------------------
  `400`         请求参数不合法，或业务校验未通过
  `401`         未登录、Token 无效，或认证失败
  `403`         已登录，但没有对应 API 权限
  `498`         access token 已过期，需要按登录鉴权约定刷新或重新登录
  `500`         服务内部错误

错误响应的具体 `msg` 以服务实际返回为准。不要仅凭 HTTP
状态码以外的文案判断业务错误。

### 4. 分页与时间

-   `page`：页码，从 `1` 开始，必填。
-   `page_size`：每页数量，必填，范围 `1`～`100`。
-   列表按记录 ID 倒序返回。
-   本文两个层级接口中的 `created_at`、`updated_at` 是 **Unix
    毫秒时间戳**，前端按毫秒解析，不要按秒处理。

## 二、接口一览

  功能           方法   外部请求路径
  -------------- ------ ---------------------------------------------
  自动层级列表   GET    `/admin/operator/member/auto-level/list`
  创建自动层级   POST   `/admin/operator/member/auto-level/create`
  固定层级列表   GET    `/admin/operator/member/fixed-level/list`
  创建固定层级   POST   `/admin/operator/member/fixed-level/create`
  修改固定层级   POST   `/admin/operator/member/fixed-level/update`
  删除固定层级   POST   `/admin/operator/member/fixed-level/delete`

当前自动层级只提供列表和创建接口，暂未提供自动层级修改、删除接口。

## 三、会员自动层级

### 1. 获取自动层级列表

**请求**

``` http
GET /admin/operator/member/auto-level/list?page=1&page_size=20&name=VIP
```

  ----------------------------------------------------------------------------------------------------
  参数           位置           必填           类型           规则与说明
  -------------- -------------- -------------- -------------- ----------------------------------------
  `page`         query          是             int64          页码，从 `1` 开始

  `page_size`    query          是             int64          每页数量，`1`～`100`

  `name`         query          否             string         层级名称模糊查询，最多 100
                                                              个字符；不传或空字符串表示不按名称筛选
  ----------------------------------------------------------------------------------------------------

**成功响应示例**

``` json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "total": 2,
    "list": [
      {
        "id": 2,
        "name": "VIP2",
        "min_deposit": 1000,
        "min_water": 500,
        "current_member_count": 12,
        "created_at": 1793059200000,
        "updated_at": 1793059200000
      },
      {
        "id": 1,
        "name": "VIP1",
        "min_deposit": 100,
        "min_water": 50,
        "current_member_count": 35,
        "created_at": 1792972800000,
        "updated_at": 1792972800000
      }
    ]
  }
}
```

**响应字段**

  字段                            类型     说明
  ------------------------------- -------- ----------------------------
  `total`                         int64    符合查询条件的总记录数
  `list`                          array    当前页自动层级列表
  `list[].id`                     int64    层级 ID
  `list[].name`                   string   层级名称
  `list[].min_deposit`            int64    晋级要求的最低累计存款金额
  `list[].min_water`              int64    晋级要求的最低累计洗码量
  `list[].current_member_count`   int64    当前归属该自动层级的会员数
  `list[].created_at`             int64    创建时间，Unix 毫秒
  `list[].updated_at`             int64    更新时间，Unix 毫秒

### 2. 创建自动层级

**请求**

``` http
POST /admin/operator/member/auto-level/create
Content-Type: application/json
```

``` json
{
  "name": "VIP3",
  "min_deposit": 3000,
  "min_water": 1500
}
```

  字段            必填   类型     规则与说明
  --------------- ------ -------- --------------------------------------
  `name`          是     string   层级名称，不能为空，最多 100 个字符
  `min_deposit`   是     int64    最低累计存款金额，必须大于或等于 `0`
  `min_water`     是     int64    最低累计洗码量，必须大于或等于 `0`

**成功响应示例**

``` json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 3
  }
}
```

**业务校验**

-   服务端会校验新增层级的两项晋级门槛。
-   按当前实现，新增层级的 `min_deposit` 和 `min_water`
    必须分别严格大于当前分站所有已有自动层级的对应门槛；任意一项与已有层级冲突都会被拒绝。
-   名称和门槛的最终校验以服务端返回为准。

## 四、会员固定层级

### 1. 获取固定层级列表

**请求**

``` http
GET /admin/operator/member/fixed-level/list?page=1&page_size=20&name=普通
```

  ----------------------------------------------------------------------------------------------------
  参数           位置           必填           类型           规则与说明
  -------------- -------------- -------------- -------------- ----------------------------------------
  `page`         query          是             int64          页码，从 `1` 开始

  `page_size`    query          是             int64          每页数量，`1`～`100`

  `name`         query          否             string         层级名称模糊查询，最多 100
                                                              个字符；不传或空字符串表示不按名称筛选
  ----------------------------------------------------------------------------------------------------

**成功响应示例**

``` json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "total": 1,
    "list": [
      {
        "id": 10,
        "name": "普通会员",
        "description": "默认固定层级",
        "withdrawal_limit": null,
        "current_member_count": 28,
        "created_at": 1793059200000,
        "updated_at": 1793059200000
      }
    ]
  }
}
```

**响应字段**

  --------------------------------------------------------------------------------------
  字段                            类型                    说明
  ------------------------------- ----------------------- ------------------------------
  `total`                         int64                   符合查询条件的总记录数

  `list`                          array                   当前页固定层级列表

  `list[].id`                     int64                   层级 ID

  `list[].name`                   string                  层级名称

  `list[].description`            string                  层级说明；无说明时为空字符串

  `list[].withdrawal_limit`       int64 / null            单笔提款金额上限；`null`
                                                          表示不额外限制

  `list[].current_member_count`   int64                   当前关联到该固定层级的会员数

  `list[].created_at`             int64                   创建时间，Unix 毫秒

  `list[].updated_at`             int64                   更新时间，Unix 毫秒
  --------------------------------------------------------------------------------------

### 2. 创建固定层级

**请求**

``` http
POST /admin/operator/member/fixed-level/create
Content-Type: application/json
```

``` json
{
  "name": "高额会员",
  "description": "高额提款层级",
  "withdrawal_limit": 50000
}
```

  -------------------------------------------------------------------------------------------
  字段                 必填              类型              规则与说明
  -------------------- ----------------- ----------------- ----------------------------------
  `name`               是                string            层级名称，不能为空，最多 100
                                                           个字符

  `description`        否                string            层级说明，最多 255
                                                           个字符；不传时按空字符串处理

  `withdrawal_limit`   否                int64 / null      单笔提款金额上限；传入时必须大于
                                                           `0`；不传或传 `null`
                                                           表示不额外限制
  -------------------------------------------------------------------------------------------

**成功响应示例**

``` json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 11
  }
}
```

### 3. 修改固定层级

**请求**

``` http
POST /admin/operator/member/fixed-level/update
Content-Type: application/json
```

``` json
{
  "id": 11,
  "name": "高额会员",
  "description": "调整后的层级说明",
  "withdrawal_limit": 80000
}
```

  -------------------------------------------------------------------------------------------
  字段                 必填              类型              规则与说明
  -------------------- ----------------- ----------------- ----------------------------------
  `id`                 是                int64             层级 ID，必须大于 `0`

  `name`               是                string            层级名称，不能为空，最多 100
                                                           个字符

  `description`        否                string            层级说明，最多 255 个字符

  `withdrawal_limit`   否                int64 / null      单笔提款金额上限；传入时必须大于
                                                           `0`
  -------------------------------------------------------------------------------------------

**成功响应示例**

``` json
{
  "code": 0,
  "msg": "ok",
  "data": {}
}
```

**重要：修改接口当前按整组字段更新，不是局部 PATCH。**

-   `id`、`name` 必须传。
-   `description` 不传时会按空字符串写入，可能清空原说明。
-   `withdrawal_limit` 不传或传 `null`
    时会清除原提款上限，即改为"不额外限制"。
-   如果要保留原提款上限，前端应提交当前的 `withdrawal_limit` 数值。

### 4. 删除固定层级

**请求**

``` http
POST /admin/operator/member/fixed-level/delete
Content-Type: application/json
```

``` json
{
  "id": 11
}
```

  字段   必填   类型    规则与说明
  ------ ------ ------- -----------------------------------
  `id`   是     int64   要删除的固定层级 ID，必须大于 `0`

**成功响应示例**

``` json
{
  "code": 0,
  "msg": "ok",
  "data": {}
}
```

删除时，服务端会在事务内解除该层级与会员的关联，再删除固定层级记录；不会删除会员本身。删除后原来关联该层级的会员不再保留这条固定层级关联。

## 五、权限说明

当前接口均定义为 `Jwt + ActionLog + Authority` 保护接口：

-   需要携带有效 access token。
-   API 权限按 HTTP 方法和服务内部路径进行校验。
-   菜单可见权限与 API 调用权限是不同的授权项，不能只配置菜单就认为 API
    已授权。

当前 Catalog 中的 API 权限登记项：

  方法   服务内部路径
  ------ ------------------------------
  GET    `/member/auto-level/list`
  POST   `/member/auto-level/create`
  GET    `/member/fixed-level/list`
  POST   `/member/fixed-level/create`
  POST   `/member/fixed-level/update`
  POST   `/member/fixed-level/delete`

前端按钮权限码（来自当前菜单 Catalog）：

  操作           权限码
  -------------- -----------------------------
  创建自动层级   `member:auto-level:create`
  创建固定层级   `member:fixed-level:create`
  修改固定层级   `member:fixed-level:update`
  删除固定层级   `member:fixed-level:delete`

按钮权限码用于前端控制操作入口；后端 API 仍按 method + path 独立校验。

## 六、对接注意事项

1.  列表接口的 `page`、`page_size` 是必填查询参数，不要只传在 JSON body
    中。
2.  `created_at`、`updated_at` 使用 Unix 毫秒时间戳。
3.  固定层级的 `withdrawal_limit: null`
    表示不额外限制；更新时省略该字段也会清除原上限。
4.  固定层级更新接口要求提交 `id` 和
    `name`；不能只提交一个要修改的字段。
5.  当前自动层级没有更新和删除接口。
6.  外部 URL 使用 `/admin/operator` 前缀；API 权限 Catalog
    中登记的是去掉该前缀后的服务内部路径。
