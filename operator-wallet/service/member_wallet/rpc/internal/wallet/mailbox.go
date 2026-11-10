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

	if err := ctx.Err(); err != nil {
		return err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.closed {
		return ErrMailboxClosed
	}

	select {
	case m.queue <- message:
		return nil
	default:
		return ErrMailboxFull
	}
}

// dequeue 从队列中获取下一条消息
func (m *mailbox) dequeue(ctx context.Context) (any, error) {
	if m == nil {
		return nil, ErrMailboxClosed
	}

	if ctx == nil {
		return nil, errNilMailboxContext
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	select {
	case message, ok := <-m.queue:
		if !ok {
			return nil, ErrMailboxClosed
		}
		return message, nil

	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// close 关闭消息队列, 已入队消息仍可继续消费
func (m *mailbox) close() {
	if m == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return
	}

	m.closed = true
	close(m.queue)
}
