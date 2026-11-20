package email

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"regexp"
	"strings"
	"time"

	"oa.98ent.com/p9/integration/rpc/internal/config"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/net/proxy"
)

// emailPattern 限定当前支持的 ASCII 邮箱格式，完整检查由 validEmail 执行。
var emailPattern = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

const maxAttachmentBytes = 10 << 20 // 全部附件解码后的总量上限：10 MiB
const maxRecipients = 100           // 主收件、抄送和密送去重后的总人数上限
const maxSubjectBytes = 8 << 10     // 主题上限：8 KiB
const maxBodyBytes = 4 << 20        // 文本与 HTML 正文合计上限：4 MiB

// chinaEmailDomains 仅在启用 BlockChinaEmail 时参与域名拦截。
var chinaEmailDomains = []string{
	"163.com", "126.com", "qq.com", "foxmail.com", "sina.com", "sina.cn", "sohu.com", "yeah.net",
	"139.com", "189.cn", "21cn.com", "aliyun.com", "tom.com", "263.net", "wo.cn", "china.com",
	"vip.163.com", "vip.126.com", "vip.qq.com", "outlook.cn", "hotmail.cn", "gmail.cn", "yandex.cn",
	"tencent.com",
}

// sendSMTP 完成 TLS 连接、认证、收件人确认和 MIME 数据提交，不在内部重试。
func sendSMTP(ctx context.Context, cfg config.EmailProvider, blacklist []string, blockChina bool, req Request) (accepted int, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	defer func() {
		if errors.Is(err, ErrProviderUnavailable) {
			logx.WithContext(ctx).Errorf("email smtp send failed host=%s port=%s: %v", cfg.Host, smtpPort(cfg), err)
		}
	}()
	// 在连接服务器前完成地址、正文和附件编码检查，失败时不提交邮件。
	clean, recipients, err := validateRequest(req, blacklist, blockChina)
	if err != nil {
		return 0, err
	}
	message, err := buildMessageWithOptions(cfg.AuthEmail, clean.Recipients, clean.CC, clean.Subject, clean.Content, clean.HTMLContent, clean.Attachments)
	if err != nil {
		return 0, err
	}
	// 单次发送默认限时 10 秒；上游更早取消时同步关闭 SMTP 连接。
	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = defaultSendTimeout
	}
	sendCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	// 先完成 TLS 连接，再使用配置中的邮箱和密码认证。
	client, err := newSMTPClient(sendCtx, cfg)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrProviderUnavailable, err)
	}
	defer client.Close()
	stop := context.AfterFunc(sendCtx, func() { _ = client.Close() })
	defer stop()
	if err := client.Auth(smtp.PlainAuth("", cfg.AuthEmail, cfg.AuthPassword, cfg.Host)); err != nil {
		return 0, smtpFailure("auth", err, false)
	}
	// 设置发送信封的发件地址，主收件、抄送和密送随后统一提交给 RCPT。
	if err := client.Mail(cfg.AuthEmail); err != nil {
		return 0, smtpFailure("mail", err, false)
	}
	// 全部 RCPT 通过后才进入 DATA；任一收件人被拒时不发送这封邮件正文。
	for _, recipient := range recipients {
		if err := client.Rcpt(recipient); err != nil {
			return 0, smtpFailure("recipient "+recipient, err, false)
		}
	}
	// 收件人全部确认后才提交 MIME 正文，正文内不包含密送地址。
	dataWriter, err := client.Data()
	if err != nil {
		return 0, smtpFailure("data", err, false)
	}
	if _, err = dataWriter.Write(message); err != nil {
		return 0, smtpFailure("write", err, false)
	}
	// Close 发送 DATA 结束标记并等待回执；此时断线可能已经受理，不能盲目重发。
	if err = dataWriter.Close(); err != nil {
		return 0, smtpFailure("close data", err, true)
	}
	// DATA 已确认受理，退出会话失败不再改变成功结果，避免诱发重复发送。
	_ = client.Quit()
	return len(recipients), nil
}

