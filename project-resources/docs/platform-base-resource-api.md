# 总站头像、头像框和 VIP 默认图标接口文档

用于前端对接总站后台的默认头像、默认头像框和 VIP 默认图标管理页面。

这三类素材都按“查询当前配置 → 上传新图片 → 预览 → 修改或保存配置 → 重新查询”使用。**上传成功只代表图片已上传，不代表默认图片配置已经修改。**

---

## 一、服务地址与测试准备

| 项目 | 当前开发环境示例 | 怎么用 |
| --- | --- | --- |
| 总站 HTTP API | `http://192.168.10.220:18001` | 下文九个 HTTP 接口的公共地址 |
| 查询头像的完整地址 | `http://192.168.10.220:18001/admin/default-avatar/list-all` | 测试工具选择 GET，填写此地址 |
| 图片公开前缀 | `http://192.168.10.220:9200/oss-bucket` | 浏览器访问图片使用，不是接口地址 |
| 登录 Token | `<access_token>` | 使用总站后台登录得到的 Token |

换环境时替换服务地址和 Token。本机运行服务时，HTTP API 可使用 `http://127.0.0.1:18001`；图片直接使用接口返回的地址，无需自行更换域名。使用网关时以提供的访问地址为准，直连本服务不要额外加 `/p9` 或 `/base`。

文中的图片文件名、ID 和响应内容仅作格式示例；实际测试请使用查询返回的 ID 及上传返回的图片地址或 Key，不要直接访问示例中的 `aaaa...png`。

### 1.1 登录与请求头

下文九个 HTTP 接口都需要登录及对应接口权限；上传和修改、保存只允许有效总站后台身份，拒绝分站用户和预览令牌。登录及 Token 获取见 [Core API 文档](./core-api.md)。

```http
Authorization: Bearer <access_token>
```

| 请求头 | 必填 | 填什么 |
| --- | --- | --- |
| `Authorization` | 是 | `Bearer ` 后面接总站登录 Token，中间保留一个空格 |
| `Content-Type` | 按请求类型 | JSON 修改、保存请求使用 `application/json`；上传选择 form-data，由工具自动生成带 boundary 的 Content-Type |
| `X-Lang` | 否 | 例如 `zh-CN`；用于提示信息，不会改变素材 ID、等级或图片 Key |
| `Host` | 无需手动设置 | 使用客户端默认值，无需填写分站后台域名 |

GET 查询不传 Body。上传不能发送 JSON，也不要手写 `multipart/form-data` 后遗漏 boundary。无需手动传站点编码、修改人、`scene`、桶名或 OSS 服务地址。

上传时总站 API 会转发到 OSS，使用同一个登录令牌继续校验。账号还需具有 OSS 的 `POST /p9/oss/file/upload` 权限；前端调用下文本站上传接口即可。

### 1.2 统一响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `code` | int | 成功为 `0`；失败按 HTTP 错误状态返回 |
| `data` | object | 本次接口的业务内容；错误响应可能没有此字段 |
| `msg` | string | 结果提示，示例使用 `ok`；不要依赖固定文案判断成功或失败 |

修改、保存成功返回空业务对象：

```json
{
  "code": 0,
  "data": {},
  "msg": "ok"
}
```

**修改、保存响应不会再次返回图片域名或列表。** 需要立即预览时使用上传返回的 `data.url`；需要确认最终配置时重新调用对应查询接口。

### 1.3 素材数量与图片规格

| 素材 | 正常配置数量 | 图片格式 | 宽 × 高 | 单张文件大小 | 保存时能否清空 |
| --- | --- | --- | --- | --- | --- |
| 默认头像 | 8 张 | PNG，文件后缀 `.png` | 200 × 200 像素 | 不超过 204800 字节，即 200 KiB | 不能 |
| 默认头像框 | 6 张 | 透明 PNG，文件后缀 `.png` | 145 × 145 像素 | 不超过 204800 字节 | 不能 |
| VIP 默认图标 | 10 组，覆盖 0–99 级 | PNG，文件后缀 `.png` | 115 × 165 像素 | 不超过 204800 字节 | 可以，`icon_url` 传 `""` |

