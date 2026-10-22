package agent

import (
	"context"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/ent/agentuser"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/dao"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"
)

// GetConfig 读取当前租户代理配置；无配置时返回全 0。
func (s *Service) GetConfig(ctx context.Context) (*agentpb.AgentConfigResp, error) {
	ctx, _, err := requireOperator(ctx)
	if err != nil {
		return nil, err
	}
	row, err := s.DB.AgentConfig.Query().Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return &agentpb.AgentConfigResp{}, nil
		}
		return nil, err
	}
	return toConfigResp(row), nil
}

// UpdateConfig upsert 当前租户代理配置。
func (s *Service) UpdateConfig(ctx context.Context, in *agentpb.UpdateAgentConfigReq) (*agentpb.AgentConfigResp, error) {
	ctx, _, err := requireOperator(ctx)
	if err != nil {
		return nil, err
	}
	if in == nil {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}

	maxRate := int64(in.MaxShareRate)
	minApply := int64(in.ApplyMinShareRate)
	maxApply := int64(in.ApplyMaxShareRate)
	if maxRate < 0 || maxRate > maxShareRate ||
		minApply < 0 || minApply > maxShareRate ||
		maxApply < 0 || maxApply > maxShareRate {
		return nil, xerr.BadRequest(i18n.AgentShareRateInvalid)
	}
	if minApply > maxApply || maxApply > maxRate {
		return nil, xerr.BadRequest(i18n.AgentShareRateOrderInvalid)
	}
	if in.CashMaxMemberCount < 0 || in.CashMaxAgentPerIp < 0 {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}

	for _, id := range []int64{
		in.CashProductionAgentId,
		in.CashDemoAgentId,
		in.CashTestAgentId,
		in.CreditDemoAgentId,
		in.CreditTestAgentId,
	} {
		if id == 0 {
			continue
		}
		ok, err := s.DB.AgentUser.Query().Where(agentuser.IDEQ(id)).Exist(ctx)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, xerr.BadRequest(i18n.AgentDesignatedAgentInvalid)
		}
	}

	var row *ent.AgentConfig
	err = dao.WithTx(ctx, s.DB, func(tx *ent.Tx) error {
		existing, qerr := tx.AgentConfig.Query().Only(ctx)
		switch {
		case qerr == nil:
			upd := tx.AgentConfig.UpdateOneID(existing.ID).
				SetMaxShareRate(maxRate).
				SetApplyMinShareRate(minApply).
				SetApplyMaxShareRate(maxApply).
				SetCashMaxMemberCount(in.CashMaxMemberCount).
				SetCashMaxAgentPerIP(in.CashMaxAgentPerIp)
			applyOptionalAgentID(upd.SetCashProductionAgentID, upd.ClearCashProductionAgentID, in.CashProductionAgentId)
			applyOptionalAgentID(upd.SetCashDemoAgentID, upd.ClearCashDemoAgentID, in.CashDemoAgentId)
			applyOptionalAgentID(upd.SetCashTestAgentID, upd.ClearCashTestAgentID, in.CashTestAgentId)
			applyOptionalAgentID(upd.SetCreditDemoAgentID, upd.ClearCreditDemoAgentID, in.CreditDemoAgentId)
			applyOptionalAgentID(upd.SetCreditTestAgentID, upd.ClearCreditTestAgentID, in.CreditTestAgentId)
			row, qerr = upd.Save(ctx)
			return mapEntErr(qerr)
		case ent.IsNotFound(qerr):
			c := tx.AgentConfig.Create().
				SetMaxShareRate(maxRate).
				SetApplyMinShareRate(minApply).
				SetApplyMaxShareRate(maxApply).
				SetCashMaxMemberCount(in.CashMaxMemberCount).
				SetCashMaxAgentPerIP(in.CashMaxAgentPerIp)
			if in.CashProductionAgentId > 0 {
				c.SetCashProductionAgentID(in.CashProductionAgentId)
			}
			if in.CashDemoAgentId > 0 {
				c.SetCashDemoAgentID(in.CashDemoAgentId)
			}
			if in.CashTestAgentId > 0 {
				c.SetCashTestAgentID(in.CashTestAgentId)
			}
			if in.CreditDemoAgentId > 0 {
				c.SetCreditDemoAgentID(in.CreditDemoAgentId)
			}
			if in.CreditTestAgentId > 0 {
				c.SetCreditTestAgentID(in.CreditTestAgentId)
			}
			row, qerr = c.Save(ctx)
			return mapEntErr(qerr)
		default:
			return qerr
		}
	})
	if err != nil {
		return nil, err
	}
	return toConfigResp(row), nil
}

func applyOptionalAgentID(set func(int64) *ent.AgentConfigUpdateOne, clear func() *ent.AgentConfigUpdateOne, id int64) {
	if id > 0 {
		set(id)
		return
	}
	clear()
}

func toConfigResp(row *ent.AgentConfig) *agentpb.AgentConfigResp {
	resp := &agentpb.AgentConfigResp{
		Id:                 row.ID,
		MaxShareRate:       int32(row.MaxShareRate),
		ApplyMinShareRate:  int32(row.ApplyMinShareRate),
		ApplyMaxShareRate:  int32(row.ApplyMaxShareRate),
		CashMaxMemberCount: row.CashMaxMemberCount,
		CashMaxAgentPerIp:  row.CashMaxAgentPerIP,
	}
	if row.CashProductionAgentID != nil {
		resp.CashProductionAgentId = *row.CashProductionAgentID
	}
	if row.CashDemoAgentID != nil {
		resp.CashDemoAgentId = *row.CashDemoAgentID
	}
	if row.CashTestAgentID != nil {
		resp.CashTestAgentId = *row.CashTestAgentID
	}
	if row.CreditDemoAgentID != nil {
		resp.CreditDemoAgentId = *row.CreditDemoAgentID
	}
	if row.CreditTestAgentID != nil {
		resp.CreditTestAgentId = *row.CreditTestAgentID
	}
	return resp
}
