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

type UpdateSiteRegionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateSiteRegionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateSiteRegionsLogic {
	return &UpdateSiteRegionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UpdateSiteRegions 调用本站 RPC 保存经营地区完整勾选结果。
func (l *UpdateSiteRegionsLogic) UpdateSiteRegions(
	req *types.UpdateSiteRegionsRequest,
) (*types.UpdateSiteRegionsResponse, error) {
	// 复用认证中间件写入的登录站点身份，本站 RPC 按该身份读取分配并保存。
	if ctxdata.OperatorCodeFromCtx(l.ctx) == "" {
		return nil, xerr.Forbidden("login operator is required")
	}
	if l.svcCtx.OperatorRpc == nil {
		return nil, xerr.Err(http.StatusServiceUnavailable, "operator RPC is not configured")
	}
	// 调用本站 Operator RPC 的 UpdateSiteRegions 方法。
	result, err := l.svcCtx.OperatorRpc.UpdateSiteRegions(
		l.ctx,
		&operatorpb.UpdateSiteRegionsRequest{
			RegionCodes: req.RegionCodes, // 本站勾选的地区编码
		},
	)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, xerr.InternalServerError("update regions RPC returned no response")
	}
	// 以 RPC 实际保存结果为准；全部关闭时返回 []，保持前端数组类型稳定。
	return &types.UpdateSiteRegionsResponse{
		RegionCodes: append([]string{}, result.GetRegionCodes()...),
	}, nil
}