图片必须能完整解码，不能只把 JPG 文件改名成 PNG。上传不自动裁剪、缩放或把头像与头像框合成一张图。头像框还要求**中心像素完全透明，且整张图片至少存在一个非全透明像素**；这不是对整个中央区域透明度的检查。

头像与头像框只修改已有记录，没有新增、删除和排序接口。`id` 必须使用查询返回的真实主键，不能假定头像 ID 永远是 1–8、头像框 ID 永远是 1–6。

---

## 二、接口索引

| 素材 | 方法 | 路径 | 说明 | 请求体 |
| --- | --- | --- | --- | --- |
| 头像 | GET | `/admin/default-avatar/list-all` | 查询全部默认头像 | 不传 |
| 头像 | POST | `/admin/default-avatar/upload` | 上传一张新头像，尚未修改配置 | form-data |
| 头像 | POST | `/admin/default-avatar/update` | 按已有 ID 修改头像图片 | JSON |
| 头像框 | GET | `/admin/default-avatar-frame/list-all` | 查询全部默认头像框 | 不传 |
| 头像框 | POST | `/admin/default-avatar-frame/upload` | 上传一个新头像框，尚未修改配置 | form-data |
| 头像框 | POST | `/admin/default-avatar-frame/update` | 按已有 ID 修改头像框图片 | JSON |
| VIP | GET | `/admin/vip-default-icon/list-all` | 查询固定十组默认图标 | 不传 |
| VIP | POST | `/admin/vip-default-icon/upload` | 上传某一等级组的图标，尚未保存配置 | form-data |
| VIP | POST | `/admin/vip-default-icon/save` | 保存本次提交的等级组 | JSON |

九个 HTTP 接口均不分页，无需 `page`、`page_size`。

---

## 三、默认头像

### 3.1 查询全部头像

**用途：页面回显，取得每条头像的真实 ID 和可展示的图片地址。**

```http
GET http://192.168.10.220:18001/admin/default-avatar/list-all
Authorization: Bearer <access_token>
```

**Query / Body：** 都不传。

**成功响应：HTTP 200。** 以下示例第一张已配置，其余槽位尚未配置：

```json
{
  "code": 0,
  "data": {
    "list": [
      {
        "id": 1,
        "image_url": "http://192.168.10.220:9200/oss-bucket/avatar-resource/avatars/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.png"
      },
      { "id": 2, "image_url": "" },
      { "id": 3, "image_url": "" },
      { "id": 4, "image_url": "" },
      { "id": 5, "image_url": "" },
      { "id": 6, "image_url": "" },
      { "id": 7, "image_url": "" },
      { "id": 8, "image_url": "" }
    ]
  },
  "msg": "ok"
}
```

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `data.list` | array | 按主键升序返回全部已有头像，正常为 8 条；没有记录时为 `[]` |
| `data.list[].id` | int64 | 头像素材记录主键，修改时提交这个值，不是会员 ID |
| `data.list[].image_url` | string | 完整图片 URL，可直接展示；`""` 表示未配置图片 |

查询只用于页面回显，不改变已有配置或会员头像。

### 3.2 上传一张头像

**用途：上传新头像，取得预览地址和保存用的 Key。这个接口不会修改任何头像槽位。**

```http
POST http://192.168.10.220:18001/admin/default-avatar/upload
Authorization: Bearer <access_token>
```

**Query：** 不传。**Body：** 选择 `form-data`，只传下面一个文件字段。

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `file` | File | 是 | 一张有效 PNG，200 × 200 像素，不超过 204800 字节 |

在 Apifox / Apipost 中，字段名填 `file`，类型选择“文件”，再选择本地 PNG。不能用 Text 填文件路径，也不能填图片 URL。无需传 `id`、`scene` 或 `min_level`，额外表单字段会被拒绝。

