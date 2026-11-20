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

type CreateNicknameLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateNicknameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateNicknameLogic {
	return &CreateNicknameLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateNicknameLogic) CreateNickname(req *types.CreateNicknameRequest) (resp *types.CreateNicknameResponse, err error) {
	r, err := l.svcCtx.Nicknames.Create(l.ctx, &nicknamepb.CreateNicknameRequest{Name: req.Name})
	if err != nil {
		return nil, err
	}
	return &types.CreateNicknameResponse{Nickname: types.NicknameInfo{Id: r.Nickname.Id, Name: r.Nickname.Name}}, nil
}
