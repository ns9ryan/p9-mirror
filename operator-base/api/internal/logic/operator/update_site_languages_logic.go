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

type UpdateSiteLanguagesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateSiteLanguagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateSiteLanguagesLogic {
	return &UpdateSiteLanguagesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UpdateSiteLanguages 调用本站 RPC 保存开放语言完整勾选结果。
func (l *UpdateSiteLanguagesLogic) UpdateSiteLanguages(
	req *types.UpdateSiteLanguagesRequest,
) (*types.UpdateSiteLanguagesResponse, error) {
	// 复用认证中间件中的站点编码，不接受 Host 或正文指定其他站点。
	if ctxdata.OperatorCodeFromCtx(l.ctx) == "" {
		return nil, xerr.Forbidden("login operator is required")
	}
	if l.svcCtx.OperatorRpc == nil {
		return nil, xerr.Err(http.StatusServiceUnavailable, "operator RPC is not configured")
	}
	// 调用本站 Operator RPC 的 UpdateSiteLanguages 方法。
	result, err := l.svcCtx.OperatorRpc.UpdateSiteLanguages(
		l.ctx,
		&operatorpb.UpdateSiteLanguagesRequest{
			LanguageCodes: req.LanguageCodes, // 本站勾选的语言编码
		},
	)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, xerr.InternalServerError("update languages RPC returned no response")
	}
	return &types.UpdateSiteLanguagesResponse{
		LanguageCodes: append([]string{}, result.GetLanguageCodes()...),
	}, nil
}
