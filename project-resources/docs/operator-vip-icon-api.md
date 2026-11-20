# 分站 VIP 图标前端接口文档

用于前端对接分站后台的 VIP 图标页面，包含查询、上传、保存和恢复默认四个接口。

自定义图片按“查询当前配置 → 上传新图片 → 预览 → 保存 → 重新查询”使用。**上传成功只代表图片已上传，点击保存后才会修改当前图标。恢复默认则立即生效，不需要再调用保存接口。**

VIP 基础配置、晋级奖励和等级管理见 [分站 VIP API 文档](./operator-vip-api.md)。本文只说明图标功能。

---

## 一、服务地址与测试准备

| 项目 | 当前开发环境 node01 示例 | 怎么用 |
| --- | --- | --- |
| 分站 HTTP API | `http://192.168.10.220:38010` | 下文四个接口的公共地址，已确认 node01 端口映射 |
| 查询图标的完整地址 | `http://192.168.10.220:38010/member/vip/icon/list` | 测试工具选择 GET，不传 Body |
| 图片公开前缀 | `http://192.168.10.220:9200/oss-bucket` | 浏览器访问图片使用，不是上传接口地址 |
| 登录 Token | `<access_token>` | 使用目标分站后台登录得到的 Token，不使用总站或代理 Token |

换环境时替换 API 地址和 Token，图片直接使用接口返回的 URL。直连本服务时，不额外加 `/admin`、`/admin/operator`、`/api` 或 `/p9`；经网关访问时，以目标环境实际提供的转发地址为准，不自行推定前缀。

文中的 ID、站点编码 `OP_example`、图片文件名和时间仅作格式示例。实际调用使用查询返回的 `group_no`、上传返回的 Key 和 URL，不直接访问示例图片。

### 1.1 登录与请求头

四个接口都需要有效分站后台身份和对应接口权限。站点归属由服务端从登录身份获取，前端不传 `operator_code`、分站域名或修改人。

```http
Authorization: Bearer <access_token>
```

| 请求头 | 必填 | 填什么 |
| --- | --- | --- |
| `Authorization` | 是 | `Bearer ` 后接目标分站登录 Token，中间保留一个空格 |
| `Content-Type` | 按请求类型 | 保存、恢复使用 `application/json`；上传由工具自动生成带 boundary 的 multipart Content-Type |
| `X-Lang` | 否 | 例如 `zh-CN`，用于提示信息 |
| `Host` | 无需手动设置 | 使用客户端默认值，不用手工填写分站域名 |

前端只调用本文的分站 API。上传时后端会转发到分站 OSS 并透传 Token，账号还需要 OSS 的 `POST /p9/oss/file/upload` 权限。登录及 Token 获取见 [Core API 文档](./core-api.md)。

### 1.2 统一响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `code` | int | 成功为 `0`；失败时与 HTTP 错误状态一致 |
| `data` | object | 本次业务内容；错误响应无业务数据 |
| `msg` | string | 结果提示，示例使用 `ok`；实际文案随语言变化，不依赖固定文案判断成功 |

保存、恢复成功均为 HTTP 200，返回空业务对象：

```json
{
  "code": 0,
  "data": {},
  "msg": "ok"
}
```

保存、恢复响应不返回图片地址或列表。前端应重新查询，使用最新的 `icon_url` 展示。

### 1.3 图标组与图片规格

图标固定十组，不分页，不新增、删除或调整等级范围。保存和恢复使用 **`group_no`**；统一上传使用 **`scene=vipicon` 和 `min_level=group_no×10`**，不使用记录 `id` 定位。

| `group_no` | 对应 VIP 等级 | 查询返回的 `start_level` | 查询返回的 `end_level` |
| --- | --- | --- | --- |
| 0 | 0–9 级 | 0 | 9 |
| 1 | 10–19 级 | 10 | 19 |
| 2 | 20–29 级 | 20 | 29 |
| 3 | 30–39 级 | 30 | 39 |
| 4 | 40–49 级 | 40 | 49 |
| 5 | 50–59 级 | 50 | 59 |
| 6 | 60–69 级 | 60 | 69 |
| 7 | 70–79 级 | 70 | 79 |
| 8 | 80–89 级 | 80 | 89 |
| 9 | 90–99 级 | 90 | 99 |