// smtpFailure 优先依据明确的 SMTP 回执分类；提交后的无回执错误归为 unknown。
func smtpFailure(stage string, err error, submitted bool) error {
	var reply *textproto.Error
	if errors.As(err, &reply) {
		// SMTP 5xx 表示明确拒绝，上层不在当前通道重试。
		if reply.Code >= 500 {
			return fmt.Errorf("%w: smtp %s: %v", ErrProviderRejected, stage, err)
		}
		// SMTP 4xx 表示暂时失败，上层可按退避策略重试。
		if reply.Code >= 400 {
			return fmt.Errorf("%w: smtp %s: %v", ErrProviderUnavailable, stage, err)
		}
	}
	// 已提交 DATA 结束标记却没有明确回执时，邮件可能已受理，不能重发。
	if submitted {
		return fmt.Errorf("%w: smtp %s: %v", ErrDeliveryUncertain, stage, err)
	}
	// 尚未完成 DATA 提交，归为可重试的通道不可用。
	return fmt.Errorf("%w: smtp %s: %v", ErrProviderUnavailable, stage, err)
}

// validateRequest 在连接厂商前校验所有地址和正文，并在 To、CC、BCC 之间统一去重。
// 返回规范化请求和完整信封收件人列表；任一地址不合法则整封拒绝。
func validateRequest(req Request, blacklist []string, blockChina bool) (Request, []string, error) {
	clean := req
	var err error
	// 主收件、抄送和密送使用相同地址规则，任一列表有非法或被拦截地址就拒绝。
	if clean.Recipients, err = validateRecipientList(req.Recipients, blacklist, blockChina); err != nil {
		return Request{}, nil, err
	}
	if clean.CC, err = validateRecipientList(req.CC, blacklist, blockChina); err != nil {
		return Request{}, nil, err
	}
	if clean.BCC, err = validateRecipientList(req.BCC, blacklist, blockChina); err != nil {
		return Request{}, nil, err
	}
	// 三个列表共用去重集合，重复地址依次保留在主收件、抄送或密送中。
	seen := make(map[string]bool)
	clean.Recipients = uniqueRecipients(clean.Recipients, seen)
	clean.CC = uniqueRecipients(clean.CC, seen)
	clean.BCC = uniqueRecipients(clean.BCC, seen)
	all := append(append(append([]string{}, clean.Recipients...), clean.CC...), clean.BCC...)
	// 按去重后的信封收件人数检查 1～100 人限制。
	if len(all) == 0 {
		return Request{}, nil, fmt.Errorf("%w: recipients are required", ErrInvalidRequest)
	}
	if len(all) > maxRecipients {
		return Request{}, nil, fmt.Errorf("%w: maximum %d recipients", ErrInvalidRequest, maxRecipients)
	}
	// 文本和 HTML 至少一项非空，主题最多 8 KiB，正文合计最多 4 MiB。
	if strings.TrimSpace(req.Content) == "" && strings.TrimSpace(req.HTMLContent) == "" {
		return Request{}, nil, fmt.Errorf("%w: text or html content is required", ErrInvalidRequest)
	}
	if len(req.Subject) > maxSubjectBytes || len(req.Content) > maxBodyBytes || len(req.HTMLContent) > maxBodyBytes-len(req.Content) {
		return Request{}, nil, fmt.Errorf("%w: subject exceeds 8 KiB or combined text and HTML content exceeds 4 MiB", ErrInvalidRequest)
	}
	// 主题不能包含换行等控制字符，防止插入额外邮件头。
	if containsHeaderControl(req.Subject) {
		return Request{}, nil, fmt.Errorf("%w: subject contains invalid characters", ErrInvalidRequest)
	}
	return clean, all, nil
}

// uniqueRecipients 仅统一域名大小写，保留邮箱本地部分的大小写语义。
func uniqueRecipients(input []string, seen map[string]bool) []string {
	out := make([]string, 0, len(input))
	for _, recipient := range input {
		local, domain, _ := strings.Cut(recipient, "@")
		recipient = local + "@" + strings.ToLower(domain)
		if !seen[recipient] {
			seen[recipient] = true
			out = append(out, recipient)
		}
	}
	return out
}

// validateRecipients 保留仅校验地址、黑名单和输入人数的兼容入口，不启用内置域名拦截。
func validateRecipients(input, blacklist []string) ([]string, error) {
	if len(input) > maxRecipients {
		return nil, fmt.Errorf("%w: maximum %d recipients", ErrInvalidRequest, maxRecipients)
	}
	return validateRecipientList(input, blacklist, false)
}