**成功响应：HTTP 200。**

```json
{
  "code": 0,
  "data": {
    "url": "http://192.168.10.220:9200/oss-bucket/avatar-resource/avatars/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.png",
    "key": "avatar-resource/avatars/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.png",
    "scene": "default_avatar",
    "visibility": "public"
  },
  "msg": "ok"
}
```

三类上传响应的字段说明见第五节。**接下来把返回的 `data.key` 放进修改请求的 `image_url`，再提交正在编辑的头像 ID。** 上传返回的完整 `data.url` 也可提交。

### 3.3 修改已有头像

**用途：将已有头像记录替换为刚上传的图片；不新增记录，也不修改分站会员资料。**

```http
POST http://192.168.10.220:18001/admin/default-avatar/update
Authorization: Bearer <access_token>
Content-Type: application/json
```

**Query：** 不传。**Body：** 选择 JSON：

```json
{
  "id": 1,
  "image_url": "avatar-resource/avatars/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.png"
}
```

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | int64 | 是 | 查询返回的已有头像主键，必须大于 0；不存在返回 404 |
| `image_url` | string | 是 | 本类上传返回的 `key` 或完整 `url`，最长 1024 字节；不允许空字符串 |

不能提交头像框目录、临时签名地址或任意本地文件路径。地址不允许携带查询参数、片段、账号密码、转义路径或路径跳转。新上传的文件名由服务端生成，直接使用返回值，不自行拼造文件名。

**成功响应：** 见第一节的 `data: {}`，不返回图片地址。成功后调用 3.1 重新查询；ID 保持不变，`image_url` 返回完整公开地址。修改不会删除旧 OSS 图片，也不会联网检查对象是否存在。

---

## 四、默认头像框

### 4.1 查询全部头像框

**用途：页面回显，取得每个头像框的真实 ID 和图片地址。**

```http
GET http://192.168.10.220:18001/admin/default-avatar-frame/list-all
Authorization: Bearer <access_token>
```

**Query / Body：** 都不传。

**成功响应：HTTP 200。**

```json
{
  "code": 0,
  "data": {
    "list": [
      {
        "id": 1,
        "image_url": "http://192.168.10.220:9200/oss-bucket/avatar-resource/avatar-frames/bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb.png"
      },
      { "id": 2, "image_url": "" },
      { "id": 3, "image_url": "" },
      { "id": 4, "image_url": "" },
      { "id": 5, "image_url": "" },
      { "id": 6, "image_url": "" }
    ]
  },
  "msg": "ok"
}
```

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `data.list` | array | 按主键升序返回全部已有头像框，正常为 6 条；没有记录时为 `[]` |
| `data.list[].id` | int64 | 头像框记录主键，修改时使用，不是会员 ID 或头像 ID |
| `data.list[].image_url` | string | 完整公开图片 URL，`""` 表示尚未配置 |

“会员不使用头像框”由分站会员业务处理，不是从此列表删除一条记录，也不是把默认头像框图片清空。

### 4.2 上传一个头像框

**用途：上传一个新头像框供预览，随后还需修改已有记录。**

```http
POST http://192.168.10.220:18001/admin/default-avatar-frame/upload
Authorization: Bearer <access_token>
```

**Query：** 不传。**Body：** 选择 `form-data`，只传一个文件：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `file` | File | 是 | 透明 PNG，145 × 145 像素，不超过 204800 字节；透明要求见第一节 |

不传 `id`、`scene` 或其他表单字段。Content-Type 由工具自动生成。

**成功响应：HTTP 200。**

```json
{
  "code": 0,
  "data": {
    "url": "http://192.168.10.220:9200/oss-bucket/avatar-resource/avatar-frames/bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb.png",
    "key": "avatar-resource/avatar-frames/bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb.png",
    "scene": "default_avatar_frame",
    "visibility": "public"
  },
  "msg": "ok"
}
```

