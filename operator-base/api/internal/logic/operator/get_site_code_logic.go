package operator

import (
	"context"
	"net/http"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-base/api/internal/svc"
	"oa.98ent.com/p9/operator-base/api/internal/types"
	"oa.98ent.com/p9/operator-base/rpc/pb/operatorbaserpc/operatorpb"
)

type GetSiteCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSiteCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSiteCodeLogic {
	return &GetSiteCodeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetSiteCode 只识别站点编码，不查询配置，也不校验登录身份或接口权限。
func (l *GetSiteCodeLogic) GetSiteCode(req *types.GetSiteCodeRequest) (*types.GetSiteCodeResponse, error) {
	if req == nil || req.DomainName == "" {
		return nil, xerr.BadRequest("domain_name is required")
	}
	if l.svcCtx.OperatorRpc == nil {
		return nil, xerr.Err(http.StatusServiceUnavailable, "operator RPC is not configured")
	}
	// 调用本站 Operator RPC 的 GetSiteCode 方法。
	result, err := l.svcCtx.OperatorRpc.GetSiteCode(
		l.ctx,
		&operatorpb.GetSiteCodeRequest{
			DomainName: req.DomainName, // 后台域名
		},
	)
	if err != nil {
		return nil, err
	}
	// 下游返回空响应或空编码都不能当作查询成功，避免前端使用无效站点身份。
	if result.GetOperatorCode() == "" {
		return nil, xerr.InternalServerError("site code RPC returned no code")
	}
	return &types.GetSiteCodeResponse{
		OperatorCode: result.GetOperatorCode(),
	}, nil
}
