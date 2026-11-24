package agent

import (
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/ent/agentdepositorder"
	"oa.98ent.com/p9/operator-crm/ent/agentdepositorderaudit"
	"oa.98ent.com/p9/operator-crm/ent/agentuser"
	entmember "oa.98ent.com/p9/operator-crm/ent/member"
	"oa.98ent.com/p9/operator-crm/ent/memberauth"
	"oa.98ent.com/p9/operator-crm/ent/memberautolevel"
	"oa.98ent.com/p9/operator-crm/ent/predicate"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/constants"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentdepositorderpb"
)

// AddAgentDepositOrder 创建存款订单。
// deposit_currency 固定为 USDT；deposit_type 由调用方传入（ 1 加密货币 / 2 人工入账）。
func (s *Service) AddAgentDepositOrder(ctx context.Context, in *agentdepositorderpb.AddAgentDepositOrderReq, orderNo string) (*agentdepositorderpb.AddAgentDepositOrderResp, error) {
	ctx, _, err := requireOperator(ctx)
	if err != nil {
		return nil, err
	}
	if in == nil {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	if in.MemberId <= 0 || in.DepositAmount <= 0 {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}

	// 从会员归属推断代理 ID
	m, merr := s.DB.Member.Query().
		Where(entmember.IDEQ(in.MemberId)).
		Select(entmember.FieldID, entmember.FieldAgentID).
		Only(ctx)
	if merr != nil {
		if ent.IsNotFound(merr) {
			return nil, xerr.BadRequest(i18n.InvalidParam)
		}
		return nil, mapEntErr(merr)
	}
	agentID := m.AgentID
	if agentID <= 0 {
		return nil, xerr.BadRequest(i18n.AgentNotFound)
	}

	// 读取代理基础信息填充订单归属字段
	agentRow, aerr := s.FecthAgentByID(ctx, agentID,
		agentuser.FieldID,
		agentuser.FieldAgentType,
		agentuser.FieldAgentLineCode,
		agentuser.FieldParentID,
	)
	if aerr != nil {
		if ent.IsNotFound(aerr) {
			return nil, xerr.BadRequest(i18n.AgentNotFound)
		}
		return nil, aerr
	}

	// 计算奖励金额：deposit_amount × bonus_rate / 100，向下取整
	bonusAmount := int64(float64(in.DepositAmount) * in.BonusRate / 100)

	// 到账金额 = 存款金额 + 奖金金额；实际到账金额 = 到账金额 - 手续费（创建时手续费为 0）
	receivedAmount := in.DepositAmount + bonusAmount
	actualReceivedAmount := receivedAmount // fee_amount 默认 0，创建时与 received_amount 相等

	// 计算审核阶段数：存款金额超过财务二审阈值则需二审，否则一审。
	// TODO: 有财务设置功能后从财务配置表读取【财务二审金额】字段替换 constants.FinanceSecondAuditThreshold。
	auditStages := int8(1)
	if in.DepositAmount > constants.FinanceSecondAuditThreshold {
		auditStages = int8(2)
	}
	// 创建时初始状态固定为待一审（11）
	initialStatus := constants.DepositOrderStatusPending1st

	// 从 ctx 中读取操作人身份，填写创建人信息
	creatorID, creatorType := claimsOperatorIDs(ctx)

	c := s.DB.AgentDepositOrder.Create().
		SetMemberID(in.MemberId).
		SetDepositType(int8(in.DepositType)).
		SetDepositAmount(in.DepositAmount).
		SetDepositCurrency(constants.DepositCurrency).
		SetSiteCurrency(constants.DepositCurrency).
		SetRolloverMultiplier(in.RolloverMultiplier).
		SetBonusRate(in.BonusRate).
		SetBonusAmount(bonusAmount).
		SetReceivedAmount(receivedAmount).
		SetActualReceivedAmount(actualReceivedAmount).
		SetAgentID(agentID).
		SetAgentType(int8(agentRow.AgentType)).
		SetAgentLineCode(agentRow.AgentLineCode).
		SetAuditStages(auditStages).
		SetStatus(initialStatus).
		SetCreatorType(creatorType).
		SetCreatorID(creatorID)

	if orderNo != "" {
		c.SetOrderNo(orderNo)
	}

	row, err := c.Save(ctx)
	if err != nil {
		return nil, mapEntErr(err)
	}
	return &agentdepositorderpb.AddAgentDepositOrderResp{Id: row.ID}, nil
}

// GetMemberTodayDeposit 统计指定会员今日累计有效充值金额及笔数。
// 今日范围：UTC 零点至当前时刻；仅统计 status=4（通过）的订单。
func (s *Service) GetMemberTodayDeposit(ctx context.Context, memberID int64) (totalAmount int64, count int64, err error) {
	if memberID <= 0 {
		return 0, 0, xerr.BadRequest(i18n.InvalidParam)
	}
	now := time.Now().UTC()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	preds := []predicate.AgentDepositOrder{
		agentdepositorder.MemberIDEQ(memberID),
		agentdepositorder.StatusEQ(constants.DepositOrderStatusApproved),
		agentdepositorder.CreatedAtGTE(todayStart),
	}

	rows, err := s.DB.AgentDepositOrder.Query().
		Where(preds...).
		Select(agentdepositorder.FieldDepositAmount).
		All(ctx)
	if err != nil {
		return 0, 0, mapEntErr(err)
	}
	for _, row := range rows {
		totalAmount += row.DepositAmount
	}
	return totalAmount, int64(len(rows)), nil
}

// CountMemberDepositOrders 统计该会员人工存款订单数量，
// 用于活动侧 DepositPayload.deposit_count（首充/N 充条件判断）。
func (s *Service) CountMemberDepositOrders(ctx context.Context, memberID int64, status int8) (int, error) {
	return s.DB.AgentDepositOrder.Query().
		Where(
			agentdepositorder.MemberIDEQ(memberID),
			agentdepositorder.StatusEQ(status),
		).
		Count(ctx)
}

// ListAgentDepositOrders 分页查询存款订单列表，并关联查询会员账号和层级名称。
func (s *Service) ListAgentDepositOrders(ctx context.Context, in *agentdepositorderpb.ListAgentDepositOrdersReq) (*agentdepositorderpb.ListAgentDepositOrdersResp, error) {
	ctx, _, err := requireOperator(ctx)
	if err != nil {
		return nil, err
	}
	if in == nil {
		in = &agentdepositorderpb.ListAgentDepositOrdersReq{}
	}

	page := max(in.Page, 1)
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	preds := make([]predicate.AgentDepositOrder, 0, 12)

	if in.Status > 0 {
		valid := (in.Status >= 1 && in.Status <= 5) || in.Status == 11 || in.Status == 12
		if !valid {
			return nil, xerr.BadRequest(i18n.InvalidParam)
		}
		preds = append(preds, agentdepositorder.StatusEQ(int8(in.Status)))
	}
	if in.MemberId > 0 {
		preds = append(preds, agentdepositorder.MemberIDEQ(in.MemberId))
	}
	if in.DepositType > 0 {
		if in.DepositType < 1 || in.DepositType > 3 {
			return nil, xerr.BadRequest(i18n.InvalidParam)
		}
		preds = append(preds, agentdepositorder.DepositTypeEQ(int8(in.DepositType)))
	}
	if in.TimeStart > 0 {
		preds = append(preds, agentdepositorder.CreatedAtGTE(time.Unix(in.TimeStart, 0)))
	}
	if in.TimeEnd > 0 {
		preds = append(preds, agentdepositorder.CreatedAtLTE(time.Unix(in.TimeEnd, 0)))
	}

	// 存款货币
	if in.DepositCurrency != nil && *in.DepositCurrency != "" {
		preds = append(preds, agentdepositorder.DepositCurrencyEQ(*in.DepositCurrency))
	}
	// 订单号（精确）
	if in.OrderNo != nil && *in.OrderNo != "" {
		preds = append(preds, agentdepositorder.OrderNoEQ(*in.OrderNo))
	}
	// 三方订单号（精确）
	if in.ThirdPartyOrderNo != nil && *in.ThirdPartyOrderNo != "" {
		preds = append(preds, agentdepositorder.ThirdPartyOrderNoEQ(*in.ThirdPartyOrderNo))
	}
	// 领取人过滤：指定 reviewer_id 时只查该领取人名下的单（"我的待审核"）
	if in.ReviewerId != nil && *in.ReviewerId > 0 {
		preds = append(preds, agentdepositorder.ReviewerIDEQ(*in.ReviewerId))
	}
	// 领取状态过滤：1=未领取(reviewer_id=0) 2=已领取(reviewer_id>0)
	if in.ClaimStatus != nil && *in.ClaimStatus > 0 {
		switch *in.ClaimStatus {
		case 1:
			preds = append(preds, agentdepositorder.ReviewerIDEQ(0))
		case 2:
			preds = append(preds, agentdepositorder.ReviewerIDGT(0))
		default:
			return nil, xerr.BadRequest(i18n.InvalidParam)
		}
	}
	// 会员账号：先查 member_auth 得到 member_id，再加入过滤
	if in.MemberAccount != nil && *in.MemberAccount != "" {
		authRow, authErr := s.DB.MemberAuth.Query().
			Where(memberauth.UsernameEQ(*in.MemberAccount)).
			Select(memberauth.FieldMemberID).
			First(ctx)
		if authErr != nil {
			if ent.IsNotFound(authErr) {
				// 账号不存在，直接返回空列表
				return &agentdepositorderpb.ListAgentDepositOrdersResp{List: nil, Total: 0}, nil
			}
			return nil, mapEntErr(authErr)
		}
		preds = append(preds, agentdepositorder.MemberIDEQ(authRow.MemberID))
	}

	q := s.DB.AgentDepositOrder.Query()
	if len(preds) > 0 {
		q = q.Where(preds...)
	}
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, mapEntErr(err)
	}
	rows, err := q.
		Order(agentdepositorder.ByID(sql.OrderDesc())).
		Offset(int((page - 1) * pageSize)).
		Limit(int(pageSize)).
		All(ctx)
	if err != nil {
		return nil, mapEntErr(err)
	}

	// 收集去重 member_id，批量查会员账号和层级
	memberAccounts, memberLevelNames := s.batchLoadMemberInfo(ctx, rows)
	// 批量加载状态 4/5 订单的最终审核记录
	auditMap := s.batchLoadAuditInfo(ctx, rows)

	items := make([]*agentdepositorderpb.AgentDepositOrderListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, toDepositOrderListItem(row, memberAccounts, memberLevelNames, auditMap))
	}
	return &agentdepositorderpb.ListAgentDepositOrdersResp{
		List:  items,
		Total: int64(total),
	}, nil
}

