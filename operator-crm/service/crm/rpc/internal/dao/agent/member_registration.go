package agent

import (
	"context"

	"oa.98ent.com/p9/operator-crm/ent"
)

// DefaultMemberRegistrationAgentIDTx 读取当前厅的现金生产代理配置，未配置时返回 0。
func DefaultMemberRegistrationAgentIDTx(ctx context.Context, tx *ent.Tx) (int64, error) {
	cfg, err := tx.AgentConfig.Query().Only(ctx)
	if ent.IsNotFound(err) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	if cfg.CashProductionAgentID == nil {
		return 0, nil
	}
	return *cfg.CashProductionAgentID, nil
}
