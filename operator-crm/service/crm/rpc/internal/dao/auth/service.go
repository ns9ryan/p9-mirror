package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net"
	"strings"
	"time"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/utils"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/ent/agentuser"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/dao"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/authpb"

	"github.com/redis/go-redis/v9"
)

const (
	statusNormal    int64 = 1
	blacklistPrefix       = "crm:agent:jwt:bl:"
	saltByteLen           = 8
	udidMaxLen            = 128
)

// Config 是签发令牌使用的密钥和有效期。
type Config struct {
	AccessSecret  string
	AccessExpire  int64
	RefreshSecret string
	RefreshExpire int64
}

// Service 处理代理登录、刷新、退出和令牌校验。
type Service struct {
	DB     *ent.Client
	Redis  redis.UniversalClient
	Config Config
}

// New 创建鉴权服务。
func New(db *ent.Client, rdb redis.UniversalClient, c Config) *Service {
	return &Service{DB: db, Redis: rdb, Config: c}
}

// Login 校验代理账号并签发令牌对。
func (s *Service) Login(ctx context.Context, in *authpb.LoginReq) (*authpb.LoginResp, error) {
	if in == nil {
		in = &authpb.LoginReq{}
	}
	account := strings.TrimSpace(in.Account)
	password := in.Password
	operatorCode := strings.TrimSpace(in.OperatorCode)
	if account == "" || password == "" {
		return nil, xerr.BadRequest(i18n.AuthUsernamePasswordRequired)
	}
	if operatorCode == "" {
		return nil, xerr.BadRequest(i18n.AuthOperatorCodeRequired)
	}

	ctx = withOperator(ctx, operatorCode)
	if ctxdata.ClientIPFromCtx(ctx) == "" && strings.TrimSpace(in.ClientIp) != "" {
		ctx = ctxdata.WithClientIP(ctx, in.ClientIp)
	}

	u, err := s.DB.AgentUser.Query().
		Where(agentuser.AccountEQ(account)).
		WithProfile().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, xerr.Unauthorized(i18n.AuthPasswordIncorrect)
		}
		return nil, err
	}
	if !dao.CheckPassword(u.LoginPasswordHash, password) {
		return nil, xerr.Unauthorized(i18n.AuthPasswordIncorrect)
	}
	if u.Status != statusNormal {
		return nil, xerr.Forbidden(i18n.AuthUserDisabled)
	}

	if err := s.touchLogin(ctx, u, in.Udid); err != nil {
		return nil, err
	}
	token, err := s.signPair(ctx, u)
	if err != nil {
		return nil, err
	}
	return &authpb.LoginResp{Token: token, User: toPublic(u)}, nil
}

// Refresh 用刷新令牌换发新的令牌对，并拉黑旧刷新令牌。
func (s *Service) Refresh(ctx context.Context, in *authpb.RefreshReq) (*authpb.LoginResp, error) {
	if in == nil {
		in = &authpb.RefreshReq{}
	}
	raw := StripBearer(in.RefreshToken)
	c, err := ParseTyped(s.Config.RefreshSecret, raw, tokenRefresh)
	if err != nil {
		return nil, xerr.Unauthorized(i18n.Unauthorized)
	}
	if err := s.ensureActive(ctx, raw); err != nil {
		return nil, err
	}
	if err := checkClientIP(ctx, c); err != nil {
		return nil, err
	}
	u, err := s.sessionUser(ctx, c)
	if err != nil {
		return nil, err
	}
	exp := expiresUnix(c)
	if err := s.blacklist(ctx, raw, exp); err != nil {
		return nil, err
	}
	token, err := s.signPair(ctx, u)
	if err != nil {
		return nil, err
	}
	return &authpb.LoginResp{Token: token, User: toPublic(u)}, nil
}

// Logout 拉黑当前访问令牌，以及同一代理、同一盐的刷新令牌。
func (s *Service) Logout(ctx context.Context, in *authpb.LogoutReq) error {
	if in == nil {
		in = &authpb.LogoutReq{}
	}
	claims := ctxdata.ClaimsFromCtx(ctx)
	if claims == nil {
		return xerr.Unauthorized(i18n.Unauthorized)
	}
	rawAccess := ctxdata.RawTokenFromCtx(ctx)
	if rawAccess == "" {
		return xerr.Unauthorized(i18n.Unauthorized)
	}
	if err := s.blacklist(ctx, rawAccess, claims.ExpiresAt); err != nil {
		return err
	}
	rawRefresh, exp, err := parseOwnRefresh(s.Config.RefreshSecret, claims, in.RefreshToken)
	if err != nil {
		return err
	}
	return s.blacklist(ctx, rawRefresh, exp)
}

