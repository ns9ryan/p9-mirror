# 邮件实测步骤

请求字段和调用示例见 [SendEmail](RPC.md#sendemail)，账号及多通道配置见 [邮件通道配置](OPS.md#邮件通道配置)。

在 `integration` 模块目录运行自动化测试：

```powershell
go test ./...
go vet ./rpc/internal/message/email ./rpc/internal/logic/message ./rpc/service/message
```

测试使用 SMTP 模拟服务器和模拟云响应，不会发送真实邮件。覆盖 implicit TLS、STARTTLS、完整 SMTP 会话、抄送与密送、临时与永久拒绝、DATA 确认丢失、QUIT 失败、请求取消、云通道限制、Unicode MIME、配置更新及并发发送。

## 准备真实通道

1. 在当前 RPC 配置中设置 `Message.Email.Enabled: true`。保留 `false` 会关闭邮件，即使 Etcd 或账号完整。
2. 配置一个真实 SMTP `Provider`，或启用 Etcd 并写入对应 key；采用 `Providers` 数组时，每个需要使用的通道都设置 `Enabled: true`。
3. 按 [OPS.md](OPS.md) 部署并启动 RPC，调用工具连接部署环境提供的服务地址。
4. 如果只使用 Etcd，首次读取是异步的；账号加载完成前发送会返回 `Unavailable`。
5. 先发送到一个你可以查看的国外邮箱，再分别验证 Gmail、Outlook 等实际需要支持的邮箱。

## 检查实际邮件

| 场景 | 预期 |
| --- | --- |
| 纯文本邮件 | 主题、正文正确，收到一封 |
| HTML 邮件 | HTML 显示正常，可同时提供纯文本正文 |
| Unicode 主题和附件名 | 邮件标题、文件名正确，附件内容完整 |
| 多收件人、抄送、密送 | 各目标收到邮件；其他收件人看不到密送地址 |
| 域名黑名单 | 返回 `InvalidArgument`，整封拒发 |
| 总开关或通道关闭 | 总开关关闭返回 `Unavailable`；精确指定已关闭的通道也返回 `Unavailable`；自动选择时跳过关闭通道，仍可使用其它有效通道 |
| 修改有效 Etcd 配置 | 后续请求使用新配置 |
| 修改为无效 Etcd 配置 | 包括空值、未知字段、多份 YAML 文档；拒绝更新，继续使用上一有效配置 |
| 删除 Etcd key | 后续发送返回 `Unavailable` |

启用 Etcd 后，以指定 key 为配置来源：启动和监听恢复时 key 缺失都禁用邮件，
不继续使用 YAML 中预设的账号；读取失败或无效更新保留上一有效配置。
配置模板显式使用 `/email-service/providers/smtp`；未填写 Etcd.Key 时默认使用
`/integration/message/email/providers/smtp`。部署时应明确填写实际使用的 Etcd.Key，并确认该路径存在。
主题最多 8 KiB，纯文本和 HTML 正文合计最多 4 MiB，附件解码后合计最多 10 MiB。
整个 gRPC 请求的 protobuf 消息最多 16 MiB；附件 Base64 会膨胀，需同时计算正文及其它字段，整包超限返回 `ResourceExhausted`。
SMTP 任一收件人在 RCPT 阶段被拒绝时不提交整封邮件，返回错误；
永久拒绝不重试同一通道，但未指定单个通道时可切换到下一通道。

`SendEmailResp.Accepted` 只代表服务商接受了多少个收件人。测试时同时检查收件箱、垃圾邮件目录和服务商控制台，判断真实投递结果。

## 云通道

- 阿里云使用对应区域的 HTTPS endpoint；发件身份需符合服务商账号配置。必须提供主题，当前实现不支持附件、抄送和密送；候选列表还有其它通道时才会切换。精确指定阿里云通道名称时，不会切换，遇到这些限制返回 `FailedPrecondition`。
- AWS SES 需要当前区域的有效凭证和已验证发件身份；沙箱账号还受收件人验证限制。单封最多 50 个收件人，最终 MIME 大小不超过 10,000,000 字节。超限且候选列表还有其它通道时切换，否则返回 `FailedPrecondition`；精确指定 SES 通道名称时不会切换。
- 需要验证指定通道时，请求的 `Provider` 填通道名称。填写精确名称会限定该通道，便于分别测试。

## 失败和重试

- 明确未提交邮件的暂时故障会重试；永久拒绝不在同一通道重复尝试。
- SMTP 服务端已确认接受邮件后，QUIT 失败仍返回成功，不重发。
- 已提交但无法确认结果时返回 `FailedPrecondition`，错误详情 `google.rpc.ErrorInfo.reason` 为 `EMAIL_DELIVERY_UNCERTAIN`，`metadata` 包含 `delivery_id` 和 `provider`。先核实服务商记录，避免直接重新发送。
- 云厂商返回 HTTP 408、请求超时错误码或无法分类的错误时，结果保持不确定，不自动重试或切换通道。
- RPC 客户端自身的超时也可能发生在服务端完成发送之后；单凭超时不能判断邮件未发送。
- 内部 `Email.Records()` 可查看近期 `sent`、`failed`、`unknown` 状态和服务商邮件 ID；这些记录只保存在内存，重启后清空。RecordLimit 限制近期完成记录及同时在途请求数；在途记录不淘汰，容量不足时提交前返回 Unavailable。完成记录按完成顺序淘汰。
- 最终失败或不确定会记录 delivery_id、通道、状态和尝试次数的业务日志，不记录正文、收件人或账号密钥。
- 关闭邮件服务会取消进行中的请求，并拒绝后续发送；已提交但无法确认的请求仍保持不确定状态。
