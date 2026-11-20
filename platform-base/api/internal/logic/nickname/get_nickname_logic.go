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

type GetNicknameLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetNicknameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNicknameLogic {
	return &GetNicknameLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetNicknameLogic) GetNickname(req *types.GetNicknameRequest) (resp *types.GetNicknameResponse, err error) {
	r, err := l.svcCtx.Nicknames.Get(l.ctx, &nicknamepb.GetNicknameRequest{Id: req.Id})
	if err != nil {
		return nil, err
	}
	return &types.GetNicknameResponse{Nickname: types.NicknameInfo{Id: r.Nickname.Id, Name: r.Nickname.Name}}, nil
}
