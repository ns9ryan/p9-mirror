package agent

import (
	"context"
	"strconv"
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
	loginLogKeywordAgentID  int32 = 1 // 代理ID
	loginLogKeywordLoginIP  int32 = 2 // 登录IP
	loginLogKeywordUdid     int32 = 3 // 设备码
	loginLogKeywordLocation int32 = 4 // 登录地点
	loginLogKeywordAccount  int32 = 5 // 登录账号
)

// ListAgentLoginLogs 分页查询代理登录日志。
func (s *Service) ListAgentLoginLogs(ctx context.Context, in *agentpb.ListAgentLoginLogsReq) (*agentpb.ListAgentLoginLogsResp, error) {
	ctx, _, err := requireOperator(ctx)
	if err != nil {
		return nil, err
	}
	if in == nil {
		in = &agentpb.ListAgentLoginLogsReq{}
	}

	page := max(in.Page, 1)
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	preds, err := buildLoginLogPredicates(in)
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

	items := make([]*agentpb.AgentLoginLogItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, toLoginLogItem(row))
	}
	return &agentpb.ListAgentLoginLogsResp{Items: items, Total: int64(total)}, nil
}

func buildLoginLogPredicates(in *agentpb.ListAgentLoginLogsReq) ([]predicate.AgentLoginLog, error) {
	preds := make([]predicate.AgentLoginLog, 0, 6)
	if in.TimeStart > 0 {
		preds = append(preds, agentloginlog.LoginAtGTE(time.Unix(in.TimeStart, 0)))
	}
	if in.TimeEnd > 0 {
		preds = append(preds, agentloginlog.LoginAtLTE(time.Unix(in.TimeEnd, 0)))
	}
	if in.LoginType > 0 {
		if in.LoginType != 1 && in.LoginType != 2 {
			return nil, xerr.BadRequest(i18n.InvalidParam)
		}
		preds = append(preds, agentloginlog.LoginTypeEQ(int64(in.LoginType)))
	}
	if in.DeviceType > 0 {
		if in.DeviceType < 1 || in.DeviceType > 5 {
			return nil, xerr.BadRequest(i18n.InvalidParam)
		}
		preds = append(preds, agentloginlog.DeviceTypeEQ(int64(in.DeviceType)))
	}

	if in.ParentId > 0 {
		preds = append(preds, agentloginlog.ParentIDEQ(in.ParentId))
	}

	keyword := strings.TrimSpace(in.Keyword)
	if keyword == "" || in.KeywordType == 0 {
		return preds, nil
	}
	switch in.KeywordType {
	case loginLogKeywordAgentID:
		id, err := strconv.ParseInt(keyword, 10, 64)
		if err != nil || id <= 0 {
			return nil, xerr.BadRequest(i18n.InvalidParam)
		}
		preds = append(preds, agentloginlog.AgentIDEQ(id))
	case loginLogKeywordLoginIP:
		preds = append(preds, agentloginlog.LoginIPEQ(keyword))
	case loginLogKeywordUdid:
		preds = append(preds, agentloginlog.UdidEQ(keyword))
	case loginLogKeywordLocation:
		preds = append(preds, agentloginlog.LoginLocationContainsFold(keyword))
	case loginLogKeywordAccount:
		preds = append(preds, agentloginlog.AccountContainsFold(keyword))
	default:
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	return preds, nil
}

func toLoginLogItem(row *ent.AgentLoginLog) *agentpb.AgentLoginLogItem {
	if row == nil {
		return &agentpb.AgentLoginLogItem{}
	}
	item := &agentpb.AgentLoginLogItem{
		Id:         row.ID,
		Account:    row.Account,
		LoginType:  int32(row.LoginType),
		DeviceType: int32(row.DeviceType),
		Result:     int32(row.Result),
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
