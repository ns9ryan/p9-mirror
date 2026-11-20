# integration 运维与部署

HTTP 见 [API.md](API.md)，其它服务 gRPC 见 [RPC.md](RPC.md)。总览见 [integration.md](integration.md)。仓库索引见 [../README.md](../README.md)。

---

## 部署准备

- 构建环境需提供 Go 及项目依赖，容器运行环境需提供 Docker 和 Docker Compose。
- RPC 需要可访问的 PostgreSQL；启用消息 Etcd 配置时，还需提供可访问的 Etcd 及通道账号配置。
- HTTP API 还需要连接 integration RPC 和 core RPC。
- 服务地址、数据库连接和 Etcd 地址均按部署环境填写。调用方可通过 `Endpoints` 直连 RPC，或配置 Etcd 服务发现。

文档中的 `*.example.com` 为示例域名，需替换为实际部署地址。

配置模板：`api/etc/integration.yaml`、`rpc/etc/integration.yaml`。RPC 通过配置中的 `DB.DSN` 读取数据库连接串，部署时填写对应环境的数据库连接信息。

Docker 部署使用 `deploy/etc/` 下的配置文件，挂载方式见下文。离线库路径相对 RPC 工作目录，需按实际部署路径挂载。

---



## Docker 部署

`Dockerfile-*` + `deploy/docker-compose.yml`。

### 构建镜像

在 `integration/` 根目录（需能访问内网 Go module）：

```bash
make docker          # linux amd64 二进制 + 打镜像
# make publish-docker  # 推到 DOCKER_REPO
```

镜像名：`${DOCKER_REPO}/integration-rpc:${VERSION}`、`integration-api`。默认仓库见 `Makefile` / `deploy/.env`。

### 启动

```bash
cd deploy
docker compose --env-file .env up -d
```


| 端口      | 服务                                                 |
| ------- | -------------------------------------------------- |
| `18005` | HTTP API                                           |
| `19005` | gRPC RPC                                           |
| `15433` | PostgreSQL（库名 `integration`，避免与 core 的 `15432` 冲突） |


配置挂载：

- `deploy/etc/integration-rpc.yaml` → 容器 `etc/integration.yaml`
- `deploy/etc/integration-api.yaml` → 容器 `etc/integration.yaml`

API 的 `CoreRpc` 默认 `host.docker.internal:19000`（宿主机上的 core-rpc）。若 core 也在 Docker 且共用网络，可改成 `core-rpc:19000`，并把两个 compose 的 `p9-network` 设为同一 external network。

离线 mmdb/xdb：在 `docker-compose.yml` 取消注释 `../assets:/app/assets`，并在 rpc yaml 里 `IsEnabled: true`。

---



## 配置

HTTP 用 `IntegrationRpc.Endpoints`（填写部署环境的 RPC 地址）和 `CoreRpc.Endpoints`**（JWT + Casbin，必填）**。启动时向 core `RegisterCatalog` 注册本服务 API。RPC 用 `DB`、`IpGeo` 和 `Message`；消息配置说明见下文。

`DB.DSN` **必填**。`DB.Driver` 若填则必须是 `postgres`。库不存在会自动创建；表 `ip_geo_cache`、`ip_geo_provider`、`ip_geo_provider_policy`、`ip_geo_local` 由 Ent `Schema.Create` 建出（与 oss 相同）。

`ip_geo_provider` 为空时，把 yaml `IpGeo.Providers` 写入库作为**全局服务商**（默认 `IsEnabled`/`Priority`/`Tag`）。已有数据不覆盖；服务商无管理接口，改连接配置需清空该表再重启。种子里的 `local` 对应本地 IP 表，数据走 `POST /p9/integration/ipgeo/local/`*。

查询 IP：合法公网按租户策略选出启用源，每个源先查缓存（KV，键 `name|IP`）。默认 `IpGeo.CacheDriver: table` 写 `ip_geo_cache`（`payload` 一整包 JSON）；`memory` 不落库；`redis` 用 `IpGeo.CacheRedis`（`GET`/`SET`/`EXPIRE`，键前缀默认 `ipgeo`）。改策略不用清缓存，停用的源不会再被读到。`POST /policy/list` 按登录 `operator_code` 取策略；`POST /policy/save` 写租户策略；`POST /policy/default/save` 写 Provider 默认启用/优先级。其它服务可在 gRPC metadata `x-operator-code` 上传租户编码，不要放请求体。

