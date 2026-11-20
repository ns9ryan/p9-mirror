# 其它服务 gRPC

HTTP 管理后台见 [API.md](API.md)。业务进程请直连本服务 RPC，不要只走 HTTP 再绕一圈（可以，但多一跳）。

定义：改 `rpc/desc/*.proto`（`base.proto`、`ipgeo.proto`、`message.proto`），goctls 合并为 `rpc/integration.proto`。

| | |
| --- | --- |
| package | `integration` |
| service | `Integration` |
| 客户端 | `oa.98ent.com/p9/integration/rpc/integrationclient` |
| 服务地址 | 使用部署环境提供的 RPC 地址，示例为 `integration-rpc.example.com:19005` |

调用方使用已部署的 integration RPC 服务，文档中的 `integration-rpc.example.com:19005` 需替换为实际服务地址。配置与部署见 [OPS.md](OPS.md)。未启用服务反射时，使用 `rpc/integration.proto` 生成客户端或导入调用工具。

IP/GEO 租户：不要传 `operator_code`。取自 ctx claims / incoming metadata `x-operator-code`。`list` 不会是 `null`。查询顺序、缓存、私网行为与 HTTP 相同，见 [API.md](API.md) / [ip_search_pfd.md](ip_search_pfd.md)。邮件和短信目前使用服务级通道配置，不按该 metadata 选择租户账号。

### 方法一览

proto 方法名（grpcurl 用 `integration.Integration/<Name>`，Go client 为 PascalCase）：