// LogoutAll 轮换令牌盐，使该代理已签发的令牌全部失效。
func (s *Service) LogoutAll(ctx context.Context) error {
	claims := ctxdata.ClaimsFromCtx(ctx)
	if claims == nil || claims.UserID == 0 {
		return xerr.Unauthorized(i18n.Unauthorized)
	}
	return s.rotateSalt(ctx, claims.UserID)
}

// CheckToken 校验访问令牌，并返回当前代理身份。
func (s *Service) CheckToken(ctx context.Context, raw string) (*authpb.CheckTokenResp, error) {
	raw = StripBearer(raw)
	c, err := Parse(s.Config.AccessSecret, raw)
	if err != nil {
		if IsExpired(err) {
			return nil, xerr.TokenExpired(i18n.TokenExpired)
		}
		return nil, xerr.Unauthorized(i18n.Unauthorized)
	}
	if c.TokenType == tokenRefresh {
		return nil, xerr.Forbidden(i18n.AuthInvalidToken)
	}
	if err := checkClientIP(ctx, c); err != nil {
		return nil, err
	}
	if err := s.ensureActive(ctx, raw); err != nil {
		return nil, err
	}
	u, err := s.agentByID(withOperator(ctx, c.OperatorCode), c.UserID, false)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, xerr.Forbidden(i18n.AuthUserNotFound)
		}
		return nil, err
	}
	if u.Status != statusNormal {
		return nil, xerr.Forbidden(i18n.AuthUserDisabled)
	}
	if u.Salt != c.Salt {
		return nil, xerr.Forbidden(i18n.AuthUserSaltMismatch)
	}
	if strPtr(u.OperatorCode) == "" || strPtr(u.OperatorCode) != c.OperatorCode {
		return nil, xerr.Unauthorized(i18n.Unauthorized)
	}
	return &authpb.CheckTokenResp{
		UserId:       u.ID,
		Account:      u.Account,
		OperatorCode: strPtr(u.OperatorCode),
		Salt:         u.Salt,
		ExpiresAt:    expiresUnix(c),
		TokenType:    c.TokenType,
	}, nil
}

func (s *Service) sessionUser(ctx context.Context, c *Claims) (*ent.AgentUser, error) {
	u, err := s.agentByID(withOperator(ctx, c.OperatorCode), c.UserID, true)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, xerr.Unauthorized(i18n.Unauthorized)
		}
		return nil, err
	}
	if u.Status != statusNormal || u.Salt != c.Salt {
		return nil, xerr.Unauthorized(i18n.Unauthorized)
	}
	if strPtr(u.OperatorCode) == "" || strPtr(u.OperatorCode) != c.OperatorCode {
		return nil, xerr.Unauthorized(i18n.Unauthorized)
	}
	return u, nil
}

func (s *Service) agentByID(ctx context.Context, id int64, withProfile bool) (*ent.AgentUser, error) {
	q := s.DB.AgentUser.Query().Where(agentuser.ID(id))
	if withProfile {
		q = q.WithProfile()
	}
	return q.Only(ctx)
}

// 更新代理登录信息
func (s *Service) touchLogin(ctx context.Context, u *ent.AgentUser, udid string) error {
	now := time.Now()
	upd := s.DB.AgentUser.UpdateOneID(u.ID).
		SetLastLoginAt(now).
		AddLoginCount(1)
	if ip := utils.NormalizeIP(ctxdata.ClientIPFromCtx(ctx)); net.ParseIP(ip) != nil {
		upd.SetLastLoginIP(ip)
		u.LastLoginIP = &ip
	}
	if udid = clip(strings.TrimSpace(udid), udidMaxLen); udid != "" {
		upd.SetLastLoginUdid(udid)
		u.LastLoginUdid = &udid
	}
	if err := upd.Exec(ctx); err != nil {
		return err
	}
	u.LastLoginAt = &now
	u.LoginCount++
	return nil
}

