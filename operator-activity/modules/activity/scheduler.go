package promotion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"runtime/debug"
	"strings"

	"oa.98ent.com/p9/operator-activity/modules/activity/perr"
	"oa.98ent.com/p9/operator-activity/modules/activity/plugin"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type binding struct {
	plugin   plugin.Plugin
	dispatch plugin.DispatchFunc
}

// Scheduler 插件调度器：按插件名投递，统一 WrapDispatch
type Scheduler struct {
	serviceName string
	plugins     []plugin.Plugin
	byName      map[string]binding
}

func NewScheduler(serviceName string) *Scheduler {
	return &Scheduler{
		serviceName: serviceName,
		plugins:     make([]plugin.Plugin, 0),
		byName:      make(map[string]binding),
	}
}

// Register 注册插件，并统一包装 Dispatch
func (s *Scheduler) Register(p plugin.Plugin) {
	if p == nil {
		return
	}
	s.plugins = append(s.plugins, p)
	s.byName[p.Name()] = binding{plugin: p, dispatch: WrapDispatch(p)}
}

// Plugins 返回已注册插件
func (s *Scheduler) Plugins() []plugin.Plugin {
	return s.plugins
}

// Dispatch 按插件名投递 Kafka 消息。事件解析由插件自己完成。
func (s *Scheduler) Dispatch(ctx context.Context, pluginName string, payload types.Payload) error {
	b, ok := s.byName[pluginName]
	if !ok {
		return fmt.Errorf("scheduler: plugin %s not registered", pluginName)
	}
	if err := b.dispatch(ctx, payload); err != nil {
		err = classifyBuildError(err)
		return perr.Classify(fmt.Errorf("plugin %s dispatch: %w", b.plugin.Name(), err))
	}
	return nil
}

func classifyBuildError(err error) error {
	var syn *json.SyntaxError
	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &syn) || errors.As(err, &typeErr) {
		return perr.Permanent(err)
	}
	msg := err.Error()
	if strings.Contains(msg, "invalid character") || strings.Contains(msg, "unexpected end of JSON") {
		return perr.Permanent(err)
	}
	return err
}

// WrapDispatch 包装插件 Dispatch，防止 panic
func WrapDispatch(p plugin.Plugin) plugin.DispatchFunc {
	name := "unknown"
	if p != nil {
		name = p.Name()
	}
	return func(ctx context.Context, payload types.Payload) (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = perr.Retryable(fmt.Errorf("plugin %s panic: %v\n%s", name, r, debug.Stack()))
				logx.Error(err)
			}
		}()
		if p == nil {
			return nil
		}
		return p.Dispatch(ctx, payload)
	}
}
