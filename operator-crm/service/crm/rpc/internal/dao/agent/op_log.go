package agent

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/ent/agentoplog"
	"oa.98ent.com/p9/operator-crm/ent/predicate"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"entgo.io/ent/dialect/sql"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	opResultSuccess int64 = 1
	opResultFail    int64 = 2

	opActionShareRate              = "share_rate"
	opActionPayoutRestricted       = "payout_restricted"
	opActionStatus                 = "status"
	opActionRemark                 = "remark"
	opActionResetLoginPassword     = "reset_login_password"
	opActionUpdateLoginPassword    = "update_login_password"
	opActionUpdateWithdrawPassword = "update_withdraw_password"
	opActionProfile                = "profile"
	opActionProfileAvatar          = "profile.avatar"
	opActionProfileRealname        = "profile.realname"
	opActionProfileTelegram        = "profile.telegram"
	opActionProfileFacebook        = "profile.facebook"
	opActionProfileWhatsapp        = "profile.whatsapp"
	opActionProfileMobile          = "profile.mobile"
	opActionProfileNickname        = "profile.nickname"
	opActionProfileIDDocument      = "profile.id_document"
	opActionProfileZalo            = "profile.zalo"
	opActionProfileTwitter         = "profile.twitter"
	opActionProfileLine            = "profile.line"
	opActionProfileEmail           = "profile.email"

	opPasswordMask  = "***"
	opFailReasonMax = 512
	opValueMax      = 2000
)

var profileOpActions = []string{
	opActionProfileAvatar,
	opActionProfileRealname,
	opActionProfileTelegram,
	opActionProfileFacebook,
	opActionProfileWhatsapp,
	opActionProfileMobile,
	opActionProfileNickname,
	opActionProfileIDDocument,
	opActionProfileZalo,
	opActionProfileTwitter,
	opActionProfileLine,
	opActionProfileEmail,
}

// ListAgentOpLogs 分页查询代理资料变更操作日志。
func (s *Service) ListAgentOpLogs(ctx context.Context, in *agentpb.ListAgentOpLogsReq) (*agentpb.ListAgentOpLogsResp, error) {
	ctx, _, err := requireOperator(ctx)
	if err != nil {
		return nil, err
	}
	if in == nil || in.GetAgentId() <= 0 {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}

	page := max(in.Page, 1)
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	preds := []predicate.AgentOpLog{agentoplog.AgentIDEQ(in.GetAgentId())}
	if in.TimeStart > 0 {
		preds = append(preds, agentoplog.OperatedAtGTE(time.Unix(in.TimeStart, 0)))
	}
	if in.TimeEnd > 0 {
		preds = append(preds, agentoplog.OperatedAtLTE(time.Unix(in.TimeEnd, 0)))
	}
	if in.Result == 1 || in.Result == 2 {
		preds = append(preds, agentoplog.ResultEQ(int64(in.Result)))
	}

	q := s.DB.AgentOpLog.Query().Where(preds...)
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := q.
		Order(agentoplog.ByOperatedAt(sql.OrderDesc()), agentoplog.ByID(sql.OrderDesc())).
		Offset(int((page - 1) * pageSize)).
		Limit(int(pageSize)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]*agentpb.AgentOpLogItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, toOpLogItem(row))
	}
	return &agentpb.ListAgentOpLogsResp{Items: items, Total: int64(total)}, nil
}

func toOpLogItem(row *ent.AgentOpLog) *agentpb.AgentOpLogItem {
	item := &agentpb.AgentOpLogItem{
		Id:              row.ID,
		AgentId:         row.AgentID,
		OperatorAccount: row.OperatorAccount,
		Action:          row.Action,
		OperatedAt:      row.OperatedAt.Unix(),
		Result:          int32(row.Result),
	}
	if row.OperatorID != nil {
		item.OperatorId = *row.OperatorID
	}
	if row.BeforeValue != nil {
		item.BeforeValue = *row.BeforeValue
	}
	if row.AfterValue != nil {
		item.AfterValue = *row.AfterValue
	}
	if row.ClientIP != nil {
		item.ClientIp = *row.ClientIP
	}
	if row.LoginDevice != nil {
		item.LoginDevice = *row.LoginDevice
	}
	if row.Udid != nil {
		item.Udid = *row.Udid
	}
	if row.FailReason != nil {
		item.FailReason = *row.FailReason
	}
	return item
}