func (s *Service) signPair(ctx context.Context, u *ent.AgentUser) (*authpb.TokenInfo, error) {
	base := tokenClaims(ctx, u)
	access, accessExp, err := Sign(s.Config.AccessSecret, s.Config.AccessExpire, base.WithType(tokenAccess))
	if err != nil {
		return nil, err
	}
	refresh, refreshExp, err := Sign(s.Config.RefreshSecret, s.Config.RefreshExpire, base.WithType(tokenRefresh))
	if err != nil {
		return nil, err
	}
	return &authpb.TokenInfo{
		AccessToken:   access,
		RefreshToken:  refresh,
		Expire:        accessExp,
		RefreshExpire: refreshExp,
	}, nil
}

func (s *Service) rotateSalt(ctx context.Context, userID int64) error {
	salt, err := randomSalt()
	if err != nil {
		return err
	}
	n, err := s.DB.AgentUser.Update().
		Where(agentuser.ID(userID)).
		SetSalt(salt).
		Save(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		return xerr.NotFound(i18n.AuthUserNotFound)
	}
	return nil
}

func (s *Service) blacklist(ctx context.Context, token string, expUnix int64) error {
	if token == "" {
		return xerr.Unauthorized(i18n.Unauthorized)
	}
	if s.Redis == nil {
		return xerr.InternalServerError(i18n.InternalError)
	}
	ttl := time.Until(time.Unix(expUnix, 0))
	if ttl <= 0 {
		ttl = time.Minute
	}
	if err := s.Redis.Set(ctx, blacklistPrefix+token, "1", ttl).Err(); err != nil {
		return err
	}
	return nil
}

// 确保令牌未被拉黑
func (s *Service) ensureActive(ctx context.Context, token string) error {
	if token == "" {
		return xerr.Unauthorized(i18n.Unauthorized)
	}
	if s.Redis == nil {
		return xerr.InternalServerError(i18n.InternalError)
	}
	n, err := s.Redis.Exists(ctx, blacklistPrefix+token).Result()
	if err != nil {
		return err
	}
	if n > 0 {
		return xerr.Unauthorized(i18n.Unauthorized)
	}
	return nil
}

func tokenClaims(ctx context.Context, u *ent.AgentUser) Claims {
	return Claims{
		UserID:       u.ID,
		Username:     u.Account,
		OperatorCode: strPtr(u.OperatorCode),
		Salt:         u.Salt,
		ClientIP:     utils.NormalizeIP(ctxdata.ClientIPFromCtx(ctx)),
	}
}

// 解析自己的刷新令牌
func parseOwnRefresh(secret string, claims *ctxdata.Claims, refreshToken string) (string, int64, error) {
	raw := StripBearer(refreshToken)
	if raw == "" {
		return "", 0, xerr.BadRequest(i18n.AuthRefreshTokenRequired)
	}
	c, err := ParseTyped(secret, raw, tokenRefresh)
	if err != nil || c.UserID != claims.UserID || c.Salt != claims.Salt {
		return "", 0, xerr.Unauthorized(i18n.Unauthorized)
	}
	return raw, expiresUnix(c), nil
}

func checkClientIP(ctx context.Context, c *Claims) error {
	got := utils.NormalizeIP(c.ClientIP)
	want := utils.NormalizeIP(ctxdata.ClientIPFromCtx(ctx))
	if got == "" || got != want {
		return xerr.Unauthorized(i18n.AuthIPMismatch)
	}
	return nil
}

func toPublic(u *ent.AgentUser) *authpb.UserPublic {
	out := &authpb.UserPublic{
		Id:           u.ID,
		Account:      u.Account,
		Status:       int32(u.Status),
		OperatorCode: strPtr(u.OperatorCode),
		AgentType:    int32(u.AgentType),
		Level:        int32(u.Level),
		CreatedAt:    u.CreatedAt.Unix(),
	}
	if u.LastLoginAt != nil {
		out.LastLoginAt = u.LastLoginAt.Unix()
	}
	if p := u.Edges.Profile; p != nil {
		out.Nickname = strPtr(p.Nickname)
		out.Realname = strPtr(p.Realname)
		out.Mobile = strPtr(p.Mobile)
		out.Email = strPtr(p.Email)
	}
	return out
}

func withOperator(ctx context.Context, code string) context.Context {
	return ctxdata.WithOperatorCode(ctx, code)
}

func randomSalt() (string, error) {
	b := make([]byte, saltByteLen)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func expiresUnix(c *Claims) int64 {
	if c == nil || c.ExpiresAt == nil {
		return 0
	}
	return c.ExpiresAt.Unix()
}

func strPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func clip(s string, n int) string {
	if n <= 0 || len(s) <= n {
		return s
	}
	return s[:n]
}