`group_no=0` 是合法的第一组，不能省略。图片必须是有效 PNG，文件后缀 `.png`，**宽 115 × 高 165 像素，不超过 204800 字节（200 KiB）**。不能把 JPG 改名成 PNG；接口不自动裁剪或缩放。

---

## 二、接口索引

| 用途 | 方法 | 路径 | 请求体 |
| --- | --- | --- | --- |
| 查询十组图标 | GET | `/member/vip/icon/list` | 不传 |
| 上传一张新图标 | POST | `/resource/upload` | form-data |
| 保存某一组自定义图标 | POST | `/member/vip/icon/update` | JSON |
| 某一组恢复总站默认图 | POST | `/member/vip/icon/reset` | JSON |

四个接口都不传 Query，也不传 `page`、`page_size`；当前没有批量保存或批量恢复接口。

---

## 三、查询十组图标

**用途：页面回显，同时取得当前实际展示的图标和总站实时默认图。查询只读，不初始化或保存配置。**

```http
GET http://192.168.10.220:38010/member/vip/icon/list
Authorization: Bearer <access_token>
```

**Query / Body：** 都不传。

**成功响应：HTTP 200。** 为便于阅读，以下只展示第一组，实际固定返回十组：

```json
{
  "code": 0,
  "data": {
    "list": [
      {
        "id": 23,
        "group_no": 0,
        "start_level": 0,
        "end_level": 9,
        "default_icon_url": "http://192.168.10.220:9200/oss-bucket/vipicon/0-9/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.png",
        "icon_url": "http://192.168.10.220:9200/oss-bucket/vipicon/OP_example/0-9/bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb.png",
        "updated_at": 1794268800000
      }
    ]
  },
  "msg": "ok"
}
```

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `data.list` | array | 固定十组，按 `group_no` 从 0 到 9 排列 |
| `data.list[].id` | int64 | 本站记录主键；本站尚无此组记录时为 `0`，不是上传、保存或恢复的请求参数 |
| `data.list[].group_no` | int64 | 图标组编号，0–9 |
| `data.list[].start_level` | int64 | 本组等级下限，含本级，等于 `group_no × 10` |
| `data.list[].end_level` | int64 | 本组等级上限，含本级，等于 `start_level + 9` |
| `data.list[].default_icon_url` | string | 总站本组实时默认完整 URL；总站未配置时为 `""` |
| `data.list[].icon_url` | string | 当前实际展示的完整 URL；优先用本站自定义图，否则使用总站默认图；两者都无图时为 `""` |
| `data.list[].updated_at` | int64 | 本站图标记录的更新时间，Unix 毫秒；本站尚无记录时为 `0`，不是总站换图时间 |

前端展示使用 `icon_url`，不需要自行回退或拼接域名；`default_icon_url` 可用于默认图预览。空 URL 展示为空状态，不渲染损坏图片。

**两个 URL 相同不代表本站一定处于“跟随默认”状态。** 本站可能保存了与总站相同的地址，当前响应没有单独的继承状态字段。需要跟随总站时，调用恢复接口；不要仅因 URL 相同就认为无需恢复。

分站 API 分别调用 CRM 查询本站图标、调用总站查询默认图，然后在 API 中合并，本站图片优先。CRM 的 VIP 查询只读本站数据库，不调用总站；缺少本站图片时返回空字符串供 API 补默认图。

每次 HTTP 列表查询都需要总站默认图服务，即使十组都有自定义图也一样；总站调用最多等待 3 秒，不自动重试。查询失败不能当作空列表或已清空配置，应保留现有页面并提示重试。

后端部署时，`operator-admin-api` 挂载的配置文件需要增加下列配置；地址按实际环境填写。该客户端可选、非阻塞初始化，缺配置时 API 仍可启动，但图标列表返回 HTTP 503；仅支持 `Endpoints` 直连，不使用 Etcd。

```yaml
# 总站默认VIP图标由分站API查询，直连地址按部署环境填写。
PlatformBaseRpc:
  Endpoints:
    - 192.168.10.220:19001
  NonBlock: true  # 总站暂时不可达时不阻止API启动。
  Timeout: 3000   # 请求超时上限，毫秒。
```

