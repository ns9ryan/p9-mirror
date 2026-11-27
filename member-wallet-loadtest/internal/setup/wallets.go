package setup

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc/managepb"
)

// WalletConfig 测试钱包准备配置
type WalletConfig struct {
	OperatorCode  string        // 测试厅编码
	CurrencyCode  string        // 钱包结算币种
	MemberIDStart int64         // 测试会员起始ID
	Count         int           // 测试钱包数量
	Concurrency   int           // 钱包准备并发数
	Timeout       time.Duration // 单次RPC请求超时时间
}

// Wallet 测试钱包信息
type Wallet struct {
	MemberID       int64  // 会员ID
	MemberGlobalID string // 会员全局唯一标识
	WalletGlobalID string // 钱包全局唯一标识
	CurrencyCode   string // 钱包结算币种

	// 初始状态
	Balance          int64 // 可用余额
	FrozenBalance    int64 // 冻结余额
	UnsettledBalance int64 // 游戏未结算余额
	Version          int64 // 钱包状态版本
}

// PrepareWallets 创建测试钱包并预热 Actor
func PrepareWallets(ctx context.Context, client memberwalletrpc.MemberWalletClient, cfg WalletConfig) ([]Wallet, error) {
	if ctx == nil {
		return nil, errors.New("prepare wallets: nil context")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if client == nil {
		return nil, errors.New("prepare wallets: nil RPC client")
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	concurrency := cfg.Concurrency
	if concurrency <= 0 {
		concurrency = 1
	}
	if concurrency > cfg.Count {
		concurrency = cfg.Count
	}

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	wallets := make([]Wallet, cfg.Count)
	jobs := make(chan int)

	var wg sync.WaitGroup
	var once sync.Once
	var firstErr error

	// 启动固定数量的 Worker 准备钱包
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()

			for index := range jobs {
				if runCtx.Err() != nil {
					return
				}

				memberID := cfg.MemberIDStart + int64(index)

				w, err := prepareWallet(runCtx, client, cfg, memberID)
				if err != nil {
					once.Do(func() {
						firstErr = fmt.Errorf("准备测试钱包失败, member_id=%d: %w", memberID, err)
						cancel()
					})
					return
				}

				// 每个Worker写入独立索引, 无需额外加锁
				wallets[index] = w
			}
		}()
	}

	// 分发钱包准备任务
	sending := true

	for i := 0; i < cfg.Count && sending; i++ {
		select {
		case jobs <- i:
		case <-runCtx.Done():
			sending = false
		}
	}

	close(jobs)
	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return wallets, nil
}

// prepareWallet 创建单个测试钱包并校验初始状态
func prepareWallet(ctx context.Context, client memberwalletrpc.MemberWalletClient, cfg WalletConfig, memberID int64) (Wallet, error) {
	memberGlobalID := testMemberGlobalID(cfg.OperatorCode, memberID)

	// 创建钱包, 重复请求复用已有钱包
	createCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)

	created, err := client.CreateWallet(createCtx, &managepb.CreateWalletRequest{
		MemberId:       memberID,
		MemberGlobalId: memberGlobalID,
		CurrencyCode:   cfg.CurrencyCode,
	})
	cancel()

	if err != nil {
		return Wallet{}, fmt.Errorf("create wallet: %w", err)
	}

	// 查询运行时余额, 提前完成Actor加载
	balanceCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)

	balance, err := client.GetBalance(balanceCtx, &managepb.GetBalanceRequest{
		MemberGlobalId: memberGlobalID,
	})
	cancel()

	if err != nil {
		return Wallet{}, fmt.Errorf("warm up wallet: %w", err)
	}

	// 确认运行时钱包与创建结果一致
	if created.GetWalletGlobalId() == "" ||
		created.GetWalletGlobalId() != balance.GetWalletGlobalId() ||
		balance.GetCurrencyCode() != cfg.CurrencyCode {
		return Wallet{}, errors.New("wallet identity mismatch")
	}

	// 初次压测必须使用未发生资金操作的钱包
	if balance.GetBalance() != 0 ||
		balance.GetFrozenBalance() != 0 ||
		balance.GetUnsettledBalance() != 0 ||
		balance.GetVersion() != 0 {
		return Wallet{}, fmt.Errorf(
			"wallet is not in initial state: wallet_global_id=%s, version=%d",
			balance.GetWalletGlobalId(),
			balance.GetVersion(),
		)
	}

	return Wallet{
		MemberID:       memberID,
		MemberGlobalID: memberGlobalID,
		WalletGlobalID: balance.GetWalletGlobalId(),
		CurrencyCode:   balance.GetCurrencyCode(),

		Balance:          balance.GetBalance(),
		FrozenBalance:    balance.GetFrozenBalance(),
		UnsettledBalance: balance.GetUnsettledBalance(),
		Version:          balance.GetVersion(),
	}, nil
}

// testMemberGlobalID 根据厅和会员ID生成稳定的测试会员UUID
func testMemberGlobalID(operatorCode string, memberID int64) string {
	name := fmt.Sprintf("p9/member-wallet-loadtest/%s/%d", operatorCode, memberID)

	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(name)).String()
}

// validate 校验测试钱包准备参数
func (c WalletConfig) validate() error {
	if c.OperatorCode == "" || strings.TrimSpace(c.OperatorCode) != c.OperatorCode {
		return errors.New("invalid operator code")
	}

	if c.CurrencyCode == "" || strings.TrimSpace(c.CurrencyCode) != c.CurrencyCode {
		return errors.New("invalid currency code")
	}

	if c.MemberIDStart <= 0 {
		return errors.New("member ID start must be greater than zero")
	}

	if c.Count <= 0 {
		return errors.New("wallet count must be greater than zero")
	}

	if int64(c.Count-1) > math.MaxInt64-c.MemberIDStart {
		return errors.New("member ID range overflow")
	}

	if c.Timeout <= 0 {
		return errors.New("timeout must be greater than zero")
	}

	return nil
}
