package middleware

import (
	"bytes"
	"context"
	"mime"
	"net/http"
	"strings"
	"time"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/utils"
	"oa.98ent.com/p9/core/common/coreadapt"
	"oa.98ent.com/p9/core/common/jwt"
	coremw "oa.98ent.com/p9/core/common/middleware"
	"oa.98ent.com/p9/core/rpc/coreclient"

	"github.com/zeromicro/go-zero/rest"
)

const maxLogBody = 16 * 1024

const multipartHint = `{"content_type":"multipart/form-data"}`

// ActionLogFromCore 有 core 时记操作日志；否则放行。
func ActionLogFromCore(cli coreclient.Core) rest.Middleware {
	if cli == nil {
		return passthrough
	}
	return wrapActionLog(coreadapt.ActionRecorder(cli))
}

func wrapActionLog(rec coremw.ActionRecorder) rest.Middleware {
	if rec == nil {
		return passthrough
	}
	jsonLog := coremw.ActionLog(rec)
	noBody := actionLogNoBody(rec)
	return func(next http.HandlerFunc) http.HandlerFunc {
		logged := jsonLog(next)
		plain := noBody(next)
		return func(w http.ResponseWriter, r *http.Request) {
			if isMultipartRequest(r) {
				plain(w, r)
				return
			}
			logged(w, r)
		}
	}
}

func isMultipartRequest(r *http.Request) bool {
	if r == nil {
		return false
	}
	media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	return err == nil && strings.HasPrefix(media, "multipart/")
}

// actionLogNoBody 记写操作但不读请求体，避免截断 multipart 上传。
func actionLogNoBody(rec coremw.ActionRecorder) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if rec == nil || !coremw.IsAdminWrite(r.Method, r.URL.Path) {
				next(w, r)
				return
			}
			cw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			start := time.Now()
			next(cw, r)
			emitAction(r, rec, cw, start)
		}
	}
}

func emitAction(r *http.Request, rec coremw.ActionRecorder, cw *statusWriter, start time.Time) {
	claims := ctxdata.ClaimsFromCtx(r.Context())
	if claims == nil || claims.UserID == 0 || claims.TokenType == jwt.TokenPreview {
		return
	}
	status := cw.status
	result := int16(2)
	if status >= 200 && status < 300 {
		result = 1
	}
	record := coremw.ActionRecord{
		UserID:         claims.UserID,
		RequestMethod:  r.Method,
		RequestPath:    r.URL.Path,
		RequestQuery:   r.URL.RawQuery,
		RequestBody:    multipartHint,
		ActionResult:   result,
		ResponseStatus: status,
		ResponseBody:   coremw.MaskJSON(clipLog(cw.body.Bytes())),
		DurationMS:     int(time.Since(start).Milliseconds()),
		ClientIP:       utils.ClientIP(r),
		UserAgent:      utils.UserAgent(r),
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		rec.RecordAction(ctxdata.WithClaims(ctx, claims), record)
	}()
}

func clipLog(b []byte) string {
	if len(b) <= maxLogBody {
		return string(b)
	}
	return string(b[:maxLogBody])
}

type statusWriter struct {
	http.ResponseWriter
	status int
	body   bytes.Buffer
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	if w.body.Len() < maxLogBody {
		remain := maxLogBody - w.body.Len()
		if len(p) > remain {
			_, _ = w.body.Write(p[:remain])
		} else {
			_, _ = w.body.Write(p)
		}
	}
	return w.ResponseWriter.Write(p)
}
