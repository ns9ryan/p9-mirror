package email

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"

	"oa.98ent.com/p9/integration/rpc/internal/config"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	aliyunerrors "github.com/aliyun/alibaba-cloud-sdk-go/sdk/errors"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dm"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	awscredentials "github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/aws/smithy-go"
)

// sendAliyun 使用阿里云 DirectMail 的 SingleSendMail 接口提交文本或 HTML 邮件。
func sendAliyun(ctx context.Context, p config.EmailProviderConfig, blacklist []string, blockChina bool, req Request) (SendResult, error) {
	cfg := p.Aliyun
	if err := validateAliyunConfig(cfg); err != nil {
		return SendResult{}, fmt.Errorf("%w: %v", ErrProviderRejected, err)
	}
	clean, _, err := validateRequest(req, blacklist, blockChina)
	if err != nil {
		return SendResult{}, err
	}
	request, err := buildAliyunRequest(cfg, clean)
	if err != nil {
		return SendResult{}, err
	}
	sendCtx, cancel := context.WithTimeout(ctx, defaultSendTimeout)
	defer cancel()
	if err := sendCtx.Err(); err != nil {
		return SendResult{}, err
	}
	return sendAliyunRequest(sendCtx, cfg, request, p.Name, len(clean.Recipients), http.DefaultTransport)
}

// sendAliyunRequest 提交一次请求，成功回执缺少编号时仍按结果不确定处理。
func sendAliyunRequest(ctx context.Context, cfg config.AliyunEmailConfig, request *dm.SingleSendMailRequest, provider string, accepted int, transport http.RoundTripper) (SendResult, error) {
	if err := ctx.Err(); err != nil {
		return SendResult{}, err
	}
	// 禁用 SDK 自动重试，由业务层根据是否明确受理决定是否再次提交。
	clientConfig := sdk.NewConfig().WithTimeout(defaultSendTimeout).WithAutoRetry(false)
	clientConfig.Scheme = "HTTPS"
	clientConfig.Transport = contextTransport{ctx: ctx, base: transport}
	client, err := dm.NewClientWithOptions(cfg.Region, clientConfig, credentials.NewAccessKeyCredential(cfg.AccessKeyID, cfg.AccessKeySecret))
	if err != nil {
		return SendResult{}, fmt.Errorf("%w: aliyun client: %v", ErrProviderRejected, err)
	}
	response, err := client.SingleSendMail(request)
	if err != nil {
		return SendResult{}, cloudFailure("aliyun", err)
	}
	if response == nil || !response.IsSuccess() {
		return SendResult{}, fmt.Errorf("%w: aliyun returned no successful response", ErrDeliveryUncertain)
	}
	messageID := response.EnvId
	if messageID == "" {
		messageID = response.RequestId
	}
	if messageID == "" {
		return SendResult{}, fmt.Errorf("%w: aliyun returned no acknowledgement ID", ErrDeliveryUncertain)
	}
	return SendResult{
		Accepted:          accepted,  // 厂商受理的主收件人数
		Provider:          provider,  // 本次通道名称
		ProviderMessageID: messageID, // 优先取 EnvId，否则取 RequestId
	}, nil
}

// contextTransport 将业务请求的取消和超时传递给未直接接收 Context 的 SDK。
type contextTransport struct {
	ctx  context.Context   // 本次发送的取消和超时信号
	base http.RoundTripper // 执行底层 HTTP 请求
}

// RoundTrip 为 SDK 请求绑定业务上下文，并阻止带签名的请求自动跳转。
func (transport contextTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := transport.base.RoundTrip(request.WithContext(transport.ctx))
	if err == nil && response != nil && response.StatusCode >= 300 && response.StatusCode < 400 {
		// SDK 的 HTTP 客户端会自动跳转；在此拦截，避免再次提交及泄露签名请求。
		// 首次请求的受理状态无法确认，交由上层按 unknown 处理，禁止切换重发。
		if response.Body != nil {
			_ = response.Body.Close()
		}
		return nil, errors.New("aliyun redirect response blocked")
	}
	return response, err
}

