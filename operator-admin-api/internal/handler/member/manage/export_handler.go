// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package manage

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/zeromicro/go-zero/rest/httpx"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-admin-api/internal/logic/member/manage"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
)

func ExportHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 16*1024)
		req, err := decodeExportRequest(r.Body)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := manage.NewExportLogic(r.Context(), svcCtx)
		resp, err := l.Export(req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

// decodeExportRequest 在类型解码前拒绝重复字段、未知字段及 null，避免静默覆盖筛选条件。
func decodeExportRequest(body io.Reader) (*types.ExportMembersRequest, error) {
	invalidRequestError := func() error { return xerr.BadRequest("invalid export JSON or unsupported filter") }
	decoder := json.NewDecoder(body)
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return nil, invalidRequestError()
	}
	requestFields := map[string]json.RawMessage{}
	allowedFieldNames := map[string]bool{}
	requestType := reflect.TypeOf(types.ExportMembersRequest{})
	for fieldIndex := 0; fieldIndex < requestType.NumField(); fieldIndex++ {
		fieldName, _, _ := strings.Cut(requestType.Field(fieldIndex).Tag.Get("json"), ",")
		allowedFieldNames[fieldName] = true
	}
	for decoder.More() {
		token, err = decoder.Token()
		if err != nil {
			return nil, invalidRequestError()
		}
		fieldName, ok := token.(string)
		if !ok {
			return nil, invalidRequestError()
		}
		if _, exists := requestFields[fieldName]; exists {
			return nil, invalidRequestError()
		}
		if !allowedFieldNames[fieldName] {
			return nil, invalidRequestError()
		}
		var fieldValue json.RawMessage
		if err = decoder.Decode(&fieldValue); err != nil {
			return nil, invalidRequestError()
		}
		if fieldName == "range_today_deposit" || fieldName == "range_today_withdraw" {
			return nil, xerr.BadRequest("today amount filters are unavailable: no daily statistics source")
		}
		if bytes.Equal(fieldValue, []byte("null")) {
			return nil, invalidRequestError()
		}
		requestFields[fieldName] = fieldValue
	}
	if _, err = decoder.Token(); err != nil {
		return nil, invalidRequestError()
	}
	// 对象结束后不允许追加第二个 JSON 值或其他数据。
	if decoder.Decode(new(any)) != io.EOF {
		return nil, invalidRequestError()
	}
	requestJSON, err := json.Marshal(requestFields)
	if err != nil {
		return nil, invalidRequestError()
	}
	decoder = json.NewDecoder(bytes.NewReader(requestJSON))
	decoder.DisallowUnknownFields()
	var request types.ExportMembersRequest
	if err = decoder.Decode(&request); err != nil {
		return nil, invalidRequestError()
	}
	return &request, nil
}
