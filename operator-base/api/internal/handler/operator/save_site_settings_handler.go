package operator

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"oa.98ent.com/p9/common/response"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-base/api/internal/logic/operator"
	"oa.98ent.com/p9/operator-base/api/internal/svc"
	"oa.98ent.com/p9/operator-base/api/internal/types"
)

// SaveSiteSettingsHandler 读取完整网站设置，交给 Logic 按登录站点保存。
func SaveSiteSettingsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := parseSiteSettings(w, r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := operator.NewSaveSiteSettingsLogic(r.Context(), svcCtx)
		result, err := l.SaveSiteSettings(req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		response.OkCtx(r.Context(), w, result)
	}
}

// parseSiteSettings 要求八项字段完整提交；0 和空 Logo 合法，不能当成漏传。
func parseSiteSettings(w http.ResponseWriter, r *http.Request) (*types.SaveSiteSettingsRequest, error) {
	var req types.SaveSiteSettingsRequest
	// 限制 JSON 正文不超过 16 KiB；此处只解析参数，身份和归属交给 Logic 校验。
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16*1024))
	// 不接受 operator_id/operator_code/domain_name/updated_by 等归属及身份字段。
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		return nil, xerr.BadRequest("invalid site settings request")
	}
	if req.SiteTitle == nil ||
		req.SiteLogoUrl == nil ||
		req.EmailLoginLimit == nil ||
		req.SmsLoginLimit == nil ||
		req.MemberPasswordErrorLimit == nil ||
		req.IpMemberRegisterLimit == nil ||
		req.RegisterCaptchaMode == nil ||
		req.LoginCaptchaMode == nil {
		return nil, xerr.BadRequest("all eight site settings fields are required")
	}
	// 正文只能包含一个 JSON 对象，拒绝其后拼接其他 JSON 或无效内容。
	if decoder.Decode(new(any)) != io.EOF {
		return nil, xerr.BadRequest("site settings body must be a single JSON object")
	}
	return &req, nil
}