CRM 的 `ResourceBaseURL` 仍用于本站图片 Key 拼接；CRM 原有头像、头像框共用的 `PlatformBaseRpc` 配置不属于本次 VIP 调整，不要随之删除。此次 CRM 图标 RPC 的 `default_icon_url` 保留原字段编号和类型，但始终返回空串，`icon_url` 只表示本站已保存图片；其他直接调用该 RPC 的调用方也需自行补默认图。HTTP 方法、路径、JSON 字段及图片展示规则保持不变。发布时先为 API 配置总站地址并更新 API，再更新 CRM；先更新 CRM 会让旧 API 暂时无法补默认图片。

---

## 四、上传一张图标

使用分站统一入口，完整规则及其他场景见 [统一上传文档](./resource-upload-api.md)。旧 VIP 专用上传入口已删除，需要重新授予 `/resource/upload` 的 POST 权限。

**用途：上传图片，取得预览 URL 和保存用的 Key；此接口不修改当前图标配置。**

```http
POST http://192.168.10.220:38010/resource/upload
Authorization: Bearer <access_token>
```

**Query：** 不传。**Body：** 在 Apifox / Postman 中选择 `form-data`，恰好提交以下三个字段，各一次：

| 字段 | 工具中的类型 | 必填 | 示例与说明 |
| --- | --- | --- | --- |
| `scene` | Text / 文本 | 是 | 固定 `vipicon` |
| `min_level` | Text / 文本 | 是 | `group_no×10`，只能为 0、10…90；0 也必须明确提交 |
| `file` | File / 文件 | 是 | 选择一张本地 PNG，115 × 165 像素，最大 200 KiB |

不能用 JSON 上传，不能用 Text 填文件路径或图片 URL。Content-Type 由工具自动生成，不手写遗漏 boundary 的 `multipart/form-data`。不要额外传 `id`、`group_no`、`operator_code`、桶名或 OSS 地址；额外表单字段或重复字段会被拒绝。

**成功响应：HTTP 200。**

```json
{
  "code": 0,
  "data": {
    "url": "http://192.168.10.220:9200/oss-bucket/vipicon/OP_example/0-9/bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb.png",
    "key": "vipicon/OP_example/0-9/bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb.png",
    "scene": "vipicon",
    "visibility": "public",
    "url_expires_in": 0
  },
  "msg": "ok"
}
```

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `data.url` | string | 完整公开 URL，用于预览，也可提交保存 |
| `data.key` | string | 桶内对象 Key，不含协议、主机或桶路径；保存时放入 `icon_url` |
| `data.scene` | string | 固定 `vipicon` |
| `data.visibility` | string | 固定 `public`，图片公开访问 |
| `data.url_expires_in` | int64 | 固定 `0`，表示永久公开地址 |

本站图片目录为 `vipicon/{登录站点编码}/{等级下限}-{等级上限}/{服务端生成的文件名}.png`。前端直接使用返回值，不自行生成文件名，不把其他站点或其他等级组的 Key 用于当前组。

用 `data.url` 更新本次预览，将 `data.key` 暂存在待保存数据中。只上传不保存、取消编辑或关闭页面，都不会修改当前图标；已上传的对象不会因此自动删除。

---

## 五、保存某一组图标

**用途：将刚上传的图片保存为本站指定组的当前图标，一次只保存一组。**

```http
POST http://192.168.10.220:38010/member/vip/icon/update
Authorization: Bearer <access_token>
Content-Type: application/json
```

**Query：** 不传。**Body：** 选择 JSON：

```json
{
  "group_no": 0,
  "icon_url": "vipicon/OP_example/0-9/bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb.png"
}
```

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `group_no` | int64 | 是 | 要保存的组，0–9，0 有效 |
| `icon_url` | string | 是 | 对应组上传返回的 `data.key` 或完整 `data.url`，最长 1024 字节；不能省略、传 `null`、空字符串或纯空白 |

字段名仍叫 `icon_url`，不要改为 `key`。优先直接提交上传返回的 Key；完整公开 URL 同样可用。VIP 目录须与当前站点、等级组匹配。历史完整公开 URL 继续兼容，但不要提交本地路径、临时签名地址或自行拼造的图片路径。

**成功响应：** 第一节的 `data: {}`。成功后重新查询列表，取得最终完整图片地址；不会在保存响应中返回 URL。

保存只更改本站目标组的图片和更新时间，保留原 ID、创建时间以及其他组的已有图片。本站缺少图标组时，会在同一事务内补齐缺失的空记录；写入任一步失败均回滚。保存不修改总站配置、会员等级或其他站点，也不删除旧 OSS 图片。