// validateRecipientList 校验地址格式及域名黑名单，不查询邮箱是否真实存在。
func validateRecipientList(input, blacklist []string, blockChina bool) ([]string, error) {
	out := make([]string, 0, len(input))
	for _, recipient := range input {
		recipient = strings.TrimSpace(recipient)
		// 先检查地址格式，再按域名黑名单和可选内置名单判断是否拦截。
		if !validEmail(recipient) {
			return nil, fmt.Errorf("%w: %s", ErrInvalidRecipient, recipient)
		}
		if blacklisted(recipient, blacklist) || (blockChina && isChinaEmail(recipient)) {
			return nil, fmt.Errorf("%w: %s", ErrRecipientBlocked, recipient)
		}
		out = append(out, recipient)
	}
	return out, nil
}

// isChinaEmail 仅匹配内置域名及其子域名，不按用户所在地判断。
func isChinaEmail(value string) bool {
	parts := strings.Split(value, "@")
	if len(parts) != 2 {
		return false
	}
	domain := strings.ToLower(strings.TrimSpace(parts[1]))
	for _, blocked := range chinaEmailDomains {
		if domain == blocked || strings.HasSuffix(domain, "."+blocked) {
			return true
		}
	}
	return false
}

// validEmail 检查总长度、本地部分和域名标签，不接受显示名，也不验证邮箱是否存在。
func validEmail(value string) bool {
	if len(value) > 254 || value == "" || !emailPattern.MatchString(value) {
		return false
	}
	local, domain, _ := strings.Cut(value, "@")
	// 本地部分最多 64 字节，每个域名标签最多 63 字节且不能以连字符开头或结尾。
	if len(local) > 64 {
		return false
	}
	for _, label := range strings.Split(domain, ".") {
		if label == "" || len(label) > 63 || strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			return false
		}
	}
	// 再用标准库解析，要求解析结果与原地址一致，拒绝显示名等额外内容。
	address, err := mail.ParseAddress(value)
	return err == nil && address.Address == value
}

// blacklisted 匹配黑名单域名及其子域名；无效地址按拦截处理。
func blacklisted(value string, blacklist []string) bool {
	parts := strings.Split(value, "@")
	if len(parts) != 2 {
		return true
	}
	domain := strings.ToLower(strings.TrimSpace(parts[1]))
	for _, blocked := range blacklist {
		if domain == blocked || strings.HasSuffix(domain, "."+blocked) {
			return true
		}
	}
	return false
}

// newSMTPClient 支持连接即 TLS 和 STARTTLS 升级；STARTTLS 不可用时直接失败。
func newSMTPClient(ctx context.Context, cfg config.EmailProvider) (*smtp.Client, error) {
	mode := strings.ToLower(strings.TrimSpace(cfg.TLSMode))
	if mode == "" {
		mode = "implicit" // 未指定时从连接开始就启用 TLS
	}
	host := strings.TrimSpace(cfg.Host)
	port := smtpPort(cfg)
	address := net.JoinHostPort(host, port)
	tlsConfig := &tls.Config{
		ServerName:         host,              // 用于 SNI 和服务器证书校验
		InsecureSkipVerify: cfg.TLSSkipVerify, // 由配置决定是否跳过证书校验
	}
	dial := func() (net.Conn, error) { return dialSMTP(ctx, cfg, address) }
	switch mode {
	case "implicit":
		// TCP 拨号后立即执行 TLS 握手，握手失败时关闭连接。
		conn, err := dial()
		if err != nil {
			return nil, err
		}
		stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
		defer stop()
		tlsConn := tls.Client(conn, tlsConfig)
		if err := tlsConn.HandshakeContext(ctx); err != nil {
			_ = conn.Close()
			return nil, err
		}
		// 将请求截止时间绑定到连接，使后续 SMTP 读写也受超时限制。
		if deadline, ok := ctx.Deadline(); ok {
			_ = tlsConn.SetDeadline(deadline)
		}
		client, err := smtp.NewClient(tlsConn, host)
		if err != nil {
			_ = tlsConn.Close()
		}
		return client, err
	case "starttls":
		// 先建立 SMTP 会话，再要求服务器升级 TLS；不允许退回明文认证。
		conn, err := dial()
		if err != nil {
			return nil, err
		}
		stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
		defer stop()
		if deadline, ok := ctx.Deadline(); ok {
			_ = conn.SetDeadline(deadline)
		}
		client, err := smtp.NewClient(conn, host)
		if err != nil {
			_ = conn.Close()
			return nil, err
		}
		// 不支持 STARTTLS 或升级失败均关闭连接，不继续提交认证信息。
		ok, _ := client.Extension("STARTTLS")
		if !ok {
			_ = client.Close()
			return nil, fmt.Errorf("smtp server does not support STARTTLS")
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			_ = client.Close()
			return nil, err
		}
		return client, nil
	default:
		return nil, fmt.Errorf("unsupported tls_mode %q", cfg.TLSMode)
	}
}