// batchLoadMemberInfo 根据订单行批量加载会员账号和自动层级名称。
// 分三步：① member_auth 查账号；② member 查 auto_level_id；③ member_auto_level 查层级名称。
// 任何一步查询失败只记录缺失，不中断主流程。
func (s *Service) batchLoadMemberInfo(ctx context.Context, rows []*ent.AgentDepositOrder) (accounts map[int64]string, levelNames map[int64]string) {
	accounts = make(map[int64]string)
	levelNames = make(map[int64]string)

	seen := make(map[int64]bool, len(rows))
	memberIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		if row.MemberID > 0 && !seen[row.MemberID] {
			memberIDs = append(memberIDs, row.MemberID)
			seen[row.MemberID] = true
		}
	}
	if len(memberIDs) == 0 {
		return
	}

	// ① 批量查会员登录账号（member_auth）
	authRows, err := s.DB.MemberAuth.Query().
		Where(memberauth.MemberIDIn(memberIDs...)).
		Select(memberauth.FieldMemberID, memberauth.FieldUsername).
		All(ctx)
	if err == nil {
		for _, a := range authRows {
			if a.Username != nil {
				accounts[a.MemberID] = *a.Username
			}
		}
	}

	// ② 批量查会员 auto_level_id（member）
	memberRows, err := s.DB.Member.Query().
		Where(entmember.IDIn(memberIDs...)).
		Select(entmember.FieldID, entmember.FieldAutoLevelID).
		All(ctx)
	if err != nil {
		return
	}

	// 收集去重 auto_level_id，建 levelID → []memberID 反查映射
	levelIDSet := make(map[int64]struct{}, len(memberRows))
	memberToLevel := make(map[int64]int64, len(memberRows))
	for _, m := range memberRows {
		memberToLevel[m.ID] = m.AutoLevelID
		levelIDSet[m.AutoLevelID] = struct{}{}
	}
	levelIDs := make([]int64, 0, len(levelIDSet))
	for lid := range levelIDSet {
		levelIDs = append(levelIDs, lid)
	}

	// ③ 批量查自动层级名称（member_auto_level）
	levelRows, err := s.DB.MemberAutoLevel.Query().
		Where(memberautolevel.IDIn(levelIDs...)).
		Select(memberautolevel.FieldID, memberautolevel.FieldName).
		All(ctx)
	if err != nil {
		return
	}
	nameByLevelID := make(map[int64]string, len(levelRows))
	for _, lv := range levelRows {
		nameByLevelID[lv.ID] = lv.Name
	}

	// 将层级名称映射回 member_id
	for memberID, levelID := range memberToLevel {
		if name, ok := nameByLevelID[levelID]; ok {
			levelNames[memberID] = name
		}
	}
	return
}

