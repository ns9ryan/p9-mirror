package perr_test

import (
	"errors"
	"testing"

	"oa.98ent.com/p9/operator-activity/ent"

	"oa.98ent.com/p9/operator-activity/modules/activity/perr"
)

func TestRetryablePermanent(t *testing.T) {
	r := perr.Retryable(errors.New("db down"))
	p := perr.Permanent(errors.New("bad json"))

	if !perr.IsRetryable(r) || perr.IsPermanent(r) {
		t.Fatalf("retryable misclassified: %v", r)
	}
	if !perr.IsPermanent(p) || perr.IsRetryable(p) {
		t.Fatalf("permanent misclassified: %v", p)
	}
}

func TestClassify(t *testing.T) {
	if !perr.IsPermanent(perr.Classify(&ent.NotFoundError{})) {
		t.Fatal("record not found should be permanent")
	}
	if !perr.IsPermanent(perr.Classify(&ent.ConstraintError{})) {
		t.Fatal("constraint error should be permanent")
	}
	if !perr.IsPermanent(perr.Classify(errors.New("duplicate key value violates unique constraint"))) {
		t.Fatal("postgres duplicate key should be permanent")
	}
	if !perr.IsRetryable(perr.Classify(errors.New("connection reset"))) {
		t.Fatal("unknown should default retryable")
	}
	p := perr.Permanent(errors.New("x"))
	if got := perr.Classify(p); !perr.IsPermanent(got) {
		t.Fatal("classify should keep permanent")
	}
}

func TestUnwrap(t *testing.T) {
	inner := errors.New("inner")
	err := perr.Retryable(inner)
	if !errors.Is(err, inner) {
		t.Fatal("should unwrap to inner")
	}
}
