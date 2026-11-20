# 总站、分站 OSS 图片上传前端接口文档

用于前端对接总站和分站后台的统一图片上传接口，包含服务地址、支持的场景、表单字段、图片规格、返回值及上传后的保存方式。

**上传成功只代表 OSS 图片已创建，不会自动保存为 VIP 图标、默认头像、头像框或网站 Logo。前端用返回的 URL 预览，用户确认后再调用对应保存接口。上传和保存都不会自动删除旧图片。**

本文统一说明上传协议；分站 VIP 查询、保存和恢复见 [分站 VIP 图标接口](./operator-vip-icon-api.md)，总站素材查询和修改见 [总站素材接口](./platform-base-resource-api.md)，网站设置保存见 [网站管理接口](./operator-base-api.md)。

---

## 一、服务地址与请求准备

| 项目 | 当前开发环境示例 | 怎么用 |
| --- | --- | --- |
| 总站上传完整地址 | `http://192.168.10.220:18001/resource/upload` | 使用总站后台登录 Token |
| 分站 node01 上传完整地址 | `http://192.168.10.220:38010/resource/upload` | 使用目标分站后台登录 Token |
| 图片公开前缀示例 | `http://192.168.10.220:9200/oss-bucket` | 浏览器访问图片使用，不是上传接口地址 |
| 登录 Token | `<access_token>` | 对应系统正常登录得到的访问令牌，不使用刷新令牌 |

两个接口均为 **POST**，路径同为 `/resource/upload`，但服务地址和登录身份不同。换环境时替换 API 地址和 Token，图片直接使用返回的 URL。直连上述服务时，不额外加 `/admin`、`/api` 或 `/p9`；经网关访问时，以实际转发地址为准。

本文的站点编码 `OP_example`、文件名、日期和记录 ID 仅作格式示例，实际调用使用上传或查询接口返回的数据。

### 1.1 登录与请求头

```http
Authorization: Bearer <access_token>
```

| 请求头 | 必填 | 填什么 |
| --- | --- | --- |
| `Authorization` | 是 | `Bearer ` 后接当前系统后台 Token，中间保留一个空格 |
| `Content-Type` | 自动生成 | Body 选择 form-data，由客户端生成 `multipart/form-data; boundary=...`，不要手工固定 boundary |
| `X-Lang` | 否 | 例如 `zh-CN`，用于提示信息 |
| `Host` | 无需手动设置 | 使用客户端默认值，不用填写分站域名 |

总站入口只允许有效总站后台身份；分站入口只允许本站运营后台身份，不接受总站、代理或预览身份。分站归属取自认证后的 Token，前端不指定站点编码、域名或修改人。登录流程见 [Core API 文档](./core-api.md)。

账号需有当前 API 的 **`POST /resource/upload`** 权限，以及目标 OSS 的 **`POST /p9/oss/file/upload`** 权限。新上传入口需单独授权，原上传路径的授权不会自动转换；接口目录登记成功不代表已给角色授权。

### 1.2 总站、分站支持范围

| `scene` | 用途 | 总站入口 | 分站入口 | 是否提交 `min_level` |
| --- | --- | --- | --- | --- |
| `vipicon` | VIP 图标；总站为默认图，分站为本站自定义图 | 支持 | 支持 | 必须提交 |
| `default_avatar` | 总站默认头像 | 支持 | 不支持 | 禁止提交 |
| `default_avatar_frame` | 总站默认头像框 | 支持 | 不支持 | 禁止提交 |
| `logo` | 分站网站 Logo | 不支持 | 支持 | 禁止提交 |

`scene` 必填，没有默认值，按表中小写名称原样提交。`public`、`banner` 或其他字符串目前不在这两个 API 的允许范围内；即使 OSS 配置了同名场景，也不代表后台统一入口已经支持。

---

## 二、公共表单字段与图片规格

**Query：** 不传。**Body：** 在 Apifox / Apipost 中选择 `form-data`。

| 字段 | 工具中的类型 | 必填 | 示例与说明 |
| --- | --- | --- | --- |
| `file` | File / 文件 | 是 | 只选择一张本地图片；不能用 Text 填文件路径或图片 URL |
| `scene` | Text / 文本 | 是 | 选择当前系统支持的场景，例如 `vipicon` |
| `min_level` | Text / 文本 | 仅 VIP 必填 | 只能为 `0`、`10`、`20`…`90`；其他场景禁止提交此字段，即使空值也不允许 |

