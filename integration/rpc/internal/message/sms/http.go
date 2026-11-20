package sms

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"

	"oa.98ent.com/p9/integration/rpc/internal/config"

	"github.com/nyaruka/phonenumbers"
)

// sendProvider 将统一号码和内容转换为各厂商请求，不在适配层执行故障切换。
func (s *Service) sendProvider(ctx context.Context, c config.SMSConf, provider, phone, content string) outcome {
	switch provider {
	case "submail":
		// 此通道不支持 +86；当前实现会给正文补固定【logo】前缀。
		number, _ := phonenumbers.Parse(phone, "")
		if number.GetCountryCode() == 86 {
			return outcome{status: "failed", message: "submail international sms does not support +86"}
		}
		content = strings.TrimSpace(content)
		if !strings.HasPrefix(content, "【logo】") {
			content = "【logo】" + content
		}
		values := url.Values{
			"appid":     {c.Submail.AppID},  // 应用编号
			"signature": {c.Submail.AppKey}, // 应用密钥，仅提交给厂商
			"to":        {phone},            // 已规范化的国际号码
			"content":   {content},          // 当前实现补齐前缀后的正文
		}
		if sender := strings.TrimSpace(c.Submail.Sender); sender != "" {
			values.Set("sender", sender)
		}
		return s.sendHTTP(ctx, provider, c.Submail.Endpoint, strings.NewReader(values.Encode()), "application/x-www-form-urlencoded", "", "", "")
	case "byteplus":
		payload, _ := json.Marshal(map[string]string{
			"PhoneNumbers": phone,             // 单个国际号码
			"Content":      content,           // 短信正文
			"From":         c.Byteplus.Sender, // 配置的发件标识
		})
		return s.sendHTTP(ctx, provider, c.Byteplus.Endpoint, strings.NewReader(string(payload)), "application/json;charset=utf-8", c.Byteplus.Username, c.Byteplus.Password, "")
	case "gotone":
		payload, _ := json.Marshal(map[string]string{
			"recipient": phone,             // 单个国际号码
			"message":   content,           // 短信正文
			"sender_id": c.Gotone.SenderID, // 配置的发件标识
			"type":      c.Gotone.Type,     // 厂商要求的短信类型
		})
		return s.sendHTTP(ctx, provider, c.Gotone.Endpoint, strings.NewReader(string(payload)), "application/json", "", "", c.Gotone.APIToken)
	case "aws":
		return s.sendAWS(ctx, c.AWS, phone, content)
	case "ali":
		return sendAliyun(ctx, c.Aliyun, phone, content, s.client.Transport)
	default:
		return outcome{status: "failed", message: "unsupported sms provider"}
	}
}