// auditSummary 存款订单最终审核记录摘要（仅 status=4/5 时填充）。
type auditSummary struct {
	AuditUID      int64
	AuditUsername string
	AuditAt       time.Time
}

// batchLoadAuditInfo 批量加载 status=4/5 订单的最终审核记录（approve/reject）。
// 对每个 order_id 取 id 最大（即最新）的一条 approve/reject 审核记录。
// 任何查询失败只记录缺失，不中断主流程。
func (s *Service) batchLoadAuditInfo(ctx context.Context, rows []*ent.AgentDepositOrder) map[int64]*auditSummary {
	result := make(map[int64]*auditSummary)

	// 筛选终态订单 ID（status=4 通过，status=5 拒绝）
	seen := make(map[int64]bool, len(rows))
	orderIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		if (row.Status == constants.DepositOrderStatusApproved || row.Status == constants.DepositOrderStatusRejected) && !seen[row.ID] {
			orderIDs = append(orderIDs, row.ID)
			seen[row.ID] = true
		}
	}
	if len(orderIDs) == 0 {
		return result
	}

	// 查 approve/reject 审核记录，取 order_id IN (...) & action IN (1, 2)
	auditRows, err := s.DB.AgentDepositOrderAudit.Query().
		Where(
			agentdepositorderaudit.OrderIDIn(orderIDs...),
			agentdepositorderaudit.ActionIn(constants.AuditActionApprove, constants.AuditActionReject),
		).
		Select(
			agentdepositorderaudit.FieldID,
			agentdepositorderaudit.FieldOrderID,
			agentdepositorderaudit.FieldAuditUID,
			agentdepositorderaudit.FieldAuditUsername,
			agentdepositorderaudit.FieldAuditAt,
		).
		All(ctx)
	if err != nil {
		return result
	}

	// 按 order_id 分组，取 id 最大的一条（最新审核记录）
	type idAndAudit struct {
		id    int64
		audit *auditSummary
	}
	latest := make(map[int64]idAndAudit, len(auditRows))
	for _, ar := range auditRows {
		if prev, ok := latest[ar.OrderID]; !ok || ar.ID > prev.id {
			latest[ar.OrderID] = idAndAudit{
				id: ar.ID,
				audit: &auditSummary{
					AuditUID:      ar.AuditUID,
					AuditUsername: ar.AuditUsername,
					AuditAt:       ar.AuditAt,
				},
			}
		}
	}
	for orderID, v := range latest {
		result[orderID] = v.audit
	}
	return result
}