使用 `data.url` 预览，把 `data.key` 或 `data.url` 交给下面的修改接口。只上传不修改，列表中的原头像框保持不变。

### 4.3 修改已有头像框

**用途：按已有 ID 替换头像框图片。**

```http
POST http://192.168.10.220:18001/admin/default-avatar-frame/update
Authorization: Bearer <access_token>
Content-Type: application/json
```

**Query：** 不传。**Body：** 选择 JSON：

```json
{
  "id": 1,
  "image_url": "avatar-resource/avatar-frames/bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb.png"
}
```

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | int64 | 是 | 查询返回的已有头像框主键，大于 0；不存在返回 404 |
| `image_url` | string | 是 | 头像框上传返回的 `key` 或完整 `url`，最长 1024 字节；不能清空，不能使用头像目录 |

URL 格式限制与 3.3 相同。修改人和更新时间由后端维护，不需要提交。

**成功响应：** 第一节的 `data: {}`。随后调用 4.1 刷新列表。修改不删除旧文件，也不替换会员已保存的头像框选择。

---

## 五、VIP 默认图标

### 5.1 查询十组默认图标

**用途：页面回显；按等级组获取总站统一配置的默认 VIP 图片。**

```http
GET http://192.168.10.220:18001/admin/vip-default-icon/list-all
Authorization: Bearer <access_token>
```

**Query / Body：** 都不传。

**成功响应：HTTP 200。** 以下示例仅 0–9 组已配置：

```json
{
  "code": 0,
  "data": {
    "list": [
      {
        "min_level": 0,
        "max_level": 9,
        "icon_url": "http://192.168.10.220:9200/oss-bucket/vipicon/0-9/cccccccccccccccccccccccccccccccc.png"
      },
      { "min_level": 10, "max_level": 19, "icon_url": "" },
      { "min_level": 20, "max_level": 29, "icon_url": "" },
      { "min_level": 30, "max_level": 39, "icon_url": "" },
      { "min_level": 40, "max_level": 49, "icon_url": "" },
      { "min_level": 50, "max_level": 59, "icon_url": "" },
      { "min_level": 60, "max_level": 69, "icon_url": "" },
      { "min_level": 70, "max_level": 79, "icon_url": "" },
      { "min_level": 80, "max_level": 89, "icon_url": "" },
      { "min_level": 90, "max_level": 99, "icon_url": "" }
    ]
  },
  "msg": "ok"
}
```

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `data.list` | array | 固定十组，按 `min_level` 升序排列；缺少配置的组仅在响应中补齐，不写库 |
| `data.list[].min_level` | int64 | 等级下限，含本级，只能为 0、10、20、30、40、50、60、70、80、90 |
| `data.list[].max_level` | int64 | 等级上限，含本级，等于 `min_level + 9` |
| `data.list[].icon_url` | string | 完整公开图片地址；`""` 表示该组未配置或已清空 |

例如 VIP 23 使用 20–29 组。本接口只提供总站默认素材，不读取或修改分站自定义 VIP 图标。

### 5.2 上传某一组 VIP 图标

**用途：为指定等级组上传图片，尚未保存为该组默认图标。**

```http
POST http://192.168.10.220:18001/admin/vip-default-icon/upload
Authorization: Bearer <access_token>
```

**Query：** 不传。**Body：** 选择 `form-data`，恰好传两个字段，各一次：

| 字段 | 工具中的类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `file` | File / 文件 | 是 | PNG，115 × 165 像素，不超过 204800 字节 |
| `min_level` | Text / 文本 | 是 | 本组等级下限，只能为 0、10、20、30、40、50、60、70、80、90 |

测试 0–9 组时，明确填写文本 `min_level=0`，0 是合法值，不能省略。上传时不传 `max_level`、`scene` 或其他字段，Content-Type 由客户端生成。

