package ipgeo

import (
	"context"

	"oa.98ent.com/p9/integration/api/internal/convert"
	"oa.98ent.com/p9/integration/api/internal/svc"
	"oa.98ent.com/p9/integration/api/internal/types"
	"oa.98ent.com/p9/integration/rpc/types/integration"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveDefaultPolicyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSaveDefaultPolicyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveDefaultPolicyLogic {
	return &SaveDefaultPolicyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SaveDefaultPolicyLogic) SaveDefaultPolicy(req *types.SavePolicyReq) (*types.PolicyListData, error) {
	out, err := l.svcCtx.Integration.SaveDefaultPolicies(l.ctx, &integration.SavePolicyReq{
		List: convert.PolicyItems(req.List),
	})
	if err != nil {
		return nil, err
	}
	return convert.PolicyListData(out), nil
}
