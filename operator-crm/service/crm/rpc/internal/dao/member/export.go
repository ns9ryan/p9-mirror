package member

import (
	"context"
	"fmt"
	"net"
	"strings"

	"entgo.io/ent/dialect/sql"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/ent/agentuser"
	m "oa.98ent.com/p9/operator-crm/ent/member"
	"oa.98ent.com/p9/operator-crm/ent/memberauth"
	"oa.98ent.com/p9/operator-crm/ent/memberautolevel"
	"oa.98ent.com/p9/operator-crm/ent/memberfixedlevel"
	"oa.98ent.com/p9/operator-crm/ent/memberfixedlevelassignment"
	"oa.98ent.com/p9/operator-crm/ent/memberprofile"
	"oa.98ent.com/p9/operator-crm/ent/memberstats"
	"oa.98ent.com/p9/operator-crm/ent/memberviplevel"
	"oa.98ent.com/p9/operator-crm/ent/predicate"
	"oa.98ent.com/p9/operator-crm/pkg/agentline"
	contract "oa.98ent.com/p9/operator-crm/pkg/memberexport"
)

func ValidateExportFilters(in ExportFilters) error {
	_, err := exportWhere(in, "validation")
	return err
}

const ExportMaxRows = contract.MaxRows
const ExportBatchSize = 100

type ExportFilters struct {
	List                       ListParams
	IP, Device, ParentUsername string
	AgentType                  int32
	IncludeDemo                bool
}

// Same base builder as List. Additional export predicates are all ANDed.
func exportWhere(in ExportFilters, operator string) ([]predicate.Member, error) {
	if len(in.IP) > 45 || (in.IP != "" && net.ParseIP(in.IP) == nil) || len(in.Device) > 128 || (in.ParentUsername != "" && strings.TrimSpace(in.ParentUsername) == "") || len(in.ParentUsername) > 64 || in.AgentType < 0 || in.AgentType > 2 {
		return nil, fmt.Errorf("invalid export filter")
	}
	filters, err := listWhere(in.List, operator)
	if err != nil {
		return nil, err
	}
	// 试玩身份取决于所属代理线路；只有显式选择包含试玩时才移除排除条件。
	// 无论是否包含试玩，都保留同厅约束及其他筛选条件。
	agentFilters := []predicate.AgentUser{agentuser.OperatorCodeEQ(operator)}
	if !in.IncludeDemo {
		agentFilters = append(agentFilters, agentuser.AgentLineCodeNotIn(agentline.CashDemo, agentline.CreditDemo))
	}
	if in.AgentType != 0 {
		agentFilters = append(agentFilters, agentuser.AgentTypeEQ(int64(in.AgentType)))
	}
	filters = append(filters, m.HasAgentWith(agentFilters...))
	if in.IP != "" {
		filters = append(filters, m.Or(m.RegisterIPEQ(in.IP), m.LastLoginIPEQ(in.IP)))
	}
	if in.Device != "" {
		filters = append(filters, m.Or(m.RegisterDeviceCodeEQ(in.Device), m.LastLoginDeviceCodeEQ(in.Device)))
	}
	if in.ParentUsername != "" {
		filters = append(filters, m.HasInviterWith(m.OperatorCodeEQ(operator), m.HasAuthWith(memberauth.OperatorCodeEQ(operator), memberauth.UsernameEqualFold(strings.TrimSpace(in.ParentUsername)))))
	}
	return filters, nil
}

// A single bounded membership probe. The IDs freeze only membership, not row values.
func (s *Service) ExportIDs(ctx context.Context, in ExportFilters) ([]int64, error) {
	filters, err := exportWhere(in, ctxdata.OperatorCodeFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	return s.exportDB.Member.Query().Where(filters...).Order(m.ByID(sql.OrderDesc())).Limit(ExportMaxRows + 1).IDs(ctx)
}

// Selected IDs are immutable task membership. Never re-run UI List or COUNT.
func (s *Service) ExportBatch(ctx context.Context, ids []int64) ([]*ent.Member, error) {
	op := ctxdata.OperatorCodeFromCtx(ctx)
	if op == "" || len(ids) == 0 || len(ids) > ExportBatchSize {
		return nil, fmt.Errorf("invalid export batch")
	}
	rows, err := s.exportDB.Member.Query().Where(m.OperatorCodeEQ(op), m.IDIn(ids...)).Order(m.ByID(sql.OrderDesc())).
		WithAuth(func(q *ent.MemberAuthQuery) {
			q.Where(memberauth.OperatorCodeEQ(op)).Select(memberauth.FieldID, memberauth.FieldMemberID, memberauth.FieldUsername, memberauth.FieldPhoneCiphertext, memberauth.FieldEmailCiphertext)
		}).
		WithProfile(func(q *ent.MemberProfileQuery) {
			q.Where(memberprofile.OperatorCodeEQ(op)).Select(memberprofile.FieldID, memberprofile.FieldMemberID, memberprofile.FieldNickname)
		}).
		WithAgent(func(q *ent.AgentUserQuery) {
			q.Where(agentuser.OperatorCodeEQ(op)).Select(agentuser.FieldID, agentuser.FieldAccount, agentuser.FieldAgentLineCode)
		}).
		WithAutoLevel(func(q *ent.MemberAutoLevelQuery) {
			q.Where(memberautolevel.OperatorCodeEQ(op)).Select(memberautolevel.FieldID, memberautolevel.FieldName)
		}).
		WithVipLevel(func(q *ent.MemberVipLevelQuery) {
			q.Where(memberviplevel.OperatorCodeEQ(op)).Select(memberviplevel.FieldID, memberviplevel.FieldLevel)
		}).
		WithFixedLevelAssignments(func(q *ent.MemberFixedLevelAssignmentQuery) {
			q.Where(memberfixedlevelassignment.HasMemberWith(m.OperatorCodeEQ(op))).WithFixedLevel(func(q *ent.MemberFixedLevelQuery) {
				q.Where(memberfixedlevel.OperatorCodeEQ(op)).Select(memberfixedlevel.FieldID, memberfixedlevel.FieldName)
			})
		}).
		WithStats(func(q *ent.MemberStatsQuery) {
			q.Where(memberstats.OperatorCodeEQ(op)).Select(memberstats.FieldID, memberstats.FieldMemberID, memberstats.FieldTotalDeposit, memberstats.FieldTotalWithdrawal)
		}).All(ctx)
	if err != nil {
		return nil, err
	}
	if len(rows) != len(ids) {
		return nil, fmt.Errorf("export membership changed")
	}
	return rows, nil
}
