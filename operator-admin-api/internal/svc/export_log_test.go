package svc

import (
	"context"
	"oa.98ent.com/p9/common/errorlog"
	"oa.98ent.com/p9/core/common/middleware"
	"testing"
)

type exportLogCapture struct {
	action  middleware.ActionRecord
	failure errorlog.Record
}

func (c *exportLogCapture) RecordAction(_ context.Context, in middleware.ActionRecord) { c.action = in }
func (c *exportLogCapture) RecordError(_ context.Context, in errorlog.Record)          { c.failure = in }
func TestExportPayloadsRemovedFromLogs(t *testing.T) {
	recorder := &exportLogCapture{}
	contactActionRecorder{recorder}.RecordAction(context.Background(), middleware.ActionRecord{RequestPath: "/member/export/v2", RequestQuery: "filter=private", RequestBody: "sensitive-filter", ResponseBody: "signed-download-url"})
	if recorder.action.RequestBody != "" || recorder.action.RequestQuery != "" || recorder.action.ResponseBody != "" {
		t.Fatal("export payload logged")
	}
	exportErrorRecorder{recorder}.RecordError(context.Background(), errorlog.Record{RequestPath: "/member/export/v2", RequestQuery: "private", RequestBody: "private", ResponseBody: "private", Detail: "private"})
	if recorder.failure.RequestBody != "" || recorder.failure.RequestQuery != "" || recorder.failure.ResponseBody != "" || recorder.failure.Detail != "" {
		t.Fatal("export error payload logged")
	}
	contactActionRecorder{recorder}.RecordAction(context.Background(), middleware.ActionRecord{RequestPath: "/unrelated", RequestBody: "unchanged"})
	if recorder.action.RequestBody != "unchanged" {
		t.Fatal("unrelated log behavior changed")
	}
}