// buildAliyunRequest 在提交前拒绝该接口不支持的抄送、密送和附件，避免静默丢字段。
func buildAliyunRequest(cfg config.AliyunEmailConfig, req Request) (*dm.SingleSendMailRequest, error) {
	if err := validateAliyunConfig(cfg); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrProviderRejected, err)
	}
	if len(req.CC) > 0 || len(req.BCC) > 0 || len(req.Attachments) > 0 {
		return nil, fmt.Errorf("%w: aliyun SingleSendMail does not support CC/BCC or attachments", ErrProviderUnsupported)
	}
	if strings.TrimSpace(req.Subject) == "" {
		return nil, fmt.Errorf("%w: aliyun requires a subject", ErrProviderUnsupported)
	}
	request := dm.CreateSingleSendMailRequest()
	request.Scheme = "HTTPS"
	request.Domain = "dm." + cfg.Region + ".aliyuncs.com"
	if cfg.Region == "cn-hangzhou" {
		request.Domain = "dm.aliyuncs.com"
	}
	request.AccountName = cfg.AccountName                 // 已配置的发信地址
	request.AddressType = "1"                             // 使用发信地址类型
	request.ReplyToAddress = "true"                       // 使用控制台配置的回信地址
	request.ToAddress = strings.Join(req.Recipients, ",") // 厂商要求逗号分隔的收件人
	request.Subject = req.Subject                         // 邮件主题
	request.TextBody = req.Content                        // 纯文本正文
	request.HtmlBody = req.HTMLContent                    // HTML 正文
	request.FromAlias = cfg.FromAlias                     // 发件人显示名
	return request, nil
}

// sesRawEmailAPI 表示提交 MIME 邮件所需的 SES 能力。
type sesRawEmailAPI interface {
	SendRawEmail(context.Context, *ses.SendRawEmailInput, ...func(*ses.Options)) (*ses.SendRawEmailOutput, error)
}

// sendAWS 通过 SES SendRawEmail 发送完整 MIME 邮件，复用 SMTP 的正文和附件编码。
func sendAWS(ctx context.Context, p config.EmailProviderConfig, blacklist []string, blockChina bool, req Request) (SendResult, error) {
	cfg := p.AWS
	if err := validateAWSConfig(cfg); err != nil {
		return SendResult{}, fmt.Errorf("%w: %v", ErrProviderRejected, err)
	}
	clean, recipients, err := validateRequest(req, blacklist, blockChina)
	if err != nil {
		return SendResult{}, err
	}
	input, err := buildAWSInput(cfg.From, clean, recipients)
	if err != nil {
		return SendResult{}, err
	}
	sendCtx, cancel := context.WithTimeout(ctx, defaultSendTimeout)
	defer cancel()
	// MaxAttempts 包含首次调用，设为 1 表示 SDK 不再额外重试。
	options := []func(*awsconfig.LoadOptions) error{awsconfig.WithRegion(cfg.Region), awsconfig.WithRetryMaxAttempts(1)}
	if cfg.AccessKeyID != "" {
		options = append(options, awsconfig.WithCredentialsProvider(awscredentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, cfg.SessionToken)))
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(sendCtx, options...)
	if err != nil {
		return SendResult{}, fmt.Errorf("%w: aws config: %v", ErrProviderUnavailable, err)
	}
	return sendAWSConfiguredInput(sendCtx, awsCfg, p.Name, input)
}

// sendAWSConfiguredInput 在提交前取得凭证；凭证加载失败时可安全切换通道。
func sendAWSConfiguredInput(ctx context.Context, cfg aws.Config, provider string, input *ses.SendRawEmailInput) (SendResult, error) {
	if err := ctx.Err(); err != nil {
		return SendResult{}, err
	}
	if cfg.Credentials == nil {
		return SendResult{}, fmt.Errorf("%w: aws credentials unavailable before submission", ErrProviderUnavailable)
	}
	retrieved, err := cfg.Credentials.Retrieve(ctx)
	if err != nil || !retrieved.HasKeys() {
		if err := ctx.Err(); err != nil {
			return SendResult{}, err
		}
		return SendResult{}, fmt.Errorf("%w: aws credentials unavailable before submission", ErrProviderUnavailable)
	}
	// 本次提交固定使用已取得的凭证，避免 SDK 签名时再次刷新并混淆失败阶段。
	cfg.Credentials = awscredentials.NewStaticCredentialsProvider(retrieved.AccessKeyID, retrieved.SecretAccessKey, retrieved.SessionToken)
	return sendAWSInput(ctx, ses.NewFromConfig(cfg), provider, input)
}

// sendAWSInput 提交一次 SES 请求，只有取得 MessageId 才确认厂商受理。
func sendAWSInput(ctx context.Context, client sesRawEmailAPI, provider string, input *ses.SendRawEmailInput) (SendResult, error) {
	if err := ctx.Err(); err != nil {
		return SendResult{}, err
	}
	result, err := client.SendRawEmail(ctx, input)
	if err != nil {
		return SendResult{}, cloudFailure("aws", err)
	}
	if result == nil || aws.ToString(result.MessageId) == "" {
		return SendResult{}, fmt.Errorf("%w: aws returned no message ID", ErrDeliveryUncertain)
	}
	return SendResult{
		Accepted:          len(input.Destinations),        // 包含主收件、抄送和密送的受理人数
		Provider:          provider,                       // 本次通道名称
		ProviderMessageID: aws.ToString(result.MessageId), // SES 受理编号
	}, nil
}

