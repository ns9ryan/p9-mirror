package sms

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"oa.98ent.com/p9/integration/rpc/internal/config"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	aliyunerrors "github.com/aliyun/alibaba-cloud-sdk-go/sdk/errors"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	awscredentials "github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/aws/smithy-go"
)

// snsAPI 表示向国际号码提交短信所需的 SNS 能力。
type snsAPI interface {
	Publish(context.Context, *sns.PublishInput, ...func(*sns.Options)) (*sns.PublishOutput, error)
}

// sendAWS 复用 SNS 客户端，向单个国际号码提交短信。
func (s *Service) sendAWS(ctx context.Context, c config.SMSAWSConfig, phone, content string) outcome {
	// 凭证和客户端在提交前准备，初始化失败时尚未发送短信。
	client, err := s.awsClientFor(ctx, c)
	if err != nil {
		return outcome{status: "failed", message: "unable to load aws sms credentials"}
	}
	return sendAWSInput(ctx, client, awsInput(c, phone, content))
}

// awsClientFor 按配置复用 SNS 客户端和凭证缓存，配置变更后才重新构建。
func (s *Service) awsClientFor(ctx context.Context, c config.SMSAWSConfig) (snsAPI, error) {
	s.awsMu.Lock()
	defer s.awsMu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// 配置完全相同时复用客户端，避免每个号码重复初始化凭证缓存。
	if s.awsClient != nil && s.awsConfig == c {
		return s.awsClient, nil
	}
	// MaxAttempts 包含首次调用，设为 1 关闭 SDK 额外重试，避免不确定结果重复提交。
	options := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(c.Region),
		awsconfig.WithRetryMaxAttempts(1),
		awsconfig.WithHTTPClient(s.client),
	}
	// 显式密钥优先；未填写时使用配置校验允许的 SDK 默认凭证链。
	if c.AccessKeyID != "" {
		options = append(options, awsconfig.WithCredentialsProvider(awscredentials.NewStaticCredentialsProvider(c.AccessKeyID, c.SecretAccessKey, c.SessionToken)))
	}
	cfg, err := awsconfig.LoadDefaultConfig(ctx, options...)
	if err != nil {
		return nil, err
	}
	s.awsClient = sns.NewFromConfig(cfg)
	s.awsConfig = c
	return s.awsClient, nil
}

// awsInput 组装 SNS 请求，发件标识和美元价格上限仅在配置非空时提交。
func awsInput(c config.SMSAWSConfig, phone, content string) *sns.PublishInput {
	attributes := map[string]types.MessageAttributeValue{
		"AWS.SNS.SMS.SMSType": {
			DataType:    aws.String("String"), // SNS 消息属性按字符串提交
			StringValue: aws.String(c.Type),   // Transactional 事务短信或 Promotional 营销短信
		},
	}
	if c.SenderID != "" {
		attributes["AWS.SNS.SMS.SenderID"] = types.MessageAttributeValue{
			DataType:    aws.String("String"),   // 发件标识属性类型
			StringValue: aws.String(c.SenderID), // 配置的发件标识
		}
	}
	if c.MaxPrice != "" {
		attributes["AWS.SNS.SMS.MaxPrice"] = types.MessageAttributeValue{
			DataType:    aws.String("String"),   // 价格按十进制字符串提交
			StringValue: aws.String(c.MaxPrice), // 单条短信美元价格上限
		}
	}
	return &sns.PublishInput{
		PhoneNumber:       aws.String(phone),   // 已规范化的 E.164 国际号码
		Message:           aws.String(content), // 本次短信正文
		MessageAttributes: attributes,          // 短信类型、可选发件标识和价格上限
	}
}

// sendAWSInput 以 SNS 返回的 MessageId 确认受理，不代表运营商已送达手机。
func sendAWSInput(ctx context.Context, client snsAPI, input *sns.PublishInput) outcome {
	if ctx.Err() != nil {
		return outcome{status: "failed", message: "sms request canceled before submission"}
	}
	response, err := client.Publish(ctx, input)
	// SDK 返回错误时统一分类，不能把所有网络错误都当成明确失败。
	if err != nil {
		return cloudFailure(err)
	}
	// 只有非空 MessageId 才确认厂商受理，缺少回执编号时禁止自动重发。
	if response == nil || strings.TrimSpace(aws.ToString(response.MessageId)) == "" {
		return outcome{status: "unknown", message: "aws returned no sms acceptance ID"}
	}
	return outcome{status: "sent", messageID: aws.ToString(response.MessageId)}
}

// contextTransport 将整批预算和单次超时传递给阿里云 SDK 的 HTTP 请求。
type contextTransport struct {
	ctx  context.Context   // 本次发送的取消和超时信号
	base http.RoundTripper // 共用 HTTP 传输和连接池
}

// RoundTrip 为 SDK 请求绑定业务上下文，并拦截重定向响应。
func (transport contextTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := transport.base.RoundTrip(request.WithContext(transport.ctx))
	// SDK 内部客户端不能配置 CheckRedirect，在传输层拦截 3xx，避免重复提交及签名参数外泄。
	// 原请求可能已受理，返回传输错误使上层保留 unknown，不切换其他厂商。
	if err == nil && response != nil && response.StatusCode >= 300 && response.StatusCode < 400 {
		if response.Body != nil {
			_ = response.Body.Close()
		}
		return nil, errors.New("aliyun redirect response blocked")
	}
	return response, err
}

