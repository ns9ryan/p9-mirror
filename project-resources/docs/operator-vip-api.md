# 分站后台 VIP API 对接文档

- **建议存放路径：** `project-resources/docs/operator-vip-api.md`
- **适用服务：** `operator-admin-api`
- **对接范围：** VIP 基础配置、VIP 等级、VIP 图标组
- **更新日期：** 2026-10-28
- **代码依据：** `p9-mirror` 当前 `main` 分支，提交
  `8a64fe90fcdc3f378c032d45175871182ac0bf9a`

本文以当前 HTTP API 定义、路由、API Logic 及 CRM RPC
实现为依据。只列出当前 VIP HTTP API
中已定义的接口，不将尚未提供的功能写成可调用接口。

## 一、基础约定

### 1. 请求地址

VIP 接口在 `operator-admin-api` 内部的路由前缀为 `/member/vip`。

当前分站会员层级对接文档采用 `/admin/operator`
作为外部访问前缀，因此本文示例暂沿用该约定：

``` text
{分站域名}/admin/operator/member/vip/{接口路径}
```

例如：

``` http
GET https://ryan05.op.com/admin/operator/member/vip/config
```

经 Nginx 转发后，`operator-admin-api` 接收到的内部路径为：

``` text
/member/vip/config
```

直接访问 `operator-admin-api` 服务端口时，不带 `/admin/operator`
前缀。例如开发环境服务端口为 `38010` 时：

``` http
GET http://127.0.0.1:38010/member/vip/config
```

**路径前缀待核对：** `project-resources/docs/api-prefix.md`
记载的通用外部入口为 `/api`，与现有 `operator-member-level-api.md` 的
`/admin/operator` 约定不一致。本文暂按现有会员层级对接文档编写外部
URL；前端正式联调前，应核对目标环境实际生效的 Nginx
路由。无论外部前缀如何调整，本文列出的服务内部路由不变。

### 2. 请求头

业务接口均需要登录 Token，并经过 API 权限校验。

``` http
Authorization: Bearer <access_token>
Content-Type: application/json
X-Lang: zh-CN
```

| 请求头          | 说明                                                                     |
|:----------------|:-------------------------------------------------------------------------|
| `Authorization` | 登录后获得的 access token，所有 VIP 接口必需                             |
| `Content-Type`  | POST JSON 请求使用 `application/json`                                    |
| `X-Lang`        | 可选，指定错误消息语言；当前支持 `zh-CN`、`zh-HK`、`en-US`，默认 `zh-CN` |

分站归属由访问域名及服务端可信上下文确定，前端不需要、也不应自行提交
`operator_code` 来决定数据范围。

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

| HTTP 状态码 | 含义                                                  |
|:------------|:------------------------------------------------------|
| `400`       | 请求参数不合法，或业务校验未通过                      |
| `401`       | 未登录、Token 无效或认证失败                          |
| `403`       | 已登录，但没有对应 API 权限                           |
| `498`       | access token 已过期，需要按登录鉴权约定刷新或重新登录 |
| `500`       | 服务内部错误                                          |

失败响应的 `msg` 以服务实际返回为准，可能随 `X-Lang`
变化。前端不应依赖某一条固定文案来判断业务状态。

修改类接口成功后统一返回：

``` json
{
  "code": 0,
  "msg": "ok",
  "data": {}
}
```

### 4. 分页、时间和数值

- VIP 等级列表的 `page`、`page_size` 是必填 Query 参数。
- `page` 从 `1` 开始；`page_size` 范围为 `1`～`100`。
- VIP 等级列表按 `level` 升序返回。
- `updated_at`、`last_enabled_at` 等时间字段使用 **Unix
  毫秒时间戳**，前端按毫秒解析。
- 金额、等级编号、计数等字段按接口定义使用 JSON 数字。VIP API
  本身未声明金额的小数位换算规则，金额展示应遵循项目统一金额规范。

## 二、接口一览

下表的”服务内部路径”不包含外部 Nginx 前缀。