func toDepositOrderListItem(row *ent.AgentDepositOrder, accounts map[int64]string, levelNames map[int64]string, auditMap map[int64]*auditSummary) *agentdepositorderpb.AgentDepositOrderListItem {
	if row == nil {
		return nil
	}

	// 实际到账金额直接从 DB 字段读取
	item := &agentdepositorderpb.AgentDepositOrderListItem{
		Id:                   row.ID,
		OrderNo:              row.OrderNo,
		MemberId:             row.MemberID,
		MemberAccount:        accounts[row.MemberID],
		MemberLevelName:      levelNames[row.MemberID],
		CreateAt:             row.CreatedAt.Unix(),
		DepositCurrency:      row.DepositCurrency,
		DepositAmount:        row.DepositAmount,
		SiteCurrency:         row.SiteCurrency,
		SiteAmount:           row.DepositAmount, // 暂以存款金额作为站点金额（1:1）
		DepositType:          int32(row.DepositType),
		ReceivedAmount:       row.ReceivedAmount,
		FeeAmount:            row.FeeAmount,
		ActualReceivedAmount: row.ActualReceivedAmount,
		Status:               int32(row.Status),
		ReviewerId:           row.ReviewerID,
		ReviewerUtype:        int32(row.ReviewerUtype),
	}
	if row.ClaimedAt != nil {
		item.ClaimedAt = row.ClaimedAt.Unix()
	}
	// 状态 4（通过）或 5（拒绝）时填充最终审核人信息
	if row.Status == constants.DepositOrderStatusApproved || row.Status == constants.DepositOrderStatusRejected {
		if a, ok := auditMap[row.ID]; ok {
			item.AuditorUid = a.AuditUID
			item.AuditorUsername = a.AuditUsername
			item.AuditedAt = a.AuditAt.Unix()
		}
	}
	if row.ThirdPartyOrderNo != "" {
		v := row.ThirdPartyOrderNo
		item.ThirdPartyOrderNo = &v
	}
	if row.PaymentReceiveInfo != "" {
		v := row.PaymentReceiveInfo
		item.PaymentReceiveInfo = &v
	}
	if row.PaymentSendInfo != "" {
		v := row.PaymentSendInfo
		item.PaymentSendInfo = &v
	}
	return item
}
