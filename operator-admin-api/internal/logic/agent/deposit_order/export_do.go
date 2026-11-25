package deposit_order

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
	"oa.98ent.com/p9/platform-base/pkg/resourceurl"
)

// ossUploadParams 通用 CSV 上传参数。
// 所有与具体业务相关的常量均由调用方填入，UploadCSVToOSS 不直接引用任何业务常量。
type ossUploadParams struct {
	// 基础上传参数
	BaseURL               string
	Filename              string
	Scene                 string
	MaxBytes              int64
	UploadPath            string        // OSS 上传端点路径，如 /p9/oss/file/upload
	AuthScopeHeader       string        // 鉴权 scope 请求头名称
	AuthScope             string        // 鉴权 scope 值
	UploadTimeout         time.Duration // HTTP 上传客户端超时
	BusinessSuccessCode   int           // OSS 业务成功码
	MaxResponseBytes      int64         // OSS 响应体读取上限
	ResponseOverflowBytes int64         // 额外读取字节数，用于探测响应超限
	// 结果校验参数（供 validOSSResult 使用）
	Visibility     string        // 上传后文件可见性，如 private
	MaxDownloadTTL time.Duration // 下载签名最大有效期
	KeyDateLayout  string        // OSS key 中日期段的格式，如 20060102
	KeyDateGroup   string        // OSS key 正则中日期命名组名
	// 结果校验参数（供 validDepositExportOSSResult 使用）
	OperatorCode string // 运营商 code，用于校验 OSS key 路径前缀
}

type OssUploadResult struct {
	URL          string `json:"url"`
	Key          string `json:"key"`
	Scene        string `json:"scene"`
	Visibility   string `json:"visibility"`
	URLExpiresIn int64  `json:"url_expires_in"`
}

// UploadCSVToOSS 通用 CSV 上传：将文件通过 OSS 公共上传接口保存，返回原始结果供调用方校验。
func UploadCSVToOSS(ctx context.Context, params ossUploadParams, file *os.File, size int64) (*OssUploadResult, error) {
	endpoint, err := resourceurl.PublicURL(strings.TrimSpace(params.BaseURL))
	if err != nil {
		return nil, xerr.Err(http.StatusServiceUnavailable, "OSS API address unavailable")
	}
	endpoint.Path = strings.TrimRight(endpoint.Path, "/") + params.UploadPath
	if size <= 0 || size > params.MaxBytes {
		return nil, xerr.Err(http.StatusRequestEntityTooLarge, "CSV exceeds size limit")
	}
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return nil, xerr.InternalServerError("export file seek failed")
	}

	var formBuffer bytes.Buffer
	mw := multipart.NewWriter(&formBuffer)
	if err = mw.WriteField("scene", params.Scene); err != nil {
		return nil, xerr.InternalServerError("export form failed")
	}
	if _, err = mw.CreateFormFile("file", params.Filename); err != nil {
		return nil, xerr.InternalServerError("export form failed")
	}
	formHeader := append([]byte(nil), formBuffer.Bytes()...)
	formBuffer.Reset()
	if err = mw.Close(); err != nil {
		return nil, xerr.InternalServerError("export form failed")
	}
	formFooter := append([]byte(nil), formBuffer.Bytes()...)
	body := io.MultiReader(bytes.NewReader(formHeader), io.LimitReader(file, size), bytes.NewReader(formFooter))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), body)
	if err != nil {
		return nil, xerr.InternalServerError("export upload request failed")
	}
	req.ContentLength = int64(len(formHeader)) + size + int64(len(formFooter))
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+ctxdata.RawTokenFromCtx(ctx))
	req.Header.Set(params.AuthScopeHeader, params.AuthScope)
	if ip := ctxdata.ClientIPFromCtx(ctx); ip != "" {
		req.Header.Set("X-Forwarded-For", ip)
	}

	uploadClient := &http.Client{
		Timeout: params.UploadTimeout,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	uploadResp, err := uploadClient.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, xerr.Err(http.StatusGatewayTimeout, "deposit export OSS upload timeout")
		}
		return nil, xerr.Err(http.StatusBadGateway, "OSS CSV upload failed")
	}
	defer uploadResp.Body.Close()

	if uploadResp.StatusCode < http.StatusOK || uploadResp.StatusCode >= http.StatusMultipleChoices {
		httpStatus := uploadResp.StatusCode
		if httpStatus < http.StatusBadRequest || httpStatus >= http.StatusInternalServerError {
			httpStatus = http.StatusBadGateway
		}
		return nil, xerr.Err(httpStatus, "OSS CSV upload rejected")
	}

	responseBody, err := io.ReadAll(io.LimitReader(uploadResp.Body, params.MaxResponseBytes+params.ResponseOverflowBytes))
	if err != nil || len(responseBody) > int(params.MaxResponseBytes) {
		return nil, xerr.Err(http.StatusBadGateway, "invalid OSS upload response")
	}
	var reply struct {
		Code *int             `json:"code"`
		Data *OssUploadResult `json:"data"`
	}
	if json.Unmarshal(responseBody, &reply) != nil || reply.Code == nil || *reply.Code != params.BusinessSuccessCode {
		return nil, xerr.Err(http.StatusBadGateway, "invalid OSS upload response")
	}
	return reply.Data, nil
}

func ValidOSSResult(result *OssUploadResult, params ossUploadParams) bool {
	if result == nil || result.Scene != params.Scene || result.Visibility != params.Visibility || result.URLExpiresIn <= 0 || result.URLExpiresIn > int64(params.MaxDownloadTTL/time.Second) {
		return false
	}
	keyPattern := regexp.MustCompile(`^` + regexp.QuoteMeta(params.OperatorCode) + `/` + regexp.QuoteMeta(params.Scene) + `/(?P<` + params.KeyDateGroup + `>[0-9]{8})/[0-9a-f]{32}\.csv$`)
	keyMatches := keyPattern.FindStringSubmatch(result.Key)
	if keyMatches == nil {
		return false
	}
	if _, err := time.Parse(params.KeyDateLayout, keyMatches[keyPattern.SubexpIndex(params.KeyDateGroup)]); err != nil {
		return false
	}
	locator, err := url.Parse(result.URL)
	return err == nil && (locator.Scheme == "https" || locator.Scheme == "http") && locator.Host != "" && locator.User == nil && locator.Fragment == "" && strings.HasSuffix(locator.Path, "/"+result.Key)
}
