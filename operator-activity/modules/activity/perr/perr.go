package perr

import (
	"errors"
	"fmt"
	"strings"

	"oa.98ent.com/p9/operator-activity/ent"
)

// Kind 错误可重试性
type Kind int

const (
	KindRetryable Kind = iota + 1 // 瞬时故障，应重试
	KindPermanent                 // 业务/数据问题，重试无意义，应跳过
)

type classifiedError struct {
	kind Kind
	err  error
}

func (e *classifiedError) Error() string {
	if e == nil || e.err == nil {
		return ""
	}
	prefix := "retryable"
	if e.kind == KindPermanent {
		prefix = "permanent"
	}
	return fmt.Sprintf("%s: %v", prefix, e.err)
}

func (e *classifiedError) Unwrap() error { return e.err }

// Retryable 标记为可重试错误
func Retryable(err error) error {
	if err == nil {
		return nil
	}
	var ce *classifiedError
	if errors.As(err, &ce) {
		return err
	}
	return &classifiedError{kind: KindRetryable, err: err}
}

// Permanent 标记为不可重试错误
func Permanent(err error) error {
	if err == nil {
		return nil
	}
	var ce *classifiedError
	if errors.As(err, &ce) {
		return err
	}
	return &classifiedError{kind: KindPermanent, err: err}
}

// Retryablef / Permanentf 格式化包装
func Retryablef(format string, args ...any) error {
	return Retryable(fmt.Errorf(format, args...))
}

func Permanentf(format string, args ...any) error {
	return Permanent(fmt.Errorf(format, args...))
}

// IsPermanent 是否不可重试
func IsPermanent(err error) bool {
	var ce *classifiedError
	if errors.As(err, &ce) {
		return ce.kind == KindPermanent
	}
	return false
}

// IsRetryable 是否可重试（未分类默认视为可重试，避免丢消息）
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	return !IsPermanent(err)
}

// Classify 根据常见错误类型自动归类；已分类则原样返回
func Classify(err error) error {
	if err == nil {
		return nil
	}
	var ce *classifiedError
	if errors.As(err, &ce) {
		return err
	}
	if ent.IsNotFound(err) {
		return Permanent(err)
	}
	if ent.IsConstraintError(err) || isDuplicateKey(err) {
		return Permanent(err)
	}
	return Retryable(err)
}

func isDuplicateKey(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "Duplicate entry") ||
		strings.Contains(msg, "UNIQUE constraint failed") ||
		strings.Contains(msg, "Error 1062") ||
		strings.Contains(msg, "duplicate key") ||
		strings.Contains(msg, "23505")
}