| 功能              | 方法 | 服务内部路径                |
|:------------------|:-----|:----------------------------|
| 获取 VIP 基础配置 | GET  | `/member/vip/config`        |
| 修改 VIP 基础配置 | POST | `/member/vip/config/update` |
| VIP 等级列表      | GET  | `/member/vip/level/list`    |
| VIP 等级详情      | GET  | `/member/vip/level/detail`  |
| 创建 VIP 等级     | POST | `/member/vip/level/create`  |
| 修改 VIP 等级     | POST | `/member/vip/level/update`  |
| 删除 VIP 等级     | POST | `/member/vip/level/delete`  |
| VIP 图标组列表    | GET  | `/member/vip/icon/list`     |
| 修改 VIP 图标     | POST | `/member/vip/icon/update`   |
| 恢复默认 VIP 图标 | POST | `/member/vip/icon/reset`    |

以上接口当前均定义为 `Jwt + ActionLog + Authority`
保护接口，权限登记现状见”权限与联调注意事项”。

## 三、VIP 基础配置

### 1. 获取 VIP 基础配置

**请求**

``` http
GET /member/vip/config
```

无请求参数。

**成功响应示例**

``` json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "vip_enabled": true,
    "upgrade_reward_enabled": true,
    "retention_check_enabled": false,
    "reward_grant_mode": 1,
    "reward_water_multiplier": 150000,
    "last_enabled_at": 1793059200000,
    "updated_at": 1793059200000
  }
}
```

**响应字段**

| 字段                      | 类型         | 说明                                                                   |
|:--------------------------|:-------------|:-----------------------------------------------------------------------|
| `vip_enabled`             | boolean      | VIP 总开关                                                             |
| `upgrade_reward_enabled`  | boolean      | 晋级奖励开关                                                           |
| `retention_check_enabled` | boolean      | 保级考核开关                                                           |
| `reward_grant_mode`       | int64        | 奖励派发模式：`1` 手动领取，`2` 自动派发                               |
| `reward_water_multiplier` | int64        | 晋级奖励洗码倍率，存储值放大 `100000` 倍                               |
| `last_enabled_at`         | int64 / null | 最近一次从关闭切换为开启的时间，Unix 毫秒；尚无开启时间时可能为 `null` |
| `updated_at`              | int64        | 配置更新时间，Unix 毫秒                                                |

`reward_water_multiplier` 的前后端换算：

- 前端显示倍率 = 接口值 ÷ `100000`。
- 前端提交值 = 用户填写的倍率 × `100000`。
- 例如 `1.5` 倍对应接口值 `150000`。

### 2. 修改 VIP 基础配置

**请求**

``` http
POST /member/vip/config/update
Content-Type: application/json
```

**请求示例**

``` json
{
  "vip_enabled": true,
  "upgrade_reward_enabled": true,
  "retention_check_enabled": false,
  "reward_grant_mode": 1,
  "reward_water_multiplier": 150000
}
```

**请求字段**

| 字段                      | 必填 | 类型    | 规则与说明                             |
|:--------------------------|:-----|:--------|:---------------------------------------|
| `vip_enabled`             | 否   | boolean | VIP 总开关                             |
| `upgrade_reward_enabled`  | 否   | boolean | 晋级奖励开关                           |
| `retention_check_enabled` | 否   | boolean | 保级考核开关                           |
| `reward_grant_mode`       | 否   | int64   | 只能为 `1` 或 `2`                      |
| `reward_water_multiplier` | 否   | int64   | 必须大于或等于 `0`；按 `100000` 倍传值 |

**成功响应**

``` json
{
  "code": 0,
  "msg": "ok",
  "data": {}
}
```

**更新规则**

- 这是**部分更新接口**。只更新本次提交的字段；未提交的字段保持原值。
- 所有字段都不传时，当前实现会直接返回成功，不修改配置。前端正常提交时应至少包含一个实际需要修改的字段。
- `reward_grant_mode` 只允许 `1` 或 `2`。
- `reward_water_multiplier` 不允许为负数。
- `vip_enabled` 仅在从 `false` 切换为 `true` 时更新
  `last_enabled_at`；重复提交 `true` 不会刷新该时间。
- 接口不返回更新后的配置。保存成功后，如需同步页面状态，应重新调用
  `GET /member/vip/config`。

## 四、VIP 等级

### 1. 获取 VIP 等级列表

**请求**

``` http
GET /member/vip/level/list?page=1&page_size=20
```

**Query 参数**

| 参数        | 必填 | 类型  | 规则与说明                |
|:------------|:-----|:------|:--------------------------|
| `page`      | 是   | int64 | 页码，从 `1` 开始         |
| `page_size` | 是   | int64 | 每页数量，范围 `1`～`100` |

