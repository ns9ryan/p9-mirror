# 总站默认头像和头像框

## 一、功能

总站维护固定 8 个头像、6 个头像框，提供查询、上传、修改以及随机读取一个默认头像的 RPC。按已有记录 `id` 修改，不新增、删除或排序；新会员可由分站调用随机头像 RPC 获取初始头像，保存会员资料以及“不使用头像框”由分站处理。

上传只把文件写入 OSS，调用修改接口成功后才更新素材配置。替换时生成新文件名并保留旧图片，不覆盖已使用的原文件。

| 类型 | 格式与尺寸 | 单张大小 | OSS 对象目录 |
|---|---|---|---|
| 头像 | PNG，200×200 像素 | 不超过 200×1024 字节 | `avatar-resource/avatars/` |
| 头像框 | 透明 PNG，145×145 像素 | 不超过 200×1024 字节 | `avatar-resource/avatar-frames/` |

头像框沿用中心像素透明、至少存在一个可见像素的检查；不自动裁剪、缩放或把头像与框合成一张图片。

## 二、配置和数据库保存方式

总网 API 的实际 YAML 配置内部上传服务：

```yaml
OssApiBaseURL: "http://<总网 OSS API 地址>:<端口>"
```

总网 RPC 的实际 YAML 配置图片公开前缀，与 VIP 共用一项：

```yaml
ResourceBaseURL: "https://<图片公开域名>/<实际桶路径>"
```

公开前缀应与 OSS 返回的图片 URL 一致，包含必要的桶名或 CDN 路径；不是上传服务地址。仓库位置为 `rpc/etc/platform_base.yaml`，容器示例为 `deploy/docker/rpc.yaml`，运行时以实际挂载配置为准。

OSS 的有效上传配置需包含以下公开场景，保留已有其他场景：

```yaml
upload:
  scenes:
    default_avatar:
      visibility: public
    default_avatar_frame:
      visibility: public
```

表名为 `default_avatar`、`default_avatar_frame`。两张表都保留 `id`、`image_url`、`updated_by`、`created_at`、`updated_at`，不重命名原列：

- `image_url` 内部保存对象 Key，兼容已有完整 URL；初始化空字符串表示还未配置图片。
- 修改时，已知公开前缀下的 URL 裁成 Key，也可直接提交本类 Key；不能清空或混用头像与头像框。
- 查询时，Key 拼上当前公开前缀，`image_url` 对外仍返回完整 URL；空值仍为空。
- 未知前缀的旧完整 URL 原样保留；查询不自动写库或迁移历史数据。
- 前缀未配置时仅兼容完整 URL；处理 Key 时返回配置错误。配置校验发生在请求中，不增加启动依赖。
- 修改公开前缀不会迁移图片，必须确保新地址下实际存在相同 Key 的对象。

固定记录沿用已有初始化机制，只补缺失记录、不覆盖已有图片。本次存 Key 调整不改变初始化与启动顺序，也不自动转换旧数据。
已有头像达到或超过八条、头像框达到或超过六条时，不再补充，也不因数量超限阻止启动或删除记录；数据库查询、写入或事务提交失败仍按原流程报错退出。

## 三、分站后端调用 RPC

| 类型 | RPC 服务 | 查询方法 | 请求与响应类型包 |
|---|---|---|---|
| 头像 | `DefaultAvatarService` | `ListAll` | `oa.98ent.com/p9/platform-base/rpc/pb/platformbaserpc/defaultavatarpb` |
| 头像框 | `DefaultAvatarFrameService` | `ListAll` | `oa.98ent.com/p9/platform-base/rpc/pb/platformbaserpc/defaultavatarframepb` |

客户端包分别为 `rpc/client/defaultavatarservice`、`rpc/client/defaultavatarframeservice`，模块前缀为 `oa.98ent.com/p9/platform-base`。调用方沿用其已有总网 RPC 配置和 ServiceContext 初始化方式，连接总网 `platform-base-rpc`；不需要读取总网数据库或自行配置图片前缀。

跨服务调用放在调用方 API 层对应的 Logic，客户端在 ServiceContext 中按非阻塞方式初始化；请求时处理超时和不可用错误，不把总站在线作为 RPC 服务启动条件。

```go
result, err := l.svcCtx.DefaultAvatarRpc.ListAll(
    l.ctx,
    &defaultavatarpb.ListAllDefaultAvatarsRequest{},
)
if err != nil {
    return nil, err
}
```

头像框同理使用 `DefaultAvatarFrameRpc.ListAll` 和 `ListAllDefaultAvatarFramesRequest`。查询请求为空，返回 `List`，每项含 `Id`、`ImageUrl`；后者已是完整公开地址。空列表为空集合，查询不插入数据。RPC 返回 Protobuf 对象和 `error`，没有 HTTP 响应包装。

查询 RPC 不要求传入总站后台管理员令牌，仍受部署环境既有服务访问限制。修改 RPC 仅允许有效总站后台身份，拒绝分站用户和预览令牌；分站接入查询不需要调用修改。

### 新会员随机初始头像

分站通过现有 `DefaultAvatarRpc` 客户端调用 `Random`，无需新增 HTTP 接口或传入会员 ID：

```go
result, err := l.svcCtx.DefaultAvatarRpc.Random(
    l.ctx,
    &defaultavatarpb.RandomDefaultAvatarRequest{},
)
if err != nil {
    return nil, err
}
avatar := result.Avatar
```

请求类型为 `RandomDefaultAvatarRequest`，响应类型为 `RandomDefaultAvatarResponse`，均在 `defaultavatarpb` 包。响应 `Avatar` 复用 `DefaultAvatarInfo`，包含 `Id` 和可直接展示的完整 `ImageUrl`；已存 Key 使用当前 `ResourceBaseURL` 拼接，旧完整 URL 保持兼容。

只在已配置图片的槽位中等概率抽取，空图片不参与；允许不同会员抽到同一张，没有可用头像时返回 gRPC `NotFound`，数据库故障和公开前缀错误继续返回失败。此调用不检查 OSS 对象是否实际存在，也不写任何数据库或上传文件。

分站应在创建会员、确需设置初始头像时调用一次，并保存本次选中的头像资料；不要在每次查询会员时重新抽取。会员写入、失败重试及是否继续注册由分站业务处理，此 RPC 不修改分站注册流程。更新依赖后可使用新增客户端方法，已有列表和修改接口保持兼容。
