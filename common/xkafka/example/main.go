// 演示 oa.98ent.com/p9/common/xkafka 的生产与消费。
//
//	go run ./xkafka/example -b 127.0.0.1:19094
package main

import (
	"context"
	"flag"
	"os"
	"strings"
	"time"

	"oa.98ent.com/p9/common/xkafka"
	"oa.98ent.com/p9/common/xkafka/topic"

	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
)

// 示例专用 topic，避免打到业务队列。正式环境用 kafka/topic 常量。
const exampleTopic = "demo.common-kafka.example"

type helloData struct {
	UserID int64  `json:"user_id"`
	Msg    string `json:"msg"`
}

type helloExt struct {
	Source string `json:"source"`
}

func main() {
	brokersFlag := flag.String("b", "127.0.0.1:19094", "kafka brokers, comma separated")
	flag.Parse()

	logx.MustSetup(logx.LogConf{Mode: "console", Encoding: "plain"})
	brokers := splitBrokers(*brokersFlag)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	// 接收消息的通道
	got := make(chan helloData, 1)
	// 创建 kq 配置
	kqConf := kq.KqConf{
		ServiceConf: service.ServiceConf{Name: "common-kafka-example", Mode: "dev"},
		Brokers:     brokers,
		Group:       "common-kafka-example",
		Topic:       exampleTopic,
		Offset:      "last",
		Consumers:   1,
		Processors:  1,
		ForceCommit: true,
	}
	// 创建 kq 处理器
	kqHandler := func(_ context.Context, msg xkafka.Message) error {
		env, err := xkafka.DecodePayload([]byte(msg.Value))
		if err != nil {
			return err
		}
		var data helloData
		if err := env.UnmarshalData(&data); err != nil {
			return err
		}
		var ext helloExt
		_ = env.UnmarshalExt(&ext)
		logx.Infof("consumed type=%s v=%d time=%d partition=%d offset=%d key=%s data=%+v ext=%+v",
			env.Name, env.V, env.Timestamp, msg.Partition, msg.Offset, msg.Key, data, ext)
		select {
		case got <- data:
		default:
		}
		return nil
	}
	// 创建 kq 消费者
	kqConsume := xkafka.Consume(kqConf, kqHandler)
	defer xkafka.Stop(kqConsume)

	time.Sleep(5 * time.Second) // 等消费者加入 group

	producer := xkafka.NewProducerWithDef(brokers)
	defer producer.Close()

	payload, err := xkafka.NewPayloadExt(topic.NoticeSend, helloData{
		UserID: 10001,
		Msg:    "hello from common/xkafka",
	}, helloExt{Source: "common-kafka-example"})
	if err != nil {
		logx.Must(err)
	}
	if err := producer.PushPayloadWithKey(ctx, exampleTopic, "user-10001", payload); err != nil {
		logx.Errorf("push failed: %v", err)
		os.Exit(1)
	}
	logx.Info("produced payload")

	select {
	case data := <-got:
		logx.Infof("ok user_id=%d msg=%s\n", data.UserID, data.Msg)
	case <-ctx.Done():
		logx.Error("timeout waiting consume, check kafka broker and topic auto-create")
		os.Exit(1)
	}
}

func splitBrokers(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