// dialSMTP 选择直连或 SOCKS5 代理，连接过程必须支持请求取消。
func dialSMTP(ctx context.Context, cfg config.EmailProvider, address string) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: timeoutOf(cfg)}
	// 未配置代理时直连，代理与直连都继承请求取消信号。
	if strings.TrimSpace(cfg.Proxy.Address) == "" {
		return dialer.DialContext(ctx, "tcp", address)
	}
	proxyDialer, err := newSMTPProxyDialer(cfg.Proxy, dialer)
	if err != nil {
		return nil, err
	}
	// 仅接受支持 Context 的代理拨号器，避免取消后仍持续等待连接。
	if contextDialer, ok := proxyDialer.(proxy.ContextDialer); ok {
		return contextDialer.DialContext(ctx, "tcp", address)
	}
	return nil, fmt.Errorf("smtp proxy does not support context cancellation")
}

// newSMTPProxyDialer 构建 SOCKS5 代理拨号器，填写凭证时启用代理认证。
func newSMTPProxyDialer(cfg config.EmailProxyConfig, forward proxy.Dialer) (proxy.Dialer, error) {
	if !strings.EqualFold(strings.TrimSpace(cfg.Type), "socks5") {
		return nil, fmt.Errorf("unsupported smtp proxy type %q", cfg.Type)
	}
	if strings.TrimSpace(cfg.Address) == "" {
		return nil, fmt.Errorf("smtp proxy address is required")
	}
	var auth *proxy.Auth
	// 任一代理凭证非空时传递认证参数，不复用 SMTP 邮箱密码。
	if cfg.Username != "" || cfg.Password != "" {
		auth = &proxy.Auth{User: cfg.Username, Password: cfg.Password}
	}
	return proxy.SOCKS5("tcp", cfg.Address, auth, forward)
}

// smtpPort 优先使用显式端口，否则 STARTTLS 使用 587，其余使用 465。
func smtpPort(cfg config.EmailProvider) string {
	if strings.TrimSpace(cfg.Port) != "" {
		return strings.TrimSpace(cfg.Port)
	}
	if strings.EqualFold(strings.TrimSpace(cfg.TLSMode), "starttls") {
		return defaultStartTLSPort
	}
	return defaultImplicitPort
}

// timeoutOf 返回 SMTP 拨号超时，未配置时使用 5 秒；发送整体另有超时限制。
func timeoutOf(cfg config.EmailProvider) time.Duration {
	if cfg.TimeoutSeconds > 0 {
		return time.Duration(cfg.TimeoutSeconds) * time.Second
	}
	return defaultDialTimeout
}

// buildMessage 保留纯文本加附件的 MIME 构建入口，不处理抄送和 HTML。
func buildMessage(from string, recipients []string, subject, body string, attachments []Attachment) ([]byte, error) {
	return buildMessageWithOptions(from, recipients, nil, subject, body, "", attachments)
}

