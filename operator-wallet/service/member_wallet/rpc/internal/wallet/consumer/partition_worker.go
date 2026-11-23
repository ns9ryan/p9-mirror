package consumer

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/twmb/franz-go/pkg/kgo"
)

var (
	errInvalidPartition       = errors.New("wallet consumer: invalid partition")
	errInvalidWorkerCapacity  = errors.New("wallet consumer: invalid worker capacity")
	errNilRecord              = errors.New("wallet consumer: nil record")
	errPartitionMismatch      = errors.New("wallet consumer: partition mismatch")
	errPartitionWorkerStopped = errors.New("wallet consumer: partition worker stopped")
)

// partitionWorker 串行处理指定 Partition 的资金事件
type partitionWorker struct {
	partition int32            // 所属 Kafka Partition
	queue     chan *kgo.Record // 有界 FIFO 消息队列
	handler   Handler          // 资金事件处理函数
	ctx       context.Context  // Worker 执行上下文

	mu      sync.Mutex    // 保护停止状态和入队操作
	stopped bool          // 是否停止接收新消息
	stopCh  chan struct{} // 停止接收信号
	spaceCh chan struct{} // 队列释放容量通知
	done    chan struct{} // Worker 退出信号
	runErr  error         // Worker 退出错误
}

// newPartitionWorker 创建并启动指定 Partition 的 Worker
func newPartitionWorker(ctx context.Context, partition int32, capacity int, handler Handler) (*partitionWorker, error) {
	if ctx == nil {
		return nil, errNilContext
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if partition < 0 {
		return nil, errInvalidPartition
	}
	if capacity <= 0 {
		return nil, errInvalidWorkerCapacity
	}
	if handler == nil {
		return nil, errNilHandler
	}

	w := &partitionWorker{
		partition: partition,                        // 所属 Kafka Partition
		queue:     make(chan *kgo.Record, capacity), // 有界 FIFO 消息队列
		handler:   handler,                          // 资金事件处理函数
		ctx:       ctx,                              // Worker 执行上下文
		stopCh:    make(chan struct{}),              // 停止接收信号
		spaceCh:   make(chan struct{}, 1),           // 队列释放容量通知
		done:      make(chan struct{}),              // Worker 退出信号
	}

	// 按入队顺序串行处理当前 Partition 的消息
	go w.run()

	return w, nil
}

// enqueue 将消息加入有界队列, 队列满时等待可用容量
func (w *partitionWorker) enqueue(ctx context.Context, record *kgo.Record) error {
	if ctx == nil {
		return errNilContext
	}
	if record == nil {
		return errNilRecord
	}
	if record.Partition != w.partition {
		return fmt.Errorf("%w: expected=%d, actual=%d", errPartitionMismatch, w.partition, record.Partition)
	}

	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := w.ctx.Err(); err != nil {
			return errPartitionWorkerStopped
		}

		w.mu.Lock()

		// 已关闭
		if w.stopped {
			w.mu.Unlock()
			return errPartitionWorkerStopped
		}

		select {
		case w.queue <- record:
			w.mu.Unlock()
			return nil

		default:
			w.mu.Unlock()
		}

		// 队列已满, 等待 Worker 取走消息或停止
		select {
		case <-w.spaceCh:
		case <-w.stopCh:
			return errPartitionWorkerStopped
		case <-w.ctx.Done():
			return errPartitionWorkerStopped
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// run 按入队顺序串行处理当前 Partition 的消息
func (w *partitionWorker) run() {
	defer func() {
		w.stop()
		close(w.done)
	}()

	for {
		// Worker Context 取消时退出, 未处理消息保留在 Kafka 中
		if err := w.ctx.Err(); err != nil {
			w.setRunErr(err)
			return
		}

		select {
		case record := <-w.queue:
			if err := w.handle(record); err != nil {
				w.setRunErr(err)
				return
			}

		case <-w.stopCh:
			// 正常停止时排空已经接受的消息
			for {
				if err := w.ctx.Err(); err != nil {
					w.setRunErr(err)
					return
				}

				select {
				case record := <-w.queue:
					if err := w.handle(record); err != nil {
						w.setRunErr(err)
						return
					}

				default:
					return
				}
			}

		case <-w.ctx.Done():
			w.setRunErr(w.ctx.Err())
			return
		}
	}
}

// handle 执行一条消息, 并通知等待入队的调用方
func (w *partitionWorker) handle(record *kgo.Record) error {
	select {
	case w.spaceCh <- struct{}{}:
	default:
	}

	if err := w.handler(w.ctx, record); err != nil {
		return fmt.Errorf("handle partition=%d offset=%d: %w", w.partition, record.Offset, err)
	}

	return nil
}

// stop 停止接收新消息, 已入队消息由 Worker 正常排空
func (w *partitionWorker) stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 已关闭直接返回
	if w.stopped {
		return
	}

	w.stopped = true
	close(w.stopCh)
}

// wait 等待 Worker 退出, Context 只控制等待过程
func (w *partitionWorker) wait(ctx context.Context) error {
	if ctx == nil {
		return errNilContext
	}

	select {
	case <-w.done:
		w.mu.Lock()
		defer w.mu.Unlock()
		return w.runErr

	case <-ctx.Done():
		return ctx.Err()
	}
}

// setRunErr 保存 Worker 的退出错误
func (w *partitionWorker) setRunErr(err error) {
	w.mu.Lock()
	w.runErr = err
	w.mu.Unlock()
}
