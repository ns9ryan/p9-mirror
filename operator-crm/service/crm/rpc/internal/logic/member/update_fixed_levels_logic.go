package memberlogic

import (
	"context"
	"oa.98ent.com/p9/common/xerr"

	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateFixedLevelsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateFixedLevelsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateFixedLevelsLogic {
	return &UpdateFixedLevelsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateFixedLevelsLogic) UpdateFixedLevels(in *memberpb.UpdateMemberFixedLevelsRequest) (*memberpb.UpdateMemberLevelsResponse, error) {
	if err := l.svcCtx.Member.UpdateFixedLevels(l.ctx, in.GetMemberId(), in.GetFixedLevelIds()); err != nil {
		return nil, xerr.RpcErr(err)
	}

	return &memberpb.UpdateMemberLevelsResponse{}, nil
}
