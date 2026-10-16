package middleware

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"oa.98ent.com/p9/common/ctxdata"
	coremw "oa.98ent.com/p9/core/common/middleware"
)

type recStub struct {
	ch chan coremw.ActionRecord
}

func (s recStub) RecordAction(_ context.Context, rec coremw.ActionRecord) {
	s.ch <- rec
}

func TestWrapActionLog_multipart记日志且不截断(t *testing.T) {
	payload := strings.Repeat("x", 20*1024)
	ch := make(chan coremw.ActionRecord, 1)
	mw := wrapActionLog(recStub{ch: ch})
	h := mw(func(w http.ResponseWriter, r *http.Request) {
		got, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != payload {
			t.Fatalf("got %d", len(got))
		}
		_, _ = w.Write([]byte(`{"code":0}`))
	})
	req := httptest.NewRequest(http.MethodPost, "/p9/oss/file/upload", strings.NewReader(payload))
	req.Header.Set("Content-Type", "multipart/form-data; boundary=----x")
	req = req.WithContext(ctxdata.WithClaims(req.Context(), &ctxdata.Claims{UserID: 7}))
	h(httptest.NewRecorder(), req)
	select {
	case rec := <-ch:
		if rec.UserID != 7 || rec.ActionResult != 1 || rec.RequestBody != multipartHint {
			t.Fatalf("%+v", rec)
		}
		if rec.RequestPath != "/p9/oss/file/upload" {
			t.Fatalf("path %s", rec.RequestPath)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}
}
