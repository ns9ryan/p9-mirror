# operator-base API 文档

用于对接分站后台的站点配置页，以及登录前的域名识别。

---

## 一、服务地址与测试准备

| 项目 | 当前开发环境示例 | 怎么用 |
| --- | --- | --- |
| 服务地址 | `http://192.168.0.15:38001` | 各接口地址前面的公共部分 |
| 查询配置的完整地址 | `http://192.168.0.15:38001/operator/info` | Apipost 选择 GET，填入这个地址 |
| 测试站点后台域名 | `ceshi3.back.com` | 直连 IP 测试时填在请求头 `Host` 中 |
| 测试站点编码 | `OP_57BE18` | 登录时使用；保存配置时不放进 Body |
| 登录 Token | `<access_token>` | 使用上述站点登录得到的 Token，替换示例占位符 |

下文请求均按当前开发环境填写。换环境时替换服务地址、已登记的后台域名和对应 Token。图片地址、记录 ID 和时间戳只是响应示例。

**Host 填域名，Authorization 填这个站点的登录 Token，两者要属于同一个站点。** `/operator/code` 是登录前查询，不需要 Token；其他五个接口需要登录和对应的接口权限。

浏览器会自动带 Host，应通过已登记的后台域名访问；Apipost 直连 IP 时才手动指定 Host。服务实际路径为 `/operator/...`，直连时不要额外加 `/base`。使用网关时按部署方提供的外部地址访问。

登录和刷新 Token 的接口见 [Core API 文档](./core-api.md)。

### 所有响应最外层的三个字段

| 字段 | 类型 | 示例 | 含义 |
| --- | --- | --- | --- |
| `code` | int | `0` | 成功为 `0`；失败时为对应错误状态码，例如 `400` |
| `data` | object | `{"operator_code":"OP_57BE18"}` | 本次接口返回的业务内容；错误响应可能没有此字段 |
| `msg` | string | `"ok"` | 结果描述；失败时说明原因，前端不要依赖固定文案判断错误类型 |

六个接口均不分页。下面每个接口都列出自己的请求和响应，可直接按对应小节对接。

---

## 二、接口索引

| 方法 | 路径 | 说明 | 登录 |
| --- | --- | --- | --- |
| GET | `/operator/code` | 登录前：用后台域名查询站点编码 | 不需要 |
| GET | `/operator/info` | 页面回显：查询语言、地区、货币、时区、标题、Logo 等配置 | 需要 |
| PUT | `/operator/settings` | 保存标题、Logo、四个阈值、两个验证码模式，共八项 | 需要 |
| PUT | `/operator/regions` | 开关本站地区：传入要开启的地区，未传的已分配地区关闭 | 需要 |
| PUT | `/operator/languages` | 开关本站语言：传入要开启的语言，未传的已分配语言关闭 | 需要 |
| POST | `/operator/logo` | 上传一张 Logo，返回图片地址，随后还要保存设置 | 需要 |

---

## 三、登录前查询站点编码

### GET /operator/code

**用途：登录前，通过后台域名获取站点编码，再把编码交给原登录接口。这个接口本身不登录。**

请求示例：

```http
GET http://192.168.0.15:38001/operator/code?domain_name=ceshi3.back.com
```

**请求头：** 不需要 Authorization，也不需要手动把 Host 改成站点域名。

**Body：** 不传。

**Query 参数：**

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `domain_name` | string | 是 | 例如 `ceshi3.back.com`；只传已登记的后台域名，不带协议、端口、路径、IP 或通配符 |

域名会去除首尾空白、忽略大小写，并去除末尾的一个点。仅查询类型为 `1` 的后台域名。本接口由 Query 中的域名决定查询目标，不要求调用者把 Host 改成目标站点。

成功响应：

```json
{
  "code": 0,
  "data": {
    "operator_code": "OP_57BE18"
  },
  "msg": "ok"
}
```

`data` 字段：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `operator_code` | string | 域名所属站点的唯一业务编码；例如 `OP_57BE18`，登录时使用 |

**错误响应示例：后台域名未登记，HTTP 404。**

```json
{
  "code": 404,
  "msg": "operator domain not found"
}
```

域名格式错误返回 `400`，后台域名未登记返回 `404`。本接口只读，不创建设置记录；查询成功不代表已经登录，也不代表站点当前一定允许登录。

前端登录流程：读取 `window.location.hostname` → 调用本接口 → 使用返回的 `operator_code`，按原登录接口要求提交账号和密码。

---

## 四、查询本站配置

### GET /operator/info

**用途：查询本站的语言、地区、结算货币、时区，以及标题、Logo、登录和注册相关设置，用于配置页面回显。**