func (s *Service) recordOpOK(ctx context.Context, agentID int64, action, before, after string) {
	s.recordOp(ctx, agentID, action, before, after, opResultSuccess, "")
}

func (s *Service) recordOpFail(ctx context.Context, agentID int64, action, before, after string, err error) {
	if agentID <= 0 || err == nil {
		return
	}
	s.recordOp(ctx, agentID, action, before, after, opResultFail, failReasonOf(err))
}

func (s *Service) recordOp(ctx context.Context, agentID int64, action, before, after string, result int64, failReason string) {
	if s == nil || s.DB == nil || agentID <= 0 || action == "" {
		return
	}
	create := s.DB.AgentOpLog.Create().
		SetAgentID(agentID).
		SetAction(action).
		SetOperatedAt(time.Now()).
		SetResult(result).
		SetOperatorAccount("")

	if c := ctxdata.ClaimsFromCtx(ctx); c != nil {
		if c.UserID > 0 {
			create.SetOperatorID(c.UserID)
		}
		create.SetOperatorAccount(clipRunes(c.Username, 64))
	}
	if v := clipRunes(before, opValueMax); v != "" {
		create.SetBeforeValue(v)
	}
	if v := clipRunes(after, opValueMax); v != "" {
		create.SetAfterValue(v)
	}
	if ip := strings.TrimSpace(ctxdata.ClientIPFromCtx(ctx)); ip != "" {
		create.SetClientIP(ip)
	}
	if ua := clipRunes(ctxdata.UserAgentFromCtx(ctx), 512); ua != "" {
		create.SetLoginDevice(ua)
	}
	if udid := clipRunes(ctxdata.UdidFromCtx(ctx), 128); udid != "" {
		create.SetUdid(udid)
	}
	if result == opResultFail {
		if fr := clipRunes(failReason, opFailReasonMax); fr != "" {
			create.SetFailReason(fr)
		}
	}

	if _, err := create.Save(ctx); err != nil {
		logx.WithContext(ctx).Errorf("agent op log save action=%s agent_id=%d: %v", action, agentID, err)
	}
}

func failReasonOf(err error) string {
	if err == nil {
		return ""
	}
	var e *xerr.Error
	if errors.As(err, &e) && e != nil && e.Message != "" {
		return e.Message
	}
	return err.Error()
}

func clipRunes(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	if maxLen <= 0 || s == "" {
		return s
	}
	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}
	return string([]rune(s)[:maxLen])
}

func formatInt64(v int64) string {
	return strconv.FormatInt(v, 10)
}

func ptrStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func profileFieldValues(p *ent.AgentProfile) []string {
	if p == nil {
		return make([]string, len(profileOpActions))
	}
	return []string{
		ptrStr(p.AvatarURL),
		ptrStr(p.Realname),
		ptrStr(p.Telegram),
		ptrStr(p.Facebook),
		ptrStr(p.Whatsapp),
		ptrStr(p.Mobile),
		ptrStr(p.Nickname),
		ptrStr(p.IDDocumentNo),
		ptrStr(p.Zalo),
		ptrStr(p.Twitter),
		ptrStr(p.Line),
		ptrStr(p.Email),
	}
}

func profileJSONDiff(before, after []string) (string, string) {
	bMap := map[string]string{}
	aMap := map[string]string{}
	for i, action := range profileOpActions {
		bv, av := "", ""
		if i < len(before) {
			bv = before[i]
		}
		if i < len(after) {
			av = after[i]
		}
		if bv == av {
			continue
		}
		key := strings.TrimPrefix(action, "profile.")
		bMap[key] = bv
		aMap[key] = av
	}
	if len(bMap) == 0 {
		return "", ""
	}
	bb, _ := json.Marshal(bMap)
	aa, _ := json.Marshal(aMap)
	return string(bb), string(aa)
}

func profileActionOf(ft int32) string {
	if ft == profileFieldAll {
		return opActionProfile
	}
	if ft >= profileFieldAvatar && ft <= profileFieldEmail {
		return profileOpActions[ft-1]
	}
	return opActionProfile
}
