// Package constants 定义 operator-admin-api 各业务模块共享的常量。
// 金额分/主币单位换算、存款订单状态等，取值与 operator-crm RPC 层约定保持一致。
package constants

import "time"

// DefaultAmountCentsFactor 金额分/主币单位的换算比例默认值。
// 实际倍率应从 platform-base CurrencyService.GetByCode → CurrencyInfo.amount_factor 动态获取；
// 仅在 RPC 不可用或返回无效值时才降级使用此兜底值。
// 输入时 × factor，输出时 ÷ factor。
const DefaultAmountCentsFactor = 100
const DepositCurrency = "USDT"

// 存款方式（deposit_type）
const (
	DepositTypeCrypto     = int32(1) // 加密货币
	DepositTypeManual     = int32(2) // 人工入账（通过 operator-admin-api 新增的订单固定为此值）
	DepositTypeThirdParty = int32(3) // 第三方代收
)

// 存款订单列表 mode，决定可查询的状态范围
const (
	DepositOrderModeAudit  = int32(1) // 审核列表：待审核、冻结、待一审、待二审
	DepositOrderModeManage = int32(2) // 存款管理：通过、拒绝、支付成功、支付失败、支付拒绝
)

// 存款订单状态（status），与 operator-crm constants/agent_deposit_order.go 保持一致
const (
	DepositOrderStatusPending    = int32(1)  // 待审核（通用待审状态）
	DepositOrderStatusFrozen     = int32(2)  // 冻结
	DepositOrderStatusUnfrozen   = int32(3)  // 解冻
	DepositOrderStatusApproved   = int32(4)  // 通过（ApproveAgentDepositOrder 返回此值时触发推荐人存款事件）
	DepositOrderStatusRejected   = int32(5)  // 拒绝
	DepositOrderStatusPaySuccess = int32(6)  // 支付成功
	DepositOrderStatusPayFailed  = int32(7)  // 支付失败
	DepositOrderStatusPayReject  = int32(8)  // 支付拒绝
	DepositOrderStatusPending1st = int32(11) // 待一审
	DepositOrderStatusPending2nd = int32(12) // 待二审
)

// 存款订单导出（CSV → OSS）相关常量
const (
	DepositExportScene                  = "deposit_order_export" // OSS scene 标识（待 OSS 注册 deposit_order_export 后替换）
	DepositExportVisibility             = "private"              // 上传后文件可见性
	DepositExportUploadPath             = "/p9/oss/file/upload"  // OSS 上传端点路径
	DepositExportOSSAuthScopeHeader     = "X-OSS-Auth-Scope"
	DepositExportOSSAuthScope           = "operator"
	DepositExportOSSBusinessSuccessCode = 0          // OSS 业务成功码
	DepositExportMaxBytes               = 32 << 20   // CSV 文件大小上限 32 MiB
	DepositExportMaxUploadResponseBytes = 1 << 20    // OSS 响应体读取上限 1 MiB
	DepositExportResponseOverflowBytes  = 1          // 额外读取 1 字节用于探测响应超限
	DepositExportKeyDateLayout          = "20060102" // OSS key 中日期段的格式
	DepositExportKeyDateGroup           = "export_date"
	DepositExportMaxDownloadTTL         = time.Hour        // 下载签名最大有效期
	DepositExportUploadTimeout          = 12 * time.Second // HTTP 上传客户端超时
	DepositExportOSSTimeout             = 10 * time.Second // 整个 OSS 阶段 context 超时
	DepositExportRPCTimeout             = 15 * time.Second // RPC 查询超时
)
