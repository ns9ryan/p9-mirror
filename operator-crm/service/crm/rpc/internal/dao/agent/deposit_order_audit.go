package agent

import (
	"context"
	"time"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/ent/agentdepositorder"
	"oa.98ent.com/p9/operator-crm/ent/agentdepositorderaudit"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/constants"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/dao"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentdepositorderpb"
)

// hasRoleCode 判断 claims.RoleCodes 中是否包含指定角色代码。
func hasRoleCode(c *ctxdata.Claims, code string) bool {
	if c == nil {
		return false
	}
	for _, rc := range c.RoleCodes {
		if rc == code {
			return true
		}
	}
	return false
}

// isSuperAdmin 判断当前用户是否为超级管理员（role_code = "super_admin"）。
func isSuperAdmin(c *ctxdata.Claims) bool {
	return hasRoleCode(c, constants.RoleSuperAdmin)
}

// claimsOperatorIDs 从 ctx 中读取已登录用户身份，返回 (userID, userType int8)。
// userType: 1=分站管理员(constants.CreateTypeOperator) 2=分站代理(constants.CreateTypeAgent)
// 用于统一填充 creator_id/creator_type、updater_id/updater_type、reviewer_id/reviewer_utype、audit_uid/audit_utype 字段。
// 若 claims 为空则返回 (0, constants.CreateTypeOperator) 作为默认值。
func claimsOperatorIDs(ctx context.Context) (int64, int8) {
	c := ctxdata.ClaimsFromCtx(ctx)
	if c == nil {
		return 0, constants.CreateTypeOperator
	}
	utype := constants.CreateTypeOperator
	if c.UserType.IsAgent() {
		utype = constants.CreateTypeAgent
	}
	return c.UserID, utype
}

// requireClaims 获取已登录用户的 Claims，要求 UserID > 0。
// 仅在需要访问 claims 中非 ID/Type 字段（如 ClientIP）时保留此函数，
// 其余情况通过 requireUID(ctx) 获取 uid/utype。
func requireClaims(ctx context.Context) (context.Context, *ctxdata.Claims, error) {
	ctx, _, err := requireOperator(ctx)
	if err != nil {
		return ctx, nil, err
	}
	c := ctxdata.ClaimsFromCtx(ctx)
	if c == nil || c.UserID <= 0 {
		return ctx, nil, xerr.Unauthorized(i18n.Unauthorized)
	}
	return ctx, c, nil
}

// requireUID 在只需要验证登录、取 uid/utype 时使用，不返回完整 claims。
func requireUID(ctx context.Context) (context.Context, int64, int8, error) {
	ctx, _, err := requireOperator(ctx)
	if err != nil {
		return ctx, 0, 0, err
	}
	uid, utype := claimsOperatorIDs(ctx)
	if uid <= 0 {
		return ctx, 0, 0, xerr.Unauthorized(i18n.Unauthorized)
	}
	return ctx, uid, utype, nil
}

// getDepositOrder 查询订单基础字段，不存在时返回 AgentDepositOrderNotFound。
func (s *Service) getDepositOrder(ctx context.Context, orderID int64, fields ...string) (*ent.AgentDepositOrder, error) {
	q := s.DB.AgentDepositOrder.Query().Where(agentdepositorder.IDEQ(orderID))
	var row *ent.AgentDepositOrder
	var err error
	if len(fields) > 0 {
		row, err = q.Select(fields...).Only(ctx)
	} else {
		row, err = q.Only(ctx)
	}
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, xerr.BadRequest(i18n.AgentDepositOrderNotFound)
		}
		return nil, mapEntErr(err)
	}
	return row, nil
}

