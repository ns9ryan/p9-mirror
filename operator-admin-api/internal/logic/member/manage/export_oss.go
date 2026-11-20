package manage

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
	memberexport "oa.98ent.com/p9/operator-crm/pkg/memberexport"
	"oa.98ent.com/p9/platform-base/pkg/resourceurl"
)

const (
	memberExportUploadPath             = "/p9/oss/file/upload"
	memberExportScene                  = "member_export"
	memberExportPrivateVisibility      = "private"
	memberExportUploadTimeout          = 12 * time.Second
	memberExportMaxUploadResponseBytes = 1 << 20
	memberExportResponseOverflowBytes  = 1 // 多读一个字节，用于识别响应体超限。
	memberExportMaxDownloadTTL         = time.Hour
	memberExportOSSBusinessSuccessCode = 0 // OSS JSON 的业务成功码，与 HTTP 状态码分别校验。
	memberExportKeyDateLayout          = "20060102"
	memberExportKeyDateGroup           = "export_date"
	memberExportOSSAuthScopeHeader     = "X-OSS-Auth-Scope"
	memberExportOSSAuthScope           = "operator"
)

type exportOSSResult struct {
	URL          string `json:"url"`
	Key          string `json:"key"`
	Scene        string `json:"scene"`
	Visibility   string `json:"visibility"`
	URLExpiresIn int64  `json:"url_expires_in"`
}

// uploadExportToOSS 将已校验的 CSV 通过 OSS 公共上传接口保存。
// 桶选择、对象创建和下载签名由 OSS 负责，此处只负责协议适配与结果校验。
func uploadExportToOSS(ctx context.Context, baseURL, operatorCode, filename string, file *os.File, size int64) (*exportOSSResult, error) {
	endpoint, err := resourceurl.PublicURL(strings.TrimSpace(baseURL))
	if err != nil {
		return nil, xerr.Err(http.StatusServiceUnavailable, "OSS API address unavailable")
	}
	endpoint.Path = strings.TrimRight(endpoint.Path, "/") + memberExportUploadPath
	if size <= 0 || size > memberexport.MaxBytes {
		return nil, xerr.Err(http.StatusRequestEntityTooLarge, "CSV exceeds size limit")
	}
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return nil, xerr.InternalServerError("export file seek failed")
	}
	// 内存中只构造 multipart 首尾边界，CSV 内容从临时文件流式读取。
	var formBuffer bytes.Buffer
	writer := multipart.NewWriter(&formBuffer)
	if err = writer.WriteField("scene", memberExportScene); err != nil {
		return nil, xerr.InternalServerError("export form failed")
	}
	if _, err = writer.CreateFormFile("file", filename); err != nil {
		return nil, xerr.InternalServerError("export form failed")
	}
	formHeader := append([]byte(nil), formBuffer.Bytes()...)
	formBuffer.Reset()
	if err = writer.Close(); err != nil {
		return nil, xerr.InternalServerError("export form failed")
	}
	formFooter := append([]byte(nil), formBuffer.Bytes()...)
	body := io.MultiReader(bytes.NewReader(formHeader), io.LimitReader(file, size), bytes.NewReader(formFooter))
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), body)
	if err != nil {
		return nil, xerr.InternalServerError("export upload request failed")
	}
	request.ContentLength = int64(len(formHeader)) + size + int64(len(formFooter))
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("Authorization", "Bearer "+ctxdata.RawTokenFromCtx(ctx))
	// 显式选择分站 Core，原有总站上传请求不受影响。
	request.Header.Set(memberExportOSSAuthScopeHeader, memberExportOSSAuthScope)
	if ip := ctxdata.ClientIPFromCtx(ctx); ip != "" {
		request.Header.Set("X-Forwarded-For", ip)
	}
	// 禁止自动重定向，避免上传请求及登录令牌被转发到其他地址。
	uploadClient := &http.Client{Timeout: memberExportUploadTimeout, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	uploadResponse, err := uploadClient.Do(request)
	if err != nil {
		if ctx.Err() != nil {
			return nil, exportHTTPError(ctx, ctx.Err())
		}
		return nil, xerr.Err(http.StatusBadGateway, "OSS CSV upload failed")
	}
	defer uploadResponse.Body.Close()
	// 保留 OSS 的客户端错误；重定向和服务端错误统一映射为网关错误。
	if uploadResponse.StatusCode < http.StatusOK || uploadResponse.StatusCode >= http.StatusMultipleChoices {
		httpStatus := uploadResponse.StatusCode
		if httpStatus < http.StatusBadRequest || httpStatus >= http.StatusInternalServerError {
			httpStatus = http.StatusBadGateway
		}
		return nil, xerr.Err(httpStatus, "OSS CSV upload rejected")
	}
	// 有界读取并额外探测一个字节，拒绝超大或被截断后仍可解析的响应。
	responseBody, err := io.ReadAll(io.LimitReader(uploadResponse.Body, memberExportMaxUploadResponseBytes+memberExportResponseOverflowBytes))
	if err != nil || len(responseBody) > memberExportMaxUploadResponseBytes {
		return nil, xerr.Err(http.StatusBadGateway, "invalid OSS upload response")
	}
	var reply struct {
		Code *int             `json:"code"`
		Data *exportOSSResult `json:"data"`
	}
	// HTTP 成功不代表上传业务成功；同时检查业务码和当前租户的私有文件契约。
	if json.Unmarshal(responseBody, &reply) != nil || reply.Code == nil || *reply.Code != memberExportOSSBusinessSuccessCode || !validExportOSSResult(reply.Data, operatorCode) {
		return nil, xerr.Err(http.StatusBadGateway, "invalid private CSV upload result")
	}
	return reply.Data, nil
}

func validExportOSSResult(result *exportOSSResult, operatorCode string) bool {
	if result == nil || result.Scene != memberExportScene || result.Visibility != memberExportPrivateVisibility || result.URLExpiresIn <= 0 || result.URLExpiresIn > int64(memberExportMaxDownloadTTL/time.Second) {
		return false
	}
	// 将厅编码作为普通文本匹配，确保返回 key 属于当前厅且日期真实有效。
	keyPattern := regexp.MustCompile(`^` + regexp.QuoteMeta(operatorCode) + `/` + memberExportScene + `/(?P<` + memberExportKeyDateGroup + `>[0-9]{8})/[0-9a-f]{32}\.csv$`)
	keyMatches := keyPattern.FindStringSubmatch(result.Key)
	if keyMatches == nil {
		return false
	}
	if _, err := time.Parse(memberExportKeyDateLayout, keyMatches[keyPattern.SubexpIndex(memberExportKeyDateGroup)]); err != nil {
		return false
	}
	locator, err := url.Parse(result.URL)
	return err == nil && (locator.Scheme == "https" || locator.Scheme == "http") && locator.Host != "" && locator.User == nil && locator.Fragment == "" && strings.HasSuffix(locator.Path, "/"+result.Key)
}