**请求地址与请求头：**

```http
GET http://192.168.0.15:38001/operator/info
Authorization: Bearer <access_token>
Host: ceshi3.back.com
X-Lang: zh-CN
```

| 请求头 | 必填 | 填什么 |
| --- | --- | --- |
| `Authorization` | 是 | `Bearer ` 后面接当前站点的登录 Token，注意中间有一个空格 |
| `Host` | 是 | 当前站点已登记的后台域名，例如 `ceshi3.back.com`；浏览器自动携带 |
| `X-Lang` | 否 | 请求语言，例如 `zh-CN`；不会把所有基础数据名称自动翻译 |

**Query / Body：** 都不传。通过 Host 和登录身份确定当前站点，无需传站点编号。

**成功响应：HTTP 200。** 以下示例开启英语、简体中文、日本、越南；繁体中文和菲律宾仍是可选项，但本站已关闭。

```json
{
  "code": 0,
  "data": {
    "site_config": {
      "operator_code": "OP_57BE18",
      "domain_name": "ceshi3.back.com",
      "language_codes": ["en-US", "zh-CN"],
      "region_codes": ["JP", "VN"],
      "settlement_currency_code": "JPY",
      "timezone_code": "Asia/Tokyo"
    },
    "languages": [
      {
        "id": 46,
        "lang": "en-US",
        "disabled": 0,
        "status": 1,
        "effective_status": 1,
        "created_at": 1792479939,
        "updated_at": 1792479939,
        "name": "English",
        "sort_no": 4,
        "i18n_key": "lang.en-US"
      },
      {
        "id": 44,
        "lang": "zh-CN",
        "disabled": 0,
        "status": 1,
        "effective_status": 1,
        "created_at": 1792479939,
        "updated_at": 1792479939,
        "name": "简体中文",
        "sort_no": 2,
        "i18n_key": "lang.zh-CN"
      },
      {
        "id": 45,
        "lang": "zh-HK",
        "disabled": 0,
        "status": 2,
        "effective_status": 2,
        "created_at": 1792479939,
        "updated_at": 1792479939,
        "name": "繁體中文",
        "sort_no": 3,
        "i18n_key": "lang.zh-HK"
      }
    ],
    "region_options": [
      {
        "region": {
          "id": 1,
          "code": "JP",
          "calling_code": "81",
          "name_key": "region.jp.name",
          "status": 1,
          "sort_no": 10
        },
        "status": 1,
        "effective_status": 1
      },
      {
        "region": {
          "id": 2,
          "code": "VN",
          "calling_code": "84",
          "name_key": "region.vn.name",
          "status": 1,
          "sort_no": 20
        },
        "status": 1,
        "effective_status": 1
      },
      {
        "region": {
          "id": 3,
          "code": "PH",
          "calling_code": "63",
          "name_key": "region.ph.name",
          "status": 1,
          "sort_no": 30
        },
        "status": 2,
        "effective_status": 2
      }
    ],
    "effective_region_codes": ["JP", "VN"],
    "effective_language_codes": ["en-US", "zh-CN"],
    "currency": {
      "id": 4,
      "code": "JPY",
      "name_key": "currency.jpy.name",
      "currency_type": 1,
      "symbol": "¥",
      "amount_factor": 1,
      "status": 1,
      "sort_no": 5
    },
    "timezone": {
      "id": 2,
      "code": "Asia/Tokyo",
      "name_key": "timezone.asia_tokyo.name",
      "status": 1,
      "sort_no": 20
    },
    "site_settings": {
      "operator_code": "OP_57BE18",
      "site_title": "ceshi3",
      "site_logo_url": "https://cdn.example.com/defaults/site-logo.png",
      "email_login_limit": 0,
      "sms_login_limit": 0,
      "member_password_error_limit": 0,
      "ip_member_register_limit": 0,
      "register_captcha_mode": 0,
      "login_captcha_mode": 0,
      "updated_by": 0,
      "created_at": 0,
      "updated_at": 0,
      "configured": false
    }
  },
  "msg": "ok"
}
```

### data 顶层字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `site_config` | object | 本站编码、后台域名及选定的基础配置 |
| `languages` | array<object> | 分配给本站的语言选项，包含本站关闭的选项 |
| `region_options` | array<object> | 分配给本站的地区选项，包含本站关闭的选项 |
| `effective_region_codes` | array<string> | 当前实际生效的地区编码；没有时为 `[]` |
| `effective_language_codes` | array<string> | 当前实际生效的语言编码；没有时为 `[]` |
| `currency` | object | 本站结算货币的基础资料 |
| `timezone` | object | 本站时区的基础资料 |
| `site_settings` | object | 本站标题、Logo、阈值、验证码模式及保存状态 |