// ClaimAgentDepositOrder 领取审核。
// 乐观并发：WHERE id=:id AND status IN(11,12) AND reviewer_id=0，影响行数=0 时报具体原因。
func (s *Service) ClaimAgentDepositOrder(ctx context.Context, orderID int64) (*agentdepositorderpb.ClaimAgentDepositOrderResp, error) {
	ctx, uid, utype, err := requireUID(ctx)
	if err != nil {
		return nil, err
	}
	if orderID <= 0 {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}

	// 先查订单，确认存在并获取当前状态
	row, err := s.getDepositOrder(ctx, orderID,
		agentdepositorder.FieldStatus,
		agentdepositorder.FieldReviewerID,
	)
	if err != nil {
		return nil, err
	}
	if row.Status != constants.DepositOrderStatusPending1st && row.Status != constants.DepositOrderStatusPending2nd {
		return nil, xerr.BadRequest(i18n.AgentDepositOrderNotClaimable)
	}
	if row.ReviewerID != 0 {
		return nil, xerr.BadRequest(i18n.AgentDepositOrderAlreadyClaimed)
	}

	now := time.Now()
	n, err := s.DB.AgentDepositOrder.Update().
		Where(
			agentdepositorder.IDEQ(orderID),
			agentdepositorder.StatusIn(constants.DepositOrderStatusPending1st, constants.DepositOrderStatusPending2nd),
			agentdepositorder.ReviewerIDEQ(0),
		).
		SetReviewerID(uid).
		SetReviewerUtype(utype).
		SetClaimedAt(now).
		SetUpdaterID(uid).
		SetUpdaterType(utype).
		Save(ctx)
	if err != nil {
		return nil, mapEntErr(err)
	}
	if n == 0 {
		// 预检通过但更新失败，说明被并发抢占
		return nil, xerr.BadRequest(i18n.AgentDepositOrderAlreadyClaimed)
	}

	return &agentdepositorderpb.ClaimAgentDepositOrderResp{
		OrderId:    orderID,
		Status:     int32(row.Status),
		ReviewerId: uid,
	}, nil
}

