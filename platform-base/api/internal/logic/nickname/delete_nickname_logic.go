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

type DeleteNicknameLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteNicknameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteNicknameLogic {
	return &DeleteNicknameLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteNicknameLogic) DeleteNickname(req *types.DeleteNicknameRequest) (resp *types.DeleteNicknameResponse, err error) {
	_, err = l.svcCtx.Nicknames.Delete(l.ctx, &nicknamepb.DeleteNicknameRequest{Id: req.Id})
	if err != nil {
		return nil, err
	}
	return &types.DeleteNicknameResponse{}, nil
}
