package ipgeo

import (
	"context"

	"oa.98ent.com/p9/integration/rpc/internal/svc"
	"oa.98ent.com/p9/integration/rpc/types/integration"

	"github.com/zeromicro/go-zero/core/logx"
)

// SaveDefaultPoliciesLogic 保存全局默认策略（RPC saveDefaultPolicies）。
type SaveDefaultPoliciesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveDefaultPoliciesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveDefaultPoliciesLogic {
	return &SaveDefaultPoliciesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SaveDefaultPolicies 更新 Provider 表全局 is_enabled/priority（非租户覆盖行）。
func (l *SaveDefaultPoliciesLogic) SaveDefaultPolicies(in *integration.SavePolicyReq) (*integration.PolicyListResp, error) {
	if l.svcCtx.Deps == nil || l.svcCtx.Deps.IpGeo == nil {
		return nil, geoNotReady()
	}
	list, err := l.svcCtx.Deps.IpGeo.SaveDefaultPolicies(l.ctx, policyItemsFromPB(in.GetList()))
	if err != nil {
		return nil, mapStoreErr(err)
	}
	return policiesToPB(list), nil
}
