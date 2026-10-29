package agent

import (
	"context"

	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/ent/agentclosure"
	"oa.98ent.com/p9/operator-crm/ent/agentstats"
)

// 插入闭包 维护代理关系树
func insertClosures(ctx context.Context, tx *ent.Tx, newID, parentID int64) error {
	if _, err := tx.AgentClosure.Create().
		SetAncestorID(newID).
		SetDescendantID(newID).
		SetDepth(0).
		Save(ctx); err != nil {
		return err
	}
	rows, err := tx.AgentClosure.Query().
		Where(agentclosure.DescendantIDEQ(parentID)).
		All(ctx)
	if err != nil {
		return err
	}
	builders := make([]*ent.AgentClosureCreate, 0, len(rows)+1)
	if len(rows) == 0 {
		// 上级缺少闭包时仍挂到直接上级
		builders = append(builders, tx.AgentClosure.Create().
			SetAncestorID(parentID).
			SetDescendantID(newID).
			SetDepth(1))
	} else {
		for _, row := range rows {
			builders = append(builders, tx.AgentClosure.Create().
				SetAncestorID(row.AncestorID).
				SetDescendantID(newID).
				SetDepth(row.Depth+1))
		}
	}
	return tx.AgentClosure.CreateBulk(builders...).Exec(ctx)
}

// 更新代理统计表 代理直属下级数及总代理数
func bumpAgentCounts(ctx context.Context, tx *ent.Tx, parentID int64) error {
	return adjustAgentCounts(ctx, tx, parentID, 1)
}

// 扣减代理统计表 代理直属下级数及总代理数（软删叶子代理后调用）
func debumpAgentCounts(ctx context.Context, tx *ent.Tx, parentID int64) error {
	return adjustAgentCounts(ctx, tx, parentID, -1)
}

func adjustAgentCounts(ctx context.Context, tx *ent.Tx, parentID int64, delta int64) error {
	if _, err := tx.AgentStats.Update().
		Where(agentstats.AgentIDEQ(parentID)).
		AddDirectAgentCount(delta).
		Save(ctx); err != nil {
		return err
	}
	rows, err := tx.AgentClosure.Query().
		Where(agentclosure.DescendantIDEQ(parentID)).
		All(ctx)
	if err != nil {
		return err
	}
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.AncestorID)
	}
	if len(ids) == 0 {
		return nil
	}
	_, err = tx.AgentStats.Update().
		Where(agentstats.AgentIDIn(ids...)).
		AddTotalAgentCount(delta).
		Save(ctx)
	return err
}

// 更新代理统计表 代理直属会员数及总会员数
func BumpMemberCounts(ctx context.Context, tx *ent.Tx, agentID int64) error {
	if _, err := tx.AgentStats.Update().
		Where(agentstats.AgentIDEQ(agentID)).
		AddDirectMemberCount(1).
		Save(ctx); err != nil {
		return err
	}
	rows, err := tx.AgentClosure.Query().
		Where(agentclosure.DescendantIDEQ(agentID)).
		All(ctx)
	if err != nil {
		return err
	}
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.AncestorID)
	}
	if len(ids) == 0 {
		return nil
	}
	_, err = tx.AgentStats.Update().
		Where(agentstats.AgentIDIn(ids...)).
		AddTotalMemberCount(1).
		Save(ctx)
	return err
}
