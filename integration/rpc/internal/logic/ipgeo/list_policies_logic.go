package ipgeo

import (
	"context"

	geo "oa.98ent.com/p9/integration/rpc/internal/ipgeo"
	"oa.98ent.com/p9/integration/rpc/internal/svc"
	"oa.98ent.com/p9/integration/rpc/types/integration"

	"github.com/zeromicro/go-zero/core/logx"
)

// ListPoliciesLogic 列出当前租户策略（RPC listPolicies）。
type ListPoliciesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListPoliciesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPoliciesLogic {
	return &ListPoliciesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListPolicies 返回叠租户策略后的服务商列表；无策略则用全局默认。
func (l *ListPoliciesLogic) ListPolicies(_ *integration.Empty) (*integration.PolicyListResp, error) {
	if l.svcCtx.Deps == nil || l.svcCtx.Deps.IpGeo == nil {
		return nil, geoNotReady()
	}
	list, err := l.svcCtx.Deps.IpGeo.ListPolicies(l.ctx, geo.AccountOperatorCode(l.ctx))
	if err != nil {
		return nil, mapStoreErr(err)
	}
	return policiesToPB(list), nil
}
