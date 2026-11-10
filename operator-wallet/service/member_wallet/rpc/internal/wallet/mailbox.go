// ============================================================================
// 钱包消息队列
// ============================================================================

package wallet

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var (
	// ErrMailboxFull 表示钱包消息队列已满
	ErrMailboxFull = errors.New("wallet mailbox is full") // 队列已达到容量上限

	// ErrMailboxClosed 表示钱包消息队列已关闭
	ErrMailboxClosed = errors.New("wallet mailbox is closed") // 队列不再接收新消息

	errNilMailboxMessage = errors.New("wallet mailbox message is nil") // 入队消息不能为空
	errNilMailboxContext = errors.New("wallet mailbox context is nil") // 上下文不能为空
)

// mailbox 是单个钱包 Actor 的有界 FIFO 消息队列
type mailbox struct {
	queue  chan any     // 消息队列
	mu     sync.RWMutex // 保护队列关闭状态
	closed bool         // 是否已关闭
}

// newMailbox 创建指定容量的钱包消息队列
func newMailbox(capacity int) (*mailbox, error) {
	if capacity <= 0 {
		return nil, fmt.Errorf("create wallet mailbox: invalid capacity: %d", capacity)
	}

	return &mailbox{
		queue: make(chan any, capacity),
	}, nil
}

// enqueue 将消息放入队列, 队列已满时立即返回
func (m *mailbox) enqueue(ctx context.Context, message any) error {
	if m == nil {
		return ErrMailboxClosed
	}

	if ctx == nil {
		return errNilMailboxContext
	}

	if message == nil {
		return errNilMailboxMessage
	}

	// 提前检查已取消的上下文, 避免继续执行入队操作
	if err := ctx.Err(); err != nil {
		return err
	}

	// 读锁允许多个入队操作并发执行, 同时阻止队列在入队期间被关闭
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.closed {
		return ErrMailboxClosed
	}

	select {
	case <-ctx.Done():
		// 处理预检查后发生的上下文取消
		return ctx.Err()

	case m.queue <- message:
		return nil

	default:
		// 使用非阻塞入队, 避免队列已满时阻塞调用方
		return ErrMailboxFull
	}
}

// dequeue 按 FIFO 顺序获取下一条消息
func (m *mailbox) dequeue(ctx context.Context) (any, error) {
	if m == nil {
		return nil, ErrMailboxClosed
	}

	if ctx == nil {
		return nil, errNilMailboxContext
	}

	// 提前检查已取消的上下文, 避免继续等待消息
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// 等待消息到达或上下文取消
	select {
	case message, ok := <-m.queue:
		if !ok {
			// 队列已关闭且没有剩余消息
			return nil, ErrMailboxClosed
		}
		return message, nil

	case <-ctx.Done():
		// 等待期间响应上下文取消
		return nil, ctx.Err()
	}
}

// close 关闭消息队列, 已入队消息仍可继续消费
func (m *mailbox) close() {
	if m == nil {
		return
	}

	// 写锁与入队操作的读锁互斥, 确保关闭时没有入队操作正在执行
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return
	}

	m.closed = true
	close(m.queue)
}