// ReleaseAgentDepositOrder 释放审核。
// 普通用户：仅已领取状态下、且自己是领取人才可释放。
// 超级管理员：在已领取状态下可强制释放他人领取的订单。
func (s *Service) ReleaseAgentDepositOrder(ctx context.Context, orderID int64) (*agentdepositorderpb.ReleaseAgentDepositOrderResp, error) {
	ctx, claims, err := requireClaims(ctx)
	if err != nil {
		return nil, err
	}
	if orderID <= 0 {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	uid, utype := claimsOperatorIDs(ctx)

	// 先查订单，确认存在并校验状态和领取人
	row, err := s.getDepositOrder(ctx, orderID,
		agentdepositorder.FieldStatus,
		agentdepositorder.FieldReviewerID,
	)
	if err != nil {
		return nil, err
	}
	// 释放只能在已领取状态（待一审/待二审）下操作，冻结状态不允许
	if row.Status != constants.DepositOrderStatusPending1st && row.Status != constants.DepositOrderStatusPending2nd {
		return nil, xerr.BadRequest(i18n.AgentDepositOrderNotClaimable)
	}
	// 非领取人时：超级管理员可强制释放，普通用户报错
	if row.ReviewerID != uid && !isSuperAdmin(claims) {
		return nil, xerr.BadRequest(i18n.AgentDepositOrderNotClaimer)
	}

	_, err = s.DB.AgentDepositOrder.Update().
		Where(agentdepositorder.IDEQ(orderID)).
		SetReviewerID(0).
		SetReviewerUtype(0).
		ClearClaimedAt().
		SetUpdaterID(uid).
		SetUpdaterType(utype).
		Save(ctx)
	if err != nil {
		return nil, mapEntErr(err)
	}

	return &agentdepositorderpb.ReleaseAgentDepositOrderResp{
		OrderId:    orderID,
		Status:     int32(row.Status),
		ReviewerId: 0,
	}, nil
}

// ApproveAgentDepositOrder 审核通过（事务内：写 audit 记录 + 更新订单状态）。
func (s *Service) ApproveAgentDepositOrder(ctx context.Context, in *agentdepositorderpb.ApproveAgentDepositOrderReq) (*agentdepositorderpb.ApproveAgentDepositOrderResp, error) {
	// 保留完整 claims 以访问 ClientIP
	ctx, claims, err := requireClaims(ctx)
	if err != nil {
		return nil, err
	}
	if in == nil || in.OrderId <= 0 {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}

	uid, utype := claimsOperatorIDs(ctx)

	// 先查订单，确认存在并校验领取人
	preRow, err := s.getDepositOrder(ctx, in.OrderId,
		agentdepositorder.FieldStatus,
		agentdepositorder.FieldReviewerID,
		agentdepositorder.FieldAuditStages,
		agentdepositorder.FieldMemberID,
		agentdepositorder.FieldDepositAmount,
		agentdepositorder.FieldOrderNo,
		agentdepositorder.FieldSiteCurrency,
		agentdepositorder.FieldActualReceivedAmount,
	)
	if err != nil {
		return nil, err
	}
	if preRow.ReviewerID != uid {
		return nil, xerr.BadRequest(i18n.AgentDepositOrderNotClaimer)
	}
	if preRow.Status != constants.DepositOrderStatusPending1st && preRow.Status != constants.DepositOrderStatusPending2nd {
		return nil, xerr.BadRequest(i18n.AgentDepositOrderNotClaimable)
	}

	stage := int8(1)
	if preRow.Status == constants.DepositOrderStatusPending2nd {
		stage = 2
	}

	var newStatus int8
	var auditAt time.Time
	err = dao.WithTx(ctx, s.DB, func(tx *ent.Tx) error {
		auditAt = time.Now()
		now := auditAt
		auditCreate := tx.AgentDepositOrderAudit.Create().
			SetOrderID(in.OrderId).
			SetAction(constants.AuditActionApprove).
			SetStage(stage).
			SetAuditUID(uid).
			SetAuditUtype(utype).
			SetAuditAt(now)
		if claims.ClientIP != "" {
			auditCreate = auditCreate.SetAuditIP(claims.ClientIP)
		}
		if in.AuditNotes != nil && *in.AuditNotes != "" {
			auditCreate = auditCreate.SetAuditNotes(*in.AuditNotes)
		}
		if name := reviewAuditor(ctx); name != "" {
			auditCreate = auditCreate.SetAuditUsername(name)
		}
		if _, terr := auditCreate.Save(ctx); terr != nil {
			return mapEntErr(terr)
		}

		if preRow.AuditStages == 2 && stage == 1 {
			newStatus = constants.DepositOrderStatusPending2nd
		} else {
			newStatus = constants.DepositOrderStatusApproved
		}

		upd := tx.AgentDepositOrder.Update().
			Where(agentdepositorder.IDEQ(in.OrderId)).
			SetStatus(newStatus).
			SetReviewerID(0).
			SetReviewerUtype(0).
			ClearClaimedAt().
			SetUpdaterID(uid).
			SetUpdaterType(utype)
		// 仅最终通过（status=4）时写入 final_audited_at
		if newStatus == constants.DepositOrderStatusApproved {
			upd = upd.SetFinalAuditedAt(auditAt)
		}
		_, terr := upd.Save(ctx)
		return mapEntErr(terr)
	})
	if err != nil {
		return nil, err
	}
	return &agentdepositorderpb.ApproveAgentDepositOrderResp{
		OrderId:               in.OrderId,
		Status:                int32(newStatus),
		MemberId:              preRow.MemberID,
		DepositAmount:         preRow.DepositAmount,
		ApprovedAt:            auditAt.Unix(),
		OrderNo:               preRow.OrderNo,
		SiteCurrency:          preRow.SiteCurrency,
		ActualReceivedAmount:  preRow.ActualReceivedAmount,
	}, nil
}

// RejectAgentDepositOrder 审核拒绝（拒绝原因必填，事务内写 audit 记录 + 更新状态）。
func (s *Service) RejectAgentDepositOrder(ctx context.Context, in *agentdepositorderpb.RejectAgentDepositOrderReq) (*agentdepositorderpb.RejectAgentDepositOrderResp, error) {
	// 保留完整 claims 以访问 ClientIP
	ctx, claims, err := requireClaims(ctx)
	if err != nil {
		return nil, err
	}
	if in == nil || in.OrderId <= 0 {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	if in.AuditNotes == "" {
		return nil, xerr.BadRequest(i18n.AgentDepositOrderRejectNotesRequired)
	}

	uid, utype := claimsOperatorIDs(ctx)

	// 先查订单，确认存在并校验领取人
	preRow, err := s.getDepositOrder(ctx, in.OrderId,
		agentdepositorder.FieldStatus,
		agentdepositorder.FieldReviewerID,
	)
	if err != nil {
		return nil, err
	}
	if preRow.ReviewerID != uid {
		return nil, xerr.BadRequest(i18n.AgentDepositOrderNotClaimer)
	}
	if preRow.Status != constants.DepositOrderStatusPending1st && preRow.Status != constants.DepositOrderStatusPending2nd {
		return nil, xerr.BadRequest(i18n.AgentDepositOrderNotClaimable)
	}

	stage := int8(1)
	if preRow.Status == constants.DepositOrderStatusPending2nd {
		stage = 2
	}

	const rejectedStatus = int8(5)
	err = dao.WithTx(ctx, s.DB, func(tx *ent.Tx) error {
		now := time.Now()
		auditCreate := tx.AgentDepositOrderAudit.Create().
			SetOrderID(in.OrderId).
			SetAction(constants.AuditActionReject).
			SetStage(stage).
			SetAuditUID(uid).
			SetAuditUtype(utype).
			SetAuditAt(now).
			SetAuditNotes(in.AuditNotes)
		if claims.ClientIP != "" {
			auditCreate = auditCreate.SetAuditIP(claims.ClientIP)
		}
		if name := reviewAuditor(ctx); name != "" {
			auditCreate = auditCreate.SetAuditUsername(name)
		}
		if _, terr := auditCreate.Save(ctx); terr != nil {
			return mapEntErr(terr)
		}

		rejectAt := time.Now()
		_, terr := tx.AgentDepositOrder.Update().
			Where(agentdepositorder.IDEQ(in.OrderId)).
			SetStatus(rejectedStatus).
			SetFinalAuditedAt(rejectAt).
			SetReviewerID(0).
			SetReviewerUtype(0).
			ClearClaimedAt().
			SetUpdaterID(uid).
			SetUpdaterType(utype).
			Save(ctx)
		return mapEntErr(terr)
	})
	if err != nil {
		return nil, err
	}
	return &agentdepositorderpb.RejectAgentDepositOrderResp{
		OrderId: in.OrderId,
		Status:  int32(rejectedStatus),
	}, nil
}

// FreezeAgentDepositOrder 冻结订单（仅已领取状态下、领取人可操作，冻结后保留 reviewer_id）。
func (s *Service) FreezeAgentDepositOrder(ctx context.Context, in *agentdepositorderpb.FreezeAgentDepositOrderReq) (*agentdepositorderpb.FreezeAgentDepositOrderResp, error) {
	ctx, claims, err := requireClaims(ctx)
	if err != nil {
		return nil, err
	}
	if in == nil || in.OrderId <= 0 {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	uid, utype := claimsOperatorIDs(ctx)

	// 先查订单，确认存在并校验领取人
	preRow, err := s.getDepositOrder(ctx, in.OrderId,
		agentdepositorder.FieldStatus,
		agentdepositorder.FieldReviewerID,
	)
	if err != nil {
		return nil, err
	}
	if preRow.ReviewerID != uid {
		return nil, xerr.BadRequest(i18n.AgentDepositOrderNotClaimer)
	}
	if preRow.Status != constants.DepositOrderStatusPending1st && preRow.Status != constants.DepositOrderStatusPending2nd {
		return nil, xerr.BadRequest(i18n.AgentDepositOrderNotClaimable)
	}

	err = dao.WithTx(ctx, s.DB, func(tx *ent.Tx) error {
		now := time.Now()
		auditCreate := tx.AgentDepositOrderAudit.Create().
			SetOrderID(in.OrderId).
			SetAction(constants.AuditActionFreeze).
			SetStage(0).
			SetAuditUID(uid).
			SetAuditUtype(utype).
			SetAuditAt(now)
		if claims.ClientIP != "" {
			auditCreate = auditCreate.SetAuditIP(claims.ClientIP)
		}
		if in.FreezeNotes != nil && *in.FreezeNotes != "" {
			auditCreate = auditCreate.SetAuditNotes(*in.FreezeNotes)
		}
		if _, terr := auditCreate.Save(ctx); terr != nil {
			return mapEntErr(terr)
		}

		// reviewer_id / reviewer_utype / claimed_at 保持不变，仅原领取人可解冻
		_, terr := tx.AgentDepositOrder.Update().
			Where(agentdepositorder.IDEQ(in.OrderId)).
			SetStatus(constants.DepositOrderStatusFrozen).
			SetUpdaterID(uid).
			SetUpdaterType(utype).
			Save(ctx)
		return mapEntErr(terr)
	})
	if err != nil {
		return nil, err
	}
	return &agentdepositorderpb.FreezeAgentDepositOrderResp{
		OrderId: in.OrderId,
		Status:  int32(constants.DepositOrderStatusFrozen),
	}, nil
}

// UnfreezeAgentDepositOrder 解冻订单（仅原领取人可操作，解冻后 reviewer_id 保持不变）。
// 目标状态由 audit 记录推断：存在 stage=1 通过记录且 audit_stages=2 → 恢复 12，否则 → 11。
func (s *Service) UnfreezeAgentDepositOrder(ctx context.Context, in *agentdepositorderpb.UnfreezeAgentDepositOrderReq) (*agentdepositorderpb.UnfreezeAgentDepositOrderResp, error) {
	ctx, claims, err := requireClaims(ctx)
	if err != nil {
		return nil, err
	}
	if in == nil || in.OrderId <= 0 {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	uid, utype := claimsOperatorIDs(ctx)

	// 先查订单，确认存在且处于冻结状态，同时取 reviewer_id 校验领取人
	preRow, err := s.getDepositOrder(ctx, in.OrderId,
		agentdepositorder.FieldStatus,
		agentdepositorder.FieldAuditStages,
		agentdepositorder.FieldReviewerID,
	)
	if err != nil {
		return nil, err
	}
	if preRow.Status != constants.DepositOrderStatusFrozen {
		return nil, xerr.BadRequest(i18n.AgentDepositOrderNotFrozen)
	}
	// 仅原领取人可解冻
	if preRow.ReviewerID != uid {
		return nil, xerr.BadRequest(i18n.AgentDepositOrderNotClaimer)
	}

	var newStatus int8
	err = dao.WithTx(ctx, s.DB, func(tx *ent.Tx) error {
		// 查该订单 stage=1 的通过记录是否存在（action=approve）
		stage1Approved, terr := tx.AgentDepositOrderAudit.Query().
			Where(
				agentdepositorderaudit.OrderIDEQ(in.OrderId),
				agentdepositorderaudit.ActionEQ(constants.AuditActionApprove),
				agentdepositorderaudit.StageEQ(1),
			).
			Exist(ctx)
		if terr != nil {
			return mapEntErr(terr)
		}

		// 推断目标状态
		if preRow.AuditStages == 2 && stage1Approved {
			newStatus = constants.DepositOrderStatusPending2nd // 12
		} else {
			newStatus = constants.DepositOrderStatusPending1st // 11
		}

		// 写 audit 记录（action=unfreeze, stage=0）
		now := time.Now()
		auditCreate := tx.AgentDepositOrderAudit.Create().
			SetOrderID(in.OrderId).
			SetAction(constants.AuditActionUnfreeze).
			SetStage(0).
			SetAuditUID(uid).
			SetAuditUtype(utype).
			SetAuditAt(now)
		if claims.ClientIP != "" {
			auditCreate = auditCreate.SetAuditIP(claims.ClientIP)
		}
		if in.UnfreezeNotes != nil && *in.UnfreezeNotes != "" {
			auditCreate = auditCreate.SetAuditNotes(*in.UnfreezeNotes)
		}
		if _, terr := auditCreate.Save(ctx); terr != nil {
			return mapEntErr(terr)
		}

		// reviewer_id / reviewer_utype / claimed_at 保持不变，解冻后订单仍归原领取人
		_, terr = tx.AgentDepositOrder.Update().
			Where(agentdepositorder.IDEQ(in.OrderId)).
			SetStatus(newStatus).
			SetUpdaterID(uid).
			SetUpdaterType(utype).
			Save(ctx)
		return mapEntErr(terr)
	})
	if err != nil {
		return nil, err
	}
	return &agentdepositorderpb.UnfreezeAgentDepositOrderResp{
		OrderId: in.OrderId,
		Status:  int32(newStatus),
	}, nil
}
