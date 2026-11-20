package ipgeo

import (
	"context"

	"oa.98ent.com/p9/integration/api/internal/convert"
	"oa.98ent.com/p9/integration/api/internal/svc"
	"oa.98ent.com/p9/integration/api/internal/types"
	"oa.98ent.com/p9/integration/rpc/types/integration"

	"github.com/zeromicro/go-zero/core/logx"
)

type SavePolicyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSavePolicyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SavePolicyLogic {
	return &SavePolicyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SavePolicyLogic) SavePolicy(req *types.SavePolicyReq) (*types.PolicyListData, error) {
	out, err := l.svcCtx.Integration.SavePolicies(l.ctx, &integration.SavePolicyReq{
		List: convert.PolicyItems(req.List),
	})
	if err != nil {
		return nil, err
	}
	return convert.PolicyListData(out), nil
}
