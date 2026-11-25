package wallet

import (
	"context"
	"errors"
	"time"
)

// operationHandler 负责执行一笔钱包资金操作
type operationHandler struct {
	ctx            context.Context // 钱包服务生命周期 Context
	producer       eventProducer   // Kafka 资金事件生产者
	publishTimeout time.Duration   // 单次 Kafka 写入超时时间
}

// newOperationHandler 创建生产资金操作 Handler
func newOperationHandler(ctx context.Context, producer eventProducer, publishTimeout time.Duration) (actorHandler, error) {
	if ctx == nil {
		return nil, errors.New("create wallet operation handler: nil context")
	}
	if producer == nil {
		return nil, errors.New("create wallet operation handler: nil producer")
	}
	if publishTimeout <= 0 {
		return nil, errors.New("create wallet operation handler: invalid publish timeout")
	}

	h := &operationHandler{
		ctx:            ctx,
		producer:       producer,
		publishTimeout: publishTimeout,
	}

	return h.handle, nil
}

// handle 串行处理钱包资金操作和余额查询
func (h *operationHandler) handle(state *State, message any) {
	switch m := message.(type) {
	case operationMessage:
		h.handleOperation(state, m)

	case balanceQueryMessage:
		if m.reply == nil {
			return
		}

		m.reply <- balanceQueryReply{
			result: newBalanceSnapshot(*state),
		}
	}
}

// handleOperation 串行处理一笔资金操作
func (h *operationHandler) handleOperation(state *State, m operationMessage) {
	if m.reply == nil {
		return
	}

	// 校验消息所属 Actor
	if m.actor == nil {
		m.reply <- operationReply{err: ErrActorClosed}
		return
	}

	// 计算候选余额和资金流水, 不修改正式 State
	prepared, err := prepareRecord(*state, m.params)
	if err != nil {
		m.reply <- operationReply{err: err}
		return
	}

	// 构造完整资金事件
	event := newWalletEvent(*state, m.params, prepared)

	// 使用服务级 Context, 不受 RPC 请求提前取消影响
	ctx, cancel := context.WithTimeout(h.ctx, h.publishTimeout)
	err = h.producer.publish(ctx, m.partition, event)
	cancel()

	if err != nil {
		// Kafka 写入结果不明确时, 保留原始事件并终止当前钱包
		if errors.Is(err, ErrKafkaWriteUncertain) {
			m.actor.failOperation(event, err)
		}

		m.reply <- operationReply{err: err}
		return
	}

	// Kafka 已可靠确认, 当前 Worker 可以安全更新正式 State
	*state = prepared.nextState

	// 返回本次资金操作完成后的钱包状态
	m.reply <- operationReply{
		result: event.operationResult(),
	}
}