VIP 上传恰好提交 **`file`、`scene`、`min_level` 三个字段**；头像、头像框、Logo 恰好提交 **`file`、`scene` 两个字段**。每个字段只提交一次，重复字段、额外字段或多个文件均会被拒绝。

不传 `group_no`、`id`、`operator_code`、`max_level`、`key`、`visibility`、桶名、目录或 OSS 地址；这些不是统一上传的请求字段。目录及对象文件名由服务端生成。

| 场景 | 格式及后缀 | 宽 × 高 | 单文件大小 | 其他要求 |
| --- | --- | --- | --- | --- |
| 总站、分站 `vipicon` | 有效 PNG，`.png` | 115 × 165 像素 | ≤ 204800 字节（200 KiB） | 明确指定等级下限 |
| 总站 `default_avatar` | 有效 PNG，`.png` | 200 × 200 像素 | ≤ 204800 字节（200 KiB） | 无需指定头像记录 ID |
| 总站 `default_avatar_frame` | 有效 PNG，`.png` | 145 × 145 像素 | ≤ 204800 字节（200 KiB） | 中心像素完全透明，且至少有一个非全透明像素；不等于校验整个中心区域透明 |
| 分站 `logo` | 有效 JPEG（`.jpg`）或 PNG（`.png`） | 宽、高分别为 1～4096 像素 | ≤ 33554432 字节（32 MiB） | 当前不接受 `.jpeg` 后缀 |

后缀比较不区分大小写，但仍会校验真实格式及完整解码结果，不能只把 JPG 改名为 PNG。接口不自动裁剪、缩放或转换格式；空文件、损坏图片、尺寸或大小不符均会失败。请求总大小还受服务和网关限制。

---

## 三、总站上传

**用途：上传总站默认 VIP 图标、默认头像或默认头像框，取得预览 URL 和保存用的 Key。**

```http
POST http://192.168.10.220:18001/resource/upload
Authorization: Bearer <总站后台access_token>
```

以下三种场景共用上面的地址，每次只选一种。

### 3.1 默认 VIP 图标

上传 60～69 级默认图时，form-data 填写：

| 字段 | 类型 | 填写内容 |
| --- | --- | --- |
| `file` | File | 选择本地 115 × 165 的 PNG |
| `scene` | Text | `vipicon` |
| `min_level` | Text | `60` |

0～9 级明确传 `min_level=0`，不能省略；等级上限由后端按下限加 9 确定，不传 `max_level`。

成功后的 Key 格式为 `vipicon/60-69/{32位随机文件名}.png`，不包含分站编码。保存时提交同一组的 `min_level=60`、`max_level=69` 和上传返回值。

### 3.2 默认头像

| 字段 | 类型 | 填写内容 |
| --- | --- | --- |
| `file` | File | 选择本地 200 × 200 的 PNG |
| `scene` | Text | `default_avatar` |

不传 `min_level` 或头像 `id`。成功后的 Key 格式为 `avatar-resource/avatars/{32位随机文件名}.png`；上传不会自动选中或增加头像记录，修改已有头像时再提交对应记录的 `id`。

### 3.3 默认头像框

| 字段 | 类型 | 填写内容 |
| --- | --- | --- |
| `file` | File | 选择本地 145 × 145、符合透明要求的 PNG |
| `scene` | Text | `default_avatar_frame` |

不传 `min_level` 或头像框 `id`。成功后的 Key 格式为 `avatar-resource/avatar-frames/{32位随机文件名}.png`；修改已有头像框时再提交对应记录的 `id`。

---

## 四、分站上传

**用途：上传登录分站的自定义 VIP 图标或网站 Logo，不修改其他站点或总站配置。**

```http
POST http://192.168.10.220:38010/resource/upload
Authorization: Bearer <目标分站后台access_token>
```

### 4.1 本站 VIP 图标

上传 60～69 级图标时，form-data 填写：

| 字段 | 类型 | 填写内容 |
| --- | --- | --- |
| `file` | File | 选择本地 115 × 165 的 PNG |
| `scene` | Text | `vipicon` |
| `min_level` | Text | `60` |

本站 Key 格式为 `vipicon/{登录站点编码}/60-69/{32位随机文件名}.png`。站点编码来自 Token，前端不传。

**上传使用等级下限 `min_level`；分站保存、恢复使用组编号 `group_no`，两者不要混用。**

