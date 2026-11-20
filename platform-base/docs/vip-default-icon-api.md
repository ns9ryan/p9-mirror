# 总网默认 VIP 图标

## 一、功能说明

总网统一维护十组 VIP 默认图标，供总网后台管理和分站后端读取。

| 等级组 | 等级下限 `min_level` | 等级上限 `max_level` |
|---|---|---|
| 0–9 | 0 | 9 |
| 10–19 | 10 | 19 |
| 20–29 | 20 | 29 |
| 30–39 | 30 | 39 |
| 40–49 | 40 | 49 |
| 50–59 | 50 | 59 |
| 60–69 | 60 | 69 |
| 70–79 | 70 | 79 |
| 80–89 | 80 | 89 |
| 90–99 | 90 | 99 |

其他服务通过 RPC 读取十组默认图标。响应中的 `IconUrl` 已是完整公开地址，调用方无需自行拼接。

- 上传只生成图片资源，保存后才更新默认图标配置。
- 每次可以保存一至十组；未提交的组保持原值。
- 查询始终按等级升序返回十组，尚未配置的组返回空图片地址。
- 总网更新默认图标不会直接修改分站数据库。分站 API 已实现默认图标合并：有本站图片时优先展示，本站图片为空时使用总站实时默认图；恢复默认只清空本站本组图片配置。
- 替换或清空配置不删除 OSS 旧图片，避免影响仍在使用旧地址的调用方。

本文只说明后端 RPC 接入、返回值和相关配置。

## 二、配置说明

以下地址均需替换成对应环境的实际地址，示例中的尖括号内容为占位符。

### 2.1 服务地址

| 用途 | 服务 | 仓库配置默认端口 | 地址要求 |
|---|---|---|---|
| 后端查询默认图标 | 总网 `platform-base-rpc` | 19001 | 调用方后端能够访问的 RPC 地址，不加 `http://` |
| API 转发图片上传 | 总网身份认证对应的 `oss-api` | 按实际部署 | 总网 API 能够访问的 HTTP(S) 根地址 |

端口映射、服务发现和网关前缀以实际部署为准。RPC 复用现有内网服务，不需要新增监听端口。

### 2.2 分站等调用方后端配置

分站 VIP 查询在 operator-admin-api 的实际运行 YAML 中配置总网 RPC 地址，不在 CRM VIP 逻辑中调用：

```yaml
PlatformBaseRpc:
  Endpoints:
    - "<总网 RPC 内网地址>:<实际 RPC 端口>"
  NonBlock: true
  Timeout: 3000
```

`Timeout` 单位为毫秒。VIP 查询客户端仅使用 `Endpoints`，并强制非阻塞创建；单次请求最多等待 3 秒，调用失败返回相应错误，不自动重试。

operator-admin-api 当前 VIP 客户端只接受 `Endpoints`，仅配置 Etcd 不会创建该客户端。其他调用方按自身现有 ServiceContext 初始化方式接入，不能直接套用为相同的服务发现行为。

仅调用 RPC 查询默认图标时，无需配置 OSS 地址，也无需连接总网数据库。

### 2.3 总网服务配置

总网 `platform-base-api` 的 YAML 中配置 OSS API 根地址：

```yaml
OssApiBaseURL: "http://<OSS API 地址>:<实际 OSS API 端口>"
```

该值不包含 `/p9/oss/file/upload`，程序会自动追加上传路径。未配置或地址格式无效时，上传返回 503，查询和保存仍可使用。

总网 `platform-base-rpc` 的 YAML 中配置三类默认素材共用的公开读前缀：

```yaml
ResourceBaseURL: "https://<公开图片域名>/<桶路径>"
```

前缀应与 OSS 上传响应的公开地址一致，包含实际需要的桶名或 CDN 路径，不是 `OssApiBaseURL`。没有桶路径的 CDN 地址只填写域名。VIP、头像和头像框共用这一项，不读取其他服务的 YAML，也不增加 RPC 依赖。

- 保存已知前缀下的本组 URL 时，后端核对协议、主机端口和完整桶路径，再裁成 `vipicon/0-9/随机文件名.png` 这样的 Key。
- 查询已存 Key 时，在前面拼上 `ResourceBaseURL`；空值仍返回空字符串。
- 未配置前缀时，已有完整 URL 继续使用；直接保存或读取 Key 返回配置错误，不把裸 Key 作为图片地址返回。前缀格式错误也在请求时返回错误，不改变服务启动条件。
- 已有其他前缀的完整 URL 仍原样返回，不自动换域名、不批量更新旧数据。旧 URL 重新保存为 Key 时，需要匹配当时的公开前缀；跨域名历史数据另行确认转换方案。
- 更换前缀只改变已存 Key 的查询地址，不会搬运图片。更换桶或存储服务时，必须保证对应 Key 的图片确实存在。

