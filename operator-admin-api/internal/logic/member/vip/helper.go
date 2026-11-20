package vip

import (
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"oa.98ent.com/p9/common/xerr"
)

// iconRequestError 供图标查询、修改和恢复共用，复用xerr返回HTTP错误；包含API调用CRM自身超时。
func iconRequestError(err error) error {
	switch status.Code(err) {
	case codes.DeadlineExceeded:
		// 包括总站调用超时及API等待CRM超时，统一返回504，不透出内部连接信息。
		return xerr.Err(http.StatusGatewayTimeout, "common.request_timeout")
	case codes.Unavailable, codes.FailedPrecondition:
		// 服务不可达、总站资料不合法或公开前缀未就绪时返回503。
		return xerr.Err(http.StatusServiceUnavailable, "common.service_unavailable")
	default:
		// 参数、身份、权限及其他错误继续沿用原响应转换，避免改变已有处理语义。
		return err
	}
}