**成功响应示例**

``` json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "total": 2,
    "list": [
      {
        "id": 100,
        "level": 0,
        "upgrade_deposit": 0,
        "upgrade_water": 0,
        "retention_deposit": 0,
        "retention_water": 0,
        "upgrade_reward": 0,
        "member_count": 20,
        "updated_at": 1793059200000
      },
      {
        "id": 101,
        "level": 1,
        "upgrade_deposit": 1000,
        "upgrade_water": 500,
        "retention_deposit": 500,
        "retention_water": 200,
        "upgrade_reward": 100,
        "member_count": 8,
        "updated_at": 1793059200000
      }
    ]
  }
}
```

**响应字段**

| 字段                       | 类型  | 说明                                 |
|:---------------------------|:------|:-------------------------------------|
| `total`                    | int64 | 符合条件的等级总数                   |
| `list`                     | array | 当前页 VIP 等级列表，按 `level` 升序 |
| `list[].id`                | int64 | VIP 等级 ID                          |
| `list[].level`             | int64 | VIP 等级编号，从 `0` 开始            |
| `list[].upgrade_deposit`   | int64 | 晋级所需存款                         |
| `list[].upgrade_water`     | int64 | 晋级所需洗码量                       |
| `list[].retention_deposit` | int64 | 保级所需存款                         |
| `list[].retention_water`   | int64 | 保级所需洗码量                       |
| `list[].upgrade_reward`    | int64 | 晋级奖励金额                         |
| `list[].member_count`      | int64 | 当前归属该 VIP 等级的会员数          |
| `list[].updated_at`        | int64 | 更新时间，Unix 毫秒                  |

### 2. 获取 VIP 等级详情

**请求**

``` http
GET /member/vip/level/detail?id=101
```

**Query 参数**

| 参数 | 必填 | 类型  | 规则与说明                |
|:-----|:-----|:------|:--------------------------|
| `id` | 是   | int64 | VIP 等级 ID，必须大于 `0` |

**成功响应示例**

``` json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 101,
    "level": 1,
    "upgrade_deposit": 1000,
    "upgrade_water": 500,
    "retention_deposit": 500,
    "retention_water": 200,
    "upgrade_reward": 100,
    "member_count": 8,
    "updated_at": 1793059200000
  }
}
```

详情的 `data` 字段与列表中单个 `list[]` 对象相同。

### 3. 创建 VIP 等级

**请求**

``` http
POST /member/vip/level/create
Content-Type: application/json
```

**请求示例**

``` json
{
  "upgrade_deposit": 3000,
  "upgrade_water": 1500,
  "retention_deposit": 1000,
  "retention_water": 500,
  "upgrade_reward": 200
}
```

**请求字段**

| 字段                | 必填 | 类型  | 规则与说明                         |
|:--------------------|:-----|:------|:-----------------------------------|
| `upgrade_deposit`   | 是   | int64 | 晋级所需存款，必须大于或等于 `0`   |
| `upgrade_water`     | 是   | int64 | 晋级所需洗码量，必须大于或等于 `0` |
| `retention_deposit` | 是   | int64 | 保级所需存款，必须大于或等于 `0`   |
| `retention_water`   | 是   | int64 | 保级所需洗码量，必须大于或等于 `0` |
| `upgrade_reward`    | 是   | int64 | 晋级奖励金额，必须大于或等于 `0`   |

**成功响应示例**

``` json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 102,
    "level": 2
  }
}
```

响应字段：

| 字段    | 类型  | 说明                 |
|:--------|:------|:---------------------|
| `id`    | int64 | 新建 VIP 等级 ID     |
| `level` | int64 | 服务端分配的等级编号 |

**业务规则**

- 不需要、也不能提交
  `level`；等级编号由服务端根据当前最高等级自动分配为下一个编号。
- VIP `0` 是分站初始化时创建的系统基础等级。
- 如果分站尚未初始化 VIP
  等级数据，创建接口会失败，应先检查分站初始化状态。

### 4. 修改 VIP 等级

**请求**

``` http
POST /member/vip/level/update
Content-Type: application/json
```

**请求示例：只修改晋级奖励**

``` json
{
  "id": 102,
  "upgrade_reward": 300
}
```

**请求字段**

