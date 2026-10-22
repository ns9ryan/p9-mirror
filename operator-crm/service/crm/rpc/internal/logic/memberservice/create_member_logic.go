package memberservicelogic

import (
	"context"
	"net/url"
	"strings"
	"unicode"

	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/dao"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"
)

type CreateMemberLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateMemberLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateMemberLogic {
	return &CreateMemberLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// CreateMember 校验后台输入并创建会员。
func (l *CreateMemberLogic) CreateMember(in *memberpb.CreateMemberRequest) (*memberpb.CreateMemberResponse, error) {
	if in == nil {
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
	}

	account := strings.TrimSpace(in.GetAccount())
	password := in.GetPassword()
	nickname := strings.TrimSpace(in.GetNickname())
	remark := strings.TrimSpace(in.GetRemark())
	avatarURL := strings.TrimSpace(in.GetAvatarUrl())
	frameURL := strings.TrimSpace(in.GetAvatarFrameUrl())

	if !validAccount(account) || !validPassword(password) || in.GetAgentId() <= 0 || len([]rune(remark)) > 200 {
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
	}
	if nickname == "" {
		nickname = account
	}
	if len([]rune(nickname)) > 100 {
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
	}

	if avatarURL == "" {
		avatarURL = defaultAvatars[0].URL
	}
	if frameURL == "" {
		frameURL = defaultAvatarFrames[0].URL
	}
	if !isPresetURL(avatarURL, defaultAvatars) || !isPresetURL(frameURL, defaultAvatarFrames) {
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
	}

	ctx, _, err := dao.OperatorContext(l.ctx)
	if err != nil {
		return nil, xerr.RpcErr(xerr.Unauthorized(i18n.Unauthorized))
	}
	hash, err := dao.HashPassword(password)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}

	created, err := l.svcCtx.DAO.Member.Create(ctx, dao.CreateInput{
		Account:        account,
		PasswordHash:   hash,
		Nickname:       nickname,
		AgentID:        in.GetAgentId(),
		AvatarURL:      avatarURL,
		AvatarFrameURL: frameURL,
		Remark:         remark,
	})
	if err != nil {
		switch {
		case ent.IsNotFound(err):
			return nil, xerr.RpcErr(xerr.BadRequest(i18n.MemberRelationNotFound))
		case ent.IsConstraintError(err):
			return nil, xerr.RpcErr(xerr.EntConstraintError(i18n.MemberConflict, err))
		case ent.IsValidationError(err):
			return nil, xerr.RpcErr(xerr.EntValidationError(i18n.InvalidParam, err))
		default:
			return nil, xerr.RpcErr(err)
		}
	}

	return &memberpb.CreateMemberResponse{
		Id:        created.ID,
		Code:      created.Code,
		CreatedAt: created.CreatedAt.UnixMilli(),
	}, nil
}

func validAccount(account string) bool {
	if account == "" || len([]rune(account)) > 64 {
		return false
	}
	for _, r := range account {
		if unicode.IsSpace(r) || unicode.IsControl(r) {
			return false
		}
	}
	return true
}

func validPassword(password string) bool {
	if len(password) < 8 || len(password) > 64 {
		return false
	}
	var letter, digit bool
	for _, r := range password {
		letter = letter || unicode.IsLetter(r)
		digit = digit || unicode.IsDigit(r)
	}
	return letter && digit
}

func isPresetURL(value string, options []ImageOption) bool {
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return false
	}
	for _, option := range options {
		if option.URL == value {
			return true
		}
	}
	return false
}