### site_config 字段

| 字段 | 类型 | 示例 | 含义 |
| --- | --- | --- | --- |
| `operator_code` | string | `"OP_57BE18"` | 当前站点的业务编码 |
| `domain_name` | string | `"ceshi3.back.com"` | 本次查询对应的后台域名 |
| `language_codes` | array<string> | `["en-US","zh-CN"]` | 本站勾选了英语和简体中文；不保证都实际可用 |
| `region_codes` | array<string> | `["JP","VN"]` | 本站勾选了日本和越南；不保证都实际可用 |
| `settlement_currency_code` | string | `"JPY"` | 本站结算货币为日元 |
| `timezone_code` | string | `"Asia/Tokyo"` | 本站使用东京时区 |

### languages 数组元素字段

| 字段 | 类型 | 示例 | 含义 |
| --- | --- | --- | --- |
| `id` | int64 | `46` | 当前节点 Core 中的语言资源 ID，保存选择时不用传 |
| `lang` | string | `"en-US"` | 英语的语言编码，保存语言选择时传这个值 |
| `disabled` | int32 | `0` | Core 语言开关：`0` 未禁用，`1` 已禁用；不表示本站是否勾选 |
| `status` | int64 | `1` | 本站是否勾选：`1` 开启，`2` 关闭 |
| `effective_status` | int64 | `1` | 最终能否使用：`1` 生效，`2` 未生效；需本站开启且当前节点 Core 中存在、未禁用 |
| `created_at` | int64 | `1792479939` | 语言资源创建时间，Unix 秒级时间戳 |
| `updated_at` | int64 | `1792479939` | 语言资源更新时间，Unix 秒级时间戳 |
| `name` | string | `"English"` | 用于展示的语言名称；基础资料缺失时可能为空 |
| `sort_no` | int32 | `4` | 排序号，用于页面排序 |
| `i18n_key` | string | `"lang.en-US"` | 前端翻译英语名称所用的 Key |

例如上方 `zh-HK` 的 `status=2`、`effective_status=2`，表示繁体中文已分配给本站，但本站关闭了它。`disabled` 来自当前节点 Core，不是实时读取总站语言开关。

### region_options 数组元素字段

| 字段 | 类型 | 示例 | 含义 |
| --- | --- | --- | --- |
| `region` | object | `{"code":"JP",...}` | 日本的基础资料，完整字段见下表 |
| `status` | int64 | `1` | 本站是否勾选该地区：`1` 开启，`2` 关闭 |
| `effective_status` | int64 | `1` | 最终能否使用：本站开启且基础数据存在、开启时为 `1`，否则为 `2` |

`region` 对象字段：

| 字段 | 类型 | 示例 | 含义 |
| --- | --- | --- | --- |
| `id` | int64 | `1` | 地区基础数据 ID，保存选择时不用传 |
| `code` | string | `"JP"` | 日本的地区编码，保存地区选择时传这个值 |
| `calling_code` | string | `"81"` | 日本国际电话区号，按字符串处理 |
| `name_key` | string | `"region.jp.name"` | 前端翻译地区名称所用的 Key |
| `status` | int64 | `1` | 基础数据是否开启：`1` 开启，`2` 关闭；不是本站选择 |
| `sort_no` | int64 | `10` | 地区排序号 |

例如上方菲律宾的 `region.status=1`、外层 `status=2`，表示基础数据允许使用菲律宾，但本站没有开启它，所以 `effective_status=2`。

### currency 字段

| 字段 | 类型 | 示例 | 含义 |
| --- | --- | --- | --- |
| `id` | int64 | `4` | 货币基础数据 ID |
| `code` | string | `"JPY"` | 日元的货币编码 |
| `name_key` | string | `"currency.jpy.name"` | 前端翻译货币名称所用的 Key |
| `currency_type` | int64 | `1` | `1` 法定货币，`2` 虚拟货币 |
| `symbol` | string | `"¥"` | 展示金额时使用的货币符号 |
| `amount_factor` | int64 | `1` | 金额换算倍率，例如 `JPY=1`、`USD=100`；不是货币汇率 |
| `status` | int64 | `1` | 货币基础数据状态：`1` 开启，`2` 关闭 |
| `sort_no` | int64 | `5` | 货币排序号 |

### timezone 字段

