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

	// DepositCurrency 人工存款订单固定使用 USDT
	DepositCurrency   = "USDT"
	DepositTypeManual = int8(2) // 2:人工入账

	// 创建人/更新人类型，与 op_log 中的 adminType 保持一致
	CreateTypeOperator = int8(1) // 分站管理员
	CreateTypeAgent    = int8(2) // 代理

	// audit action 对应 agent_deposit_order_audit.action 字段
	AuditActionApprove  = int8(1) // 审核通过
	AuditActionReject   = int8(2) // 审核拒绝
	AuditActionFreeze   = int8(3) // 冻结
	AuditActionUnfreeze = int8(4) // 解冻

	// RoleSuperAdmin 超级管理员角色代码，直接引用 core 包常量
	RoleSuperAdmin = coreservice.RoleSuperAdmin

	// TODO: 财务二审金额阈值 — 有财务设置功能后从财务配置表读取【财务二审金额】字段。
	// 当存款金额（分）超过此阈值时 audit_stages=2（需二审），否则 audit_stages=1。
	FinanceSecondAuditThreshold = int64(100000) // 暂定固定值 1000 元 = 100000 分
)
