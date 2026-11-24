package agent

import (
	"context"
	"time"

	"entgo.io/ent/dialect/sql"

	"oa.98ent.com/p9/common/utils/orderno"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent"
	entmember "oa.98ent.com/p9/operator-crm/ent/member"
	"oa.98ent.com/p9/operator-crm/ent/memberauth"
	"oa.98ent.com/p9/operator-crm/ent/memberautolevel"
	"oa.98ent.com/p9/operator-crm/ent/predicate"
	"oa.98ent.com/p9/operator-crm/ent/turnover"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/turnoverpb"
)

// turnoverOrderNoGen 流水订单号生成器，nodeID 自动从本机内网 IP 推导。
// 订单号前缀 TRN（Turnover）。
var turnoverOrderNoGen = orderno.NewSnowflake(0)

const turnoverOrderNoPrefix = "TRN"

// AddTurnover 创建流水记录。
// 消费货币固定 USDT，消费类型固定游戏（1），消费时间取当前时间，
// consume_amount 正数表示赢（余额增加），负数表示输（余额减少），不可为 0，
// 消费后余额 = balance_before + consume_amount，余额货币固定 USDT，
// 代理 ID 由会员归属推断，订单号内部自动生成。
func (s *Service) AddTurnover(ctx context.Context, in *turnoverpb.AddTurnoverReq) (*turnoverpb.AddTurnoverResp, error) {
	ctx, _, err := requireOperator(ctx)
	if err != nil {
		return nil, err
	}
	if in == nil {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	if in.MemberId <= 0 {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}

	// 代理 ID 从会员归属推断
	m, merr := s.DB.Member.Query().
		Where(entmember.IDEQ(in.MemberId)).
		Select(entmember.FieldAgentID).
		Only(ctx)
	if merr != nil {
		if ent.IsNotFound(merr) {
			return nil, xerr.BadRequest(i18n.InvalidParam)
		}
		return nil, mapEntErr(merr)
	}
	agentID := m.AgentID

	// 消费后余额自动计算：正数赢钱余额涨，负数输钱余额跌
	balanceAfter := in.BalanceBefore + in.ConsumeAmount

	// 订单号内部生成
	orderNo := turnoverOrderNoGen.GenerateWithPrefix(turnoverOrderNoPrefix)

	c := s.DB.Turnover.Create().
		SetMemberID(in.MemberId).
		SetConsumeAmount(in.ConsumeAmount).
		SetConsumeCurrency("USDT").
		SetConsumeType(1).
		SetConsumeAt(time.Now()).
		SetBalanceBefore(in.BalanceBefore).
		SetBalanceAfter(balanceAfter).
		SetBalanceCurrency("USDT").
		SetAgentID(agentID).
		SetOrderNo(orderNo)

	if in.Notes != nil {
		c.SetNotes(*in.Notes)
	}
	if in.GameId != nil {
		c.SetGameID(*in.GameId)
	}

	row, err := c.Save(ctx)
	if err != nil {
		return nil, mapEntErr(err)
	}
	return &turnoverpb.AddTurnoverResp{Id: row.ID}, nil
}

// ListTurnovers 分页查询流水列表。
func (s *Service) ListTurnovers(ctx context.Context, in *turnoverpb.ListTurnoversReq) (*turnoverpb.ListTurnoversResp, error) {
	ctx, _, err := requireOperator(ctx)
	if err != nil {
		return nil, err
	}
	if in == nil {
		in = &turnoverpb.ListTurnoversReq{}
	}

	page := max(in.Page, 1)
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	preds := make([]predicate.Turnover, 0, 3)
	if in.MemberId > 0 {
		preds = append(preds, turnover.MemberIDEQ(in.MemberId))
	}
	if in.TimeStart > 0 {
		preds = append(preds, turnover.ConsumeAtGTE(time.Unix(in.TimeStart, 0)))
	}
	if in.TimeEnd > 0 {
		preds = append(preds, turnover.ConsumeAtLTE(time.Unix(in.TimeEnd, 0)))
	}

	q := s.DB.Turnover.Query()
	if len(preds) > 0 {
		q = q.Where(preds...)
	}

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, mapEntErr(err)
	}
	rows, err := q.
		Order(turnover.ByID(sql.OrderDesc())).
		Offset(int((page - 1) * pageSize)).
		Limit(int(pageSize)).
		All(ctx)
	if err != nil {
		return nil, mapEntErr(err)
	}

	// 收集当前页的去重 member_id
	memberIDs := make([]int64, 0, len(rows))
	seen := make(map[int64]struct{}, len(rows))
	for _, row := range rows {
		if _, ok := seen[row.MemberID]; !ok {
			seen[row.MemberID] = struct{}{}
			memberIDs = append(memberIDs, row.MemberID)
		}
	}

	// 批量查询会员登录账号和自动层级名称
	usernameMap := make(map[int64]string, len(memberIDs))
	levelNameMap := make(map[int64]string, len(memberIDs))
	if len(memberIDs) > 0 {
		members, merr := s.DB.Member.Query().
			Where(entmember.IDIn(memberIDs...)).
			Select(entmember.FieldID, entmember.FieldAutoLevelID).
			WithAuth(func(q *ent.MemberAuthQuery) {
				q.Select(memberauth.FieldMemberID, memberauth.FieldUsername)
			}).
			WithAutoLevel(func(q *ent.MemberAutoLevelQuery) {
				q.Select(memberautolevel.FieldID, memberautolevel.FieldName)
			}).
			All(ctx)
		if merr != nil {
			return nil, mapEntErr(merr)
		}
		for _, m := range members {
			if m.Edges.Auth != nil && m.Edges.Auth.Username != nil {
				usernameMap[m.ID] = *m.Edges.Auth.Username
			}
			if m.Edges.AutoLevel != nil {
				levelNameMap[m.ID] = m.Edges.AutoLevel.Name
			}
		}
	}

	items := make([]*turnoverpb.TurnoverListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, toTurnoverListItem(row, usernameMap, levelNameMap))
	}
	return &turnoverpb.ListTurnoversResp{List: items, Total: int64(total)}, nil
}

func toTurnoverListItem(row *ent.Turnover, usernameMap, levelNameMap map[int64]string) *turnoverpb.TurnoverListItem {
	if row == nil {
		return nil
	}
	item := &turnoverpb.TurnoverListItem{
		Id:              row.ID,
		MemberId:        row.MemberID,
		ConsumeAmount:   row.ConsumeAmount,
		ConsumeCurrency: row.ConsumeCurrency,
		ConsumeType:     int32(row.ConsumeType),
		BalanceBefore:   row.BalanceBefore,
		BalanceAfter:    row.BalanceAfter,
		BalanceCurrency: row.BalanceCurrency,
		AgentId:         row.AgentID,
		WalletId:        row.WalletID,
		CreatedAt:       row.CreatedAt.Unix(),
		MemberAccount:   usernameMap[row.MemberID],
		MemberLevelName: levelNameMap[row.MemberID],
	}
	if !row.ConsumeAt.IsZero() {
		item.ConsumeAt = row.ConsumeAt.Unix()
	}
	if row.Notes != "" {
		v := row.Notes
		item.Notes = &v
	}
	if row.GameID != 0 {
		v := row.GameID
		item.GameId = &v
	}
	if row.OrderNo != "" {
		v := row.OrderNo
		item.OrderNo = &v
	}
	return item
}