// aliyunInput 使用国际短信接口，仅在构造厂商请求时去掉 E.164 号码的加号。
func aliyunInput(c config.SMSAliyunConfig, phone, content string) *dysmsapi.SendMessageToGlobeRequest {
	request := dysmsapi.CreateSendMessageToGlobeRequest()
	request.Scheme = "https"                                  // 厂商请求只走 HTTPS
	request.Domain = "dysmsapi." + c.Region + ".aliyuncs.com" // 使用配置区域的服务端点
	request.To = strings.TrimPrefix(phone, "+")               // 厂商接收不带加号的国际号码
	request.Message = content                                 // 本次短信正文
	request.From = c.Sender                                   // 发件标识
	request.Type = c.Type                                     // NOTIFY 通知或 OTP 验证码
	return request
}

// sendAliyun 提交国际短信，以 OK 和非空 MessageId 确认受理。
func sendAliyun(ctx context.Context, c config.SMSAliyunConfig, phone, content string, transport http.RoundTripper) outcome {
	// 提交前已取消时返回明确失败，不再调用厂商接口。
	if ctx.Err() != nil {
		return outcome{status: "failed", message: "sms request canceled before submission"}
	}
	// 禁用 SDK 自动重试，是否切换厂商由上层逐号码状态决定。
	clientConfig := sdk.NewConfig().WithAutoRetry(false)
	clientConfig.Scheme = "HTTPS"
	clientConfig.Transport = contextTransport{
		ctx:  ctx,       // 继承本次调用的取消和超时
		base: transport, // 复用服务的 HTTP 连接池
	}
	// 用配置区域和密钥创建客户端；创建失败发生在短信提交之前。
	client, err := dysmsapi.NewClientWithOptions(c.Region, clientConfig, credentials.NewAccessKeyCredential(c.AccessKeyID, c.AccessKeySecret))
	if err != nil {
		return outcome{status: "failed", message: "unable to initialize ali sms client"}
	}
	// 组装国际短信参数并只提交一次，后续是否切换通道由 outcome 状态决定。
	response, err := client.SendMessageToGlobe(aliyunInput(c, phone, content))
	if err != nil {
		return cloudFailure(err)
	}
	// 缺少成功响应不能证明短信未受理，保留 unknown，不自动换厂商重发。
	if response == nil || !response.IsSuccess() {
		return outcome{status: "unknown", message: "ali returned no successful sms response"}
	}
	// HTTP 请求成功后仍需检查厂商业务码，不能仅凭 HTTP 成功计为已受理。
	if response.ResponseCode != "OK" {
		// 厂商超时码可能发生在受理之后，优先于普通拒绝分类。
		if strings.Contains(strings.ToLower(response.ResponseCode), "timeout") {
			return outcome{status: "unknown", message: "ali sms timed out without delivery confirmation"}
		}
		// 非空且非超时的错误码表示明确拒绝，仅透出经过筛选的错误码。
		if response.ResponseCode != "" {
			return outcome{status: "failed", message: "ali rejected sms" + vendorCode(response.ResponseCode)}
		}
		// 缺少业务码时无法确认结果，不能假设发送失败。
		return outcome{status: "unknown", message: "ali returned no sms acceptance code"}
	}
	// 即便业务码为 OK，也要求有效 MessageId 才确认受理。
	if strings.TrimSpace(response.MessageId) == "" {
		return outcome{status: "unknown", message: "ali returned no sms acceptance ID"}
	}
	return outcome{
		status:    "sent",             // 厂商已受理，不代表手机已送达
		messageID: response.MessageId, // 保留厂商受理编号
	}
}

// cloudFailure 将明确拒绝和超时分开：超时即使带 4xx 也可能已经提交。
func cloudFailure(err error) outcome {
	// AWS 明确返回非超时的客户端错误时，标记失败，允许上层尝试其他通道。
	var apiError smithy.APIError
	if errors.As(err, &apiError) && apiError.ErrorFault() == smithy.FaultClient &&
		!strings.Contains(strings.ToLower(apiError.ErrorCode()), "timeout") {
		return outcome{status: "failed", message: "aws rejected sms" + vendorCode(apiError.ErrorCode())}
	}
	var serverError *aliyunerrors.ServerError
	// 阿里云超时可能发生在受理之后，不能因为响应带有 4xx 就重新发送。
	if errors.As(err, &serverError) && strings.Contains(strings.ToLower(serverError.ErrorCode()), "timeout") {
		return outcome{status: "unknown", message: "ali sms timed out without delivery confirmation"}
	}
	// 排除超时后，其余明确的阿里云 4xx 拒绝才允许切换通道。
	if errors.As(err, &serverError) && serverError.HttpStatus() >= 400 && serverError.HttpStatus() < 500 &&
		serverError.HttpStatus() != http.StatusRequestTimeout {
		return outcome{status: "failed", message: "ali rejected sms" + vendorCode(serverError.ErrorCode())}
	}
	// 无明确业务回执时，再区分提交前连接失败和提交状态不明。
	return transportFailure(err)
}
