package agent

import (
	"context"
	"database/sql"
	"strings"
	"unicode/utf8"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/utils"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/ent/agentstats"
	"oa.98ent.com/p9/operator-crm/ent/agentuser"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/dao"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"
)

const (
	profileFieldAll        = iota // 全部
	profileFieldAvatar            // 头像
	profileFieldRealname          // 真实姓名
	profileFieldTelegram          // 电报
	profileFieldFacebook          // 脸书
	profileFieldWhatsapp          // 微信
	profileFieldMobile            // 手机号
	profileFieldNickname          // 昵称
	profileFieldIDDocument        // 身份证号
	profileFieldZalo              // 扎拉
	profileFieldTwitter           // 推特
	profileFieldLine              // 连信
	profileFieldEmail             // 邮箱

	payoutNormal          int64 = 1 // 不限制出款
	payoutOwnerRestricted int64 = 2 // 限制本人出款
	payoutLineRestricted  int64 = 3 // 限制全线出款
)

// 获取代理资料
func (s *Service) GetProfile(ctx context.Context, req *agentpb.GetAgentProfileReq) (*agentpb.AgentProfile, error) {
	if req == nil {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	row, err := s.loadAgent(ctx, req.GetAgentId())
	if err != nil {
		return nil, err
	}
	return s.toProfile(ctx, row)
}

// 重置登录密码 随机生成密码并返回
func (s *Service) ResetLoginPassword(ctx context.Context, req *agentpb.ResetLoginPasswordReq) (*agentpb.ResetLoginPasswordResp, error) {
	if req == nil {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	agentID := req.GetAgentId()
	// 操作日志参数
	opParams := &RecordOpParams{
		targetType: opTargetTypeAgent,
		targetID:   agentID,
		action:     opActionResetLoginPassword,
		after:      opPasswordMask,
		before:     opPasswordMask,
	}
	// 加载代理
	row, err := s.loadAgent(ctx, agentID)
	if err != nil {
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	// 操作日志参数 代理账号
	opParams.targetAccount = row.Account
	// 生成随机密码
	plain, err := utils.GenRandomPassword(16)
	if err != nil {
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	hashed, err := dao.HashPassword(plain)
	if err != nil {
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	if err = s.DB.AgentUser.UpdateOneID(row.ID).
		SetLoginPasswordHash(hashed).
		Exec(ctx); err != nil {
		err = mapEntErr(err)
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	s.RecordOpOK(ctx, opParams)
	return &agentpb.ResetLoginPasswordResp{Password: plain}, nil
}

// UpdateLoginPassword 当前代理修改登录密码。校验旧密码后只更新哈希，不改令牌盐。
func (s *Service) UpdateLoginPassword(ctx context.Context, req *agentpb.UpdateLoginPasswordReq) (*agentpb.Empty, error) {
	if req == nil {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	oldPw := strings.TrimSpace(req.GetOldPassword())
	newPw := strings.TrimSpace(req.GetNewPassword())
	if oldPw == "" || newPw == "" {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	c := ctxdata.ClaimsFromCtx(ctx)
	if c == nil || c.UserID <= 0 || strings.TrimSpace(c.OperatorCode) == "" || c.Salt == "" {
		return nil, xerr.Unauthorized(i18n.Unauthorized)
	}
	agentID := c.UserID
	// 操作日志参数
	opParams := &RecordOpParams{
		targetType: opTargetTypeAgent,
		targetID:   agentID,
		action:     opActionUpdateLoginPassword,
		after:      opPasswordMask,
		before:     opPasswordMask,
	}
	// 加载代理
	row, err := s.loadAgent(ctx, agentID)
	if err != nil {
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	// 操作日志参数 代理账号
	opParams.targetAccount = row.Account
	// 校验代理状态和盐值
	if row.Status != statusNormal || row.Salt != c.Salt {
		err = xerr.Unauthorized(i18n.Unauthorized)
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	if !dao.CheckPassword(row.LoginPasswordHash, oldPw) {
		err = xerr.Unauthorized(i18n.AuthPasswordIncorrect)
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	hashed, err := dao.HashPassword(newPw)
	if err != nil {
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	if err = s.DB.AgentUser.UpdateOneID(row.ID).
		SetLoginPasswordHash(hashed).
		Exec(ctx); err != nil {
		err = mapEntErr(err)
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	s.RecordOpOK(ctx, opParams)
	return &agentpb.Empty{}, nil
}

// 重置提现密码
func (s *Service) UpdateWithdrawPassword(ctx context.Context, req *agentpb.UpdateWithdrawPasswordReq) (*agentpb.Empty, error) {
	if req == nil || strings.TrimSpace(req.GetPassword()) == "" {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	agentID := req.GetAgentId()
	// 操作日志参数
	opParams := &RecordOpParams{
		targetType: opTargetTypeAgent,
		targetID:   agentID,
		action:     opActionUpdateWithdrawPassword,
		after:      opPasswordMask,
		before:     opPasswordMask,
	}
	// 加载代理
	row, err := s.loadAgent(ctx, agentID)
	if err != nil {
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	// 操作日志参数 代理账号
	opParams.targetAccount = row.Account
	// 生成提现密码哈希
	hashed, err := dao.HashPassword(strings.TrimSpace(req.GetPassword()))
	if err != nil {
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	if err = s.DB.AgentUser.UpdateOneID(row.ID).
		SetWithdrawPasswordHash(hashed).
		Exec(ctx); err != nil {
		err = mapEntErr(err)
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	s.RecordOpOK(ctx, opParams)
	return &agentpb.Empty{}, nil
}

// 更新代理占成比例
func (s *Service) UpdateShareRate(ctx context.Context, req *agentpb.UpdateAgentShareRateReq) (*agentpb.Empty, error) {
	if req == nil {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	agentID := req.GetAgentId()
	share := int64(req.GetShareRate())
	// 操作日志参数
	opParams := &RecordOpParams{
		targetType: opTargetTypeAgent,
		targetID:   agentID,
		action:     opActionShareRate,
		after:      formatInt64(share),
	}
	if share < 0 || share > maxShareRate {
		err := xerr.BadRequest(i18n.AgentShareRateInvalid)
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	row, err := s.loadAgent(ctx, agentID)
	if err != nil {
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	// 操作日志参数 更新前占成比例
	opParams.before = formatInt64(row.ShareRate)
	// 操作日志参数 代理账号
	opParams.targetAccount = row.Account
	if row.ParentID != nil {
		parent, err := s.DB.AgentUser.Query().
			Where(agentuser.IDEQ(*row.ParentID)).
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				err = xerr.BadRequest(i18n.AgentParentNotFound)
			}
			s.RecordOpFail(ctx, opParams, err)
			return nil, err
		}
		if share > parent.ShareRate {
			err = xerr.BadRequest(i18n.AgentShareRateExceedParent)
			s.RecordOpFail(ctx, opParams, err)
			return nil, err
		}
	}
	var childMax []struct {
		Max sql.NullInt64 `json:"max"`
	}
	if err = s.DB.AgentUser.Query().
		Where(agentuser.ParentIDEQ(row.ID)).
		Aggregate(ent.As(ent.Max(agentuser.FieldShareRate), "max")).
		Scan(ctx, &childMax); err != nil {
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	if len(childMax) > 0 && childMax[0].Max.Valid && share < childMax[0].Max.Int64 {
		err = xerr.BadRequest(i18n.AgentShareRateBelowChild)
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	if err = s.DB.AgentUser.UpdateOneID(row.ID).
		SetShareRate(share).
		Exec(ctx); err != nil {
		err = mapEntErr(err)
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	if _, op, oerr := requireOperator(ctx); oerr == nil {
		s.invalidateTree(op)
	}
	s.RecordOpOK(ctx, opParams)
	return &agentpb.Empty{}, nil
}

// 更新代理提现限制
func (s *Service) UpdatePayoutRestricted(ctx context.Context, req *agentpb.UpdateAgentPayoutRestrictedReq) (*agentpb.Empty, error) {
	if req == nil {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	agentID := req.GetAgentId()
	v := int64(req.GetPayoutRestricted())
	// 操作日志参数
	opParams := &RecordOpParams{
		targetType: opTargetTypeAgent,
		targetID:   agentID,
		action:     opActionPayoutRestricted,
		after:      formatInt64(v),
	}
	if v < payoutNormal || v > payoutLineRestricted {
		err := xerr.BadRequest(i18n.InvalidParam)
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	row, err := s.loadAgent(ctx, agentID)
	if err != nil {
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	// 操作日志参数 更新前提现限制
	opParams.before = formatInt64(row.PayoutRestricted)
	// 操作日志参数 代理账号
	opParams.targetAccount = row.Account
	if err = s.DB.AgentUser.UpdateOneID(row.ID).
		SetPayoutRestricted(v).
		Exec(ctx); err != nil {
		err = mapEntErr(err)
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	s.RecordOpOK(ctx, opParams)
	return &agentpb.Empty{}, nil
}

// UpdateStatus 更新代理账号状态：1正常 2冻结 3禁用。
func (s *Service) UpdateStatus(ctx context.Context, req *agentpb.UpdateAgentStatusReq) (*agentpb.Empty, error) {
	if req == nil {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	agentID := req.GetAgentId()
	v := int64(req.GetStatus())
	// 操作日志参数
	opParams := &RecordOpParams{
		targetType: opTargetTypeAgent,
		targetID:   agentID,
		action:     opActionStatus,
		after:      formatInt64(v),
	}
	if v < statusNormal || v > statusDisabled {
		err := xerr.BadRequest(i18n.InvalidParam)
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	row, err := s.loadAgent(ctx, agentID)
	if err != nil {
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	// 操作日志参数 更新前状态
	opParams.before = formatInt64(row.Status)
	// 操作日志参数 代理账号
	opParams.targetAccount = row.Account
	// 更新代理状态
	if err = s.DB.AgentUser.UpdateOneID(row.ID).
		SetStatus(v).
		Exec(ctx); err != nil {
		err = mapEntErr(err)
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	s.RecordOpOK(ctx, opParams)
	return &agentpb.Empty{}, nil
}

// 更新代理备注
func (s *Service) UpdateRemark(ctx context.Context, req *agentpb.UpdateAgentRemarkReq) (*agentpb.Empty, error) {
	if req == nil {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	agentID := req.GetAgentId()
	remark := strings.TrimSpace(req.GetRemark())
	// 操作日志参数
	opParams := &RecordOpParams{
		targetType: opTargetTypeAgent,
		targetID:   agentID,
		action:     opActionRemark,
		after:      remark,
	}
	if utf8.RuneCountInString(remark) > 1000 {
		err := xerr.BadRequest(i18n.InvalidParam)
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	row, err := s.loadAgent(ctx, agentID)
	if err != nil {
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	// 操作日志参数 更新前备注
	opParams.before = ptrStr(row.Remark)
	// 操作日志参数 代理账号
	opParams.targetAccount = row.Account
	// 更新代理备注
	upd := s.DB.AgentUser.UpdateOneID(row.ID)
	if remark == "" {
		upd.ClearRemark()
	} else {
		upd.SetRemark(remark)
	}
	if err = upd.Exec(ctx); err != nil {
		err = mapEntErr(err)
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	s.RecordOpOK(ctx, opParams)
	return &agentpb.Empty{}, nil
}

// 更新代理资料
func (s *Service) UpdateProfile(ctx context.Context, req *agentpb.UpdateAgentProfileReq) (*agentpb.AgentProfile, error) {
	if req == nil {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	agentID := req.GetAgentId()
	ft := req.GetFieldType()
	// 操作日志参数
	opParams := &RecordOpParams{
		targetType: opTargetTypeAgent,
		targetID:   agentID,
		action:     profileActionOf(ft),
	}
	if ft < profileFieldAll || ft > profileFieldEmail {
		err := xerr.BadRequest(i18n.InvalidParam)
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	if (ft == profileFieldMobile || ft == profileFieldAll) && req.GetMobile() != "" {
		if !utils.ValidPhone(req.GetMobile()) {
			err := xerr.BadRequest(i18n.InvalidPhone)
			if ft == profileFieldMobile {
				opParams.after = req.GetMobile()
			}
			s.RecordOpFail(ctx, opParams, err)
			return nil, err
		}
	}
	row, err := s.loadAgent(ctx, agentID)
	if err != nil {
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	// 操作日志参数 代理账号
	opParams.targetAccount = row.Account
	// 加载代理资料
	profile, err := row.QueryProfile().Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			err = xerr.NotFound(i18n.AgentNotFound)
		}
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	beforeVals := profileFieldValues(profile)
	upd := s.DB.AgentProfile.UpdateOneID(profile.ID)
	fields := profileUpdates(req)
	if ft == profileFieldAll {
		for _, item := range fields {
			if err = item.apply(upd); err != nil {
				s.RecordOpFail(ctx, opParams, err)
				return nil, err
			}
		}
	} else if err = fields[ft-1].apply(upd); err != nil {
		opParams.before = beforeVals[ft-1]
		opParams.after = strings.TrimSpace(fields[ft-1].value)
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	if err = upd.Exec(ctx); err != nil {
		err = mapEntErr(err)
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	row, err = s.loadAgent(ctx, row.ID)
	if err != nil {
		s.RecordOpFail(ctx, opParams, err)
		return nil, err
	}
	afterVals := profileFieldValues(row.Edges.Profile)
	if ft == profileFieldAll {
		opParams.before, opParams.after = profileJSONDiff(beforeVals, afterVals)
		s.RecordOpOK(ctx, opParams)
	} else {
		opParams.action = profileOpActions[ft-1]
		opParams.before = beforeVals[ft-1]
		opParams.after = afterVals[ft-1]
		s.RecordOpOK(ctx, opParams)
	}
	return s.toProfile(ctx, row)
}

type profileField struct {
	value string
	max   int
	set   func(*ent.AgentProfileUpdateOne, string) *ent.AgentProfileUpdateOne
	clear func(*ent.AgentProfileUpdateOne) *ent.AgentProfileUpdateOne
}

func (f profileField) apply(upd *ent.AgentProfileUpdateOne) error {
	v := strings.TrimSpace(f.value)
	if v == "" {
		if f.clear != nil {
			f.clear(upd)
		}
		return nil
	}
	if utf8.RuneCountInString(v) > f.max {
		return xerr.BadRequest(i18n.InvalidParam)
	}
	f.set(upd, v)
	return nil
}

func profileUpdates(req *agentpb.UpdateAgentProfileReq) []profileField {
	return []profileField{
		{req.GetAvatarUrl(), 255, (*ent.AgentProfileUpdateOne).SetAvatarURL, (*ent.AgentProfileUpdateOne).ClearAvatarURL},
		{req.GetRealname(), 100, (*ent.AgentProfileUpdateOne).SetRealname, (*ent.AgentProfileUpdateOne).ClearRealname},
		{req.GetTelegram(), 64, (*ent.AgentProfileUpdateOne).SetTelegram, (*ent.AgentProfileUpdateOne).ClearTelegram},
		{req.GetFacebook(), 128, (*ent.AgentProfileUpdateOne).SetFacebook, (*ent.AgentProfileUpdateOne).ClearFacebook},
		{req.GetWhatsapp(), 64, (*ent.AgentProfileUpdateOne).SetWhatsapp, (*ent.AgentProfileUpdateOne).ClearWhatsapp},
		{req.GetMobile(), 32, (*ent.AgentProfileUpdateOne).SetMobile, (*ent.AgentProfileUpdateOne).ClearMobile},
		{req.GetNickname(), 100, (*ent.AgentProfileUpdateOne).SetNickname, (*ent.AgentProfileUpdateOne).ClearNickname},
		{req.GetIdDocumentNo(), 64, (*ent.AgentProfileUpdateOne).SetIDDocumentNo, (*ent.AgentProfileUpdateOne).ClearIDDocumentNo},
		{req.GetZalo(), 64, (*ent.AgentProfileUpdateOne).SetZalo, (*ent.AgentProfileUpdateOne).ClearZalo},
		{req.GetTwitter(), 64, (*ent.AgentProfileUpdateOne).SetTwitter, (*ent.AgentProfileUpdateOne).ClearTwitter},
		{req.GetLine(), 64, (*ent.AgentProfileUpdateOne).SetLine, (*ent.AgentProfileUpdateOne).ClearLine},
		{req.GetEmail(), 255, (*ent.AgentProfileUpdateOne).SetEmail, (*ent.AgentProfileUpdateOne).ClearEmail},
	}
}

func (s *Service) loadAgent(ctx context.Context, id int64) (*ent.AgentUser, error) {
	if id <= 0 {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	ctx, op, err := requireOperator(ctx)
	if err != nil {
		return nil, err
	}
	row, err := s.DB.AgentUser.Query().
		Where(agentuser.IDEQ(id), agentuser.OperatorCodeEQ(op)).
		WithProfile().
		Only(ctx)
	if ent.IsNotFound(err) {
		return nil, xerr.NotFound(i18n.AgentNotFound)
	}
	return row, err
}

func (s *Service) toProfile(ctx context.Context, row *ent.AgentUser) (*agentpb.AgentProfile, error) {
	out := &agentpb.AgentProfile{
		AgentId:          row.ID,
		Status:           int32(row.Status),
		RegisterSource:   int32(row.RegisterSource),
		PromoCode:        row.PromoCode,
		ShareRate:        int32(row.ShareRate),
		LoginCount:       row.LoginCount,
		PayoutRestricted: int32(row.PayoutRestricted),
		AgentType:        int32(row.AgentType),
	}
	if row.Remark != nil {
		out.Remark = *row.Remark
	}
	if row.RegisterAt != nil {
		out.RegisterAt = row.RegisterAt.Unix()
	}
	if row.LastLoginAt != nil {
		out.LastLoginAt = row.LastLoginAt.Unix()
	}
	if row.LastLoginIP != nil {
		out.LastLoginIp = *row.LastLoginIP
	}
	if row.LastLoginUdid != nil {
		out.LastLoginUdid = *row.LastLoginUdid
	}
	if row.ParentID != nil {
		parent, err := s.DB.AgentUser.Query().
			Where(agentuser.IDEQ(*row.ParentID)).
			Only(ctx)
		if err != nil && !ent.IsNotFound(err) {
			return nil, err
		}
		if parent != nil {
			out.ParentAccount = parent.Account
		}
	}
	stats, err := s.DB.AgentStats.Query().Where(agentstats.AgentIDEQ(row.ID)).Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, err
	}
	if stats != nil {
		out.InviteMemberCount = stats.DirectMemberCount
	}
	profile := row.Edges.Profile
	if profile == nil {
		return out, nil
	}
	assign := func(dst *string, src *string) {
		if src != nil {
			*dst = *src
		}
	}
	assign(&out.AvatarUrl, profile.AvatarURL)
	assign(&out.Realname, profile.Realname)
	assign(&out.Telegram, profile.Telegram)
	assign(&out.Facebook, profile.Facebook)
	assign(&out.Whatsapp, profile.Whatsapp)
	assign(&out.Mobile, profile.Mobile)
	assign(&out.Nickname, profile.Nickname)
	assign(&out.IdDocumentNo, profile.IDDocumentNo)
	assign(&out.Zalo, profile.Zalo)
	assign(&out.Twitter, profile.Twitter)
	assign(&out.Line, profile.Line)
	assign(&out.Email, profile.Email)
	return out, nil
}