// sendHTTP 统一认证和响应读取；HTTP 成功后仍需厂商业务字段明确确认受理。
func (s *Service) sendHTTP(ctx context.Context, provider, endpoint string, body io.Reader, contentType, username, password, token string) outcome {
	if ctx.Err() != nil {
		return outcome{status: "failed", message: "sms request canceled before submission"}
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, body)
	if err != nil {
		return outcome{status: "failed", message: "invalid sms endpoint"}
	}
	request.Header.Set("Content-Type", contentType)
	request.Header.Set("Accept", "application/json")
	// Byteplus 使用 Basic 认证，Gotone 使用 Bearer 令牌；Submail 凭证已在表单中。
	if username != "" {
		request.SetBasicAuth(username, password)
	}
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response, err := s.client.Do(request)
	if err != nil {
		return transportFailure(err)
	}
	defer response.Body.Close()
	if response.StatusCode >= 400 && response.StatusCode < 500 && response.StatusCode != http.StatusRequestTimeout {
		return outcome{status: "failed", message: "sms provider rejected HTTP request: " + response.Status}
	}
	// 3xx、5xx 和 408 均不能证明短信未提交，保守标为 unknown，阻止通道重发。
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return outcome{status: "unknown", message: "sms provider returned HTTP error without delivery confirmation"}
	}
	// 多读一个字节识别超限，避免只截取前 1 MiB 后误当作完整响应。
	data, err := io.ReadAll(io.LimitReader(response.Body, (1<<20)+1))
	if err != nil || len(data) > 1<<20 {
		return outcome{status: "unknown", message: "sms provider response was incomplete or too large"}
	}
	// 最多读取 1 MiB，只保留用于确认受理的字段及白名单附加数据。
	switch provider {
	case "submail":
		var value struct {
			Status string `json:"status"`  // success 受理或 error 拒绝
			SendID string `json:"send_id"` // 成功时要求非空受理编号
			Code   string `json:"code"`    // 厂商错误码，仅经筛选后返回
		}
		if json.Unmarshal(data, &value) != nil {
			return outcome{status: "unknown", message: "invalid submail response"}
		}
		if value.Status == "error" {
			return outcome{status: "failed", message: "submail rejected sms" + vendorCode(value.Code)}
		}
		if value.Status == "success" && strings.TrimSpace(value.SendID) != "" {
			var safe struct {
				Status         string  `json:"status"`          // 厂商受理状态
				SendID         string  `json:"send_id"`         // 厂商受理编号
				Fee            float64 `json:"fee"`             // 厂商返回的本次费用
				AccountBalance float64 `json:"account_balance"` // 厂商返回的账户余额
			}
			if json.Unmarshal(data, &safe) != nil {
				return outcome{status: "sent", messageID: value.SendID}
			}
			sanitized, _ := json.Marshal(safe)
			return outcome{status: "sent", messageID: value.SendID, data: string(sanitized)}
		}
	case "byteplus":
		var value struct {
			ResponseMetadata struct {
				Error *struct {
					Code string `json:"Code"` // 厂商拒绝原因码
				} `json:"Error"` // 非空表示厂商明确返回错误
			} `json:"ResponseMetadata"` // 请求处理元数据
			Result struct {
				MessageID []string `json:"MessageID"` // 单号码提交时要求恰有一个非空编号
			} `json:"Result"` // 厂商受理结果
		}
		if json.Unmarshal(data, &value) != nil {
			return outcome{status: "unknown", message: "invalid byteplus response"}
		}
		if value.ResponseMetadata.Error != nil {
			return outcome{status: "failed", message: "byteplus rejected sms" + vendorCode(value.ResponseMetadata.Error.Code)}
		}
		if len(value.Result.MessageID) == 1 && strings.TrimSpace(value.Result.MessageID[0]) != "" {
			return outcome{status: "sent", messageID: value.Result.MessageID[0]}
		}
	case "gotone":
		var value struct {
			Status string `json:"status"` // success 表示受理，error 表示拒绝
			Data   struct {
				UID string `json:"uid"` // 厂商回执编号，当前允许为空
			} `json:"data"` // 回执数据，不透传完整响应
		}
		if json.Unmarshal(data, &value) != nil {
			return outcome{status: "unknown", message: "invalid gotone response"}
		}
		if value.Status == "error" {
			return outcome{status: "failed", message: "gotone rejected sms"}
		}
		if value.Status == "success" {
			return outcome{status: "sent", messageID: value.Data.UID}
		}
	}
	return outcome{status: "unknown", message: "sms provider returned no explicit acceptance confirmation"}
}

// transportFailure 仅将解析、拨号等提交前失败归为 failed，其余连接错误保留 unknown。
func transportFailure(err error) outcome {
	var operation *net.OpError
	var dnsError *net.DNSError
	// 解析、拨号或连接代理失败时尚未提交短信，允许尝试下一个通道。
	if errors.As(err, &dnsError) || errors.As(err, &operation) && (operation.Op == "dial" || operation.Op == "proxyconnect") {
		return outcome{status: "failed", message: "sms connection failed before submission"}
	}
	// 已建立连接后的读写错误可能发生在受理之后，禁止自动重发。
	return outcome{status: "unknown", message: "sms submission failed without delivery confirmation"}
}

// vendorCode 仅透出长度受限的错误码，不原样返回可能包含敏感信息的厂商响应。
func vendorCode(code string) string {
	code = strings.TrimSpace(code)
	if code == "" || len(code) > 100 {
		return ""
	}
	for _, character := range code {
		if character < 'a' || character > 'z' {
			if (character < 'A' || character > 'Z') && (character < '0' || character > '9') && character != '_' && character != '-' && character != '.' {
				return ""
			}
		}
	}
	return " (code=" + code + ")"
}
