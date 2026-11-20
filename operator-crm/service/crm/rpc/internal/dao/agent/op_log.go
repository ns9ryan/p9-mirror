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
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"entgo.io/ent/dialect/sql"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	opResultSuccess int32 = 1 // 成功
	opResultFail    int32 = 2 // 失败

	adminTypeOperator int32 = 1 // 管理员
	adminTypeAgent    int32 = 2 // 代理

	opTargetTypeAgent int32 = 1 // 代理
	opTargetTypeUser  int32 = 2 // 玩家

	opActionCreate                 = "agent.opAction.create"                   // 创建
	opActionDelete                 = "agent.opAction.delete"                   // 删除
	opActionShareRate              = "agent.opAction.share_rate"               // 分享率
	opActionPayoutRestricted       = "agent.opAction.payout_restricted"        // 提现限制
	opActionStatus                 = "agent.opAction.status"                   // 状态
	opActionRemark                 = "agent.opAction.remark"                   // 备注
	opActionResetLoginPassword     = "agent.opAction.reset_login_password"     // 重置登录密码
	opActionUpdateLoginPassword    = "agent.opAction.update_login_password"    // 更新登录密码
	opActionUpdateWithdrawPassword = "agent.opAction.update_withdraw_password" // 更新提现密码
	opActionProfile                = "agent.opAction.profile"                  // 资料
	opActionProfileAvatar          = "agent.opAction.profile.avatar"           // 头像
	opActionProfileRealname        = "agent.opAction.profile.realname"         // 真实姓名
	opActionProfileTelegram        = "agent.opAction.profile.telegram"         // Telegram
	opActionProfileFacebook        = "agent.opAction.profile.facebook"         // Facebook
	opActionProfileWhatsapp        = "agent.opAction.profile.whatsapp"         // WhatsApp
	opActionProfileMobile          = "agent.opAction.profile.mobile"           // 手机号
	opActionProfileNickname        = "agent.opAction.profile.nickname"         // 昵称
	opActionProfileIDDocument      = "agent.opAction.profile.id_document"      // 身份证
	opActionProfileZalo            = "agent.opAction.profile.zalo"             // Zalo
	opActionProfileTwitter         = "agent.opAction.profile.twitter"          // Twitter
	opActionProfileLine            = "agent.opAction.profile.line"             // Line
	opActionProfileEmail           = "agent.opAction.profile.email"            // 邮箱

	opPasswordMask  = "***" // 密码掩码
	opFailReasonMax = 512   // 失败原因最大长度
	opValueMax      = 2000  // 值最大长度
)

var profileOpActions = []string{ // 资料操作动作列表
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

	page := max(in.Page, 1)
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	// preds 查询条件
	preds := []predicate.AgentOpLog{}
	if in.GetOpTargetId() > 0 {
		preds = append(preds, agentoplog.OpTargetIDEQ(in.GetOpTargetId()))
	}
	if in.GetOpTargetType() > 0 {
		preds = append(preds, agentoplog.OpTargetTypeEQ(in.GetOpTargetType()))
	}
	if in.TimeStart > 0 {
		preds = append(preds, agentoplog.OperatedAtGTE(time.Unix(in.TimeStart, 0)))
	}
	if in.TimeEnd > 0 {
		preds = append(preds, agentoplog.OperatedAtLTE(time.Unix(in.TimeEnd, 0)))
	}
	if in.Result == opResultSuccess || in.Result == opResultFail {
		preds = append(preds, agentoplog.ResultEQ(in.Result))
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
	return &agentpb.ListAgentOpLogsResp{List: items, Total: int64(total)}, nil
}

func toOpLogItem(row *ent.AgentOpLog) *agentpb.AgentOpLogItem {
	item := &agentpb.AgentOpLogItem{
		Id:              row.ID,
		OpTargetId:      row.OpTargetID,
		OpTargetType:    row.OpTargetType,
		OpTargetAccount: row.OpTargetAccount,
		AdminId:         row.AdminID,
		AdminType:       row.AdminType,
		AdminAccount:    row.AdminAccount,
		Action:          row.Action,
		OperatedAt:      row.OperatedAt.Unix(),
		Result:          int32(row.Result),
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

// RecordOpParams 记录操作参数。
type RecordOpParams struct {
	targetType    int32
	targetID      int64
	targetAccount string
	action        string
	before        string
	after         string
}

// recordOpOK 记录操作成功。
func (s *Service) RecordOpOK(ctx context.Context, p *RecordOpParams) {
	s.recordOp(ctx, p, opResultSuccess, "")
}

// recordOpFail 记录操作失败。
func (s *Service) RecordOpFail(ctx context.Context, p *RecordOpParams, err error) {
	if p == nil || err == nil {
		return
	}
	s.recordOp(ctx, p, opResultFail, failReasonOf(err))
}

// recordOp 记录操作。
func (s *Service) recordOp(ctx context.Context, p *RecordOpParams, result int32, failReason string) {
	if s == nil || s.DB == nil || p == nil || p.action == "" {
		return
	}
	create := s.DB.AgentOpLog.Create().
		SetOpTargetType(p.targetType).
		SetOpTargetID(p.targetID).
		SetAction(p.action).
		SetOperatedAt(time.Now()).
		SetResult(result).
		SetOpTargetAccount(p.targetAccount)

	if c := ctxdata.ClaimsFromCtx(ctx); c != nil {
		if c.UserID > 0 {
			create.SetAdminID(c.UserID)
		}
		// 操作人类型
		if c.UserType.IsAgent() {
			create.SetAdminType(adminTypeAgent)
		} else {
			create.SetAdminType(adminTypeOperator)
		}
		create.SetAdminAccount(clipRunes(c.Username, 64))
	}
	if v := clipRunes(p.before, opValueMax); v != "" {
		create.SetBeforeValue(v)
	}
	if v := clipRunes(p.after, opValueMax); v != "" {
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
		logx.WithContext(ctx).Errorf("agent op log save action=%s target_type=%d target_id=%d: %v", p.action, p.targetType, p.targetID, err)
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
