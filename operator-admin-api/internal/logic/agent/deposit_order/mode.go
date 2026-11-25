package deposit_order

import agentconst "oa.98ent.com/p9/operator-admin-api/internal/constants"

// modeAllowedSet 定义各 mode 允许的状态集合。
// DepositOrderModeAudit  审核列表：待审核、冻结、待一审、待二审
// DepositOrderModeManage 存款管理：通过、拒绝、支付成功、支付失败、支付拒绝
var modeAllowedSet = map[int32]map[int32]struct{}{
	agentconst.DepositOrderModeAudit: {
		agentconst.DepositOrderStatusPending:    {},
		agentconst.DepositOrderStatusFrozen:     {},
		agentconst.DepositOrderStatusPending1st: {},
		agentconst.DepositOrderStatusPending2nd: {},
	},
	agentconst.DepositOrderModeManage: {
		agentconst.DepositOrderStatusApproved:   {},
		agentconst.DepositOrderStatusRejected:   {},
		agentconst.DepositOrderStatusPaySuccess: {},
		agentconst.DepositOrderStatusPayFailed:  {},
		agentconst.DepositOrderStatusPayReject:  {},
	},
}

// modeAllowedStatuses 返回指定 mode 允许的状态集合；mode=0 或不存在返回 nil（不限）。
func modeAllowedStatuses(mode int32) map[int32]struct{} {
	return modeAllowedSet[mode]
}

// filterStatuses 静默剔除不在 allowed 中的值；allowed=nil 时原样返回。
func filterStatuses(ss []int32, allowed map[int32]struct{}) []int32 {
	if allowed == nil {
		return ss
	}
	out := ss[:0:0]
	for _, s := range ss {
		if _, ok := allowed[s]; ok {
			out = append(out, s)
		}
	}
	return out
}

// filterStatus 若 s 不在 allowed 中则置 0；allowed=nil 时原样返回。
func filterStatus(s int32, allowed map[int32]struct{}) int32 {
	if allowed == nil || s == 0 {
		return s
	}
	if _, ok := allowed[s]; ok {
		return s
	}
	return 0
}

// modeAllowedKeys 返回 allowed 集合中所有状态值的切片，用于兜底填充。
func modeAllowedKeys(allowed map[int32]struct{}) []int32 {
	out := make([]int32, 0, len(allowed))
	for k := range allowed {
		out = append(out, k)
	}
	return out
}