数据库沿用 `icon_url` 列，长度、默认值和约束不变，保存内容改为 Key，并兼容旧完整 URL。此次不重命名数据库列、不执行数据迁移。

OSS 的有效配置中需包含以下场景；已有其他上传场景时保留并追加：

```yaml
upload:
  scenes:
    vipicon:
      visibility: public
```

配置写入实际加载的 YAML 或 Etcd。OSS 需连接能够验证总网登录令牌的 Core，图片公开地址需配置为后端业务调用方可访问的地址。

总网 `platform-base-api`、`platform-base-rpc` 和 `oss-api` 需使用包含本功能的版本。图标表 `vip_default_icon_group` 沿用 RPC 原有自动迁移流程创建，无需额外手动建表。默认图片通过上传、保存进行配置。

## 三、后端通过 RPC 读取默认图标

### 3.1 服务与方法

| 项目 | 内容 |
|---|---|
| 服务 | 总网 `platform-base-rpc` |
| RPC 服务名 | `VipDefaultIconService` |
| 查询方法 | `ListAll` |
| Go 客户端包 | `oa.98ent.com/p9/platform-base/rpc/client/vipdefaulticonservice` |
| 请求、响应类型包 | `oa.98ent.com/p9/platform-base/rpc/pb/platformbaserpc/vipdefaulticonpb` |
| 请求类型 | `ListAllVipDefaultIconsRequest`，空请求 |
| 响应类型 | `ListAllVipDefaultIconsResponse`，包含 `List` |

调用方需更新 `oa.98ent.com/p9/platform-base` 依赖到包含该客户端的版本。以下示例仅调用查询，不要求传入总网管理员令牌或站点编码；连接方式沿用部署环境现有的 RPC 配置。保存 RPC 会校验总网用户身份，分站读取默认图标无需调用保存方法。

### 3.2 Go 调用示例

分站在 operator-admin-api 的配置结构体中保留可选的 `PlatformBaseRpc zrpc.RpcClientConf`，由现有配置流程读取第二节的 YAML。
`internal/svc/service_context.go` 的 `newVipDefaultIconClient` 仅接受 `Endpoints`，设置 `NonBlock=true` 后通过 `zrpc.NewClient` 创建客户端；缺配置时返回空客户端，创建失败时记录日志并返回空客户端，不增加总站启动依赖。

跨服务调用直接位于 API 的 `internal/logic/member/vip/list_icon_groups_logic.go`，CRM RPC 仅查询本站数据库。以下是查询 Logic 的实际调用片段，使用该包现有的错误转换方法：

```go
// API 直接查询总站实时默认图；缺客户端返回 503，单次调用最多等待 3 秒。
if l.svcCtx.VipDefaultIconRpc == nil {
	return nil, iconRequestError(status.Error(codes.Unavailable, "common.service_unavailable"))
}
ctx, cancel := context.WithTimeout(l.ctx, 3*time.Second)
defer cancel()
defaults, err := l.svcCtx.VipDefaultIconRpc.ListAll(ctx, &vipdefaulticonpb.ListAllVipDefaultIconsRequest{})
if err != nil {
	return nil, l.defaultIconError(err)
}
```

调用成功后校验十组等级范围和图片地址，再按组编号与本站查询结果合并；只补 HTTP 响应，不将总站默认图写回分站数据库。总站超时返回 504，不可用或返回资料不合法时返回 503。

### 3.3 RPC 返回字段

RPC 返回 Protobuf 对象，不带 HTTP 的 `code`、`msg`、`data` 包装。调用成功时读取 `response.List`，失败时处理返回的 `error`。

| Go 字段 | 类型 | 含义 |
|---|---|---|
| `List` | `[]*VipDefaultIconGroupInfo` | 固定十组图标，按等级下限升序排列 |
| `List[i].MinLevel` | `int64` | 等级下限，含本级 |
| `List[i].MaxLevel` | `int64` | 等级上限，含本级 |
| `List[i].IconUrl` | `string` | 默认图标公开 URL；空字符串表示该组未配置 |

例如 VIP 23 对应 `MinLevel = 20`、`MaxLevel = 29` 的记录。分站读取该记录的 `IconUrl` 作为默认图片地址；地址为空时由分站处理空状态。RPC 调用失败不能视为“总网清空了图片”，应按调用方既有错误处理或缓存策略处理。
