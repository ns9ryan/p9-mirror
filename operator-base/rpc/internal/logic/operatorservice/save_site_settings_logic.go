package operatorservicelogic

import (
	"context"
	"database/sql"
	"net/url"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/operator-base/rpc/ent/operatorsitesetting"
	operatorservice "oa.98ent.com/p9/operator-base/rpc/internal/operator"
	"oa.98ent.com/p9/operator-base/rpc/internal/svc"
	"oa.98ent.com/p9/operator-base/rpc/pb/operatorbaserpc/operatorpb"
)

type SaveSiteSettingsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveSiteSettingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveSiteSettingsLogic {
	return &SaveSiteSettingsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SaveSiteSettings 完整保存八项配置，当前不参与登录、注册或验证码校验。
func (l *SaveSiteSettingsLogic) SaveSiteSettings(in *operatorpb.SaveSiteSettingsRequest) (
	*operatorpb.SiteSettingsModel, error,
) {
	// 登录身份只读取并校验一次，事务内归属核对和修改人记录复用该身份。
	claims := ctxdata.ClaimsFromCtx(l.ctx)
	if claims == nil || claims.UserID <= 0 || strings.TrimSpace(claims.OperatorCode) == "" {
		return nil, status.Error(codes.PermissionDenied, "login user and operator are required")
	}
	title, logo, err := validateSiteSettings(in)
	if err != nil {
		return nil, err
	}
	// 事务前检查数据库配置，站点归属在事务内按认证上下文查询。
	if l.svcCtx.DB == nil {
		return nil, status.Error(codes.FailedPrecondition, "operator database is not configured")
	}
	// 前端仍提交完整 URL；只将受信任公开前缀下的本站图片转换为 Key 保存。
	logo, _, err = operatorservice.LogoValue(logo, claims.OperatorCode, l.svcCtx.Config)
	if err != nil {
		return nil, err
	}
	// 开启可串行化事务，后续归属复核、配置写入和结果查询统一使用该事务。
	tx, err := l.svcCtx.DB.BeginTx(l.ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, siteSettingsError(l.ctx, l.Logger, err)
	}
	// 提交前任何一步出错返回，都会回滚；提交成功后再次 Rollback 的错误可忽略。
	defer func() { _ = tx.Rollback() }()
	// 从事务获取 Ent 客户端，使用 db 执行的数据库操作都属于 tx。
	db := tx.Client()
	// 复用共享站点查询，在同一事务内确认登录站点，禁止从请求选择其他站点。
	owner, err := lookupSiteOwner(l.ctx, db, l.Logger)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	// 本站无设置时新增；已有记录按 operator_id 更新，全部操作属于当前事务。
	// 冲突时 UpdateXxx 使用上方 SetXxx 的值覆盖八项设置及审计信息，保留归属和首次创建时间。
	// 下方仅组装一条写入语句，不是每个 Update 都执行一次 SQL。
	err = db.OperatorSiteSetting.Create().
		SetOperatorID(owner.ID).                                       // 使用事务内已核对的站点 ID
		SetSiteTitle(title).                                           // 已校验并去除首尾空白的标题
		SetSiteLogoURL(logo).                                          // 可识别地址存 Key；兼容旧地址；空串清除
		SetEmailLoginLimit(in.GetEmailLoginLimit()).                   // 邮箱登录阈值，0 不限制；仅保存配置
		SetSmsLoginLimit(in.GetSmsLoginLimit()).                       // 短信登录阈值，0 不限制
		SetMemberPasswordErrorLimit(in.GetMemberPasswordErrorLimit()). // 密码错误次数上限，0 不限制
		SetIPMemberRegisterLimit(in.GetIpMemberRegisterLimit()).       // 同 IP 注册会员数上限，0 不限制
		SetRegisterCaptchaMode(in.GetRegisterCaptchaMode()).           // 注册验证：0 不验证、1 拖拽、2 图形
		SetLoginCaptchaMode(in.GetLoginCaptchaMode()).                 // 登录验证，模式含义与注册相同
		SetUpdatedBy(claims.UserID).                                   // 修改人来自登录上下文
		SetUpdatedAt(now).                                             // 修改时间由服务端生成
		OnConflictColumns(operatorsitesetting.FieldOperatorID).        // 同站唯一键冲突时更新已有记录
		UpdateSiteTitle().                                             // 用本次标题覆盖旧标题
		UpdateSiteLogoURL().                                           // 用本次 Logo 覆盖旧值，允许空串清除
		UpdateEmailLoginLimit().                                       // 覆盖邮箱阈值，0 恢复不限制
		UpdateSmsLoginLimit().                                         // 覆盖短信阈值，0 恢复不限制
		UpdateMemberPasswordErrorLimit().                              // 覆盖错误次数上限，不修改实际计数
		UpdateIPMemberRegisterLimit().                                 // 覆盖注册数上限，不修改已有会员
		UpdateRegisterCaptchaMode().                                   // 覆盖注册验证码模式
		UpdateLoginCaptchaMode().                                      // 覆盖登录验证码模式
		UpdateUpdatedBy().                                             // 覆盖为本次登录用户 ID
		UpdateUpdatedAt().                                             // 更新修改时间，保留首次创建时间
		Exec(l.ctx)                                                    // 执行事务内写入；仍须 Commit 成功
	if err != nil {
		// 复用数据库错误转换，记录内部原因并保留并发冲突等语义；返回后由 defer 回滚。
		return nil, siteSettingsError(l.ctx, l.Logger, err)
	}
	// 在同一事务内读取刚写入的设置，读取失败也会回滚本次写入。
	row, err := db.OperatorSiteSetting.Query().
		Where(operatorsitesetting.OperatorIDEQ(owner.ID)).
		Only(l.ctx)
	if err != nil {
		return nil, siteSettingsError(l.ctx, l.Logger, err)
	}
	// 提交前先生成对外响应，避免保存成功后才发现公开地址配置有误。
	result, err := toSiteSettingsModel(owner, row, l.svcCtx.Config)
	if err != nil {
		return nil, err
	}
	// 全部操作成功后提交事务，只有提交成功才向客户端返回保存结果。
	if err := tx.Commit(); err != nil {
		return nil, siteSettingsError(l.ctx, l.Logger, err)
	}
	return result, nil
}

// validateSiteSettings 校验完整八项设置，区分未传字段与显式提交的 0 或空 Logo。
func validateSiteSettings(in *operatorpb.SaveSiteSettingsRequest) (string, string, error) {
	// RPC 也可能被其他后端直接调用，不能依赖 HTTP 层已检查必填字段。
	if !hasSiteSettingsFields(in) {
		return "", "", status.Error(codes.InvalidArgument, "all eight site settings fields are required")
	}
	// 标题按去除首尾空白后的字节数限制；Logo 允许空字符串，以明确清除当前图片。
	title := strings.TrimSpace(in.GetSiteTitle())
	logo := strings.TrimSpace(in.GetSiteLogoUrl())
	if title == "" || len(title) > 100 {
		return "", "", status.Error(codes.InvalidArgument, "site_title must contain 1 to 100 bytes")
	}
	if len(logo) > 1024 || !validSiteLogoURL(logo) {
		return "", "", status.Error(codes.InvalidArgument, "site_logo_url must be an HTTP(S) URL or absolute path")
	}
	// 四项次数限制均为 0～100，0 表示不限制，不表示未提交。
	limits := []int64{
		in.GetEmailLoginLimit(),
		in.GetSmsLoginLimit(),
		in.GetMemberPasswordErrorLimit(),
		in.GetIpMemberRegisterLimit(),
	}
	for _, limit := range limits {
		if limit < 0 || limit > 100 {
			return "", "", status.Error(codes.InvalidArgument, "limits must be between 0 and 100; 0 means unlimited")
		}
	}
	// 验证码模式：0 不验证、1 拖拽验证、2 图形验证；此处只保存设置，不执行验证码校验。
	for _, mode := range []int64{in.GetRegisterCaptchaMode(), in.GetLoginCaptchaMode()} {
		if mode < 0 || mode > 2 {
			return "", "", status.Error(codes.InvalidArgument, "captcha modes must be 0, 1 or 2")
		}
	}
	return title, logo, nil
}

// hasSiteSettingsFields 检查完整保存的八项字段是否提交，0 和空 Logo 都属于已提交的合法值。
func hasSiteSettingsFields(in *operatorpb.SaveSiteSettingsRequest) bool {
	if in == nil || in.SiteTitle == nil || in.SiteLogoUrl == nil {
		return false
	}
	fields := []*int64{
		in.EmailLoginLimit,
		in.SmsLoginLimit,
		in.MemberPasswordErrorLimit,
		in.IpMemberRegisterLimit,
		in.RegisterCaptchaMode,
		in.LoginCaptchaMode,
	}
	for _, field := range fields {
		if field == nil {
			return false
		}
	}
	return true
}

// validSiteLogoURL 允许 HTTP(S) 地址或站内绝对路径，空字符串表示清除 Logo。
func validSiteLogoURL(value string) bool {
	if value == "" {
		return true
	}
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed.User != nil || strings.HasPrefix(value, "//") {
		return false
	}
	if parsed.Scheme == "" {
		return strings.HasPrefix(value, "/") && parsed.Host == ""
	}
	return (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Hostname() != ""
}
