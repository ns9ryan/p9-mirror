package httpresponse

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"oa.98ent.com/p9/common/response"
)

// SetupHTTPX 统一站点接口的错误响应，并隐藏内部依赖和数据库错误细节。
func SetupHTTPX() {
	httpx.SetErrorHandlerCtx(func(ctx context.Context, err error) (int, any) {
		// 先复用公共错误转换，再补齐本站对依赖故障、超时和事务冲突的 HTTP 映射。
		converted := response.FromError(err)
		httpStatus := converted.Status
		switch status.Code(err) {
		case codes.Unavailable, codes.FailedPrecondition:
			httpStatus = http.StatusServiceUnavailable
		case codes.DeadlineExceeded:
			httpStatus = http.StatusGatewayTimeout
		case codes.Canceled:
			httpStatus = http.StatusRequestTimeout
		case codes.Aborted:
			// 并发修改冲突返回 409，由调用方重新读取配置后决定是否重试。
			httpStatus = http.StatusConflict
		}
		message := converted.Message
		if httpStatus >= http.StatusInternalServerError {
			// 内部原因只写入日志，对外返回通用状态说明，避免泄露 SQL 或连接信息。
			logx.WithContext(ctx).Errorw("站点 API 请求失败", logx.Field("error", err.Error()))
			message = http.StatusText(httpStatus)
		}
		// 错误响应不带 data，业务 code 与 HTTP 状态码保持一致。
		return httpStatus, map[string]any{"code": httpStatus, "msg": message}
	})
}
