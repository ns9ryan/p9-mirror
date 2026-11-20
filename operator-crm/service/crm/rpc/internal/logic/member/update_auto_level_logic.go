package memberlogic

import (
	"context"
	"oa.98ent.com/p9/common/xerr"

	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateAutoLevelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateAutoLevelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAutoLevelLogic {
	return &UpdateAutoLevelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateAutoLevelLogic) UpdateAutoLevel(in *memberpb.UpdateMemberAutoLevelRequest) (*memberpb.UpdateMemberLevelsResponse, error) {
	if err := l.svcCtx.Member.UpdateAutoLevel(l.ctx, in.GetMemberId(), in.GetAutoLevelId()); err != nil {
		return nil, xerr.RpcErr(err)
	}

	return &memberpb.UpdateMemberLevelsResponse{}, nil
}
