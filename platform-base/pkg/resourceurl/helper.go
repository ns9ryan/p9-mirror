// Package resourceurl 处理总站公开素材的对象 Key 和访问地址，不访问 OSS 或数据库。
package resourceurl

import (
	"net/url"
	"path"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"oa.98ent.com/p9/platform-base/pkg/i18nkey"
)

// Value 返回数据库保存值和公开地址。已配置前缀下的 URL 转为 Key，其他旧 URL 保持原样。
// Key 必须由调用方继续校验素材目录；空值不拼接，缺少前缀时不把 Key 当作完整地址返回。
func Value(value, baseURL string) (string, string, error) {
	// 空图片保持为空；先限制保存值长度，避免超出原字段长度约定。
	value = strings.TrimSpace(value)
	if value == "" {
		return "", "", nil
	}
	if len(value) > 1024 {
		return "", "", status.Error(codes.InvalidArgument, i18nkey.ValidationError)
	}
	// 区分完整 URL 和桶内 Key，分别检查公开地址格式与相对路径安全性。
	locator, err := url.Parse(value)
	if err != nil {
		return "", "", status.Error(codes.InvalidArgument, i18nkey.ValidationError)
	}
	if locator.IsAbs() {
		locator, err = PublicURL(value)
		if err != nil {
			return "", "", err
		}
	} else if !validKey(value) {
		return "", "", status.Error(codes.InvalidArgument, i18nkey.ValidationError)
	}
	// 未配置公开前缀时只兼容完整 URL；Key 无法拼接，返回配置未就绪错误。
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		if locator.IsAbs() {
			return value, value, nil
		}
		return "", "", status.Error(codes.FailedPrecondition, i18nkey.ServiceUnavailable)
	}
	// 配置错误在当前素材请求中返回，不在这里初始化连接或改变启动条件。
	base, err := PublicURL(baseURL)
	if err != nil {
		return "", "", status.Error(codes.FailedPrecondition, i18nkey.ServiceUnavailable)
	}
	// 完整 URL 只有协议、主机端口和桶路径前缀全部匹配时才裁成 Key。
	key := value
	if locator.IsAbs() {
		prefix := strings.TrimRight(base.Path, "/") + "/"
		if !sameOrigin(locator, base) || !strings.HasPrefix(locator.Path, prefix) {
			return value, value, nil
		}
		key = strings.TrimPrefix(locator.Path, prefix)
	}
	if !validKey(key) {
		return "", "", status.Error(codes.InvalidArgument, i18nkey.ValidationError)
	}
	// 使用当前公开前缀拼出访问地址；这里只转换字符串，不检查或迁移实际对象。
	base.Path = strings.TrimRight(base.Path, "/") + "/" + key
	accessURL := base.String()
	if len(accessURL) > 1024 {
		return "", "", status.Error(codes.FailedPrecondition, i18nkey.ServiceUnavailable)
	}
	return key, accessURL, nil
}

// PublicURL 只接受无签名参数、无用户凭据和无歧义路径的 HTTP(S) 永久地址。
func PublicURL(value string) (*url.URL, error) {
	parsed, err := url.ParseRequestURI(value)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, i18nkey.ValidationError)
	}
	if !validOrigin(parsed) || !plainURL(parsed, value) || !cleanPath(parsed.Path) {
		return nil, status.Error(codes.InvalidArgument, i18nkey.ValidationError)
	}
	return parsed, nil
}

// validOrigin 排除非 HTTP(S) 协议、缺失主机和包含登录凭据的地址。
func validOrigin(parsed *url.URL) bool {
	return (parsed.Scheme == "http" || parsed.Scheme == "https") &&
		parsed.Hostname() != "" && parsed.User == nil && parsed.Opaque == ""
}

// plainURL 拒绝签名、片段和转义路径，避免 URL 校验与实际对象路径产生歧义。
func plainURL(parsed *url.URL, value string) bool {
	return parsed.RawQuery == "" && !parsed.ForceQuery && parsed.RawPath == "" &&
		!strings.Contains(value, "#") && !strings.Contains(value, "%")
}

// cleanPath 拒绝反斜杠、重复分隔符以及 .、.. 路径段；根路径允许末尾斜杠。
func cleanPath(value string) bool {
	trimmed := strings.TrimSuffix(value, "/")
	return !strings.Contains(value, `\`) && !strings.Contains(value, "//") &&
		(trimmed == "" || path.Clean(trimmed) == trimmed)
}

// validKey 仅接受桶内相对路径，具体素材目录和文件名由业务规则校验。
func validKey(value string) bool {
	return value != "" && !strings.HasPrefix(value, "/") &&
		value != "." && value != ".." && !strings.HasPrefix(value, "../") &&
		!strings.ContainsAny(value, "?#%:\\") && cleanPath(value) && !strings.HasSuffix(value, "/")
}

// sameOrigin 同时比较协议和主机端口，不能仅按路径裁掉相似域名的地址前缀。
func sameOrigin(left, right *url.URL) bool {
	return left.Scheme == right.Scheme && strings.EqualFold(left.Host, right.Host)
}