| 字段                | 必填 | 类型  | 规则与说明                               |
|:--------------------|:-----|:------|:-----------------------------------------|
| `id`                | 是   | int64 | VIP 等级 ID，必须大于 `0`                |
| `upgrade_deposit`   | 否   | int64 | 晋级所需存款，传入时必须大于或等于 `0`   |
| `upgrade_water`     | 否   | int64 | 晋级所需洗码量，传入时必须大于或等于 `0` |
| `retention_deposit` | 否   | int64 | 保级所需存款，传入时必须大于或等于 `0`   |
| `retention_water`   | 否   | int64 | 保级所需洗码量，传入时必须大于或等于 `0` |
| `upgrade_reward`    | 否   | int64 | 晋级奖励金额，传入时必须大于或等于 `0`   |

**成功响应**

``` json
{
  "code": 0,
  "msg": "ok",
  "data": {}
}
```

**重要：这是部分更新接口。**

- `id` 必须传入；其余字段只更新实际提交的字段。
- `id` 之外至少要提供一个待修改字段；只提交 `id` 会被拒绝。
- VIP `0` 的晋级、保级及奖励条件均不可修改。
- 如果当前等级已有会员，实际修改晋级或保级门槛会被拒绝。当前实现尚未支持修改门槛后自动重新计算关联会员等级。
- 对于非 VIP `0` 等级，单独修改 `upgrade_reward` 不受上述会员数限制。
- 保存成功后如需刷新页面数据，应重新获取等级列表或详情。

### 5. 删除 VIP 等级

**请求**

``` http
POST /member/vip/level/delete
Content-Type: application/json
```

**请求 Body**

``` json
{
  "id": 102
}
```

| 字段 | 必填 | 类型  | 规则与说明                         |
|:-----|:-----|:------|:-----------------------------------|
| `id` | 是   | int64 | 要删除的 VIP 等级 ID，必须大于 `0` |

**成功响应**

``` json
{
  "code": 0,
  "msg": "ok",
  "data": {}
}
```

**删除限制**

- VIP `0` 是系统基础等级，禁止删除。
- 只允许删除当前最高等级，以保持等级编号连续。
- 当前等级关联会员数必须为 `0`；有会员归属时，删除请求会被拒绝。
- 删除失败时以服务端返回的错误响应为准。前端可以根据列表中的
  `member_count` 和 `level` 提前控制按钮状态，但后端校验仍然有效。

## 五、VIP 图标组

### 1. 获取 VIP 图标组列表

**请求**

``` http
GET /member/vip/icon/list
```

无请求参数，列表按 `group_no` 升序返回。

**成功响应示例**

``` json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "list": [
      {
        "id": 1,
        "group_no": 0,
        "start_level": 0,
        "end_level": 9,
        "default_icon_url": "https://cdn.example.com/vip/default-0.png",
        "icon_url": "https://cdn.example.com/vip/custom-0.png",
        "updated_at": 1793059200000
      },
      {
        "id": 2,
        "group_no": 1,
        "start_level": 10,
        "end_level": 19,
        "default_icon_url": "https://cdn.example.com/vip/default-1.png",
        "icon_url": "https://cdn.example.com/vip/default-1.png",
        "updated_at": 1793059200000
      }
    ]
  }
}
```

**响应字段**

| 字段                      | 类型   | 说明                                         |
|:--------------------------|:-------|:---------------------------------------------|
| `list`                    | array  | VIP 图标组列表                               |
| `list[].id`               | int64  | 图标组 ID                                    |
| `list[].group_no`         | int64  | 图标组编号                                   |
| `list[].start_level`      | int64  | 起始 VIP 等级，当前按 `group_no × 10` 计算   |
| `list[].end_level`        | int64  | 结束 VIP 等级，当前按 `start_level + 9` 计算 |
| `list[].default_icon_url` | string | 系统默认图标 URL                             |
| `list[].icon_url`         | string | 当前实际使用的图标 URL                       |
| `list[].updated_at`       | int64  | 更新时间，Unix 毫秒                          |

`default_icon_url` 是系统默认值，`icon_url`
是当前使用值。前端展示当前图标时应使用 `icon_url`。

### 2. 修改 VIP 图标

**请求**

``` http
POST /member/vip/icon/update
Content-Type: application/json
```

**请求 Body**

``` json
{
  "group_no": 0,
  "icon_url": "https://cdn.example.com/vip/custom-0.png"
}
```

