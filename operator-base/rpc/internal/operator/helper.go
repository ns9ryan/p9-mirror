package operator

import (
	"net/url"
	"path"
	"regexp"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"oa.98ent.com/p9/operator-base/rpc/internal/config"
)

var siteLogoKeyPattern = regexp.MustCompile(
	`^(?:logo/([a-zA-Z0-9][a-zA-Z0-9_-]{0,63})|([a-zA-Z0-9][a-zA-Z0-9_-]{0,63})/logo)/[0-9]{8}/[0-9a-f]{32}\.(?:png|jpg)$`,
)

var defaultLogoKeyPattern = regexp.MustCompile(`^defaults/[a-zA-Z0-9][a-zA-Z0-9_.-]*\.(?:png|jpg)$`)

// LogoValue 返回数据库保存值和对外访问地址，不调用 RPC、不读写数据库。
// 已知公开前缀下的本站 Logo 保存为 Key；旧外链和站内绝对路径保持兼容，不擅自替换域名。
func LogoValue(value, operatorCode string, c config.Config) (string, string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", "", nil
	}
	locator, err := url.Parse(value)
	if err != nil {
		return "", "", status.Error(codes.InvalidArgument, "invalid site logo address")
	}
	baseURL := strings.TrimSpace(c.SiteLogoBaseURL)
	// 未配置公开前缀时，完整 URL 和原有站内路径按原样保存、返回。
	if baseURL == "" {
		if locator.IsAbs() || strings.HasPrefix(value, "/") {
			return value, value, nil
		}
		return "", "", status.Error(codes.FailedPrecondition, "site logo public base URL is not configured")
	}
	base, err := siteLogoBase(baseURL)
	if err != nil {
		return "", "", err
	}
	// 完整地址只在匹配当前或已配置旧公开前缀、且目录符合 Logo 规则时裁成 Key。
	key := value
	if locator.IsAbs() || strings.HasPrefix(value, "/") {
		key, err = siteLogoURLKey(value, locator, base, c.SiteLogoLegacyBaseURLs)
		if err != nil {
			return "", "", err
		}
		if key == "" {
			// 未识别的旧外链和站内路径保持原值；不会自动迁移对象或替换其访问域名。
			return value, value, nil
		}
	}
	owner, ok := siteLogoKeyOwner(key)
	if !ok {
		return "", "", status.Error(codes.InvalidArgument, "invalid site logo object key")
	}
	// 公共默认图片不带站点归属，其他 Logo 必须属于当前登录站点。
	if owner != "" && owner != operatorCode {
		return "", "", status.Error(codes.PermissionDenied, "site logo does not belong to login operator")
	}
	// 返回值依次为入库 Key、供前端展示的完整地址；更改前缀配置本身不会搬迁 OSS 图片。
	base.Path = strings.TrimRight(base.Path, "/") + "/" + key
	accessURL := base.String()
	if len(accessURL) > 1024 {
		return "", "", status.Error(codes.FailedPrecondition, "site logo public URL exceeds 1024 bytes")
	}
	return key, accessURL, nil
}

// siteLogoURLKey 从当前或明确配置的旧公开前缀提取合法 Logo Key，未识别时保留给调用方兼容。
// 按原顺序匹配，只有当前前缀未命中时才解析旧前缀，不访问 OSS 或修改数据库。
func siteLogoURLKey(value string, locator, base *url.URL, legacy []string) (string, error) {
	prefixes := append([]string{base.String()}, legacy...)
	for index, prefix := range prefixes {
		// 当前公开前缀已校验并解析，只有遍历到旧前缀时才额外解析。
		trusted := base
		if index > 0 {
			var err error
			trusted, err = siteLogoBase(prefix)
			if err != nil {
				return "", err
			}
		}
		// 同时核对协议、主机（含端口）和完整路径段，防止相似域名或桶名前缀误匹配。
		pathPrefix := strings.TrimRight(trusted.Path, "/") + "/"
		if locator.Scheme != trusted.Scheme || !strings.EqualFold(locator.Host, trusted.Host) ||
			locator.User != nil || locator.RawQuery != "" || locator.ForceQuery ||
			locator.Fragment != "" || strings.Contains(value, "#") || locator.RawPath != "" ||
			!strings.HasPrefix(locator.Path, pathPrefix) {
			continue
		}
		candidate := strings.TrimPrefix(locator.Path, pathPrefix)
		if _, ok := siteLogoKeyOwner(candidate); ok {
			return candidate, nil
		}
	}
	return "", nil
}

// siteLogoBase 只解析公开读地址，不复用内部上传服务地址，也不改变服务启动流程。
func siteLogoBase(value string) (*url.URL, error) {
	base, err := url.ParseRequestURI(strings.TrimSpace(value))
	if err != nil || (base.Scheme != "http" && base.Scheme != "https") || base.Hostname() == "" ||
		base.User != nil || base.RawQuery != "" || base.ForceQuery || base.RawPath != "" ||
		strings.Contains(value, "#") {
		return nil, status.Error(codes.FailedPrecondition, "site logo public base URL is invalid")
	}
	base.Path = strings.TrimRight(base.Path, "/")
	if strings.Contains(base.Path, `\`) || (base.Path != "" && path.Clean(base.Path) != base.Path) {
		return nil, status.Error(codes.FailedPrecondition, "site logo public base path is invalid")
	}
	return base, nil
}

// siteLogoKeyOwner 识别 OSS 当前和旧版 Logo 目录，以及可共用的默认图片目录。
func siteLogoKeyOwner(key string) (string, bool) {
	if defaultLogoKeyPattern.MatchString(key) {
		return "", true
	}
	parts := siteLogoKeyPattern.FindStringSubmatch(key)
	if len(parts) != 3 {
		return "", false
	}
	if parts[1] != "" {
		return parts[1], true
	}
	return parts[2], true
}