// buildAWSRawEmail 复用 MIME 编码，保留文本、HTML、抄送和附件，密送不进入邮件头。
func buildAWSRawEmail(from string, req Request) ([]byte, error) {
	return buildMessageWithOptions(from, req.Recipients, req.CC, req.Subject, req.Content, req.HTMLContent, req.Attachments)
}

// buildAWSInput 在本地检查 SES 的收件人数、编码后大小和行长限制。
func buildAWSInput(from string, req Request, destinations []string) (*ses.SendRawEmailInput, error) {
	if len(destinations) > 50 {
		return nil, fmt.Errorf("%w: aws ses supports at most 50 recipients per message", ErrProviderUnsupported)
	}
	raw, err := buildAWSRawEmail(from, req)
	if err != nil {
		return nil, err
	}
	if len(raw) > 10_000_000 {
		return nil, fmt.Errorf("%w: aws ses message exceeds 10 MB after MIME encoding", ErrProviderUnsupported)
	}
	for _, line := range strings.Split(string(raw), "\r\n") {
		if len(line) > 998 {
			return nil, fmt.Errorf("%w: aws ses MIME line exceeds 998 bytes", ErrProviderUnsupported)
		}
	}
	return &ses.SendRawEmailInput{
		Source:       aws.String(from),                       // 已配置的 SES 发件地址
		Destinations: append([]string(nil), destinations...), // 独立的完整信封收件人列表
		RawMessage: &types.RawMessage{
			Data: raw, // 已校验大小和行长的 MIME 邮件
		},
	}, nil
}

// cloudFailure 区分提交前失败、明确拒绝和无确定回执；超时不能直接认定未发送。
func cloudFailure(provider string, err error) error {
	// 建立连接前就失败，尚未提交邮件，允许按不可用策略重试或切换通道。
	var dialError *net.OpError
	if errors.As(err, &dialError) && dialError.Op == "dial" {
		return fmt.Errorf("%w: %s connection failed before submission", ErrProviderUnavailable, provider)
	}
	// 域名解析失败也发生在提交前，与拨号失败使用相同的重试分类。
	var dnsError *net.DNSError
	if errors.As(err, &dnsError) {
		return fmt.Errorf("%w: %s endpoint resolution failed before submission", ErrProviderUnavailable, provider)
	}
	// 阿里云回执优先识别超时，再识别限流和明确拒绝，避免把超时误当成可重发。
	var aliyunError *aliyunerrors.ServerError
	if errors.As(err, &aliyunError) {
		// HTTP 408 或厂商超时码无法证明未受理，停止所有重试和通道切换。
		if aliyunError.HttpStatus() == http.StatusRequestTimeout || strings.Contains(strings.ToLower(aliyunError.ErrorCode()), "timeout") {
			return fmt.Errorf("%w: %s request timed out without acknowledgement", ErrDeliveryUncertain, provider)
		}
		// 厂商明确限流，按暂时不可用处理，允许上层执行退避重试。
		if isThrottling(aliyunError.ErrorCode()) {
			return fmt.Errorf("%w: %s: %s", ErrProviderUnavailable, provider, aliyunError.ErrorCode())
		}
		// 其余 4xx 视为明确拒绝，不在当前通道重试，由上层尝试下一通道。
		if aliyunError.HttpStatus() >= 400 && aliyunError.HttpStatus() < 500 {
			return fmt.Errorf("%w: %s: %s", ErrProviderRejected, provider, aliyunError.ErrorCode())
		}
	}
	// AWS 使用 SDK 的业务错误分类，不直接依赖 HTTP 状态码判断受理结果。
	var apiError smithy.APIError
	if errors.As(err, &apiError) {
		// 超时可能发生在受理之后，必须保留结果不确定状态。
		if strings.Contains(strings.ToLower(apiError.ErrorCode()), "timeout") {
			return fmt.Errorf("%w: %s request timed out without acknowledgement", ErrDeliveryUncertain, provider)
		}
		// 明确限流允许按不可用策略重试，不能按普通拒绝直接跳过当前通道。
		if isThrottling(apiError.ErrorCode()) {
			return fmt.Errorf("%w: %s: %s", ErrProviderUnavailable, provider, apiError.ErrorCode())
		}
		// 除超时和限流外，SDK 明确归为客户端错误时按厂商拒绝处理。
		if apiError.ErrorFault() == smithy.FaultClient {
			return fmt.Errorf("%w: %s: %s", ErrProviderRejected, provider, apiError.ErrorCode())
		}
	}
	// 未归类的断线或服务端故障无法确认是否已受理，保守停止重发。
	return fmt.Errorf("%w: %s request did not receive a definitive acknowledgement", ErrDeliveryUncertain, provider)
}

// isThrottling 识别厂商限流错误，供上层按不可用策略退避重试。
func isThrottling(code string) bool {
	code = strings.ToLower(code)
	return strings.Contains(code, "throttl") || code == "toomanyrequestsexception"
}
