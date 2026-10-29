package agent

import (
	"context"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/ent/agentclosure"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/dao"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"
)

// DeleteAgent 软删除代理（仅无下级、非根、非线路指定代理）。
func (s *Service) DeleteAgent(ctx context.Context, req *agentpb.DeleteAgentReq) (*agentpb.Empty, error) {
	if req == nil {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	agentID := req.GetAgentId()
	opParams := &RecordOpParams{
		targetType: opTargetTypeAgent,
		targetID:   agentID,
		action:     opActionDelete,
	}

	row, err := s.LoadAgent(ctx, agentID)
	if err != nil {
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	opParams.targetAccount = row.Account

	if row.Level < 1 {
		err = xerr.BadRequest(i18n.AgentCannotDeleteRoot)
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	// 统计父级代理用户数量,如果大于0,则不能删除
	childCount, err := s.CountAgentByParentID(ctx, row.ID)
	if err != nil {
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	if childCount > 0 {
		err = xerr.BadRequest(i18n.AgentHasChildren)
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}

	if designated, derr := s.IsDesignatedAgent(ctx, row.ID); derr != nil {
		s.RecordOpFail(ctx, opParams, derr)
		return nil, derr
	} else if designated {
		err = xerr.BadRequest(i18n.AgentIsDesignated)
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}

	parentID := int64(0)
	if row.ParentID != nil {
		parentID = *row.ParentID
	}

	err = dao.WithTx(ctx, s.DB, func(tx *ent.Tx) error {
		if cerr := tx.AgentUser.DeleteOneID(row.ID).Exec(ctx); cerr != nil {
			return mapEntErr(cerr)
		}
		if _, cerr := tx.AgentClosure.Delete().
			Where(agentclosure.DescendantIDEQ(row.ID)).
			Exec(ctx); cerr != nil {
			return mapEntErr(cerr)
		}
		if parentID > 0 {
			if cerr := debumpAgentCounts(ctx, tx, parentID); cerr != nil {
				return cerr
			}
		}
		return nil
	})
	if err != nil {
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}

	if _, op, oerr := requireOperator(ctx); oerr == nil {
		s.invalidateTree(op)
	}
	s.RecordOpOK(ctx, opParams)
	return &agentpb.Empty{}, nil
}
