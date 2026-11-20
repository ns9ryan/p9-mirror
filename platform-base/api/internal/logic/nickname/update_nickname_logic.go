// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package nickname

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/platform-base/api/internal/svc"
	"oa.98ent.com/p9/platform-base/api/internal/types"
	"oa.98ent.com/p9/platform-base/rpc/pb/platformbaserpc/nicknamepb"
)

type UpdateNicknameLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateNicknameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateNicknameLogic {
	return &UpdateNicknameLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateNicknameLogic) UpdateNickname(req *types.UpdateNicknameRequest) (resp *types.UpdateNicknameResponse, err error) {
	_, err = l.svcCtx.Nicknames.Update(l.ctx, &nicknamepb.UpdateNicknameRequest{Id: req.Id, Name: req.Name})
	if err != nil {
		return nil, err
	}
	return &types.UpdateNicknameResponse{}, nil
}