| 等级范围 | 上传 `min_level` | 分站保存、恢复 `group_no` |
| --- | --- | --- |
| 0～9 | 0 | 0 |
| 10～19 | 10 | 1 |
| 20～29 | 20 | 2 |
| 30～39 | 30 | 3 |
| 40～49 | 40 | 4 |
| 50～59 | 50 | 5 |
| 60～69 | 60 | 6 |
| 70～79 | 70 | 7 |
| 80～89 | 80 | 8 |
| 90～99 | 90 | 9 |

换算为 `min_level=group_no×10`。例如 Key 位于 `60-69/`，保存必须传 `group_no=6`；传 `group_no=5` 会因目录与等级组不符返回 400，不会保存。

### 4.2 本站 Logo

| 字段 | 类型 | 填写内容 |
| --- | --- | --- |
| `file` | File | 选择符合规格的本地 `.jpg` 或 `.png` |
| `scene` | Text | `logo` |

不传 `min_level`、站点编码或网站设置字段。后端先查询本站配置并核对登录站点，通过后才打开、解码和上传图片；站点不匹配或查询失败时不上传。

当前 `OP_` 站点的 Key 通常为 `logo/{登录站点编码}/{yyyyMMdd}/{32位随机文件名}.{后缀}`，同时兼容原站点目录规则。前端直接使用返回值，不自行拼接目录或文件名。

---

## 五、成功响应与字段说明

两种系统、全部场景使用相同响应结构，成功为 **HTTP 200，`code=0`**。下面以分站 60～69 级 VIP 上传为例：

```json
{
  "code": 0,
  "data": {
    "url": "http://192.168.10.220:9200/oss-bucket/vipicon/OP_example/60-69/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.png",
    "key": "vipicon/OP_example/60-69/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.png",
    "scene": "vipicon",
    "visibility": "public",
    "url_expires_in": 0
  },
  "msg": "ok"
}
```

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `code` | int | 成功为 `0`；失败时与 HTTP 错误状态一致 |
| `data` | object | 本次已上传素材，失败时无业务数据 |
| `data.url` | string | 完整永久公开 URL，用于预览；使用实际返回值，不自行拼接域名 |
| `data.key` | string | 桶内对象 Key，不含协议、域名或 `oss-bucket/` 桶路径；VIP、头像、头像框可直接用它保存 |
| `data.scene` | string | 本次场景，与请求 `scene` 一致 |
| `data.visibility` | string | 固定 `public`，图片公开访问；不是请求参数 |
| `data.url_expires_in` | int64 | 固定 `0`，表示永久公开地址；不是零秒后失效，也不表示对象永远不能被删除 |
| `msg` | string | 结果提示，示例使用 `ok`；实际文案随语言变化，不依赖固定文案判断成功 |

URL 为永久地址不代表自动保存业务配置。用户取消编辑或只上传不保存时，原配置保持不变，已经上传的对象保留。

---

## 六、上传后的保存方式

上传只负责返回素材；保存仍使用各业务原有接口和权限。不要把上传响应整体提交给保存接口，只取允许的字段。

| 素材 | 保存完整地址示例 | JSON 字段及规则 |
| --- | --- | --- |
| 总站默认 VIP | `POST http://192.168.10.220:18001/admin/vip-default-icon/save` | `list` 中提交 `min_level`、`max_level`、`icon_url`；图片可用对应组的 Key 或 URL |
| 总站默认头像 | `POST http://192.168.10.220:18001/admin/default-avatar/update` | 已有记录的 `id` 和 `image_url`；图片使用头像 Key 或 URL |
| 总站默认头像框 | `POST http://192.168.10.220:18001/admin/default-avatar-frame/update` | 已有记录的 `id` 和 `image_url`；图片使用头像框 Key 或 URL |
| 分站 VIP | `POST http://192.168.10.220:38010/member/vip/icon/update` | `group_no` 和 `icon_url`；本站 Key 目录必须与站点及等级组一致 |
| 分站 Logo | `PUT http://192.168.10.220:38001/operator/settings` | `site_logo_url` 使用上传返回的完整 URL，同时提交其余七项当前网站设置，不能仅提交 Logo |

保存 Body 使用 JSON，设置 `Content-Type: application/json`，Token 与当前系统一致。分站 Logo 保存入口属于 operator-base，端口与上传入口不同。

例如分站刚上传的 60～69 级图标，保存 Body 为：

```json
{
  "group_no": 6,
  "icon_url": "vipicon/OP_example/60-69/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.png"
}
```

总站保存同一等级默认图时，Body 为：

```json
{
  "list": [
    {
      "min_level": 60,
      "max_level": 69,
      "icon_url": "vipicon/60-69/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.png"
    }
  ]
}
```