| 字段 | 类型 | 示例 | 含义 |
| --- | --- | --- | --- |
| `id` | int64 | `2` | 时区基础数据 ID |
| `code` | string | `"Asia/Tokyo"` | 东京时区编码 |
| `name_key` | string | `"timezone.asia_tokyo.name"` | 前端翻译时区名称所用的 Key |
| `status` | int64 | `1` | 时区基础数据状态：`1` 开启，`2` 关闭 |
| `sort_no` | int64 | `20` | 时区排序号 |

### site_settings 字段

| 字段 | 类型 | 示例 | 含义 |
| --- | --- | --- | --- |
| `operator_code` | string | `"OP_57BE18"` | 当前站点编码 |
| `site_title` | string | `"ceshi3"` | 页面显示的站点标题；历史站点未保存时默认取分站名称 |
| `site_logo_url` | string | `"https://cdn.example.com/defaults/site-logo.png"` | 展示 Logo 使用的地址；也可能是站内绝对路径或空字符串 |
| `email_login_limit` | int64 | `0` | 邮箱登录阈值；`0` 不限制，允许范围 `0`～`100` |
| `sms_login_limit` | int64 | `0` | 短信登录阈值；`0` 不限制，允许范围 `0`～`100` |
| `member_password_error_limit` | int64 | `0` | 会员密码错误次数上限；`0` 不限制，允许范围 `0`～`100` |
| `ip_member_register_limit` | int64 | `0` | 同一 IP 最多注册会员数；`0` 不限制，允许范围 `0`～`100` |
| `register_captcha_mode` | int64 | `0` | 注册页验证码：`0` 不验证，`1` 拖拽，`2` 图形 |
| `login_captcha_mode` | int64 | `0` | 登录页验证码：`0` 不验证，`1` 拖拽，`2` 图形 |
| `updated_by` | int64 | `0` | 最后修改人的用户 ID；例如 `21` 为用户 21 修改，`0` 为初始化或尚未保存 |
| `created_at` | int64 | `0` | 设置记录创建时间，Unix 秒；无记录时为 `0` |
| `updated_at` | int64 | `0` | 设置记录最后修改时间，Unix 秒；无记录时为 `0` |
| `configured` | bool | `false` | 是否已有设置记录；`false` 为尚未保存、使用默认值，`true` 为已有记录 |

货币和时区在此接口中只查询，下面的地区、语言和八项设置保存接口均不修改结算货币或时区。

### 错误响应

例如 Host 对应的站点与 Token 中的站点不一致，返回 HTTP 403：

```json
{
  "code": 403,
  "msg": "operator domain does not match login operator"
}
```

未登录或 Token 失效通常为 `401`，无查询权限为 `403`，后台域名未登记为 `404`。具体错误描述以实际响应为准。

### 默认值与页面回显

- 新建分站时自动创建设置记录：标题取分站名称，Logo 取 RPC 配置 `DefaultSiteLogoURL`，未配置时为空；四个阈值和两个验证码模式为 `0`，`updated_by=0`，`configured=true`。
- 历史站点没有设置记录时，查询返回同样的默认值，但 `configured=false`，修改人和时间为 `0`。查询不会自动插入设置记录，首次完整保存后才创建记录。
- 已保存的标题、Logo 和其他设置不会因默认配置调整而被覆盖。
- 复选框应按本站 `status` 或 `site_config` 中的选择列表回显；实际可用范围使用 `effective_*`。不要用实际生效列表覆盖本站已保存的选择。
- 基础数据临时关闭时，可以出现本站 `status=1`、`effective_status=2`；恢复基础数据后可重新生效。
- 已分配的语言或地区基础资料缺失时，仍可返回仅保留编码的占位对象：其他字符串为空、数值为 `0`，实际状态为 `2`。语言占位对象的 `disabled=0` 也不代表可用，应以 `effective_status` 为准。
- 结算货币或时区尚未配置、或对应基础资料不存在时，对象中的编码保留配置值，其他字段可能为空或为 `0`；此时 `status=0` 是缺少资料的占位值，不是正常的开启或关闭状态。
- `name_key`、`i18n_key` 用于前端翻译；不要假定响应额外包含翻译后的 `name`、UTC 偏移等未定义字段。

---

## 五、保存本站设置

### PUT /operator/settings

**用途：保存站点标题、Logo、四个阈值和两个验证码模式。只修改其中一项时，也要把其他七项的当前值一起提交。**

**请求地址与请求头：**

```http
PUT http://192.168.0.15:38001/operator/settings
Authorization: Bearer <access_token>
Host: ceshi3.back.com
Content-Type: application/json
```