**清空图片不能调用此保存接口，恢复默认必须使用下一节的 `/icon/reset`。** 服务端保存时校验地址与归属，不联网确认图片存在，因此应使用实际上传成功的返回值。

---

## 六、恢复总站默认图

**用途：清空本站指定组的自定义图片配置，之后随总站默认图变化。点击恢复即保存，不需要再点击保存按钮。**

```http
POST http://192.168.10.220:38010/member/vip/icon/reset
Authorization: Bearer <access_token>
Content-Type: application/json
```

**Query：** 不传。**Body：** 选择 JSON：

```json
{
  "group_no": 0
}
```

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `group_no` | int64 | 是 | 要恢复的单个组，0–9；0–9 级传 `0`，10–19 级传 `1` |

不传 `id`、`icon_url` 或 `default_icon_url`。**不要先查询默认 URL，再调用保存接口写入那个 URL，这会保存一份固定图片配置，不能保证后续跟随总站换图。**

**成功响应：** 第一节的 `data: {}`。成功后重新调用查询接口刷新当前图标，清除页面上该组尚未保存的上传预览和 Key。

恢复的实际效果：

- 只将登录站点目标组保存的 `icon_url` 清为空字符串，并更新该组修改时间；不删除数据库记录或 OSS 图片。
- 缺少本站图标组时补齐空记录，已有其他组的图片保持不变；其他站点、总站默认配置、会员及等级数据不受此操作影响。
- 后续查询使用总站本组最新默认图；总站换图后再次查询即可看到新图，无需再次恢复。
- 总站也没有默认图片时，恢复仍可成功，后续查询的当前图片为空。

恢复写入本身不查询总站；成功后的列表刷新仍依赖总站。如果恢复返回成功而刷新失败，应提示“恢复成功，列表刷新失败，请重试查询”，不要提示恢复未生效或重新保存旧图。恢复或保存请求自身超时则结果可能尚不明确，应重新查询确认，不直接按失败回填旧配置。

---

## 七、常见错误与前端对接顺序

### 7.1 常见错误

| 状态 | 常见原因 | 前端处理 |
| --- | --- | --- |
| HTTP 400 | 参数缺失、组编号越界、表单重复或额外字段、图片规格或地址不合法 | 检查对应字段表，不提示成功 |
| HTTP 401 / 403 / 498 | 登录无效、登录过期、身份不符或无接口权限 | 按现有分站登录与权限流程处理；上传还需 OSS 上传权限 |
| HTTP 502 | 上传连接失败、超时或 OSS 返回异常结果 | 提示上传失败，保留当前已保存配置 |
| HTTP 503 | CRM、总站默认图服务不可用，或 OSS 地址、图片公开前缀等配置未就绪 | 提示暂不可用；持续失败反馈后端，不自行替换图片域名 |
| HTTP 504 | 图标查询、保存或恢复调用 RPC 超时 | 查询失败时重试查询；写请求结果不明确时先重新查询确认 |
| HTTP 500 | 服务内部异常 | 保留待编辑内容，反馈后端排查 |

上传时 OSS 返回的 4xx 会作为失败状态返回，OSS 5xx、响应不合法或上传连接超时通常转换为 502。网关也可能返回自身的大小限制或超时错误，以实际响应为准。

错误响应示例，提示文字仅用于说明：

```json
{
  "code": 400,
  "msg": "参数不符合要求"
}
```

### 7.2 对接顺序

1. 使用目标分站后台账号登录，确认四个 VIP 接口和 OSS 上传接口权限。
2. 打开图标页时查询十组，使用 `group_no` 对应每个编辑项，使用 `icon_url` 展示当前图片。
3. 用户选择图片后校验规格，上传时提交 `scene=vipicon + min_level=group_no×10 + file`；上传成功用 `data.url` 预览，暂存 `data.key`。
4. 用户点击保存时，将同一组的 `group_no + icon_url` 提交给保存接口，`icon_url` 使用刚暂存的 Key。
5. 用户点击恢复默认时，只提交目标 `group_no` 到恢复接口，成功后不再调用保存接口。
6. 保存或恢复成功后重新查询。对同一组的保存、恢复请求应串行处理，避免旧的待保存图片覆盖刚完成的恢复。

前端不直接编辑数据库，不调用 OSS 删除接口，也不需要调用总站素材接口。所有操作都以接口返回结果为准，不仅凭上传预览或按钮点击判断配置已生效。