分站上传的 Key 不能用于总站默认图保存，也不能用于其他分站或其他等级组。头像、头像框同样不能跨类保存。保存成功后重新调用对应查询接口，取得最终展示地址。

分站 VIP 恢复默认仍调用 `POST /member/vip/icon/reset`，只传 `group_no`，不需要重新上传。该操作清空本站目标组自定义配置，不删除 OSS 图片，详细语义见分站 VIP 图标文档。

---

## 七、常见错误与对接顺序

### 7.1 常见错误

| 状态 | 常见原因 | 前端处理 |
| --- | --- | --- |
| HTTP 400 | 缺字段、当前系统不支持的场景、重复或额外字段、VIP 等级下限不合法、图片规格不符 | 按场景核对字段和图片，不提示成功 |
| HTTP 401 / 403 / 498 | 登录过期、系统或身份不符、缺 API 或 OSS 上传权限 | 按现有登录及权限流程处理，不混用总站、分站 Token |
| HTTP 404 | Logo 上传前核对的本站记录不存在等 | 核对站点配置，不自行改为其他站点编码 |
| HTTP 413 | 请求超过网关或服务的总大小限制 | 缩小文件；文件大小校验也可能返回 400 |
| HTTP 502 | OSS 连接失败、上传超时、下游服务错误、响应格式或素材归属不符 | 提示上传失败，不继续保存；不能仅凭图片可能存在判断成功 |
| HTTP 503 | 上传地址未配置、Logo 所需本站 RPC 不可用等 | 提示暂不可用，持续失败交后端排查 |
| HTTP 504 | Logo 上传前查询本站配置的 RPC 超时 | 本次不继续上传，稍后重试 |
| HTTP 500 | 内部处理异常 | 保留当前已保存配置，反馈后端 |

OSS 返回的 HTTP 4xx 会保留对应失败状态，但隐藏内部错误；其他下游错误通常转换为 502。网关、认证组件也可能返回自身错误，以实际 HTTP 状态和业务码为准。

错误响应示例，提示文字仅用于说明：

```json
{
  "code": 400,
  "msg": "图片格式、尺寸或大小不符合场景要求"
}
```

失败时无业务 `data`；调试环境可能额外返回 `debug`，前端不要依赖其内部字段。上传失败或结果不明确时不要调用保存接口，也不要覆盖当前已保存图片；重新上传可能产生新的对象。

### 7.2 对接顺序

1. 确认属于总站还是分站，选择对应上传地址和后台 Token。
2. 按业务选择支持的 `scene`，校验图片规格；只有 VIP 额外填写 `min_level`。
3. 用 form-data 发送请求，各字段只提交一次，不传 JSON 或任意目录参数。
4. 上传成功后用 `data.url` 预览，暂存 `data.key`；此时业务配置尚未改变。
5. 用户确认后调用对应保存接口；分站 VIP 核对 `group_no` 与 Key 等级目录，Logo 完整提交网站八项设置。
6. 保存成功后重新查询并更新页面；取消编辑时不保存，也不调用 OSS 删除接口。

---

## 八、旧入口切换

以下专用上传入口已删除，统一替换为当前系统的 `POST /resource/upload`：

| 原入口 | 新入口所属系统 | 需要填写的场景 |
| --- | --- | --- |
| `/admin/vip-default-icon/upload` | 总站 | `scene=vipicon`，保留 `min_level` |
| `/admin/default-avatar/upload` | 总站 | `scene=default_avatar` |
| `/admin/default-avatar-frame/upload` | 总站 | `scene=default_avatar_frame` |
| `/member/vip/icon/upload` | 分站 operator-admin | `scene=vipicon`，原 `group_no` 改传 `min_level=group_no×10` |
| operator-base `/operator/logo` | 分站 operator-admin | `scene=logo`，改用分站上传服务地址 |

查询、保存、修改和恢复接口继续使用各自原协议。新的上传入口权限需单独配置；统一上传不会自动迁移原权限或改写已有图片配置。

两个 API 使用各自的 `OssApiBaseURL` 转发上传，分站 OSS 必须认证同一分站 Core 的 Token，对应 OSS 场景须配置为公开。以上由后端部署配置，前端不传 OSS 地址或 `public/private`。

后续增加图片类型仍可使用统一入口，但必须由后端明确加入场景、图片规则和 OSS 配置后才能使用，前端不能只填一个新 `scene` 就自动开通。
