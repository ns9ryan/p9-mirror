// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package manage

import (
	"context"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"

	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateFixedLevelsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateFixedLevelsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateFixedLevelsLogic {
	return &UpdateFixedLevelsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateFixedLevelsLogic) UpdateFixedLevels(req *types.UpdateMemberFixedLevelsRequest) (resp *types.EmptyResponse, err error) {
	_, err = l.svcCtx.MemberRpc.UpdateFixedLevels(l.ctx, &memberpb.UpdateMemberFixedLevelsRequest{MemberId: req.MemberId, FixedLevelIds: req.FixedLevelIds})
	if err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}
