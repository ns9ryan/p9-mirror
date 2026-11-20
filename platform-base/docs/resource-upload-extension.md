# 新增统一上传场景

总站统一上传入口是 `POST /resource/upload`，目前只接入公开图片。新增图片类型时，先确定唯一的 `scene` 名称、图片规格，以及上传后由哪个业务接口保存返回的 `key` 或 `url`。下文的 `banner` 仅为新场景示例，当前尚未接入。

## 接入位置

1. 在 `pkg/resourceupload/form.go` 增加场景常量。只有需要公共表单规则时才在这里增加解析字段；现有 `min_level` 只属于 `vipicon`。
2. 在 `api/internal/logic/resource/upload_resource_logic.go` 的场景白名单中加入新场景。总站只允许总站素材，不能在这里放行分站专属场景。
3. 在 `pkg/resourceupload/upload.go` 的 `validImage` 中加入图片规则，在 `readImage` 中按需区分大小上限，在 `validKey` 中加入返回目录规则。Key 校验应核对完整目录结构、日期和随机文件名，不能只判断 `banner/` 前缀。需要业务专用校验函数时放在同文件，不为几行逻辑增加包装层。

   总站当前图片上限为 200 KiB。如果新场景需要更大文件，须同时核对 `api/desc/resource.api` 的 `maxBytes`、`api/internal/handler/resource/upload_resource_handler.go` 传入的请求上限、公共包单文件上限和 OSS 上传上限；请求上限还需包含 multipart 开销，不要直接放大其他场景的限制。
4. 如果新场景需要特殊目录，在 `oss/common/object/object.go` 增加对应的 Key 规划；普通公开图片可以沿用通用规划。Key 必须固定以新场景目录开头，并由服务端生成随机文件名。
5. OSS 上传入口在 `oss/api/internal/logic/upload/upload_logic.go` 的 `UploadWithMinLevel`。有业务专用图片或身份限制时，在该操作对应文件增加校验并接入；只有通用格式、大小和目录要求时可以沿用现有处理。最终由 `putBytes` 写入对象存储，不能只在 API 校验却允许直连 OSS 绕过必要的业务限制。
6. 在 OSS API 的实际生效配置中，在已有 `upload.scenes` 下追加新项，保留其他配置。仓库示例是 `oss/api/etc/oss.yaml`，运行环境应核对实际加载来源或挂载路径。以下只展示新增片段：

   ```yaml
   upload:
     scenes:
       banner: # 新增公开 Banner 场景，示例名称。
         visibility: public # 使用现有公开桶。
   ```

   配置中的场景名必须与 API 白名单完全一致。当前统一入口的 `validResult` 只接受 `visibility=public`、`url_expires_in=0`；私密素材不能只把此处改为 `private`，需要另行设计访问授权和响应契约。

7. 在 `api/desc/resource.api` 更新 `scene` 注释；修改 API 源定义后，按仓库现有 `make api` 生成配套代码。调用方参数和响应契约同步更新到后端接口定义及相关服务说明。新业务另外提供保存入口，校验素材归属后保存，不在上传接口顺带写库。

## 处理边界

- Handler 只解析 multipart 表单；场景权限、图片检查、OSS 转发和返回素材归属检查放在 Logic 或公共上传包。
- 上传只创建 OSS 对象并返回素材信息，不自动写入业务配置，也不删除旧对象。
- 新场景不能通过任意客户端字符串自动创建目录，必须同时经过总站 API 白名单、图片校验、Key 校验和 OSS `upload.scenes` 配置。
- 如果新场景属于分站或需要分站身份，不要加入总站白名单，应按分站文档接入。
- 公共包由总站和分站共用，新增分支只对新场景生效。当前两个入口各只有一个上传权限；若新业务要求单独授权，需要在 API Logic 明确接入对应权限判断，不能默认 `scene` 自带权限隔离。