| RPC | 请求 | 返回 | 说明 |
| --- | --- | --- | --- |
| `ping` | `Empty` | `Empty` | 健康检查（`base`） |
| `sendSms` | `SendSmsReq` | `SendSmsResp` | 国际短信、指定通道或加权路由；见 [SendSms](#sendsms) |
| `searchIp` | `SearchIpReq` | `SearchIpResp` | 批量查 IP 地理信息 |
| `listPolicies` | `Empty` | `PolicyListResp` | 当前租户策略（未配则用默认） |
| `savePolicies` | `SavePolicyReq` | `PolicyListResp` | 保存当前租户策略覆盖 |
| `saveDefaultPolicies` | `SavePolicyReq` | `PolicyListResp` | 写服务商默认启用/优先级 |
| `listLocalIps` | `LocalIPListReq` | `LocalIPListResp` | 本地 IP 分页列表 |
| `getLocalIp` | `IDReq` | `LocalIP` | 本地 IP 详情 |
| `createLocalIp` | `LocalIP` | `LocalIP` | 新增本地 IP |
| `updateLocalIp` | `LocalIP` | `LocalIP` | 更新本地 IP（必填 `id`） |
| `deleteLocalIp` | `IDReq` | `Empty` | 删除本地 IP |
| `sendEmail` | `SendEmailReq` | `SendEmailResp` | SMTP / 阿里云 DirectMail / AWS SES；见 [SendEmail](#sendemail) |

说明：本地 IP 相关方法名用 `LocalIp` 而非 `LocalIP`，避免 goctl 生成怪异文件名。

### 公共消息

| 消息 | 字段 | 说明 |
| --- | --- | --- |
| `Empty` | （无） | 空请求/空响应 |
| `IDReq` | `id` int64 | 主键 id |

## 消息域接入

邮件和短信由 `Integration` RPC 提供。调用工具选择 gRPC 请求，连接部署环境提供的 RPC 地址，方法分别选 `integration.Integration/sendEmail`、`integration.Integration/sendSms`。不支持反射时导入 `rpc/integration.proto`。下列 JSON 使用 proto 原字段名，部分客户端会显示为 lowerCamelCase。

其它 Go 服务使用当前版本的 `rpc/integrationclient`。按现有 go-zero 方式，将配置放入调用方 yaml：

```yaml
IntegrationRpc:
  Endpoints:
    - integration-rpc.example.com:19005
  NonBlock: true
  Timeout: 30000
```

调用方配置结构增加 `IntegrationRpc zrpc.RpcClientConf`，在 `ServiceContext` 中保存 `integrationclient.Integration`；启动时初始化一次、后续复用：

```go
// imports:
// "github.com/zeromicro/go-zero/zrpc"
// "oa.98ent.com/p9/integration/rpc/integrationclient"
cli := integrationclient.NewIntegration(zrpc.MustNewClient(c.IntegrationRpc))
```

业务方法使用调用方的 `ctx`，通过 `cli.SendEmail` / `cli.SendSms` 发送，字段示例见下文。`Timeout` 单位为毫秒，还需结合服务端 `Timeout` 和 ctx deadline 设置；调大短信业务预算时，也要检查这两端限制。短信会根据上游截止时间提前结束提交，预留剩余时间的 10%（最多 1 秒）用于汇总和返回逐号码结果。RPC 的连接配置 `IntegrationRpc.Etcd` 属于服务发现；`Message.Email.Etcd` / `Message.SMS.Etcd` 则用于读取通道账号，二者用途不同。其它机器或容器调用时，`Endpoints` 要改成实际可达的 RPC 地址。

调用方只传收件人和内容，账号由 integration 管理。`delivery_id` 不是幂等键，重复调用会产生新投递；目前没有按 ID 查询投递记录、最终送达或已读状态的 RPC。客户端超时或断网不能证明未发送，调用方不要对整次发送盲目重试。

## SendEmail

`sendEmail` 是服务内部 gRPC 能力。其它微服务应调用客户端的 `cli.SendEmail`，不直接访问邮件服务商，也不携带邮件账号密码。支持 SMTP、阿里云 DirectMail、AWS SES，提供多个收件人、主题、纯文本及 HTML 正文、Base64 附件、抄送和密送功能，各通道限制见下表。

RPC yaml 的 `Message.Email.Enabled` 是总开关，必须显式设置为 `true` 才启用邮件。`false` 时，即使账号或 Etcd 配置完整也不会发信。`Providers` 列表中的每个通道也必须显式设置 `Enabled: true`；仅填写 SMTP 账号不会自动启用通道。最小 YAML 配置示例：

```yaml
Message:
  Email:
    Enabled: true
    Provider:
      Host: smtp.example.test
      Port: "465"
      AuthEmail: sender@example.test
      AuthPassword: replace-with-smtp-password
      TLSMode: implicit
```

SMTP 配置可从 RPC yaml 的 `Message.Email.Provider` 读取；总开关启用且配置 `Message.Email.Etcd.Hosts` 后会从 Etcd 读取并监听 `Message.Email.Etcd.Key`，有效 Etcd 配置会覆盖 yaml 快照。代码默认 key 为 `/integration/message/email/providers/smtp`；当前配置模板显式使用 `/email-service/providers/smtp`，部署时应填写实际存在的 key。首次加载为异步，尚未获得有效账号时返回 `Unavailable`；启动或监听恢复时 key 缺失均禁用邮件，无效更新或读取失败保留上一份有效配置，删除配置会禁用邮件。单 SMTP 的 Etcd value 可使用以下格式：

```yaml
smtp_provider:
  host: smtp.example.test
  port: "465"
  auth_email: sender@example.test
  auth_password: smtp-password
email_blacklist:
  - blocked.example.test
```

也可使用 `email` 嵌套格式：

```yaml
email:
  smtp:
    host: smtp.example.test
    port: "465"
    auth_email: sender@example.test
    auth_password: smtp-password
    tls_mode: implicit
    tls_skip_verify: false
    timeout_seconds: 10
  blacklist:
    - blocked.example.test
```

Etcd 顶层或 `email.enabled` 可显式设置 `false` 禁用发送。单 SMTP 格式省略开关时，完整 SMTP 配置自动视为远端启用，但仍受服务总开关限制。多通道配置使用 `providers` 数组，每项包含 `name`、`type`、`enabled`、`priority` 及相应的 `smtp`、`aliyun` 或 `aws` 配置；名称必须唯一，优先级越小越先尝试。`provider` 请求参数优先匹配通道名称，其次匹配类型；指定名称时只使用该通道。

配置非空 `Providers` / `providers` 数组后，候选通道仅使用数组中的配置。需要使用 SMTP 时，将其作为一个通道写入数组；单通道 `Provider` / `smtp_provider` 不会自动追加。完整多通道示例见 [邮件通道配置](OPS.md#邮件通道配置)。

`tls_mode` 支持 `implicit` 和 `starttls`，默认端口分别为 `465` 和 `587`。默认验证 TLS 证书，`tls_skip_verify: true` 会跳过服务器证书校验，但仍使用 TLS 加密。SMTP 可通过 `proxy.type: socks5` 配置代理。收件人去除首尾空格、规范域名大小写，并在收件人、抄送、密送之间去重；最多 100 个唯一地址。命中域名黑名单或启用的国内邮箱规则时，整封邮件拒发。必须提供纯文本或 HTML 正文；主题最多 8 KiB，纯文本和 HTML 合计最多 4 MiB。附件 Base64 必须有效，解码后总大小不超过 10 MiB，文件名和 MIME 类型禁止控制字符。MIME 自动生成 Date、Message-ID，编码 Unicode 主题与附件名，密送地址只出现在发送信封中。

整个 gRPC 请求的 protobuf 消息上限为 16 MiB，正文、附件 Base64 字符串及其它字段合计都受此限制。Base64 通常比原文件大约三分之一，因此各字段单独未超限，组合后仍可能被 gRPC 以 `ResourceExhausted` 拒绝。调用方连接也可能有更小的消息限制。

| 通道 | 配置及限制 |
| --- | --- |
| SMTP | `host`、`auth_email`、`auth_password`；支持附件、HTML、抄送、密送 |
| 阿里云 | `region`、`access_key_id`、`access_key_secret`、`account_name`；使用对应区域的 HTTPS endpoint；必须有主题，不支持附件、抄送和密送 |
| AWS SES | `region`、`from`；可使用 AWS 默认凭证链或显式密钥对；单封最多 50 个收件人，最终 MIME 不超过 10,000,000 字节 |

未限定单个通道时，通道不能承接请求会继续尝试下一候选通道，例如超过 SES 限制的请求可交给 SMTP；不会拆分邮件或部分发送。全部候选通道都不支持时返回 `FailedPrecondition`。SMTP 任一收件人在 RCPT 阶段被拒绝时，不提交整封邮件，不会只向已接受的部分地址发送。

重试由 `retry.max_attempts`、`retry.backoff_ms`、`retry.max_backoff_ms` 控制，默认每通道最多 3 次。只自动重试明确未接受邮件的暂时失败；永久拒绝直接切换通道。SMTP 已确认接受邮件后，QUIT 失败仍返回成功。邮件提交后响应丢失或云请求发生不确定网络错误时，记录为 `unknown`，停止重试及通道切换，返回 `FailedPrecondition` 和 `EMAIL_DELIVERY_UNCERTAIN` 错误详情。调用方不得自动重发，应先核实服务商结果。

`accepted` 表示服务商接受的收件人数，不表示邮件已经进入收件箱。`delivery_id` 是本次调用的内部记录 ID；已生成 ID 的失败请求通过 gRPC `google.rpc.ErrorInfo` 的 `metadata.delivery_id` 返回。发送记录保留状态、尝试次数和独立的服务商邮件 ID，可通过内部 `Email.Records()` 获取；记录仅保存在进程内存，重启后清空，不作为持久化投递账本。`RecordLimit` 默认 100，分别限制近期完成记录数和同时在途请求数；在途记录不淘汰，容量满时提交前返回 `Unavailable`。没有邮件正文的 Redis 或 PostgreSQL 缓存。

### 请求 `SendEmailReq`

| 字段 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- |
| `email` | 条件必填 | repeated string | 主收件人；email、cc、bcc 合计至少一个地址 |
| `subject` | 否 | string | 主题，最多 8 KiB；阿里云通道要求非空 |
| `content` | 条件必填 | string | 纯文本正文；与 html_content 至少一个非空 |
| `html_content` | 条件必填 | string | HTML 正文；可与纯文本并存，二者合计最多 4 MiB |
| `attachment` | 否 | repeated EmailAttachment | 附件数组，字段名为单数；受通道能力及大小限制 |
| `cc` | 否 | repeated string | 抄送地址 |
| `bcc` | 否 | repeated string | 密送地址，不写入邮件头 |
| `provider` | 否 | string | 精确通道名称优先，其次按类型匹配；省略则按优先级尝试 |

`EmailAttachment`：

| 字段 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- |
| `filename` | 是 | string | 附件名，支持中文 |
| `base64` | 是 | string | 文件内容的标准 Base64 编码，不带 data URL 前缀；空文件可用空字符串 |
| `mime` | 是 | string | MIME 类型，例如 text/plain、application/pdf |

示例使用支持全部字段的 SMTP 通道；地址和通道名需替换为实际值。示例附件内容是 `Hello`：

```json
{
  "email": ["recipient@example.com"],
  "cc": ["copy@example.com"],
  "bcc": ["hidden@example.com"],
  "subject": "邮件功能测试",
  "content": "纯文本备用正文",
  "html_content": "<h2>邮件测试</h2><p>请查看附件。</p>",
  "attachment": [
    { "filename": "测试.txt", "base64": "SGVsbG8=", "mime": "text/plain" }
  ],
  "provider": "smtp"
}
```

### 返回 `SendEmailResp`

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `accepted` | int32 | 服务商接受的唯一收件人数，包含 To、CC、BCC |
| `message` | string | 成功为 success；失败使用 gRPC status |
| `delivery_id` | string | integration 内部投递记录 ID，不是服务商邮件 ID |
| `provider` | string | 实际使用的通道名称 |

成功响应示例（仅说明结构，不代表真实投递测试已通过）：

```json
{
  "accepted": 3,
  "message": "success",
  "delivery_id": "example-email-delivery-id",
  "provider": "smtp"
}
```

### Go 调用

复用上文初始化的 `cli`，在业务方法中调用：

```go
// 日志示例使用 "github.com/zeromicro/go-zero/core/logx"。
out, err := cli.SendEmail(ctx, &integrationclient.SendEmailReq{
    Email:       []string{"recipient@example.com"},
    Subject:     "邮件功能测试",
    Content:     "纯文本备用正文",
    HtmlContent: "<p>HTML 正文</p>",
    Attachment: []*integrationclient.EmailAttachment{
        {Filename: "测试.txt", Base64: "SGVsbG8=", Mime: "text/plain"},
    },
    Provider: "smtp",
})
if err != nil {
    // 交给业务层处理 gRPC status；不确定结果不得直接重发。
    return err
}
logx.Infof("email delivery_id=%s provider=%s accepted=%d",
    out.GetDeliveryId(), out.GetProvider(), out.GetAccepted())
```

如需识别不确定结果，可在错误分支检查详情：

```go
// imports:
// "google.golang.org/genproto/googleapis/rpc/errdetails"
// "google.golang.org/grpc/status"
for _, detail := range status.Convert(err).Details() {
    info, ok := detail.(*errdetails.ErrorInfo)
    if ok && info.GetReason() == "EMAIL_DELIVERY_UNCERTAIN" {
        deliveryID := info.GetMetadata()["delivery_id"]
        // 记录 deliveryID 并核实服务商结果，不直接重新投递。
        _ = deliveryID
    }
}
```

真实通道准备和验收见 [EMAIL_TESTING.md](EMAIL_TESTING.md)。支持通道不等于账号已开通或已完成真实收信测试。

## SendSms

`sendSms` 用于国际短信发送，Go 客户端方法为 `cli.SendSms`。服务端必须开启 `Message.SMS.Enabled: true`，并配置至少一个可用通道。账号、发件标识、地区权限和余额均由对应短信厂商管理。不会因为接口返回成功就保证手机最终收到短信。

### 请求 `SendSmsReq`

| 字段 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- |
| `phone` | 是 | repeated string | 输入 1–100 个国际号码，建议明确使用 +国际区号；任一非法号码使整批提交前拒绝，规范化后去重 |
| `content` | 是 | string | 非空 UTF-8 正文，最多 16000 字节；厂商限制、拆分计费规则继续生效 |
| `provider` | 否 | string | 指定厂商，仅走该通道；省略使用 router 加权选择 |

| provider | 通道 | 前置条件及限制 |
| --- | --- | --- |
| `byteplus` | BytePlus | 用户名、密码，按账号权限配置 sender 和目的地区 |
| `submail` | Submail 国际短信 | appid、app_key；当前实现不向 +86 号码提交；去除正文首尾空白后，如无 `【logo】` 前缀会自动添加该固定前缀 |
| `gotone` | Gotone | api_token、获准的 sender_id 和 type |
| `aws` | AWS SNS | 显式凭证或主动启用默认凭证链；区域、权限和短信资质有效 |
| `ali` / `aliyun` | 阿里云国际短信 | 国际短信凭证和权限；使用 SendMessageToGlobe |

请求示例，号码必须替换为自己持有且可接收测试短信的真实号码：

```json
{
  "phone": ["+12025550123"],
  "content": "Your verification code is 123456.",
  "provider": "byteplus"
}
```

指定通道无需 router；不填 provider 时，router 必须含已配置且权重大于零的通道，按权重随机生成候选顺序，不是邮件的 priority 排序。`router.max_attempts` 限制本次候选通道数量；小于等于零或超过候选数量时使用全部候选。每个号码在每个候选通道最多提交一次；只有明确失败才尝试下一候选，sent 和 unknown 都不再提交。

### 返回 `SendSmsResp`

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `error_code` | int32 | 0：全部厂商接受；2：全部或部分失败；3：至少一个号码投递不确定 |
| `msg` | string | 本次发送结果说明 |
| `accepted` | int32 | 厂商接受的唯一号码数，不表示最终送达数 |
| `delivery_id` | string | 本次调用标识，不是查询接口参数或幂等键 |
| `results` | repeated SmsRecipientResult | 每个规范化、去重后的号码的最终结果 |
| `errors` | repeated string | 仅最终 failed / unknown 的号码及原因，不包含 fallback 成功前的历史失败 |
| `data` | string | results 的 JSON 字符串；建议直接读取结构化字段 results |

`SmsRecipientResult`：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `phone` | string | 规范化后的 E.164 号码 |
| `provider` | string | 最后实际尝试的通道；未提交时可能为空 |
| `status` | string | sent：厂商已接受；failed：明确失败或尚未提交；unknown：已提交但无法确定结果 |
| `message_id` | string | 厂商确认 ID；部分通道明确成功时也可能为空 |
| `message` | string | 失败或不确定的原因 |
| `data` | string | 厂商返回的白名单附加信息，可能为空 |

成功响应示例（省略 data 字段；不是实际送达回执）：

```json
{
  "error_code": 0,
  "msg": "send completed",
  "accepted": 1,
  "delivery_id": "example-sms-delivery-id",
  "errors": [],
  "results": [
    {
      "phone": "+12025550123",
      "provider": "byteplus",
      "status": "sent",
      "message_id": "example-provider-message-id"
    }
  ]
}
```

发送开始后的厂商失败、部分成功或不确定，通常通过正常 gRPC 响应携带上述结果；**err 为 nil 后仍须检查 error_code 和 results**。非法请求、未配置等提交前错误使用 gRPC status，见下表。手机未收到时，用对应厂商的 `message_id` 查平台记录或联系账号管理员核实。

### Go 调用

```go
out, err := cli.SendSms(ctx, &integrationclient.SendSmsReq{
    Phone:    []string{"+12025550123"}, // 替换为真实测试号码。
    Content:  "Your verification code is 123456.",
    Provider: "byteplus",
})
if err != nil {
    return err
}
for _, item := range out.GetResults() {
    switch item.GetStatus() {
    case "sent":
        // 厂商已接受；可用 item.GetMessageId() 关联平台记录。
    case "failed":
        // 检查原因，按业务规则处理该号码，避免整批重发。
    case "unknown":
        // 先核实厂商结果，禁止自动重发该号码。
    }
}
```

单通道提交超时 `TimeoutSeconds` 默认 10 秒，整请求预算上限 `RequestTimeoutSeconds` 默认 30 秒；存在上游截止时间时，会进一步缩短预算，为返回部分结果留出时间。主动取消或网络中断时，仍可能无法收到响应，不能据此认定短信未发送。正文与逐号码结果不存入 Redis 或 PostgreSQL。integration RPC 的普通访问日志和慢调用日志均不记录邮件、短信请求正文；调用方应另行配置自身的日志保护。启动方式见 [OPS.md](OPS.md)；实测时先指定一个已配置的通道，向自己持有的号码发送，再同时核对响应与手机收信情况。

厂商账号、YAML 配置、Etcd key 与热更新规则见 [短信通道配置](OPS.md#短信通道配置)。

## 失败（gRPC status）

| 情况 | code | msg |
| --- | --- | --- |
| 超过 100 个 IP / 不支持的 lang | `InvalidArgument` | `max ips is 100` / `unsupported lang` |
| 策略列表空 / 超过 100 条 / 重复 id / 非法项 | `InvalidArgument` | `policy list is required` / `max policies is 100` / `duplicate provider_id` / `invalid provider` |
| 服务商不存在 | `NotFound` | `provider not found` |
| 本地 IP 不存在 | `NotFound` | `local ip not found` |
| ip 非法 / 已存在 | `InvalidArgument` | `ip is required` / `invalid ip` / `ip already exists` |
| 邮件未配置或 SMTP 不可用 | `Unavailable` | `email provider not configured` / `email provider unavailable` |
| 邮件收件人、正文、附件不合法或命中黑名单 | `InvalidArgument` | 具体校验错误 |
| 邮件 protobuf 整包超过服务端或调用方限制 | `ResourceExhausted` | 服务端上限 16 MiB，计入 Base64 编码后的附件和正文；可能在业务校验前拒绝 |
| 邮件通道限制、永久拒绝或投递结果不确定 | `FailedPrecondition` | 具体原因；不确定时附带 `EMAIL_DELIVERY_UNCERTAIN` 和内部发送记录 ID |
| 邮件请求被取消或超时 | `Canceled` / `DeadlineExceeded` | 已提交但无法确认结果时改为 `FailedPrecondition`，避免自动重发 |
| 短信号码、正文非法或 provider 不受支持 | `InvalidArgument` | 整批提交前拒绝，具体原因见错误消息 |
| 短信停用、指定通道缺少账号或 router 无可用通道 | `Unavailable` | `sms is not configured` 或具体配置原因 |
| 短信提交前请求被取消或超时 | `Canceled` / `DeadlineExceeded` | 发送开始后的结果以逐号码 results 为准；客户端自身断开时可能收不到响应 |
| 查询引擎未就绪 | `Unavailable` | `ip geo not ready` |

---

## SearchIp

合法公网按该租户策略选出启用源（没有对应策略则用 Provider 默认），按 `priority` 从小到大；每个源先读缓存（KV，键 `name|IP`），miss 再查源。无效 IP 仍占一位，只回 `query`。私网 / 回环：`country` 为 `私有IP`，`source` 为 `private`。单 IP 某源失败继续下一条。

### 请求 `SearchIpReq`

| 字段 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- |
| `ips` | 是 | `repeated string` | 待查 IP，最多 100 |
| `lang` | 否 | `string` | 本地化语言，空则 `en`；允许：`en` `zh-CN` `pt-BR` `ru` `fr` `ja` `es` `de` |

```json
{ "ips": ["8.8.8.8", "1.1.1.1"], "lang": "zh-CN" }
```

### 返回 `SearchIpResp`

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `list` | `repeated IpInfo` | 与 `ips` 等长；不会是 `null` |

### `IpInfo`

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `query` | string | 请求中的原始 IP |
| `continent` | string | 大洲名（随 `lang`） |
| `continent_code` | string | 大洲代码，如 `NA` |
| `country` | string | 国家名（随 `lang`） |
| `country_code` | string | 国家代码，如 `US` |
| `province` | string | 省/州 |
| `province_iso_code` | string | 省/州 ISO，如 `CA` |
| `city` | string | 城市 |
| `zip` | string | 邮编（与 `postal` 同源） |
| `lat` | double | 纬度 |
| `lon` | double | 经度 |
| `timezone` | string | 时区 |
| `isp` | string | ISP |
| `org` | string | 组织 |
| `as` | string | ASN 文本 |
| `postal` | string | 邮编 |
| `source` | string | 命中服务商 `name`，或 `private` |

---

## 策略

服务商连接配置来自 yaml 种子，RPC 不管理。每个租户可覆盖 `is_enabled`、`priority`。该租户没有策略行时，用 `ip_geo_provider` 上的默认启用/优先级。

`tag`：`在线`（`driver=http`）、`本地`（`driver=local`）、`离线`（`mmdb`、`ip2region`）。列表按有效 `priority`、`id` 升序，一次返回全部服务商。

### `PolicyView`（列表回显）

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `provider_id` | int64 | 全局服务商 id，保存时回传 |
| `name` | string | 服务商名称 |
| `is_enabled` | bool | 当前租户是否启用；未配策略时为默认 |
| `priority` | int32 | 当前租户优先级，越小越先 |
| `tag` | string | `在线` / `本地` / `离线` |

### `PolicyItem`（保存入参，不含 `name`/`tag`）

| 字段 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- |
| `provider_id` | 是 | int64 | 全局服务商 id |
| `is_enabled` | 是 | bool | 是否启用 |
| `priority` | 是 | int32 | 优先级，越小越先 |

### `SavePolicyReq` / `PolicyListResp`

| 消息 | 字段 | 说明 |
| --- | --- | --- |
| `SavePolicyReq` | `list` `repeated PolicyItem` | 不能为空；最多 100；同一 `provider_id` 不能重复 |
| `PolicyListResp` | `list` `repeated PolicyView`、`total` int64 | `list` 禁止 null |

| RPC | 行为 |
| --- | --- |
| `listPolicies` | 当前租户策略；未配则用默认 |
| `savePolicies` | 租户覆盖：不存在则创建，存在则更新；不改连接配置、不改其它租户 |
| `saveDefaultPolicies` | 写入 Provider 默认；已有租户策略的行不受影响；返回的是**默认**启用/优先级 |

```json
{
  "list": [
    { "provider_id": 1, "is_enabled": true, "priority": 0 },
    { "provider_id": 2, "is_enabled": false, "priority": 10 }
  ]
}
```

---

## 本地 IP

`driver=local` 的数据源。全局一份，不按租户隔离。是否启用、优先级看当前租户策略。表中没有该 IP 时，查询继续走后面的源。增删改会清掉该 IP 的查询缓存。

`page` 从 1；`page_size` 默认 20、最大 100；`page` / `page_size` ≤0 用默认，不报错。

### `LocalIP`

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int64 | 主键；创建时忽略，更新时必填 |
| `ip` | string | IPv4/IPv6，全局唯一；创建必填 |
| `continent` / `continent_code` | string | 大洲名 / 代码 |
| `country` / `country_code` | string | 国家名 / 代码 |
| `province` / `province_iso_code` | string | 省/州及 ISO |
| `city` | string | 城市 |
| `zip` / `postal` | string | 邮编，写入同一列；填其一即可 |
| `lat` / `lon` | double | 经纬度 |
| `timezone` / `isp` / `org` / `as` | string | 其它 |

### `LocalIPListReq` / `LocalIPListResp`

| 消息 | 字段 | 说明 |
| --- | --- | --- |
| `LocalIPListReq` | `page` int32、`page_size` int32、`ip` string | `ip` 模糊筛选，可空 |
| `LocalIPListResp` | `list` `repeated LocalIP`、`total` int64 | `list` 禁止 null |

| RPC | 请求示例 | 返回 |
| --- | --- | --- |
| `listLocalIps` | `{ "page": 1, "page_size": 20, "ip": "8.8" }` | `{ list, total }` |
| `getLocalIp` | `{ "id": 1 }` | `LocalIP` |
| `createLocalIp` | `{ "ip": "8.8.8.8", "country": "中国", "city": "深圳" }` | `LocalIP` |
| `updateLocalIp` | `LocalIP`（必填 `id`） | `LocalIP` |
| `deleteLocalIp` | `{ "id": 1 }` | `Empty` |

---

## Demo（grpcurl）

```powershell
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

PowerShell 会拆坏 JSON 引号，从 stdin 传入：

邮件与短信示例（在仓库根目录执行；必须先替换收件地址或号码，通道名以实际配置为准）：

```powershell
'{"email":["recipient@example.com"],"subject":"test","content":"Hello","provider":"smtp"}' | grpcurl -plaintext -d '@' `
  -import-path rpc -proto integration.proto `
  integration-rpc.example.com:19005 integration.Integration/sendEmail
```

```powershell
'{"phone":["+12025550123"],"content":"Your verification code is 123456.","provider":"byteplus"}' | grpcurl -plaintext -d '@' `
  -import-path rpc -proto integration.proto `
  integration-rpc.example.com:19005 integration.Integration/sendSms
```

`-plaintext` 表示示例中的 RPC 连接不启用 TLS，与邮件服务商连接所使用的 SMTP TLS 或短信 HTTPS 无关。实际 RPC 连接方式按部署配置调整。

IP/GEO 示例：

```powershell
cd integration
'{"ips":["8.8.8.8"],"lang":"zh-CN"}' | grpcurl -plaintext -d '@' `
  -import-path rpc -proto integration.proto `
  integration-rpc.example.com:19005 integration.Integration/SearchIp
```

```powershell
'{"id":1}' | grpcurl -plaintext -d '@' `
  -import-path rpc -proto integration.proto `
  integration-rpc.example.com:19005 integration.Integration/GetLocalIp
```

```powershell
'{}' | grpcurl -plaintext -d '@' `
  -import-path rpc -proto integration.proto `
  integration-rpc.example.com:19005 integration.Integration/ListPolicies
```

---

## Demo（Go）

其它服务在 `go.mod` 里：

```text
require oa.98ent.com/p9/integration v0.0.0
```

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/integration/rpc/integrationclient"

	"github.com/zeromicro/go-zero/zrpc"
)

func main() {
	cli := integrationclient.NewIntegration(zrpc.MustNewClient(zrpc.RpcClientConf{
		Endpoints: []string{"integration-rpc.example.com:19005"},
		NonBlock:  true,
		Timeout:   30000,
	}))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ctx = ctxdata.WithClaims(ctx, &ctxdata.Claims{OperatorCode: "op-xxx"})

	out, err := cli.SearchIp(ctx, &integrationclient.SearchIpReq{
		Ips:  []string{"8.8.8.8"},
		Lang: "zh-CN",
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, item := range out.GetList() {
		fmt.Printf("query=%s country=%s source=%s\n",
			item.GetQuery(), item.GetCountry(), item.GetSource())
	}
}
```
