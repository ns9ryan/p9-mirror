package agent

import (
	"context"
	"strconv"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/utils"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/ent/agentapply"
	"oa.98ent.com/p9/operator-crm/ent/agentuser"
	"oa.98ent.com/p9/operator-crm/ent/predicate"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"
)

const (
	auditStatusPending      int64 = 1 // 待审核
	auditStatusApproved     int64 = 2 // 已通过
	auditStatusManualReject int64 = 3 // 人工拒绝
	auditStatusRiskReject   int64 = 4 // 风控拒绝

	reviewActionApprove      int32 = 1 // 通过
	reviewActionManualReject int32 = 2 // 人工拒绝
	reviewActionRiskReject   int32 = 3 // 风控拒绝

	accountTypeAgent  int32 = 1 // 代理账号
	accountTypeMember int32 = 2 // 会员账号
)

// AddAgentApply 新增代理申请（待审核）。
func (s *Service) AddAgentApply(ctx context.Context, in *agentpb.AddAgentApplyReq) (*agentpb.AgentApplyItem, error) {
	ctx, _, err := requireOperator(ctx)
	if err != nil {
		return nil, err
	}
	if in == nil {
		return nil, xerr.BadRequest(i18n.AgentFieldsRequired)
	}
	account := strings.TrimSpace(in.Account)
	realname := strings.TrimSpace(in.Realname)
	mobile := strings.TrimSpace(in.Mobile)
	memberAccount := strings.TrimSpace(in.MemberAccount)
	memberCode := strings.TrimSpace(in.MemberCode)
	if account == "" || realname == "" || mobile == "" || memberAccount == "" || memberCode == "" {
		return nil, xerr.BadRequest(i18n.AgentFieldsRequired)
	}
	if in.RegisterSource < 1 || in.RegisterSource > 4 {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	shareRate := int64(in.ApplyShareRate)
	if shareRate < 0 || shareRate > maxShareRate {
		return nil, xerr.BadRequest(i18n.AgentShareRateInvalid)
	}
	// 查询代理配置是否存在
	cfg, err := s.GetConfig(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, xerr.BadRequest(i18n.AgentConfigMissing)
		}
		return nil, err
	}
	if shareRate < int64(cfg.GetApplyMinShareRate()) || shareRate > int64(cfg.GetApplyMaxShareRate()) {
		return nil, xerr.BadRequest(i18n.AgentShareRateInvalid)
	}
	// 查询代理账号是否存在
	exists, err := s.DB.AgentUser.Query().
		Where(agentuser.AccountEQ(account)).
		Exist(ctx)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, xerr.BadRequest(i18n.AgentApplyAccountUsedByOther)
	}
	// 查询是否存在相同的申请
	oldApply, err := s.DB.AgentApply.Query().
		Where(
			agentapply.Or(agentapply.AccountEQ(account), agentapply.MemberCodeEQ(memberCode)),
			agentapply.AuditStatusIn(auditStatusPending, auditStatusApproved),
		).
		Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, err
	}
	if oldApply != nil {
		if oldApply.MemberCode == memberCode {
			if oldApply.AuditStatus == auditStatusPending {
				return nil, xerr.BadRequest(i18n.AgentApplyPendingExists)
			}
			return nil, xerr.BadRequest(i18n.AgentApplyMemberCodeExists)
		}
		if oldApply.Account == account {
			return nil, xerr.BadRequest(i18n.AgentApplyAccountUsedByOther)
		}
	}

	now := time.Now()
	c := s.DB.AgentApply.Create().
		SetAccount(account).
		SetRealname(realname).
		SetMobile(mobile).
		SetRegisterSource(int64(in.RegisterSource)).
		SetApplyShareRate(shareRate).
		SetMemberAccount(memberAccount).
		SetMemberCode(memberCode).
		SetAuditStatus(auditStatusPending).
		SetRegisterAt(now)
	if v := strings.TrimSpace(in.ApplyRemark); v != "" {
		c.SetApplyRemark(v)
	}
	row, err := c.Save(ctx)
	if err != nil {
		return nil, mapEntErr(err)
	}
	return toApplyItem(row), nil
}

// ListAgentApplies 分页查询代理申请。
func (s *Service) ListAgentApplies(ctx context.Context, in *agentpb.ListAgentAppliesReq) (*agentpb.ListAgentAppliesResp, error) {
	ctx, _, err := requireOperator(ctx)
	if err != nil {
		return nil, err
	}
	if in == nil {
		in = &agentpb.ListAgentAppliesReq{}
	}

	page := max(in.Page, 1)
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	preds := make([]predicate.AgentApply, 0, 4)
	if in.AuditStatus > 0 {
		if in.AuditStatus < 1 || in.AuditStatus > 4 {
			return nil, xerr.BadRequest(i18n.InvalidParam)
		}
		preds = append(preds, agentapply.AuditStatusEQ(int64(in.AuditStatus)))
	}
	keyword := strings.TrimSpace(in.Account)
	if keyword != "" {
		switch in.AccountType {
		case accountTypeAgent:
			preds = append(preds, agentapply.AccountContainsFold(keyword))
		case accountTypeMember:
			preds = append(preds, agentapply.MemberAccountContainsFold(keyword))
		case 0:
			preds = append(preds, agentapply.Or(
				agentapply.AccountContainsFold(keyword),
				agentapply.MemberAccountContainsFold(keyword),
			))
		default:
			return nil, xerr.BadRequest(i18n.InvalidParam)
		}
	}
	if in.TimeStart > 0 {
		preds = append(preds, agentapply.RegisterAtGTE(time.Unix(in.TimeStart, 0)))
	}
	if in.TimeEnd > 0 {
		preds = append(preds, agentapply.RegisterAtLTE(time.Unix(in.TimeEnd, 0)))
	}

	q := s.DB.AgentApply.Query()
	if len(preds) > 0 {
		q = q.Where(preds...)
	}
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := q.
		Order(agentapply.ByID(sql.OrderDesc())).
		Offset(int((page - 1) * pageSize)).
		Limit(int(pageSize)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]*agentpb.AgentApplyItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, toApplyItem(row))
	}
	return &agentpb.ListAgentAppliesResp{Items: items, Total: int64(total)}, nil
}

