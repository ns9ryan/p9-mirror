package svc

import (
	"context"

	"oa.98ent.com/p9/common/errorlog"
)

// Only export request/response payloads are removed; unrelated logging is unchanged.
type exportErrorRecorder struct{ errorlog.Recorder }

func (r exportErrorRecorder) RecordError(ctx context.Context, in errorlog.Record) {
	if in.RequestPath == "/member/export/v2" {
		in.RequestBody = ""
		in.RequestQuery = ""
		in.ResponseBody = ""
		in.Detail = ""
	}
	r.Recorder.RecordError(ctx, in)
}
