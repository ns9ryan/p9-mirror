package agent

import (
	"context"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/ent/agentloginlog"
	"oa.98ent.com/p9/operator-crm/ent/predicate"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"
)

const (
	assocKeywordLoginIP int32 = 1
	assocKeywordUdid    int32 = 2
)

// ListAgentLoginAssociations 关联检索登录日志，附带同 IP / 同设备代理数。
func (s *Service) ListAgentLoginAssociations(ctx context.Context, in *agentpb.ListAgentLoginAssociationsReq) (*agentpb.ListAgentLoginAssociationsResp, error) {
	ctx, _, err := requireOperator(ctx)
	if err != nil {
		return nil, err
	}
	if in == nil {
		in = &agentpb.ListAgentLoginAssociationsReq{}
	}

	page := max(in.Page, 1)
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	preds, err := buildAssociationListPreds(in)
	if err != nil {
		return nil, err
	}

	q := s.DB.AgentLoginLog.Query()
	if len(preds) > 0 {
		q = q.Where(preds...)
	}
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := q.
		Order(agentloginlog.ByLoginAt(sql.OrderDesc()), agentloginlog.ByID(sql.OrderDesc())).
		Offset(int((page - 1) * pageSize)).
		Limit(int(pageSize)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]*agentpb.AgentLoginAssociationItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, toAssociationItem(row))
	}
	if err := s.attachAssociationCounts(ctx, in, items); err != nil {
		return nil, err
	}
	return &agentpb.ListAgentLoginAssociationsResp{Items: items, Total: int64(total)}, nil
}

