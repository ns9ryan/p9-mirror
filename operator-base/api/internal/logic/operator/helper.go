package operator

import (
	"oa.98ent.com/p9/operator-base/api/internal/types"
	"oa.98ent.com/p9/operator-base/rpc/pb/operatorbaserpc/operatorpb"
)

// toSiteSettingsInfo 保持保存响应与完整资料中的设置字段一致，0 和空值原样返回。
func toSiteSettingsInfo(result *operatorpb.SiteSettingsModel) *types.SiteSettingsInfo {
	return &types.SiteSettingsInfo{
		OperatorCode:             result.GetOperatorCode(),             // 分站编码
		SiteTitle:                result.GetSiteTitle(),                // 网站标题
		SiteLogoUrl:              result.GetSiteLogoUrl(),              // 网站 Logo 地址
		EmailLoginLimit:          result.GetEmailLoginLimit(),          // 邮箱登录阈值，0 不限制
		SmsLoginLimit:            result.GetSmsLoginLimit(),            // 短信登录阈值，0 不限制
		UpdatedBy:                result.GetUpdatedBy(),                // 最后修改用户ID
		MemberPasswordErrorLimit: result.GetMemberPasswordErrorLimit(), // 会员密码错误上限，0 不限制
		IpMemberRegisterLimit:    result.GetIpMemberRegisterLimit(),    // 同一 IP 注册上限，0 不限制
		RegisterCaptchaMode:      result.GetRegisterCaptchaMode(),      // 注册验证方式
		LoginCaptchaMode:         result.GetLoginCaptchaMode(),         // 登录验证方式
		CreatedAt:                result.GetCreatedAt(),                // 创建时间，Unix 秒
		UpdatedAt:                result.GetUpdatedAt(),                // 更新时间，Unix 秒
		Configured:               result.GetConfigured(),               // 是否已有设置记录
	}
}
