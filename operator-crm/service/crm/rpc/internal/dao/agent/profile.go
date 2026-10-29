package agent

import (
	"context"
	"strings"
	"unicode/utf8"

	"oa.98ent.com/p9/common/utils"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/ent/agentstats"
	"oa.98ent.com/p9/operator-crm/ent/agentuser"
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
)

// 获取代理资料
func (s *Service) GetProfile(ctx context.Context, req *agentpb.GetAgentProfileReq) (*agentpb.AgentProfile, error) {
	if req == nil {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	row, err := s.LoadAgent(ctx, req.GetAgentId())
	if err != nil {
		return nil, err
	}
	return s.toProfile(ctx, row)
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
	row, err := s.LoadAgent(ctx, agentID)
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
	row, err = s.LoadAgent(ctx, row.ID)
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
		// 获取上级代理账号
		parent, err := s.FecthAgentByID(ctx, *row.ParentID, agentuser.FieldAccount)
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
