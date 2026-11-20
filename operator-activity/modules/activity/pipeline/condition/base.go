package condition

import (
	"oa.98ent.com/p9/operator-activity/modules/activity/pipeline"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"
)

// BaseDeps 公共条件链依赖
type BaseDeps struct {
	CycleMgr *pipeline.Manager
	Apps     ApplicationChecker
	Limits   LimitChecker
}

// BuildBaseChain 组装各业务共用的条件节点（状态/时间/用户/周期/参与/领取限制）
func BuildBaseChain(activity *types.Activity, deps BaseDeps) *Chain {
	return NewChain(
		&ActivityStatusCondition{},
		&ActivityTimeCondition{},
		&FixedLevelCondition{Allowed: activity.AllowedFixedLevel},
		&AutoLevelCondition{Allowed: activity.AllowedAutoLevel},
		&ClientTypeCondition{Allowed: activity.AllowedClients},
		&CycleModeCondition{Manager: deps.CycleMgr},
		&ApplyLimitCondition{Limits: activity.ApplyLimit, Repo: deps.Limits},
		&ApplyModeCondition{Mode: activity.ApplyMode, Repo: deps.Apps},
		&ClaimLimitCondition{Limits: activity.ClaimLimit, Repo: deps.Limits},
	)
}
