package ipgeo

import (
	"context"

	"oa.98ent.com/p9/integration/api/internal/convert"
	"oa.98ent.com/p9/integration/api/internal/svc"
	"oa.98ent.com/p9/integration/api/internal/types"
	"oa.98ent.com/p9/integration/rpc/types/integration"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListPolicyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListPolicyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPolicyLogic {
	return &ListPolicyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListPolicyLogic) ListPolicy() (*types.PolicyListData, error) {
	out, err := l.svcCtx.Integration.ListPolicies(l.ctx, &integration.Empty{})
	if err != nil {
		return nil, err
	}
	return convert.PolicyListData(out), nil
}