| 请求头 | 必填 | 填什么 |
| --- | --- | --- |
| `Authorization` | 是 | 当前站点的 `Bearer <access_token>` |
| `Host` | 是 | 当前站点的后台域名，浏览器自动携带 |
| `Content-Type` | 是 | `application/json`，Apipost 的 Body 选择 JSON |
| `X-Lang` | 否 | 请求语言，例如 `zh-CN` |

**Query：** 不传。

**Body：** 只传下面八个字段，未修改的字段取自 `/operator/info` 返回的 `data.site_settings`。

```json
{
  "site_title": "ceshi3",
  "site_logo_url": "https://cdn.example.com/logo/OP_57BE18/20261027/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.png",
  "email_login_limit": 0,
  "sms_login_limit": 5,
  "member_password_error_limit": 10,
  "ip_member_register_limit": 3,
  "register_captcha_mode": 1,
  "login_captcha_mode": 2
}
```

Body 参数：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `site_title` | string | 是 | 标题，去除首尾空白后为 `1`～`100` 字节；不是 100 个汉字 |
| `site_logo_url` | string | 是 | Logo 地址，最多 `1024` 字节；HTTP(S) URL 或以 `/` 开头的站内路径；`""` 表示清除 |
| `email_login_limit` | int64 | 是 | 邮箱登录阈值，`0`～`100`；`0` 不限制 |
| `sms_login_limit` | int64 | 是 | 短信登录阈值，`0`～`100`；`0` 不限制 |
| `member_password_error_limit` | int64 | 是 | 会员密码错误次数上限，`0`～`100`；`0` 不限制 |
| `ip_member_register_limit` | int64 | 是 | 同一 IP 最多注册会员数，`0`～`100`；`0` 不限制 |
| `register_captcha_mode` | int64 | 是 | 注册页验证模式：`0` 不验证，`1` 拖拽，`2` 图形 |
| `login_captcha_mode` | int64 | 是 | 登录页验证模式：`0` 不验证，`1` 拖拽，`2` 图形 |

校验规则：

- 任一字段省略或传 `null` 返回 `400`；数值必须是整数，不能使用字符串、布尔值或小数。
- 不允许额外字段，包括 `operator_id`、`operator_code`、`domain_name`、`updated_by`；归属及修改人由后端确定。
- Body 只能有一个 JSON 对象，最大 `16 KiB`。
- Logo 不能使用以 `//` 开头的地址或包含用户名、密码的 URL。

**成功响应：HTTP 200，返回实际保存后的设置。**

```json
{
  "code": 0,
  "data": {
    "operator_code": "OP_57BE18",
    "site_title": "ceshi3",
    "site_logo_url": "https://cdn.example.com/logo/OP_57BE18/20261027/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.png",
    "email_login_limit": 0,
    "sms_login_limit": 5,
    "member_password_error_limit": 10,
    "ip_member_register_limit": 3,
    "register_captcha_mode": 1,
    "login_captcha_mode": 2,
    "updated_by": 21,
    "created_at": 1793089648,
    "updated_at": 1793090000,
    "configured": true
  },
  "msg": "ok"
}
```

**响应 data 字段：**

| 字段 | 类型 | 示例 | 含义 |
| --- | --- | --- | --- |
| `operator_code` | string | `"OP_57BE18"` | 本次保存所属的站点编码，由后端确定 |
| `site_title` | string | `"ceshi3"` | 保存后的站点标题 |
| `site_logo_url` | string | `"https://cdn.example.com/logo/OP_57BE18/20261027/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.png"` | 保存后的 Logo 地址；清除后为空字符串 |
| `email_login_limit` | int64 | `0` | 保存后的邮箱登录阈值；`0` 不限制 |
| `sms_login_limit` | int64 | `5` | 保存后的短信登录阈值为 `5`；`0` 不限制 |
| `member_password_error_limit` | int64 | `10` | 保存后的会员密码错误次数上限为 `10`；`0` 不限制 |
| `ip_member_register_limit` | int64 | `3` | 保存后的同一 IP 最多注册会员数为 `3`；`0` 不限制 |
| `register_captcha_mode` | int64 | `1` | 注册页使用拖拽验证；`0` 不验证，`2` 图形验证 |
| `login_captcha_mode` | int64 | `2` | 登录页使用图形验证；`0` 不验证，`1` 拖拽验证 |
| `updated_by` | int64 | `21` | 本次修改的登录用户 ID，客户端不用传 |
| `created_at` | int64 | `1793089648` | 设置记录首次创建时间，Unix 秒；后续更新保留 |
| `updated_at` | int64 | `1793090000` | 本次保存时间，Unix 秒 |
| `configured` | bool | `true` | 已有设置记录；保存成功后为 `true` |