`IpGeo.CacheTTL`：过期小时数；查询命中会按该时长续期（表驱动只 UPDATE `expires_at`；Redis 用 `EXPIRE`）；`0` = 不过期。默认 yaml 为 `168`。切 Redis：`CacheDriver: redis`，并保证 `CacheRedis.Host` 可连（填写实际 Redis 地址）。

生产若用 etcd 服务发现：RPC 配 `Etcd.Hosts` / `Key`，API 把 `IntegrationRpc` 改成同样的 `Etcd`（不要再写 `Endpoints`）。

## 邮件通道配置

邮件和短信都运行在当前 RPC 中。仅调用消息接口时不需要启动 HTTP API；RPC 启动依然需要有效 `DB.DSN`。两项业务分别由 `Message.Email.Enabled`、`Message.SMS.Enabled` 控制，省略时均为 false；远端配置不能绕过服务总开关。账号配置填写在 integration，调用方不传密钥。

单 SMTP 的 YAML 和 Etcd 格式、TLS、代理、重试、记录容量说明见 [SendEmail](RPC.md#sendemail)。通过 YAML 配置账号时移除 `Message.Email.Etcd` 或将其 Hosts 清空；YAML 修改需要重启。启用 Etcd 时，配置如下，合并到现有 `Message` 下并保留 SMS 等其它字段：

```yaml
Message:
  Email:
    Enabled: true
    Etcd:
      Hosts:
        - etcd.example.com:2379
      Key: /email-service/providers/smtp
      DialTimeout: 5
```

该 key 的完整 value 可使用下面的多通道格式。示例凭证均为占位值，只加入实际开通的通道；完整替换 key 前先保存原值，不能仅上传其中一段以免覆盖现有通道：

```yaml
enabled: true
providers:
  - name: smtp-main
    type: smtp
    enabled: true
    priority: 10
    smtp:
      host: smtp.example.com
      port: "465"
      auth_email: sender@example.com
      auth_password: "<SMTP_PASSWORD>"
      tls_mode: implicit
      tls_skip_verify: false
      timeout_seconds: 10
  - name: aliyun-main
    type: aliyun
    enabled: true
    priority: 20
    aliyun:
      region: ap-southeast-1
      access_key_id: "<ALI_ACCESS_KEY_ID>"
      access_key_secret: "<ALI_ACCESS_KEY_SECRET>"
      account_name: sender@example.com
      from_alias: "Example"
  - name: aws-main
    type: aws
    enabled: true
    priority: 30
    aws:
      region: us-east-1
      access_key_id: "<AWS_ACCESS_KEY_ID>"
      secret_access_key: "<AWS_SECRET_ACCESS_KEY>"
      from: sender@example.com
email_blacklist:
  - blocked.example.com
block_china_email: false
retry:
  max_attempts: 3
  backoff_ms: 100
  max_backoff_ms: 5000
record_limit: 100
```

配置非空 `providers` 数组后，候选通道仅使用数组中的配置，单 SMTP 配置 `smtp_provider` 不会自动加入。上例已显式列出 SMTP；请求精确填写 `provider: "smtp-main"`、`"aliyun-main"` 或 `"aws-main"` 可逐通道验证，不填则按 priority 从小到大选择。名称必须唯一；某个启用通道配置不合法，会拒绝整份更新并保留上一份有效配置。

阿里云需要已开通 DirectMail 和对应区域的发件身份，必须提供主题，当前不支持附件及 CC/BCC。AWS SES 需要当前区域的权限和已验证发件身份，沙箱还可能要求验证收件人。AWS 临时凭证可增加 `session_token`；使用默认凭证链时同时省略 `access_key_id` 与 `secret_access_key`，并在运行环境配置有效身份。

`email_blacklist` 是域名列表，同时拦截子域；`block_china_email` 仅按代码内列出的邮箱域名拦截，如 163.com、qq.com，不是按收件人所在地判断。不要将该域名黑名单理解为单个邮箱地址名单。

有效 Etcd 更新自动生效。无效 YAML、未知字段、不完整账号或读取失败保留上一有效配置；启动或监听重连时 key 缺失、运行中删除 key 均禁用邮件。恢复完整有效值后可继续发送。删除 key 与写入无效值的行为不同。

## 短信通道配置

请求与返回字段、路由和失败处理见 [SendSms](RPC.md#sendsms)。只配置指定厂商即可发送；省略请求 provider 时，必须另配含有效正权重通道的 Router。

不使用 Etcd 时，移除 `Message.SMS.Etcd` 或清空 Hosts，直接在 RPC YAML 配置的 `Message.SMS` 下填写账号。下面是单 BytePlus 示例，需合并到已有 Message，保留邮件配置：

```yaml
Message:
  SMS:
    Enabled: true
    TimeoutSeconds: 10
    RequestTimeoutSeconds: 30
    Byteplus:
      Username: "<BYTEPLUS_USERNAME>"
      Password: "<BYTEPLUS_PASSWORD>"
      Sender: "<APPROVED_SENDER>"
    Router:
      MaxAttempts: 1
      Channels:
        - Name: byteplus
          Weight: 100
```

RPC YAML 配置使用 PascalCase：`Submail.AppID` / `AppKey` / `Sender`、`Gotone.APIToken` / `SenderID` / `Type`、`AWS.Region` / `AccessKeyID` / `SecretAccessKey`、`Aliyun.Region` / `AccessKeyID` / `AccessKeySecret` 等。完整字段定义见 `rpc/internal/config/sms.go`。TimeoutSeconds 默认 10，允许 1–120 秒；RequestTimeoutSeconds 是整批预算上限，默认 30，允许 1–300 秒。存在 RPC 或 ctx deadline 时，再预留上游剩余时间的 10%（最多 1 秒）用于返回结果，因此实际提交预算可能更短。

使用 Etcd 时，RPC YAML 使用以下连接配置；账号存放在下一表所列的独立 key 中：

```yaml
Message:
  SMS:
    Enabled: true
    TimeoutSeconds: 10
    RequestTimeoutSeconds: 30
    Etcd:
      Hosts:
        - etcd.example.com:2379
      Prefix: /sms-service
      DialTimeout: 5
```

Prefix 默认为 `/sms-service`，实际读取位置为 `<Prefix>/providers/<名称>`。下面是每个 key 对应的 YAML value，可分别写入所需的 key，无需配置所有厂商。key 的名称 `ali` 与 value 的根字段 `ali_provider` 固定；请求中的别名 `aliyun` 不会把 Etcd key 改成 aliyun。

| Etcd key | value 根字段 |
| --- | --- |
| `/sms-service/providers/router` | router |
| `/sms-service/providers/byteplus` | byteplus_provider |
| `/sms-service/providers/submail` | submail_provider |
| `/sms-service/providers/gotone` | gotone_provider |
| `/sms-service/providers/aws` | aws_provider |
| `/sms-service/providers/ali` | ali_provider |

```yaml
# /sms-service/providers/router
router:
  max_attempts: 2
  channels:
    - name: byteplus
      weight: 100
    - name: submail
      weight: 50
```

```yaml
# /sms-service/providers/byteplus
byteplus_provider:
  username: "<BYTEPLUS_USERNAME>"
  password: "<BYTEPLUS_PASSWORD>"
  sender: "<APPROVED_SENDER>"
```

```yaml
# /sms-service/providers/submail
submail_provider:
  appid: "<SUBMAIL_APPID>"
  app_key: "<SUBMAIL_APP_KEY>"
  sender: "<APPROVED_SENDER>"
```

```yaml
# /sms-service/providers/gotone
gotone_provider:
  api_token: "<GOTONE_API_TOKEN>"
  sender_id: "<APPROVED_SENDER_ID>"
  type: plain
```

```yaml
# /sms-service/providers/aws
aws_provider:
  region: us-east-1
  access_key_id: "<AWS_ACCESS_KEY_ID>"
  secret_access_key: "<AWS_SECRET_ACCESS_KEY>"
  sender_id: "<APPROVED_SENDER_ID>"
  type: Transactional
  max_price: "0.01"
```

```yaml
# /sms-service/providers/ali
ali_provider:
  region: ap-southeast-1
  access_key_id: "<ALI_ACCESS_KEY_ID>"
  access_key_secret: "<ALI_ACCESS_KEY_SECRET>"
  sender: "<APPROVED_SENDER>"
  type: NOTIFY
```

AWS 使用 SNS，type 可选 Transactional / Promotional，默认 Transactional；临时凭证增加 session_token。使用默认凭证链时省略显式凭证，并设置 `use_default_credentials: true`，不会在未配置此通道时自动启用环境账号。上例 max_price 是可选的单条价格上限示例，可能因实际价格或目的地区导致拒绝，按账号情况调整或省略。阿里云使用国际 SendMessageToGlobe，type 可选 NOTIFY / OTP / PROMOTION，默认 NOTIFY。各平台仍需开通目的地区、资质、余额及获准的发件标识。

Submail 不提交 +86 号码，并会在去除正文首尾空白后自动补 `【logo】` 固定前缀，已有相同前缀不重复添加。BytePlus、Submail、Gotone 可配置 endpoint；默认分别为 `https://sms.byteplusapi.com/sms/openapi/send_sms`、`https://api-v4.mysubmail.com/internationalsms/send.json`、`https://app.gotonesms.net/api/v3/sms/send`。外部 endpoint 要求 HTTPS，证书正常校验；不会跟随 HTTP 重定向。

短信的 Etcd 读取是异步的，远端账号尚未加载时可能返回 Unavailable。有效更新自动生效；无效配置保留对应上一有效配置；删除某个 key 会清除对应厂商或 router，启动和 Watch 恢复重读时缺失的 key 也会清除对应 YAML 预设配置。删除 router 后，指定有效 provider 仍可使用；不填 provider 则没有路由可选。读取失败保留上一配置，Watch 中断后会重连并重新读取。Etcd value 不支持 enabled 总开关，停用短信需修改 RPC YAML 中的 `Message.SMS.Enabled` 并重启，或删除相应账号 key 停用通道。

integration RPC 启动时会为邮件和短信方法设置请求正文日志排除规则，普通访问日志和慢调用日志均不会记录验证码、邮件正文、密送地址或附件；其他业务方法的日志配置保持不变。该规则不会清理已有日志，调用方自身的客户端日志也需要单独保护。

在专用测试环境先验证单号码、指定通道，再验证批量与路由；accepted / sent 仅代表厂商接受，需同时检查手机和厂商记录。unknown 或客户端超时先核实结果，避免整批重发。自动化模拟不会发送真实邮件或收费短信，可在仓库根目录执行：

```powershell
go test ./rpc/internal/message/email ./rpc/internal/message/sms ./rpc/internal/logic/message ./rpc/service/message -count=1
```

模拟通过不等于真实账号投递通过，真实验证必须分别记录厂商接受和最终收信结果。

---



## 排错


| 现象                                      | 处理                                                   |
| --------------------------------------- | ---------------------------------------------------- |
| `open etc/integration.yaml`             | 在仓库根执行 `go run ./rpc` / `go run ./api`               |
| `database "integration" does not exist` | RPC 启动时会自动建库；检查实际使用的 `DB.DSN`、PostgreSQL 连接及账号的建库权限 |
| `password authentication failed`        | 核对数据库地址、账号和密码是否与部署环境一致                     |
| Etcd 连接失败 | 先区分 RPC 服务发现与 Message 下的账号配置；消息使用 Etcd 时启动对应实例并核对 Hosts。改用 YAML 账号配置时清空对应 Message 域的 Etcd.Hosts 并重启 |
| `ip geo not ready` / RPC 连不上            | 先起 `go run ./rpc`，API yaml `Endpoints` 对上 `ListenOn` |
| Redis 连接失败         | `CacheDriver: redis` 时先起 Redis；Host/Pass/DB 与实例一致   |
| mmdb / xdb 打开失败                         | 检查 `endpoint` 相对 RPC 工作目录是否存在                        |
| 查询 `source` 空、地理字段空                     | 无启用源，或三方都失败；看 RPC 日志 `ipgeo source`                  |
| 邮件或短信 `Unavailable` | 检查服务总开关、初始化日志、账号是否加载、指定通道是否存在；短信不填 provider 时还需有效 Router；邮件在途容量满也会返回该状态 |
| 邮件 `ResourceExhausted` | 检查 protobuf 整包 16 MiB 限制和调用方消息上限，计入 Base64 膨胀；不要只计算原附件大小 |
| 邮件或短信厂商已接受但未收到 | 检查垃圾箱或手机，并凭平台邮件 ID / 短信 message_id 核实厂商记录；accepted 不等于最终送达 |


