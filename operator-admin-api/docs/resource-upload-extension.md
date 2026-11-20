# 新增统一上传场景

分站统一上传入口是 `POST /resource/upload`，目前只接入公开图片。新增图片类型时，先确定唯一的 `scene` 名称、图片规格，以及上传后由哪个业务接口保存返回的 `key` 或 `url`。下文的 `banner` 仅为新场景示例，当前尚未接入。

## 接入位置

1. 在 `platform-base/pkg/resourceupload/form.go` 增加场景常量，在本仓库 `desc/resource.api` 更新 `scene` 注释。请求仍使用 `file`、`scene`；只有 VIP 图标使用 `min_level`，新增场景不要复用或强制提交它。增加专用字段时，须同步修改公共表单解析和 API 源定义，不能删除原有额外字段、重复字段检查。
2. 在 `internal/logic/resource/upload_resource_logic.go` 的分站场景白名单中加入新场景。分站场景只能由本站后台身份使用，站点编码必须继续从已认证 Token 获取，不能从表单读取。
3. 在公共包 `platform-base/pkg/resourceupload/upload.go` 中，给 `validImage`、`readImage` 和 `validKey` 补齐新场景的格式、尺寸、大小及本站 Key 规则。公共实现和 OSS 目录接入位置见 [总站接入说明](../../platform-base/docs/resource-upload-extension.md)，不要在 Handler 复制上传逻辑。

   分站 Handler 当前请求上限约为 33 MiB，但公共包除 Logo 外的单文件上限仍为 200 KiB。如果新场景需要其他大小，按场景设置；涉及总请求上限时，同时调整 `desc/resource.api` 的 `maxBytes`、`internal/handler/resource/upload_resource_handler.go` 的解析上限和 OSS 限制，保留 multipart 开销及原场景约束。

4. 通用 OSS 规划在 `oss/common/object/object.go` 的 `PlanObjectKey`；例如新分站场景通常生成 `{站点编码}/banner/{日期}/{随机文件名}`。特殊目录才增加专用规划。公共包 `validKey` 必须核对 Token 中本站编码、完整场景目录和文件名结构，不能只判断宽泛前缀。
5. OSS 上传入口在 `oss/api/internal/logic/upload/upload_logic.go` 的 `UploadWithMinLevel`，最终由 `putBytes` 写入对象存储。有新业务专用的图片或身份限制时，在上传操作中接入对应校验；不要仅在 API 校验而让直连 OSS 绕过限制。站点编码仍取自认证上下文，不接受表单指定。
6. 在分站使用的 OSS API 实际生效配置中，在已有 `upload.scenes` 下追加新项，保留其他场景。仓库示例是 `oss/api/etc/oss.yaml`，运行环境应核对实际加载来源或挂载路径。以下只展示新增片段：

   ```yaml
   upload:
     scenes:
       banner: # 新增公开 Banner 场景，示例名称。
         visibility: public # 使用现有公开桶。
   ```

   当前公共包只接受 `visibility=public`、`url_expires_in=0`，私密素材不能只改成 `private`。分站 OSS API 必须连接与分站后台认证匹配的 Core，不能把分站 Token 转给总站 OSS API；继续复用已有 `OssApiBaseURL`。

7. 如果新场景需要站点归属或业务前置检查，在 `internal/logic/resource/upload_resource_logic.go` 中直接调用本站对应 RPC；不要把跨服务调用放进 RPC 服务内部，也不要改变服务启动依赖。业务保存接口和前端说明同步写清返回 `key`、`url` 的使用方式。

   修改 API 源定义后，按本仓库 `make api` 生成配套代码。前端参数和响应说明更新到 [图片统一上传接口](../../project-resources/docs/resource-upload-api.md)。公共包来自 `platform-base` 依赖，交付时须确认分站使用的依赖版本包含新增场景，不能只依靠本机 `go.work` 生效。

## 处理边界

- Handler 只解析 multipart 表单；身份、站点归属、场景权限、图片检查、OSS 转发和返回素材归属检查放在 Logic 或公共上传包。
- 上传只创建 OSS 对象并返回素材信息，不自动写入业务配置，也不删除旧对象。
- 新场景不能通过任意客户端字符串自动创建目录，必须同时经过分站 API 白名单、图片校验、站点 Key 校验和 OSS `upload.scenes` 配置。
- 总站专属素材不要加入分站白名单；需要总站数据的查询或合并，应在 API 层调用总站 RPC。
- 现有入口共用一个上传权限，新增 `scene` 不自动拥有独立权限；新业务要求单独授权时，在 API Logic 明确接入对应判断。