**错误响应示例：只传了一部分设置，HTTP 400。**

```json
{
  "code": 400,
  "msg": "all eight site settings fields are required"
}
```

表示八个字段未全部提供，或有字段为 `null`。标题为空、数值越界等也返回 `400`，具体原因看 `msg`；并发修改冲突返回 `409`，重新查询后再保存。

无记录时新增，已有记录时更新并保留原 `created_at`。同一站点的不同后台域名共用设置。

本接口只保存配置，不统计实际登录或注册次数，也不执行登录限制、注册限制及验证码校验。Logo 地址保存不包含图片上传或旧文件删除。

---

## 六、保存本站地区

### PUT /operator/regions

**用途：保存本站“经营地区”勾选结果。请求数组代表保存后要开启的完整名单：在数组里的开启，不在数组里的已分配地区关闭。**

**请求地址与请求头：**

```http
PUT http://192.168.0.15:38001/operator/regions
Authorization: Bearer <access_token>
Host: ceshi3.back.com
Content-Type: application/json
```

| 请求头 | 必填 | 填什么 |
| --- | --- | --- |
| `Authorization` | 是 | 当前站点的 `Bearer <access_token>` |
| `Host` | 是 | 当前站点的后台域名，浏览器自动携带 |
| `Content-Type` | 是 | `application/json` |
| `X-Lang` | 否 | 请求语言，例如 `zh-CN` |

**Query：** 不传。

**Body：** 页面勾选哪些地区，就把这些地区的编码全部放进 `region_codes`。编码取自 `/operator/info` 的 `data.region_options[].region.code`，不是地区 ID、电话区号或名称。

```json
{
  "region_codes": ["JP", "VN", "PH"]
}
```

Body 参数：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `region_codes` | array<string> | 是 | 完整开启列表；编码取自查询响应 `region_options[].region.code`；`[]` 表示全部关闭 |

**这次请求保存后的结果：**

| 地区 | 请求中是否包含 | 本站保存后的状态 |
| --- | --- | --- |
| 日本 `JP` | 包含 | 开启，`status=1` |
| 越南 `VN` | 包含 | 开启，`status=1` |
| 菲律宾 `PH` | 包含 | 开启，`status=1` |
| 本站其他已分配地区 | 未包含 | 关闭，`status=2` |

如果只是想多开启一个地区，必须保留原来仍要开启的编码，再加上新编码。不要只发送新编码，否则原来开启但本次未提交的地区会关闭。传 `[]` 是主动全部关闭，省略字段或传 `null` 则是错误请求。

**成功响应：HTTP 200。**

```json
{
  "code": 0,
  "data": {
    "region_codes": ["JP", "PH", "VN"]
  },
  "msg": "ok"
}
```

`data` 字段：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `region_codes` | array<string> | 保存后的本站开启列表，规范化后按编码排序；实际生效状态需查询 `/operator/info` |

响应中的 `["JP","PH","VN"]` 表示本站已保存开启日本、菲律宾、越南。返回顺序可能与提交顺序不同，按编码判断是否选中。

**错误响应示例：region_codes 传了 null，HTTP 400。**

```json
{
  "code": 400,
  "msg": "region_codes must be an array"
}
```

提交未分配地区返回 `403`；编码无效或新开启的地区基础数据不可用返回 `400`；并发修改冲突返回 `409`。失败时不会只保存名单中的一部分。

保存规则：

- 完整替换勾选结果，不是增量添加。提交的已分配地区设为 `status=1`，未提交的已分配地区设为 `status=2`。
- 不新增或删除地区基础数据，也不改变总站给本站分配的地区范围。
- 缺少字段或传 `null` 返回 `400`；空数组有效。
- 编码去除首尾空白并转为大写，必须是两个英文字母；规范化后重复返回 `400`。
- 提交未分配给本站的地区返回 `403`，不会部分保存。
- 从关闭改为开启时，地区基础数据必须存在且开启；否则返回 `400`，依赖不可用可能返回 `503`。
- 已勾选但基础数据临时关闭的地区可以保留本地选择，此时 `effective_status=2`。
- 并发分配或勾选变动可能返回 `409`，需要重新查询并确认后保存。

---

## 七、保存本站语言

### PUT /operator/languages

**用途：保存本站“开放语言”勾选结果。数组内的已分配语言开启，未放入数组的已分配语言关闭。**

**请求地址与请求头：**

```http
PUT http://192.168.0.15:38001/operator/languages
Authorization: Bearer <access_token>
Host: ceshi3.back.com
Content-Type: application/json
```