**成功响应：HTTP 200。**

```json
{
  "code": 0,
  "data": {
    "url": "http://192.168.10.220:9200/oss-bucket/vipicon/0-9/cccccccccccccccccccccccccccccccc.png",
    "key": "vipicon/0-9/cccccccccccccccccccccccccccccccc.png",
    "scene": "vipicon",
    "visibility": "public"
  },
  "msg": "ok"
}
```

**三类上传接口共用的返回字段说明：**

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `data.url` | string | 完整公开 URL，用于当前页面预览，也可提交给修改、保存接口 |
| `data.key` | string | 桶内对象 Key，不含协议、域名或 `oss-bucket/`；可作为 `image_url` 或 `icon_url` 提交 |
| `data.scene` | string | 头像为 `default_avatar`，头像框为 `default_avatar_frame`，VIP 为 `vipicon` |
| `data.visibility` | string | 固定 `public`，图片公开访问 |
| `data.url_expires_in` | int64，可省略 | URL 有效期，单位秒；此处只接收永久公开地址，为 `0` 或省略 |

VIP 图片按 `vipicon/{min_level}-{max_level}/{服务端生成的文件名}.png` 存放，例如 `vipicon/0-9/...png`。前端不生成文件名、不指定桶名，也不要把 0–9 组的图片用于 10–19 组。

### 5.3 保存 VIP 图标配置

**用途：保存本次提交的一至十个等级组；未提交的整个等级组保持原值。**

```http
POST http://192.168.10.220:18001/admin/vip-default-icon/save
Authorization: Bearer <access_token>
Content-Type: application/json
```

**Query：** 不传。**Body：** 选择 JSON。以下仅修改 0–9 组，其余九组不变：

```json
{
  "list": [
    {
      "min_level": 0,
      "max_level": 9,
      "icon_url": "vipicon/0-9/cccccccccccccccccccccccccccccccc.png"
    }
  ]
}
```

| 字段 | 类型 | 提交要求 | 说明 |
| --- | --- | --- | --- |
| `list` | array | 必填 | 1–10 组，不允许空数组，同一等级组不能重复 |
| `list[].min_level` | int64 | 明确提交 | 固定十组之一，0 有效；不通过 `id` 定位 |
| `list[].max_level` | int64 | 明确提交 | 必须等于 `min_level + 9` |
| `list[].icon_url` | string | 明确提交 | 对应等级组上传返回的 `key` 或 `url`，最长 1024 字节；`""` 表示清空 |

请提交组内的全部三个字段。**保持一组不变应不提交整个组，不能只漏传该组的 `icon_url`；当前字符串字段漏传会按空值处理，有清空风险。**

**等级范围只能从下面十组中选择，不能自行拆分、合并或更改上下限：**

| 可保存的等级组 | `min_level` | `max_level` |
| --- | --- | --- |
| 0–9 级 | 0 | 9 |
| 10–19 级 | 10 | 19 |
| 20–29 级 | 20 | 29 |
| 30–39 级 | 30 | 39 |
| 40–49 级 | 40 | 49 |
| 50–59 级 | 50 | 59 |
| 60–69 级 | 60 | 69 |
| 70–79 级 | 70 | 79 |
| 80–89 级 | 80 | 89 |
| 90–99 级 | 90 | 99 |

例如保存 0–9 组时必须传 `min_level: 0`、`max_level: 9`；保存 10–19 组时必须传 `min_level: 10`、`max_level: 19`。`1–10`、`0–19`、`10–29` 等自定义范围均不允许，返回参数错误。等级上下限都是数字，不传 `"0-9"` 这样的区间字符串。

**同时保存 0–9、10–19 两组的完整请求示例：** 每组使用各自上传返回的图片 Key，另外八组保持不变。