// ReviewAgentApply 审批代理申请；通过时开户并返回明文密码。
func (s *Service) ReviewAgentApply(ctx context.Context, in *agentpb.ReviewAgentApplyReq) (*agentpb.ReviewAgentApplyResp, error) {
	ctx, _, err := requireOperator(ctx)
	if err != nil {
		return nil, err
	}
	if in == nil || in.Id <= 0 {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}

	row, err := s.DB.AgentApply.Get(ctx, in.Id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, xerr.BadRequest(i18n.AgentApplyNotFound)
		}
		return nil, err
	}
	if row.AuditStatus != auditStatusPending {
		return nil, xerr.BadRequest(i18n.AgentApplyNotPending)
	}

	auditor := reviewAuditor(ctx)
	now := time.Now()

	switch in.Action {
	case reviewActionManualReject, reviewActionRiskReject:
		reason := strings.TrimSpace(in.RejectReason)
		if reason == "" {
			return nil, xerr.BadRequest(i18n.AgentApplyRejectReasonRequired)
		}
		status := auditStatusManualReject
		if in.Action == reviewActionRiskReject {
			status = auditStatusRiskReject
		}
		updated, uerr := s.DB.AgentApply.UpdateOneID(row.ID).
			SetAuditStatus(status).
			SetRejectReason(reason).
			SetAuditedAt(now).
			SetAuditor(auditor).
			Save(ctx)
		if uerr != nil {
			return nil, mapEntErr(uerr)
		}
		return &agentpb.ReviewAgentApplyResp{Apply: toApplyItem(updated)}, nil

	case reviewActionApprove:
		cfg, cerr := s.DB.AgentConfig.Query().Only(ctx)
		if cerr != nil {
			if ent.IsNotFound(cerr) {
				return nil, xerr.BadRequest(i18n.AgentConfigMissing)
			}
			return nil, cerr
		}
		if cfg.CashProductionAgentID == nil || *cfg.CashProductionAgentID <= 0 {
			return nil, xerr.BadRequest(i18n.AgentDesignatedAgentMissing)
		}
		parentID := *cfg.CashProductionAgentID
		parent, perr := s.getAgentUserByID(ctx, parentID)
		if perr != nil {
			if ent.IsNotFound(perr) {
				return nil, xerr.BadRequest(i18n.AgentDesignatedAgentInvalid)
			}
			return nil, perr
		}

		password, perr := utils.GenRandomPassword(16)
		if perr != nil {
			return nil, perr
		}
		promoCode, perr := utils.GenShortID()
		if perr != nil {
			return nil, perr
		}
		mobile := row.Mobile
		realname := row.Realname
		fields := &agentpb.AgentCreateFields{
			Account:       row.Account,
			Code:          row.Account,
			Password:      password,
			AgentType:     int32(parent.AgentType),
			AgentLineCode: parent.AgentLineCode,
			ShareRate:     int32(row.ApplyShareRate),
			PromoCode:     strings.ToUpper(promoCode),
			Mobile:        &mobile,
			Realname:      &realname,
		}
		created, cerr := s.createUnderParent(ctx, fields, parentID, row.RegisterSource, false)
		if cerr != nil {
			return nil, cerr
		}
		updated, uerr := s.DB.AgentApply.UpdateOneID(row.ID).
			SetAuditStatus(auditStatusApproved).
			SetAuditedAt(now).
			SetAuditor(auditor).
			SetAgentID(created.Id).
			ClearRejectReason().
			Save(ctx)
		if uerr != nil {
			return nil, mapEntErr(uerr)
		}
		return &agentpb.ReviewAgentApplyResp{
			Apply:    toApplyItem(updated),
			Password: password,
		}, nil

	default:
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
}

func reviewAuditor(ctx context.Context) string {
	c := ctxdata.ClaimsFromCtx(ctx)
	if c == nil {
		return ""
	}
	if v := strings.TrimSpace(c.Username); v != "" {
		return v
	}
	if v := strings.TrimSpace(c.UserCode); v != "" {
		return v
	}
	if c.UserID > 0 {
		return strconv.FormatInt(c.UserID, 10)
	}
	return ""
}

func toApplyItem(row *ent.AgentApply) *agentpb.AgentApplyItem {
	if row == nil {
		return nil
	}
	item := &agentpb.AgentApplyItem{
		Id:             row.ID,
		Account:        row.Account,
		Realname:       row.Realname,
		Mobile:         row.Mobile,
		RegisterSource: int32(row.RegisterSource),
		ApplyShareRate: int32(row.ApplyShareRate),
		MemberAccount:  row.MemberAccount,
		MemberCode:     row.MemberCode,
		AuditStatus:    int32(row.AuditStatus),
		RegisterAt:     row.RegisterAt.Unix(),
	}
	if row.RejectReason != nil {
		item.RejectReason = *row.RejectReason
	}
	if row.ApplyRemark != nil {
		item.ApplyRemark = *row.ApplyRemark
	}
	if row.AuditedAt != nil {
		item.AuditedAt = row.AuditedAt.Unix()
	}
	if row.Auditor != nil {
		item.Auditor = *row.Auditor
	}
	if row.AgentID != nil {
		item.AgentId = *row.AgentID
	}
	return item
}
