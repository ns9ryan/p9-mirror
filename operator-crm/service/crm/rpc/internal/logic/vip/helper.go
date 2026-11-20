package viplogic

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/platform-base/pkg/resourceurl"
)

var (
	iconOperatorPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}$`)
	iconFilePattern     = regexp.MustCompile(`^[0-9a-f]{32}\.png$`)
)

// iconOperator 复用认证上下文，只允许分站运营身份处理本站图标，不接受请求指定站点。
func iconOperator(ctx context.Context) (string, error) {
	claims, ok := ctxdata.OperatorClaimsFromCtx(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, i18n.Unauthorized)
	}
	if claims.IsPlatform || claims.TokenType == ctxdata.TokenTypePreview ||
		!iconOperatorPattern.MatchString(claims.OperatorCode) {
		return "", status.Error(codes.PermissionDenied, "common.forbidden")
	}
	return claims.OperatorCode, nil
}

// iconValue 返回保存值、展示URL和错误；复用公开地址转换并校验目录，不写库或迁移对象。
func iconValue(value, operatorCode string, groupNo int64, baseURL string) (string, string, error) {
	// Key按公开前缀拼接，已知前缀URL裁成Key；空值原样返回，Key缺少有效前缀时返回错误。
	stored, accessURL, err := resourceurl.Value(value, baseURL)
	if err != nil || stored == "" {
		return stored, accessURL, err
	}
	key := stored
	locator, err := url.Parse(stored)
	if err != nil {
		return "", "", status.Error(codes.InvalidArgument, i18n.InvalidParam)
	}
	// 未匹配公开前缀的完整URL仍兼容；若包含VIP目录，仍须检查站点和等级，不能绕过归属校验。
	if locator.IsAbs() {
		position := strings.LastIndex(locator.Path, "/vipicon/")
		if position < 0 {
			return stored, accessURL, nil
		}
		key = locator.Path[position+1:]
	}
	if !validIconKey(key, operatorCode, groupNo) {
		return "", "", status.Error(codes.InvalidArgument, i18n.InvalidParam)
	}
	return stored, accessURL, nil
}

// validIconKey 只允许本等级组的总站公共默认Key或登录站点Key，目录与现有OSS生成规则一致。
func validIconKey(key, operatorCode string, groupNo int64) bool {
	if groupNo < 0 || groupNo > 9 {
		return false
	}
	parts := strings.Split(key, "/")
	// 总站公共Key有三段，分站Key有四段；分站目录必须与认证身份的站点编码一致。
	if len(parts) == 4 {
		if parts[1] != operatorCode {
			return false
		}
		parts = []string{parts[0], parts[2], parts[3]}
	}
	return len(parts) == 3 && parts[0] == "vipicon" &&
		parts[1] == fmt.Sprintf("%d-%d", groupNo*10, groupNo*10+9) && iconFilePattern.MatchString(parts[2])
}

// iconDBError 供修改和恢复共用，沿用原数据库错误分类及公共错误转换，不执行数据库操作。
func iconDBError(err error) error {
	switch {
	case ent.IsNotFound(err):
		// 本站记录不存在时保留不存在语义，不转换为内部错误。
		return xerr.RpcErr(xerr.EntNotFound(i18n.InvalidParam, err))
	case ent.IsConstraintError(err):
		// 保留唯一索引等数据库约束冲突语义。
		return xerr.RpcErr(xerr.EntConstraintError(i18n.InvalidParam, err))
	case ent.IsValidationError(err):
		// Ent字段校验失败仍作为参数错误处理。
		return xerr.RpcErr(xerr.EntValidationError(i18n.InvalidParam, err))
	default:
		// 未知错误继续交给现有公共组件记录并转换，不直接用数据库错误文本作为业务提示。
		return xerr.RpcErr(err)
	}
}
