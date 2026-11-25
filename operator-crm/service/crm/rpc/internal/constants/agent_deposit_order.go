package constants

import (
	coreservice "oa.98ent.com/p9/core/rpc/service"
)

const (
	// agent_deposit_order 状态值
	DepositOrderStatusPending    = int8(1)  // 待审核（通用待审状态）
	DepositOrderStatusPending1st = int8(11) // 待一审
	DepositOrderStatusPending2nd = int8(12) // 待二审
	DepositOrderStatusFrozen     = int8(2)  // 冻结
	DepositOrderStatusApproved   = int8(4)  // 通过
	DepositOrderStatusRejected   = int8(5)  // 拒绝
	DepositOrderStatusPaySuccess = int8(6)  // 支付成功
	DepositOrderStatusPayFailed  = int8(7)  // 支付失败
	DepositOrderStatusPayReject  = int8(8)  // 支付拒绝

	// DepositCurrency 人工存款订单固定使用 USDT
	DepositCurrency   = "USDT"
	DepositTypeManual = int8(2) // 2:人工入账

	// 创建人/更新人类型，与 op_log 中的 adminType 保持一致 1：分站管理员 2：代理
	CreateTypeOperator = int8(1) // 分站管理员
	CreateTypeAgent    = int8(2) // 代理

	// audit action 对应 agent_deposit_order_audit.action 字段
	AuditActionApprove  = int8(1) // 审核通过
	AuditActionReject   = int8(2) // 审核拒绝
	AuditActionFreeze   = int8(3) // 冻结
	AuditActionUnfreeze = int8(4) // 解冻

	// RoleSuperAdmin 超级管理员角色代码，直接引用 core 包常量
	RoleSuperAdmin = coreservice.RoleSuperAdmin

	// AuditStage1 / AuditStage2 审核阶段数
	AuditStage1 = int8(1) // 一审
	AuditStage2 = int8(2) // 二审

	// BonusRateDivisor bonus_rate 存储单位换算除数。
	// bonus_rate 以百分比×100 存储（如 10.5% 存为 1050），计算奖励金额时除以此值：
	// bonusAmount = depositAmount × bonusRate / BonusRateDivisor
	BonusRateDivisor = int64(10000)

	// TODO: 财务二审金额阈值 — 有财务配置功能后从配置表读取【财务二审金额】字段。
	// FinanceSecondAuditConfigAmount 为显示单位（主币金额），与货币倍率相乘后得到分为单位的阈值。
	// 当存款金额（分）超过阈值时 audit_stages=2（需二审），否则 audit_stages=1。
	FinanceSecondAuditConfigAmount = int64(2000) // 暂定 2000 主币

	// DepositExportMaxRows 存款订单导出最大行数
	DepositExportMaxRows = int(5000)

	// ClientType 客户端类型，用于 Kafka 事件 payload 的 client_type 字段
	ClientTypePCWeb     = int(1) // PC Web
	ClientTypeMobileWeb = int(2) // 移动 Web
	ClientTypeAndroid   = int(3) // Android
	ClientTypeIOS       = int(4) // iOS
)