| 请求头 | 必填 | 填什么 |
| --- | --- | --- |
| `Authorization` | 是 | 当前站点的 `Bearer <access_token>` |
| `Host` | 是 | 当前站点的后台域名，浏览器自动携带 |
| `Content-Type` | 是 | `application/json` |
| `X-Lang` | 否 | 请求语言，例如 `zh-CN` |

**Query：** 不传。

**Body：** 将页面全部勾选语言的编码放入 `language_codes`。编码取自 `/operator/info` 的 `data.languages[].lang`，不是语言 ID、名称或翻译 Key。

```json
{
  "language_codes": ["en-US", "zh-CN"]
}
```

Body 参数：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `language_codes` | array<string> | 是 | 完整开启列表；编码取自查询响应 `languages[].lang`；`[]` 表示全部关闭 |

上面的请求表示保存后开启英语和简体中文，其他已分配语言关闭。如果还要保留繁体中文，就把 `zh-HK` 一起放进数组。传 `[]` 表示全部关闭，省略字段或传 `null` 会报错。

**成功响应：HTTP 200。**

```json
{
  "code": 0,
  "data": {
    "language_codes": ["en-US", "zh-CN"]
  },
  "msg": "ok"
}
```

`data` 字段：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `language_codes` | array<string> | 保存后的本站开启列表，按编码排序；实际生效状态需查询 `/operator/info` |

**错误响应示例：language_codes 传了 null，HTTP 400。**

```json
{
  "code": 400,
  "msg": "language_codes must be an array"
}
```

提交未分配语言返回 `403`；编码无效或新开启的语言在 Core 中不可用返回 `400`；并发修改冲突返回 `409`。失败时不会只保存名单中的一部分。

保存规则：

- 完整替换勾选结果。已分配且提交的语言设为 `status=1`，未提交的已分配语言设为 `status=2`。
- 不新增、删除 Core 语言，也不改变本站的语言分配范围。
- 缺少字段或传 `null` 返回 `400`；空数组有效。
- 编码去除首尾空白，保留原编码大小写，请直接使用查询值。空编码、超过 `16` 字节或重复编码返回 `400`。
- 提交未分配给本站的语言返回 `403`，不会部分保存。
- 从关闭改为开启时，当前节点 Core 中必须存在该语言且 `disabled=0`；否则返回 `400`，依赖不可用可能返回 `503`。
- 已勾选但 Core 临时禁用的语言可以保留本地选择，此时 `effective_status=2`；Core 恢复后可以重新生效。
- 并发分配或勾选变动可能返回 `409`，需要重新查询并确认后保存。

例如，原来开启 `en-US`、`zh-CN`、`zh-HK`，提交示例中的两个编码后，`zh-HK` 本站状态变为 `2`，仍会出现在查询的语言选项中。

---

## 八、上传本站 Logo

### POST /operator/logo

**用途：上传本站的一张 Logo 图片，取得公开图片地址。成功后还要把返回的 url 填进站点设置保存，才会成为本站使用的 Logo。**

**请求地址与请求头：**

```http
POST http://192.168.0.15:38001/operator/logo
Authorization: Bearer <access_token>
Host: ceshi3.back.com
```

| 请求头 | 必填 | 填什么 |
| --- | --- | --- |
| `Authorization` | 是 | 当前站点的 `Bearer <access_token>` |
| `Host` | 是 | 当前站点的后台域名，浏览器自动携带 |
| `Content-Type` | 自动生成 | 选择 form-data 后由客户端生成 `multipart/form-data; boundary=...`，不要手写 |
| `X-Lang` | 否 | 请求语言，例如 `zh-CN` |

**Query：** 不传。

**Body：** 选择 form-data，只添加下面一个文件字段。

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `file` | file | 是 | 选择本地图片，例如 `logo.png`；必须是文件类型，不能填文件路径字符串或图片 URL |

在 Apipost 中选择 `Body → form-data`，字段名填 `file`，类型选文件，再选择图片。不要发送 JSON，也不要手动固定 Content-Type boundary。

上传限制：

- 只允许一个 `file` 文件字段，不接受额外文本字段或多个文件；无需提交 `scene`、站点编码或站点 ID。
- 支持 `.jpg`、`.png`，后缀比较不区分大小写；当前不接受 `.jpeg`。
- 单文件最大 `32 MiB`，宽、高分别不超过 `4096` 像素；后端会校验格式、后缀和完整图片内容。
- 上传成功只返回文件地址，尚未保存为站点 Logo；需继续调用 `PUT /operator/settings` 提交完整八项设置。

**成功响应：HTTP 200。**