| 字段       | 必填 | 类型   | 规则与说明                                           |
|:-----------|:-----|:-------|:-----------------------------------------------------|
| `group_no` | 是   | int64  | 图标组编号，必须大于或等于 `0`，且对应图标组必须存在 |
| `icon_url` | 是   | string | 自定义图标 URL；不能为空白字符串                     |

**成功响应**

``` json
{
  "code": 0,
  "msg": "ok",
  "data": {}
}
```

服务端会去除 `icon_url` 首尾空白，但当前实现不校验 URL
格式或图片是否可访问。前端应确保传入的是可用的图片地址。该接口只更新图标地址，不负责上传图片文件。

### 3. 恢复默认 VIP 图标

**请求**

``` http
POST /member/vip/icon/reset
Content-Type: application/json
```

**请求 Body**

``` json
{
  "group_no": 0
}
```

| 字段       | 必填 | 类型  | 规则与说明                                                           |
|:-----------|:-----|:------|:---------------------------------------------------------------------|
| `group_no` | 是   | int64 | 要恢复默认图标的图标组编号，必须大于或等于 `0`，且对应图标组必须存在 |

**成功响应**

``` json
{
  "code": 0,
  "msg": "ok",
  "data": {}
}
```

恢复后，服务端会将该组的 `icon_url` 重置为
`default_icon_url`。该接口一次处理一个图标组；如需刷新页面状态，重新调用
`GET /member/vip/icon/list`。

## 六、权限与联调注意事项

### 1. 接口鉴权

当前 VIP 路由均使用以下中间件：

``` text
Jwt + ActionLog + Authority
```

因此，调用接口不仅需要有效 Token，还需要 Core 中存在并授予对应 API
权限。菜单可见权限与 API 调用权限是不同的授权项。

### 2. 当前 API Catalog 登记缺口

核对当前 `operator-admin-api/internal/catalog/catalog.go` 后发现：

- 当前源码已登记会员自动层级和固定层级 API，但尚未登记本文列出的 VIP
  API。
- 当前源码也没有 VIP 菜单及对应的按钮权限码定义。
- 因此，虽然 VIP HTTP 路由已经存在，正常启用 `Authority` 时仍需确认 Core
  中是否已通过其他方式登记权限；否则可能返回 `403`。

需要补充登记的 API 项如下，路径均为服务内部路径：

| 方法 | API 路径                    |
|:-----|:----------------------------|
| GET  | `/member/vip/config`        |
| POST | `/member/vip/config/update` |
| GET  | `/member/vip/level/list`    |
| GET  | `/member/vip/level/detail`  |
| POST | `/member/vip/level/create`  |
| POST | `/member/vip/level/update`  |
| POST | `/member/vip/level/delete`  |
| GET  | `/member/vip/icon/list`     |
| POST | `/member/vip/icon/update`   |
| POST | `/member/vip/icon/reset`    |

**本文只记录当前代码现状，不预设或虚构 VIP 按钮权限码。**
菜单、按钮权限和 API Catalog 应在后续权限配置中一并确认。

### 3. 前端对接建议

1.  打开 VIP 页面时，可分别请求
    `GET /member/vip/config`、`GET /member/vip/level/list`、`GET /member/vip/icon/list`
    获取页面初始数据。
2.  修改 VIP 基础配置时，只提交本次修改的字段；保存成功后重新获取配置。
3.  创建等级时使用服务端返回的
    `id`、`level`；不要在前端自行分配等级编号。
4.  修改等级时始终传
    `id`，并注意已有会员的等级不允许修改晋级、保级门槛。
5.  删除等级前，可根据 `level` 和 `member_count`
    提示用户；最终结果以后端校验为准。
6.  修改或恢复图标后重新获取图标组列表，确保展示的是服务端当前保存的
    `icon_url`。
7.  `updated_at`、`last_enabled_at` 均按 Unix 毫秒处理。
8.  前端应根据接口返回的 `code` 和当前鉴权流程处理错误，不要仅凭 `msg`
    文案判断结果。

### 4. 当前 VIP HTTP API 未定义的功能

当前 `vip.api` 和路由中没有定义以下 HTTP
接口，不应直接按本文接口进行调用：

- VIP 会员列表或会员 VIP 详情。
- VIP 晋级、保级记录或奖励领取记录。
- 批量修改或批量恢复图标。
- 前端直接调用的 VIP 等级初始化接口。

如后续需要这些功能，应以新增的 API 定义和实现为准。
