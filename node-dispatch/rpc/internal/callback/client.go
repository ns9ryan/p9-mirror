package callback

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/rest/httpc"

	"oa.98ent.com/p9/node-dispatch/rpc/internal/config"
)

const (
	serviceName      = "node-dispatch-callback" // HTTP回调服务名称
	maxErrorBodySize = 4096                     // 错误响应最大读取长度
)

// Client HTTP回调客户端
type Client struct {
	taskResultURL string        // 任务结果回调地址
	secret        string        // 回调认证密钥
	service       httpc.Service // go-zero HTTP客户端
}

// TaskResultRequest 任务结果回调请求
type TaskResultRequest struct {
	TaskNo       string          // 调度任务编号
	Status       int64           // 任务最终状态: 3成功, 4失败
	Result       json.RawMessage // 执行结果
	ErrorMessage string          // 执行失败原因
	FinishedAt   int64           // 执行结束时间, Unix毫秒时间戳
}

// taskResultHTTPRequest HTTP任务结果回调请求
type taskResultHTTPRequest struct {
	TaskNo        string          `json:"task_no"`                // 调度任务编号
	Status        int64           `json:"status"`                 // 任务最终状态: 3成功, 4失败
	Result        json.RawMessage `json:"result,optional"`        // 执行结果
	ErrorMessage  string          `json:"error_message,optional"` // 执行失败原因
	FinishedAt    int64           `json:"finished_at"`            // 执行结束时间, Unix毫秒时间戳
	Authorization string          `header:"Authorization"`        // 回调认证信息
}

// NewClient 创建HTTP回调客户端
func NewClient(c config.CallbackConf) (*Client, error) {
	// 整理配置
	taskResultURL := strings.TrimSpace(c.TaskResultURL)
	secret := strings.TrimSpace(c.Secret)

	// 校验配置
	if taskResultURL == "" {
		return nil, fmt.Errorf("callback task result url is required")
	}
	if secret == "" {
		return nil, fmt.Errorf("callback secret is required")
	}
	if c.Timeout <= 0 {
		return nil, fmt.Errorf("callback timeout must be greater than 0")
	}

	// 创建带超时控制的HTTP客户端
	httpClient := &http.Client{
		Timeout: time.Duration(c.Timeout) * time.Millisecond,
	}

	// 创建go-zero HTTP服务客户端
	service := httpc.NewServiceWithClient(
		serviceName,
		httpClient,
	)

	// 返回HTTP回调客户端
	return &Client{
		taskResultURL: taskResultURL, // 任务结果回调地址
		secret:        secret,        // 回调认证密钥
		service:       service,       // go-zero HTTP客户端
	}, nil
}

// TaskResult 回调任务执行结果
func (c *Client) TaskResult(ctx context.Context, data TaskResultRequest) error {
	// 创建HTTP回调请求
	request := taskResultHTTPRequest{
		TaskNo:        data.TaskNo,          // 调度任务编号
		Status:        data.Status,          // 任务最终状态
		Result:        data.Result,          // 执行结果
		ErrorMessage:  data.ErrorMessage,    // 执行失败原因
		FinishedAt:    data.FinishedAt,      // 执行结束时间
		Authorization: "Bearer " + c.secret, // 回调认证信息
	}

	// 发送HTTP回调请求
	resp, err := c.service.Do(
		ctx,
		http.MethodPost,
		c.taskResultURL,
		request,
	)
	if err != nil {
		return fmt.Errorf("发送任务结果回调请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 2xx状态码表示回调成功
	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}

	// 读取有限长度的错误响应
	responseBody, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodySize))
	message := strings.TrimSpace(string(responseBody))

	if message == "" {
		return fmt.Errorf("任务结果回调失败: http_status=%d", resp.StatusCode)
	}

	return fmt.Errorf(
		"任务结果回调失败: http_status=%d response=%s",
		resp.StatusCode,
		message,
	)
}