```json
{
  "code": 0,
  "data": {
    "url": "https://cdn.example.com/logo/OP_57BE18/20261027/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.png",
    "key": "logo/OP_57BE18/20261027/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.png",
    "scene": "logo",
    "visibility": "public"
  },
  "msg": "ok"
}
```

`data` 字段：

| 字段 | 类型 | 示例 | 含义 |
| --- | --- | --- | --- |
| `url` | string | `"https://cdn.example.com/logo/OP_57BE18/20261027/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.png"` | 完整图片地址，可预览；把此值作为设置接口的 `site_logo_url` |
| `key` | string | `"logo/OP_57BE18/20261027/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.png"` | 存储路径，用于定位文件；不要把它当作 `site_logo_url` |
| `scene` | string | `"logo"` | 上传场景，由后端填写，表示本站 Logo |
| `visibility` | string | `"public"` | 公开访问图片，不需要 Token 才能展示 |
| `url_expires_in` | int | `0` 或省略 | 地址有效期，单位秒；本接口只接受永久公开地址，该字段为 `0` 时省略 |

**接下来怎么保存：** 把返回的 `data.url` 放进 `PUT /operator/settings` 的 `site_logo_url`，同时带上其余七项当前值。上传接口不会自动修改设置，也不会自动删除旧 Logo。

**错误响应示例：文件类型或图片内容不合法，HTTP 400。**

```json
{
  "code": 400,
  "msg": "logo must be a valid JPG or PNG image"
}
```

检查文件是否为有效 `.jpg` 或 `.png` 图片。大小、尺寸、字段数量不符合要求也会返回 `400`；请求体过大可能返回 `413`。OSS 地址未配置返回 `503`，上传链路或 OSS 响应异常可能返回 `502`；权限问题可能返回 `401` 或 `403`。

当前 `OP_` 站点的新上传路径为：

```text
logo/<operator_code>/<yyyyMMdd>/<32位随机十六进制文件名>.<jpg或png>
```

不同站点的图片放在同一个 `logo` 前缀下，再按站点编码和日期区分。旧路径 `<operator_code>/logo/...` 和已有默认图不会自动迁移或删除；接口兼容旧目录返回值。

后端需配置 `OssApiBaseURL`，OSS 需配置公开的 `logo` 场景，并能够使用对应站点的 Core 验证 Token 和接口权限。客户端不需要直接调用 OSS 或自行填写转发 IP 头。

---

## 九、常见错误与前端对接顺序

### 常见错误

| HTTP 状态 / code | 常见原因 | 处理方式 |
| --- | --- | --- |
| `400` | 缺少必填字段、标题为空、阈值越界、编码重复、图片不符合要求、新开启资源不可用 | 按字段规则检查请求 |
| `401` | 未登录、Token 失效或认证绑定的客户端 IP 不匹配 | 检查登录状态和访问链路 |
| `403` | 无接口权限、域名与登录站点不一致、提交未分配的地区或语言 | 核对权限、后台域名和可选范围 |
| `404` | 后台域名未登记 | 检查域名及其类型是否为后台域名 `1` |
| `408` / `504` | 请求取消或依赖超时等 | 检查请求状态，重新查询确认结果 |
| `409` | 并发修改冲突 | 重新查询最新配置，确认后提交 |
| `413` | 请求体超过服务允许大小 | 缩小请求或图片；文件校验阶段也可能返回 `400` |
| `500` | 服务端内部错误 | 联系后端查看日志 |
| `502` | OSS 请求失败、返回格式异常或未返回符合要求的公开本站 Logo | 检查 OSS 日志及响应 |
| `503` | RPC、OSS 地址等未配置，或依赖服务不可用 | 检查部署配置及依赖状态 |

OSS 返回有效 JSON 的 `4xx` 错误时，上传接口会保留相应 HTTP 状态码，并返回通用错误描述；详细原因查看服务端日志。

### 建议对接顺序

1. 登录页通过当前后台域名调用 `GET /operator/code`，取得 `operator_code` 后按原流程登录。
2. 进入站点配置页，携带 Token 调用 `GET /operator/info`，回显站点设置和地区、语言选项。
3. 地区或语言变更分别调用对应 PUT 接口，提交完整选择列表，成功后重新查询确认。
4. 更换 Logo 时先调用 `POST /operator/logo`，使用返回的 `url` 预览；保存时将其放入完整八项设置提交。
5. 调用 `PUT /operator/settings` 后使用响应更新页面；历史站点首次保存也应携带查询返回的默认标题和其他设置。

不要将整个 `/operator/info` 响应或带 `operator_code`、时间、`configured` 等只读字段的 `site_settings` 对象原样提交给设置保存接口，应只取允许的八个字段。