```json
{
  "list": [
    {
      "min_level": 0,
      "max_level": 9,
      "icon_url": "vipicon/0-9/cccccccccccccccccccccccccccccccc.png"
    },
    {
      "min_level": 10,
      "max_level": 19,
      "icon_url": "vipicon/10-19/dddddddddddddddddddddddddddddddd.png"
    }
  ]
}
```

允许同时修改多组。一次请求的所有组一起提交，任一组校验或写入失败，整次保存失败，不会只保存其中几组。修改人由后端取得，不传 `updated_by`。

清空 0–9 组的示例：

```json
{
  "list": [
    {
      "min_level": 0,
      "max_level": 9,
      "icon_url": ""
    }
  ]
}
```

清空只修改默认配置，不删除 OSS 文件。图片地址须为本组 Key 或无签名的 HTTP(S) 永久地址；`vipicon` 目录中的图片必须与等级组匹配。已有其他目录的旧公开 URL 继续兼容。保存校验地址及等级，不会联网下载或重新校验图片尺寸。

**成功响应：** 第一节的 `data: {}`，不会带回完整图片地址。保存后调用 5.1 刷新；空图片用页面空状态展示，不要把请求失败当作配置已清空。

---

## 六、图片地址怎么用

- 上传返回的 `data.url` 是完整图片地址，用于预览；不要自行拼域名或修改地址。
- 上传返回的 `data.key` 可以提交给修改、保存接口：头像和头像框填进 `image_url`，VIP 填进 `icon_url`；字段名不要改成 `key`。
- 也可以提交上传返回的完整 `data.url`，无需前端裁剪路径。
- 修改、保存成功只返回 `data: {}`；重新查询时，图片字段会返回完整地址，可以直接展示。
- 查询返回的旧地址若无法打开，应反馈后端核查，不要在前端硬编码替换域名。

---

## 七、常见错误与前端对接顺序

### 7.1 常见错误

| 状态 | 常见原因 | 检查方式 |
| --- | --- | --- |
| HTTP 400 | JSON、表单字段、图片格式/尺寸/大小、等级范围或 URL 不合法 | 按对应接口的字段表检查，尤其注意 File 类型、PNG 尺寸和 `min_level=0` |
| HTTP 401 / 403 | 未登录、令牌无效、分站身份、预览令牌或权限不足 | 使用本站总站 Token；上传同时核对 OSS 上传权限 |
| HTTP 404 | 修改的头像或头像框 ID 不存在 | 先查询列表，使用返回的真实 ID |
| HTTP 502 | 图片上传服务调用失败或返回异常 | 保留当前配置，提示上传失败；持续失败时反馈后端 |
| HTTP 503 | 图片服务配置未就绪 | 提示暂时不可用，反馈后端核查 |
| HTTP 500 | 服务内部异常 | 提示操作失败，保留待编辑内容，不提示保存成功 |

上传时，OSS 返回的 4xx 会继续作为失败状态返回；OSS 5xx 和业务失败等通常转换为 502。网关也可能因自己的请求大小限制返回 413，以实际响应为准。

错误响应格式示例，文案仅作说明：

```json
{
  "code": 400,
  "msg": "参数不符合要求"
}
```

### 7.2 对接顺序

1. 使用总站账号登录，取得 Token，并确认账号拥有相关接口权限。
2. 进入页面时调用对应 `list-all`，保留真实 ID 或 VIP 等级范围用于修改。
3. 用户选择图片后按规格检查，通过对应 `upload` 上传；头像框需保留透明效果。
4. 用上传返回的 `data.url` 更新预览，将 `data.key` 暂存在页面待保存数据里。
5. 头像、头像框点击保存时，提交 `id + image_url`；VIP 提交本次修改组的 `list`。
6. 保存成功后重新查询，使用查询返回的完整图片地址展示；失败时保留待编辑内容，不提示成功。

只上传不保存不会更改默认配置；替换或清空配置不会删除旧 OSS 图片。选择会员头像、是否使用头像框、分站 VIP 展示以及会员数据保存属于分站业务，不由这些总站素材接口自动完成。
