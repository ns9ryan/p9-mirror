package agent

// AmountCentsFactor 金额分/主币单位的换算比例。
// RPC 层金额以"分"存储，前端展示以主币单位（如 USDT）为准。
// 输入时 × AmountCentsFactor，输出时 ÷ AmountCentsFactor。
const AmountCentsFactor = 100

// DepositTypeManual 人工存款订单固定存款方式。
// RPC 层支持：1 加密货币 / 2 人工入账。
// 通过 operator-admin-api 新增的订单固定为人工入账。
const DepositTypeManual = int32(2)

// DepositOrderStatusApproved 存款订单最终通过状态。
// 仅当 ApproveAgentDepositOrder 返回此状态时，才触发推荐人存款事件。
const DepositOrderStatusApproved = int32(4)
