package ipgeo

import (
	"context"

	geo "oa.98ent.com/p9/integration/rpc/internal/ipgeo"
	"oa.98ent.com/p9/integration/rpc/internal/svc"
	"oa.98ent.com/p9/integration/rpc/types/integration"

	"github.com/zeromicro/go-zero/core/logx"
)

// SavePoliciesLogic 保存当前租户策略（RPC savePolicies）。
type SavePoliciesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSavePoliciesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SavePoliciesLogic {
	return &SavePoliciesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SavePolicies 写入租户维度启用/优先级，不影响其它租户与全局默认连接配置。
func (l *SavePoliciesLogic) SavePolicies(in *integration.SavePolicyReq) (*integration.PolicyListResp, error) {
	if l.svcCtx.Deps == nil || l.svcCtx.Deps.IpGeo == nil {
		return nil, geoNotReady()
	}
	list, err := l.svcCtx.Deps.IpGeo.SavePolicies(
		l.ctx, geo.AccountOperatorCode(l.ctx), policyItemsFromPB(in.GetList()))
	if err != nil {
		return nil, mapStoreErr(err)
	}
	return policiesToPB(list), nil
}