func buildAssociationListPreds(in *agentpb.ListAgentLoginAssociationsReq) ([]predicate.AgentLoginLog, error) {
	if in.DeviceType > 0 && (in.DeviceType < 1 || in.DeviceType > 5) {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	preds := make([]predicate.AgentLoginLog, 0, 6)
	preds = append(preds, associationScopePreds(in)...)
	if in.AgentId > 0 {
		preds = append(preds, agentloginlog.AgentIDEQ(in.AgentId))
	}
	keyword := strings.TrimSpace(in.Keyword)
	if keyword == "" || in.KeywordType == 0 {
		return preds, nil
	}
	switch in.KeywordType {
	case assocKeywordLoginIP:
		preds = append(preds, agentloginlog.LoginIPEQ(keyword))
	case assocKeywordUdid:
		preds = append(preds, agentloginlog.UdidEQ(keyword))
	default:
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	return preds, nil
}

// associationScopePreds 关联计数范围：时间、设备类型。
func associationScopePreds(in *agentpb.ListAgentLoginAssociationsReq) []predicate.AgentLoginLog {
	preds := make([]predicate.AgentLoginLog, 0, 3)
	if in.TimeStart > 0 {
		preds = append(preds, agentloginlog.LoginAtGTE(time.Unix(in.TimeStart, 0)))
	}
	if in.TimeEnd > 0 {
		preds = append(preds, agentloginlog.LoginAtLTE(time.Unix(in.TimeEnd, 0)))
	}
	if in.DeviceType > 0 {
		preds = append(preds, agentloginlog.DeviceTypeEQ(int64(in.DeviceType)))
	}
	return preds
}

func (s *Service) attachAssociationCounts(ctx context.Context, in *agentpb.ListAgentLoginAssociationsReq, items []*agentpb.AgentLoginAssociationItem) error {
	if len(items) == 0 {
		return nil
	}
	ips := make([]string, 0, len(items))
	udids := make([]string, 0, len(items))
	seenIP := make(map[string]struct{}, len(items))
	seenUdid := make(map[string]struct{}, len(items))
	for _, it := range items {
		if it == nil {
			continue
		}
		if ip := strings.TrimSpace(it.LoginIp); ip != "" {
			if _, ok := seenIP[ip]; !ok {
				seenIP[ip] = struct{}{}
				ips = append(ips, ip)
			}
		}
		if udid := strings.TrimSpace(it.Udid); udid != "" {
			if _, ok := seenUdid[udid]; !ok {
				seenUdid[udid] = struct{}{}
				udids = append(udids, udid)
			}
		}
	}

	scope := associationScopePreds(in)
	ipCounts, err := s.countDistinctAgentsByIP(ctx, scope, ips)
	if err != nil {
		return err
	}
	udidCounts, err := s.countDistinctAgentsByUdid(ctx, scope, udids)
	if err != nil {
		return err
	}
	for _, it := range items {
		if it == nil {
			continue
		}
		if ip := strings.TrimSpace(it.LoginIp); ip != "" {
			it.SameIpAgentCount = ipCounts[ip]
		}
		if udid := strings.TrimSpace(it.Udid); udid != "" {
			it.SameDeviceAgentCount = udidCounts[udid]
		}
	}
	return nil
}

func (s *Service) countDistinctAgentsByIP(ctx context.Context, scope []predicate.AgentLoginLog, ips []string) (map[string]int64, error) {
	out := make(map[string]int64, len(ips))
	if len(ips) == 0 {
		return out, nil
	}
	preds := append(append([]predicate.AgentLoginLog{}, scope...),
		agentloginlog.LoginIPIn(ips...),
		agentloginlog.AgentIDNotNil(),
	)
	rows, err := s.DB.AgentLoginLog.Query().
		Where(preds...).
		Select(agentloginlog.FieldLoginIP, agentloginlog.FieldAgentID).
		All(ctx)
	if err != nil {
		return nil, err
	}
	agents := make(map[string]map[int64]struct{}, len(ips))
	for _, row := range rows {
		if row.LoginIP == nil || row.AgentID == nil {
			continue
		}
		ip := *row.LoginIP
		set, ok := agents[ip]
		if !ok {
			set = make(map[int64]struct{})
			agents[ip] = set
		}
		set[*row.AgentID] = struct{}{}
	}
	for ip, set := range agents {
		out[ip] = int64(len(set))
	}
	return out, nil
}

func (s *Service) countDistinctAgentsByUdid(ctx context.Context, scope []predicate.AgentLoginLog, udids []string) (map[string]int64, error) {
	out := make(map[string]int64, len(udids))
	if len(udids) == 0 {
		return out, nil
	}
	preds := append(append([]predicate.AgentLoginLog{}, scope...),
		agentloginlog.UdidIn(udids...),
		agentloginlog.AgentIDNotNil(),
	)
	rows, err := s.DB.AgentLoginLog.Query().
		Where(preds...).
		Select(agentloginlog.FieldUdid, agentloginlog.FieldAgentID).
		All(ctx)
	if err != nil {
		return nil, err
	}
	agents := make(map[string]map[int64]struct{}, len(udids))
	for _, row := range rows {
		if row.Udid == nil || row.AgentID == nil {
			continue
		}
		udid := *row.Udid
		set, ok := agents[udid]
		if !ok {
			set = make(map[int64]struct{})
			agents[udid] = set
		}
		set[*row.AgentID] = struct{}{}
	}
	for udid, set := range agents {
		out[udid] = int64(len(set))
	}
	return out, nil
}

func toAssociationItem(row *ent.AgentLoginLog) *agentpb.AgentLoginAssociationItem {
	if row == nil {
		return &agentpb.AgentLoginAssociationItem{}
	}
	item := &agentpb.AgentLoginAssociationItem{
		Id:         row.ID,
		Account:    row.Account,
		DeviceType: int32(row.DeviceType),
	}
	if row.AgentID != nil {
		item.AgentId = *row.AgentID
	}
	if row.LoginIP != nil {
		item.LoginIp = *row.LoginIP
	}
	if row.LoginLocation != nil {
		item.LoginLocation = *row.LoginLocation
	}
	if row.Udid != nil {
		item.Udid = *row.Udid
	}
	if row.LoginDevice != nil {
		item.LoginDevice = *row.LoginDevice
	}
	if !row.LoginAt.IsZero() {
		item.LoginAt = row.LoginAt.Unix()
	}
	return item
}
