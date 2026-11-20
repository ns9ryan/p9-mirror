package cron

import (
	"sync"

	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/logx"
)

// InitCron 初始化定时任务调度器
// 包含以下定时任务：
// 1. RunGoogleSheetsReporter - 每小时第6分钟从Doris查询数据并导出Google Sheets快照
// 2. HandlePartnerSettlement* - Kafka 消费失败重试（分站财务结算、撤销）
// 3. HandlePaymentGateway* - Kafka 消费失败重试（支付网关 CDC 订单）
func InitCron() (*cron.Cron, error) {
	c := cron.New(cron.WithSeconds())

	// 任务 1: Google Sheets 整点快照 (每小时第6分钟)
	// 从Doris查询实时聚合数据，生成快照并导出到Google Sheets
	var sheetsMutex sync.Mutex
	if _, err := c.AddFunc("0 6 * * * *", func() {
		if !sheetsMutex.TryLock() {
			logx.Errorf("Google Sheets快照任务正在执行，跳过本次调度")
			return
		}
		defer sheetsMutex.Unlock()
		logx.Infof("开始执行 [Google Sheets 快照] 定时任务")
		// jobs.RunGoogleSheetsReporter()
		logx.Infof("[Google Sheets 快照] 定时任务执行完毕")
	}); err != nil {
		logx.Errorf("添加Google Sheets快照定时任务失败: %v", err)
		return nil, err
	}

	// 任务 2: 分站财务日结算失败重试 (每5分钟检查一次)
	// 消费 Kafka Topic: PartnerSettlementDaily
	// 处理失败重试的结算消息
	var settlementMutex sync.Mutex
	if _, err := c.AddFunc("0 */5 * * * *", func() {
		if !settlementMutex.TryLock() {
			logx.Errorf("分站财务日结算重试任务正在执行，跳过本次调度")
			return
		}
		defer settlementMutex.Unlock()
		logx.Infof("开始执行 [分站财务日结算重试] 定时任务")
		// jobs.HandlePartnerSettlementDaily(nil)
		logx.Infof("[分站财务日结算重试] 定时任务执行完毕")
	}); err != nil {
		logx.Errorf("添加分站财务日结算重试定时任务失败: %v", err)
		return nil, err
	}

	// 任务 3: 分站财务日结算撤销失败重试 (每5分钟检查一次)
	// 消费 Kafka Topic: PartnerSettlementDailyCancel
	// 处理失败重试的撤销消息
	var settlementCancelMutex sync.Mutex
	if _, err := c.AddFunc("0 */5 * * * *", func() {
		if !settlementCancelMutex.TryLock() {
			logx.Errorf("分站财务日结算撤销重试任务正在执行，跳过本次调度")
			return
		}
		defer settlementCancelMutex.Unlock()
		logx.Infof("开始执行 [分站财务日结算撤销重试] 定时任务")
		// jobs.HandlePartnerSettlementDailyCancel(nil)
		logx.Infof("[分站财务日结算撤销重试] 定时任务执行完毕")
	}); err != nil {
		logx.Errorf("添加分站财务日结算撤销重试定时任务失败: %v", err)
		return nil, err
	}

	// 任务 4: 支付网关 CDC 存款订单失败重试 (每分钟检查一次)
	// 消费 Kafka Topic: cdc_source_mysql.ns9.t_game_user_deposit_order
	// 处理失败重试的 CDC 消息
	var depositOrderMutex sync.Mutex
	if _, err := c.AddFunc("0 * * * * *", func() {
		if !depositOrderMutex.TryLock() {
			logx.Errorf("支付网关CDC存款订单重试任务正在执行，跳过本次调度")
			return
		}
		defer depositOrderMutex.Unlock()
		logx.Infof("开始执行 [支付网关CDC存款订单重试] 定时任务")
		// jobs.HandlePaymentGatewayDepositOrder(nil)
		logx.Infof("[支付网关CDC存款订单重试] 定时任务执行完毕")
	}); err != nil {
		logx.Errorf("添加支付网关CDC存款订单重试定时任务失败: %v", err)
		return nil, err
	}

	// 任务 5: 支付网关 CDC 提现订单失败重试 (每分钟检查一次)
	// 消费 Kafka Topic: cdc_source_mysql.ns9.t_game_user_withdraw_order
	// 处理失败重试的 CDC 消息
	var withdrawOrderMutex sync.Mutex
	if _, err := c.AddFunc("0 * * * * *", func() {
		if !withdrawOrderMutex.TryLock() {
			logx.Errorf("支付网关CDC提现订单重试任务正在执行，跳过本次调度")
			return
		}
		defer withdrawOrderMutex.Unlock()
		logx.Infof("开始执行 [支付网关CDC提现订单重试] 定时任务")
		// jobs.HandlePaymentGatewayWithdrawOrder(nil)
		logx.Infof("[支付网关CDC提现订单重试] 定时任务执行完毕")
	}); err != nil {
		logx.Errorf("添加支付网关CDC提现订单重试定时任务失败: %v", err)
		return nil, err
	}

	return c, nil
}
