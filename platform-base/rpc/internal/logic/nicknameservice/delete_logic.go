package nicknameservicelogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/platform-base/pkg/i18nkey"
	"oa.98ent.com/p9/platform-base/pkg/rpc/grpcerror"
	"oa.98ent.com/p9/platform-base/rpc/internal/enterror"
	"oa.98ent.com/p9/platform-base/rpc/internal/svc"
	"oa.98ent.com/p9/platform-base/rpc/pb/platformbaserpc/nicknamepb"
)

type DeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteLogic {
	return &DeleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteLogic) Delete(in *nicknamepb.DeleteNicknameRequest) (*nicknamepb.DeleteNicknameResponse, error) {
	if err := requirePlatform(l.ctx); err != nil {
		return nil, err
	}

	if in.GetId() <= 0 {
		return nil, grpcerror.InvalidArgument(i18nkey.ValidationError)
	}
	if err := l.svcCtx.DB.Nickname.DeleteOneID(in.Id).Exec(l.ctx); err != nil {
		return nil, enterror.Handle(l.Logger, err)
	}
	l.svcCtx.Nicknames.Invalidate(l.ctx)
	return &nicknamepb.DeleteNicknameResponse{}, nil
}
