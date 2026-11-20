package consumer

import (
	"context"
	"strings"

	"oa.98ent.com/p9/operator-activity/internal/config"
	"oa.98ent.com/p9/operator-activity/modules/activity/perr"
	"oa.98ent.com/p9/operator-activity/modules/activity/plugin"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/queue"
	"oa.98ent.com/p9/common/xkafka"
)

// PluginDispatcher 调度器能力：列出已注册插件并按名投递。
type PluginDispatcher interface {
	Plugins() []plugin.Plugin
	Dispatch(ctx context.Context, pluginName string, payload types.Payload) error
}

// StartKq 按 PluginKq 为已注册插件启动独立 Group 消费者。
func StartKq(sched PluginDispatcher, c config.Config) []queue.MessageQueue {
	if sched == nil {
		logx.Error("[kq] scheduler is nil")
		return nil
	}
	if len(c.KqBrokers) == 0 {
		logx.Error("[kq] brokers is empty")
		return nil
	}

	// 获取Kafka消费配置
	pluginKqConf := GetPluginKqConf(strings.Replace(c.Name, "rpc", "engine", 1))

	var queues []queue.MessageQueue
	for _, p := range sched.Plugins() {
		pluginName := p.Name()
		confs := pluginKqConf[pluginName]
		if len(confs) == 0 {
			logx.Infof("[kq] plugin has no kafka config, skip plugin=%s", pluginName)
			continue
		}
		for _, conf := range confs {
			if len(conf.Brokers) == 0 {
				conf.Brokers = c.KqBrokers
			}
			topicName := conf.Topic
			q := xkafka.Consume(conf, func(ctx context.Context, msg xkafka.Message) error {
				logx.Debugf("[kq] consume message plugin=%s topic=%s partition=%d offset=%d key=%s value=%s",
					pluginName, topicName, msg.Partition, msg.Offset, msg.Key, msg.Value)
				// 解码 payload统一封装到 xkafka.Payload 中
				payload, err := xkafka.DecodePayload([]byte(msg.Value))
				if err != nil {
					logx.Errorf("[kq] decode payload error plugin=%s topic=%s partition=%d offset=%d err=%v",
						pluginName, topicName, msg.Partition, msg.Offset, err)
					return err
				}
				// 将xkafka.Payload 转换为 types.Payload
				payloadNew := types.Payload{
					ID:        payload.ID,
					Name:      payload.Name,
					Data:      payload.Data,
					Timestamp: payload.Timestamp,
					V:         payload.V,
					Ext:       payload.Ext,
				}
				// 分发消息给插件处理
				err = sched.Dispatch(ctx, pluginName, payloadNew)
				if err != nil {
					// 如果是永久错误，则跳过消息
					if perr.IsPermanent(err) {
						logx.Errorf("[kq] permanent error, skip message plugin=%s topic=%s partition=%d offset=%d err=%v",
							pluginName, topicName, msg.Partition, msg.Offset, err)
						return nil
					}
					logx.Errorf("[kq] retryable error plugin=%s topic=%s partition=%d offset=%d err=%v",
						pluginName, topicName, msg.Partition, msg.Offset, err)
					return err
				}
				return nil
			})
			if q != nil {
				queues = append(queues, q)
			}
		}
	}
	return queues
}

// StopKq 优雅停止所有消费者。
func StopKq(queues []queue.MessageQueue) {
	xkafka.StopMany(queues)
}