// buildMessageWithOptions 构建 MIME 邮件；BCC 只进入发送信封，不写入邮件头。
// 文本和 HTML 使用 alternative，有附件时再套一层 mixed。
func buildMessageWithOptions(from string, to, cc []string, subject, textBody, htmlBody string, attachments []Attachment) ([]byte, error) {
	// 检查将写入邮件头的地址和主题，拒绝非法格式及控制字符。
	if !validEmail(from) {
		return nil, fmt.Errorf("%w: sender is invalid", ErrInvalidRequest)
	}
	if containsHeaderControl(subject) {
		return nil, fmt.Errorf("%w: subject contains invalid characters", ErrInvalidRequest)
	}
	for _, recipient := range append(append([]string(nil), to...), cc...) {
		if !validEmail(recipient) {
			return nil, fmt.Errorf("%w: %s", ErrInvalidRecipient, recipient)
		}
	}
	var buf bytes.Buffer
	// 只写 From、To 和 Cc；密送地址仅保留在发送信封中。
	if _, err := fmt.Fprintf(&buf, "From: %s\r\n", from); err != nil {
		return nil, err
	}
	if len(to) > 0 {
		writeAddressHeader(&buf, "To", to)
	}
	if len(cc) > 0 {
		writeAddressHeader(&buf, "Cc", cc)
	}
	if err := writeSubjectHeader(&buf, subject); err != nil {
		return nil, err
	}
	// 补齐发送时间与随机 Message-ID；此编号不是服务内存投递记录的 ID。
	fmt.Fprintf(&buf, "Date: %s\r\n", time.Now().Format(time.RFC1123Z))
	_, domain, _ := strings.Cut(from, "@")
	fmt.Fprintf(&buf, "Message-ID: <%s@%s>\r\n", rand.Text(), domain)
	buf.WriteString("MIME-Version: 1.0\r\n")
	// 纯文本无附件时直接编码正文，无需创建多段 MIME 容器。
	if len(attachments) == 0 && htmlBody == "" {
		// quoted-printable 编码 UTF-8 正文，处理特殊字节和邮件行长。
		buf.WriteString("Content-Type: text/plain; charset=UTF-8\r\nContent-Transfer-Encoding: quoted-printable\r\n\r\n")
		q := quotedprintable.NewWriter(&buf)
		if _, err := q.Write([]byte(textBody)); err != nil {
			return nil, err
		}
		if err := q.Close(); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}
	// 带 HTML 但无附件时，文本与 HTML 放在同一个 alternative 容器。
	if len(attachments) == 0 {
		alternative, boundary, err := buildAlternative(textBody, htmlBody)
		if err != nil {
			return nil, err
		}
		fmt.Fprintf(&buf, "Content-Type: multipart/alternative; boundary=%q\r\n\r\n", boundary)
		buf.Write(alternative)
		return buf.Bytes(), nil
	}
	// 有附件时使用 mixed 外层容器，正文作为首段，附件逐项追加。
	root := multipart.NewWriter(&buf)
	fmt.Fprintf(&buf, "Content-Type: multipart/mixed; boundary=%q\r\n\r\n", root.Boundary())
	if len(attachments) > 0 {
		bodyHeader := make(textproto.MIMEHeader)
		if htmlBody != "" {
			// 文本和 HTML 先组成 alternative，再作为 mixed 的第一部分写入。
			alternative, boundary, err := buildAlternative(textBody, htmlBody)
			if err != nil {
				return nil, err
			}
			bodyHeader.Set("Content-Type", fmt.Sprintf("multipart/alternative; boundary=%q", boundary))
			part, err := root.CreatePart(bodyHeader)
			if err != nil {
				return nil, err
			}
			if _, err := part.Write(alternative); err != nil {
				return nil, err
			}
		} else {
			// 无 HTML 时，首段直接写入经过 quoted-printable 编码的纯文本。
			bodyHeader.Set("Content-Type", "text/plain; charset=UTF-8")
			bodyHeader.Set("Content-Transfer-Encoding", "quoted-printable")
			part, err := root.CreatePart(bodyHeader)
			if err != nil {
				return nil, err
			}
			q := quotedprintable.NewWriter(part)
			if _, err := q.Write([]byte(textBody)); err != nil {
				return nil, err
			}
			if err := q.Close(); err != nil {
				return nil, err
			}
		}
	}
	// 同时限制 Base64 输入和解码后附件总量，避免超大附件占用过多内存。
	total := 0
	for _, a := range attachments {
		// 文件名和媒体类型写入 MIME 头之前，先拒绝空值和控制字符。
		if a.Filename == "" || a.MIME == "" || containsHeaderControl(a.Filename) || containsHeaderControl(a.MIME) {
			return nil, fmt.Errorf("%w: filename and valid mime are required", ErrInvalidAttachment)
		}
		mediaType, _, err := mime.ParseMediaType(a.MIME)
		if err != nil || mediaType == "" {
			return nil, fmt.Errorf("%w: invalid mime type", ErrInvalidAttachment)
		}
		// 先限制编码输入长度，再解码累计大小，全部附件合计不超过 10 MiB。
		if len(a.Base64) > base64.StdEncoding.EncodedLen(maxAttachmentBytes)+4 {
			return nil, fmt.Errorf("%w: total attachment size exceeds %d bytes", ErrInvalidAttachment, maxAttachmentBytes)
		}
		decoded, err := base64.StdEncoding.DecodeString(a.Base64)
		if err != nil {
			return nil, fmt.Errorf("%w: base64 decode: %v", ErrInvalidAttachment, err)
		}
		total += len(decoded)
		if total > maxAttachmentBytes {
			return nil, fmt.Errorf("%w: total attachment size exceeds %d bytes", ErrInvalidAttachment, maxAttachmentBytes)
		}
		// 用标准 MIME 参数编码文件名，附件单独标记为 attachment。
		h := make(textproto.MIMEHeader)
		h.Set("Content-Type", mime.FormatMediaType(mediaType, map[string]string{"name": a.Filename}))
		h.Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": a.Filename}))
		h.Set("Content-Transfer-Encoding", "base64")
		p, err := root.CreatePart(h)
		if err != nil {
			return nil, err
		}
		// 附件重新编码为标准 Base64，每行最多 76 字符。
		encoded := make([]byte, base64.StdEncoding.EncodedLen(len(decoded)))
		base64.StdEncoding.Encode(encoded, decoded)
		for i := 0; i < len(encoded); i += 76 {
			end := i + 76
			if end > len(encoded) {
				end = len(encoded)
			}
			fmt.Fprintf(p, "%s\r\n", encoded[i:end])
		}
	}
	// 写入 mixed 容器的结束边界，成功后才返回完整邮件内容。
	if err := root.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// buildAlternative 将文本和 HTML 编码为同一封邮件的两种正文格式。
func buildAlternative(textBody, htmlBody string) ([]byte, string, error) {
	var nested bytes.Buffer
	w := multipart.NewWriter(&nested)
	// 文本段先写入，供不支持 HTML 的邮件客户端显示。
	textHeader := make(textproto.MIMEHeader)
	textHeader.Set("Content-Type", "text/plain; charset=UTF-8")
	textHeader.Set("Content-Transfer-Encoding", "quoted-printable")
	p, err := w.CreatePart(textHeader)
	if err != nil {
		return nil, "", err
	}
	q := quotedprintable.NewWriter(p)
	if _, err = q.Write([]byte(textBody)); err != nil {
		return nil, "", err
	}
	if err = q.Close(); err != nil {
		return nil, "", err
	}
	// HTML 段与文本段表示同一封邮件，客户端选择适合的格式展示。
	htmlHeader := make(textproto.MIMEHeader)
	htmlHeader.Set("Content-Type", "text/html; charset=UTF-8")
	htmlHeader.Set("Content-Transfer-Encoding", "quoted-printable")
	p, err = w.CreatePart(htmlHeader)
	if err != nil {
		return nil, "", err
	}
	q = quotedprintable.NewWriter(p)
	if _, err = q.Write([]byte(htmlBody)); err != nil {
		return nil, "", err
	}
	if err = q.Close(); err != nil {
		return nil, "", err
	}
	if err = w.Close(); err != nil {
		return nil, "", err
	}
	// 返回正文和边界值，由外层写入 multipart/alternative 的 Content-Type。
	return nested.Bytes(), w.Boundary(), nil
}

// writeAddressHeader 将非空地址列表写入 To 或 Cc 头，并按地址折行。
func writeAddressHeader(buf *bytes.Buffer, name string, addresses []string) {
	fmt.Fprintf(buf, "%s: %s", name, addresses[0])
	for _, address := range addresses[1:] {
		fmt.Fprintf(buf, ",\r\n %s", address)
	}
	buf.WriteString("\r\n")
}

// writeSubjectHeader 编码非 ASCII 主题并折行，避免生成超过 MIME 行长限制的头部。
func writeSubjectHeader(buf *bytes.Buffer, subject string) error {
	// 非 ASCII 主题先编码；常规折行按 78 字符控制，单行硬上限为 998 字节。
	encoded := mime.QEncoding.Encode("UTF-8", subject)
	lineLength := len("Subject: ")
	buf.WriteString("Subject: ")
	for index, word := range strings.Split(encoded, " ") {
		if len(word)+1 > 998 {
			return fmt.Errorf("%w: subject exceeds MIME line length", ErrInvalidRequest)
		}
		if index > 0 {
			if lineLength+1+len(word) > 78 {
				buf.WriteString("\r\n ")
				lineLength = 1
			} else {
				buf.WriteByte(' ')
				lineLength++
			}
		}
		if lineLength+len(word) > 998 {
			buf.WriteString("\r\n ")
			lineLength = 1
		}
		buf.WriteString(word)
		lineLength += len(word)
	}
	buf.WriteString("\r\n")
	return nil
}

// containsHeaderControl 检查控制字符，防止主题或附件字段注入额外邮件头。
func containsHeaderControl(value string) bool {
	return strings.IndexFunc(value, func(r rune) bool { return r < 0x20 || r == 0x7f }) >= 0
}
