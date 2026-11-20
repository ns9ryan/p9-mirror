package pipeline

import (
	"context"
	"fmt"
	"strings"

	"oa.98ent.com/p9/operator-activity/modules/activity/types"

	"github.com/zeromicro/go-zero/core/logx"
)

// Result 责任链评估结果
type Result int

const (
	ResultPass  Result = 1
	ResultSkip  Result = 2
	ResultError Result = 3
)

func (r Result) String() string {
	switch r {
	case ResultPass:
		return "pass"
	case ResultSkip:
		return "skip"
	case ResultError:
		return "error"
	default:
		return "unknown"
	}
}

// EvalContext 评估上下文，贯穿整条责任链和策略
type EvalContext struct {
	Context  context.Context
	Activity *types.Activity
	Event    types.ActivityEvent
	Member   *types.MemberInfo
	Round    int
	Window   types.TimeWindow
	Metric   *types.TriggerMetric
	StepLog  *StepLogger
}

// Step 单步执行记录
type Step struct {
	Name   string
	Result string
	Detail string
}

// StepLogger 步骤日志：记录每步结果与判断值，并即时打印
type StepLogger struct {
	activityID int64
	memberID   int64
	steps      []Step
}

func NewStepLogger(activityID, memberID int64) *StepLogger {
	return &StepLogger{activityID: activityID, memberID: memberID}
}

// Step 记录一步：result 为 pass/skip/error/info，detail 为判断条件值
func (s *StepLogger) Step(name, result, detail string) {
	if s == nil {
		return
	}
	s.steps = append(s.steps, Step{Name: name, Result: result, Detail: detail})
	if detail == "" {
		logx.Infof("[activity] activity=%d member=%d step=%s result=%s",
			s.activityID, s.memberID, name, result)
		return
	}
	logx.Infof("[activity] activity=%d member=%d step=%s result=%s %s",
		s.activityID, s.memberID, name, result, detail)
}

// Stepf 格式化 detail 后记录
func (s *StepLogger) Stepf(name, result, format string, args ...any) {
	s.Step(name, result, fmt.Sprintf(format, args...))
}

// Cond 记录条件评估结果
func (s *StepLogger) Cond(name string, result Result, format string, args ...any) {
	s.Stepf(name, result.String(), format, args...)
}

// Info 兼容旧调用
func (s *StepLogger) Info(cond, action, msg string) {
	detail := msg
	if action != "" {
		detail = fmt.Sprintf("action=%s %s", action, msg)
	}
	s.Step(cond, "info", detail)
}

// Skip 兼容旧调用
func (s *StepLogger) Skip(cond, msg string) {
	s.Step(cond, ResultSkip.String(), msg)
}

// Steps 返回已记录步骤
func (s *StepLogger) Steps() []Step {
	if s == nil {
		return nil
	}
	return s.steps
}

// Flush 打印完整步骤摘要
func (s *StepLogger) Flush() {
	if s == nil || len(s.steps) == 0 {
		return
	}
	parts := make([]string, 0, len(s.steps))
	for i, st := range s.steps {
		if st.Detail == "" {
			parts = append(parts, fmt.Sprintf("%d.%s=%s", i+1, st.Name, st.Result))
			continue
		}
		parts = append(parts, fmt.Sprintf("%d.%s=%s(%s)", i+1, st.Name, st.Result, st.Detail))
	}
	logx.Infof("[activity] activity=%d member=%d summary %s",
		s.activityID, s.memberID, strings.Join(parts, " | "))
}
