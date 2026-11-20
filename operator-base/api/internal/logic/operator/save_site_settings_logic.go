package operator

import (
	"context"
	"net/http"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-base/api/internal/svc"
	"oa.98ent.com/p9/operator-base/api/internal/types"
	"oa.98ent.com/p9/operator-base/rpc/pb/operatorbaserpc/operatorpb"
)

type SaveSiteSettingsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSaveSiteSettingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveSiteSettingsLogic {
	return &SaveSiteSettingsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// SaveSiteSettings 调用本站 RPC 完整保存八项设置。
func (l *SaveSiteSettingsLogic) SaveSiteSettings(
	req *types.SaveSiteSettingsRequest,
) (*types.SaveSiteSettingsResponse, error) {
	// 校验保存所需的登录用户、站点身份及本站 RPC 配置。
	claims := ctxdata.ClaimsFromCtx(l.ctx)
	if claims == nil || claims.UserID <= 0 || claims.OperatorCode == "" {
		return nil, xerr.Forbidden("login user and operator are required")
	}
	if l.svcCtx.OperatorRpc == nil {
		return nil, xerr.Err(http.StatusServiceUnavailable, "operator RPC is not configured")
	}
	// 调用本站设置保存 RPC，归属和修改人由后端校验。
	result, err := l.svcCtx.OperatorRpc.SaveSiteSettings(
		l.ctx,
		&operatorpb.SaveSiteSettingsRequest{
			SiteTitle:                req.SiteTitle,                // 网站标题
			SiteLogoUrl:              req.SiteLogoUrl,              // 网站 Logo 地址
			EmailLoginLimit:          req.EmailLoginLimit,          // 邮箱登录阈值，0 不限制
			SmsLoginLimit:            req.SmsLoginLimit,            // 短信登录阈值，0 不限制
			MemberPasswordErrorLimit: req.MemberPasswordErrorLimit, // 会员密码错误上限，0 不限制
			IpMemberRegisterLimit:    req.IpMemberRegisterLimit,    // 同一 IP 注册上限，0 不限制
			RegisterCaptchaMode:      req.RegisterCaptchaMode,      // 注册验证方式
			LoginCaptchaMode:         req.LoginCaptchaMode,         // 登录验证方式
		},
	)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, xerr.InternalServerError("site settings RPC returned no response")
	}
	// 返回的是 RPC 已提交成功的设置，不在 API 层再次写库或重复裁剪 Logo 地址。
	// 复用查询接口的设置映射，Logo 使用 RPC 拼好的公开地址，0 和空字符串保持原义。
	return &types.SaveSiteSettingsResponse{
		SiteSettingsInfo: *toSiteSettingsInfo(result),
	}, nil
}
