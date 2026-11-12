// ============================================================================
// 钱包 Shard 管理与生命周期
// ============================================================================

package wallet

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

var (
	// ErrShardClosed 表示 Shard 已停止接收新的钱包请求
	ErrShardClosed = errors.New("wallet shard is closed")

	// ErrShardNotClosed 表示 Shard 尚未进入关闭流程
	ErrShardNotClosed = errors.New("wallet shard is not closed")

	// ErrShardWalletIdentityMismatch 表示同一个钱包标识对应的归属信息发生冲突
	ErrShardWalletIdentityMismatch = errors.New("wallet shard identity mismatch")

	errNilShard        = errors.New("wallet shard is nil")
	errNilShardContext = errors.New("wallet shard context is nil")
)

// shardActor 保存 Actor 及其不可变的钱包身份, 避免重复初始化时混用钱包
// 余额和版本不保存在此处, 它们只由 Actor Handler 管理
type shardActor struct {
	actor *actor // 当前钱包唯一的 Actor

	walletID       int64     // 数据库钱包ID
	memberGlobalID uuid.UUID // 会员全局唯一标识
	operatorCode   string    // 所属厅业务编码
	currencyCode   string    // 钱包结算币种编码
}

// shard 管理一个固定分片中的钱包 Actor
// mu 保护 Actor Map 与关闭状态, 不参与 Actor 内部的消息处理
type shard struct {
	index           int          // 固定 Shard 下标, 后续对应 Kafka Partition
	pool            *workerPool  // 共享 Worker Pool
	mailboxCapacity int          // 每个 Actor 的 Mailbox 容量
	handler         actorHandler // 钱包消息处理逻辑

	mu     sync.Mutex               // 保护 actors 和 closed
	actors map[uuid.UUID]shardActor // WalletGlobalID 到 Actor 的映射
	closed bool                     // 是否已停止创建和查找 Actor
}

// newShard 创建固定 Shard, 不提前创建钱包 Actor
func newShard(index int, pool *workerPool, mailboxCapacity int, handler actorHandler) (*shard, error) {
	if index < 0 {
		return nil, fmt.Errorf("create wallet shard: invalid index: %d", index)
	}
	if pool == nil {
		return nil, errors.New("create wallet shard: worker pool is nil")
	}
	if mailboxCapacity <= 0 {
		return nil, fmt.Errorf("create wallet shard: invalid mailbox capacity: %d", mailboxCapacity)
	}
	if handler == nil {
		return nil, errors.New("create wallet shard: handler is nil")
	}

	return &shard{
		index:           index,
		pool:            pool,
		mailboxCapacity: mailboxCapacity,
		handler:         handler,
		actors:          make(map[uuid.UUID]shardActor),
	}, nil
}

// get 查找已有 Actor, 不存在时返回 nil, nil
func (s *shard) get(walletGlobalID uuid.UUID) (*actor, error) {
	if s == nil {
		return nil, errNilShard
	}
	if walletGlobalID == uuid.Nil {
		return nil, ErrInvalidWalletGlobalID
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil, ErrShardClosed
	}

	entry, ok := s.actors[walletGlobalID]
	if !ok {
		return nil, nil
	}

	return entry.actor, nil
}

// getOrCreate 查找或创建钱包 Actor, 同一钱包最多只创建一个 Actor
// 首次创建使用传入的 State, 已存在时不覆盖 Actor 的运行时余额和版本
func (s *shard) getOrCreate(state State) (*actor, error) {
	if s == nil {
		return nil, errNilShard
	}
	if err := state.Validate(); err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil, ErrShardClosed
	}

	if entry, ok := s.actors[state.WalletGlobalID]; ok {
		// 同一个全局钱包标识不能对应不同的厅、会员、币种或数据库钱包
		if entry.walletID != state.WalletID ||
			entry.memberGlobalID != state.MemberGlobalID ||
			entry.operatorCode != state.OperatorCode ||
			entry.currencyCode != state.CurrencyCode {
			return nil, fmt.Errorf("%w: wallet_global_id=%s", ErrShardWalletIdentityMismatch, state.WalletGlobalID)
		}

		return entry.actor, nil
	}

	// 持有 mu 完成创建和登记, 防止并发请求重复创建 Actor
	a, err := newActor(s.pool, state, s.mailboxCapacity, s.handler)
	if err != nil {
		return nil, fmt.Errorf("create shard wallet actor: %w", err)
	}

	s.actors[state.WalletGlobalID] = shardActor{
		actor:          a,                    // 当前钱包唯一的 Actor
		walletID:       state.WalletID,       // 数据库钱包ID
		memberGlobalID: state.MemberGlobalID, // 会员全局唯一标识
		operatorCode:   state.OperatorCode,   // 所属厅业务编码
		currencyCode:   state.CurrencyCode,   // 钱包结算币种编码
	}

	return a, nil
}

// close 停止创建 Actor, 并要求所有已登记 Actor 排空消息后退出
func (s *shard) close() {
	if s == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return
	}

	// 在同一把锁内停止新建和关闭现有 Actor, 避免创建与关闭交错导致遗漏
	s.closed = true
	for _, entry := range s.actors {
		entry.actor.close()
	}
}

// wait 等待关闭后的全部 Actor 退出, Context 只控制本次等待
func (s *shard) wait(ctx context.Context) error {
	if s == nil {
		return errNilShard
	}
	if ctx == nil {
		return errNilShardContext
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	if !s.closed {
		s.mu.Unlock()
		return ErrShardNotClosed
	}

	// 已关闭的 Shard 不会再登记新 Actor, 复制指针后即可释放锁等待
	actors := make([]*actor, 0, len(s.actors))
	for _, entry := range s.actors {
		actors = append(actors, entry.actor)
	}
	s.mu.Unlock()

	var waitErr error
	for _, a := range actors {
		if err := a.wait(ctx); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			// 个别 Actor 出错也应等待其他 Actor 退出, 避免提前关闭 Worker Pool
			waitErr = errors.Join(waitErr, err)
		}
	}

	return waitErr
}
