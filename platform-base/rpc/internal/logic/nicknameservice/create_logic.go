package nicknameservicelogic

import (
	"context"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/platform-base/pkg/i18nkey"
	"oa.98ent.com/p9/platform-base/pkg/rpc/grpcerror"
	"oa.98ent.com/p9/platform-base/rpc/internal/enterror"
	"oa.98ent.com/p9/platform-base/rpc/internal/svc"
	"oa.98ent.com/p9/platform-base/rpc/pb/platformbaserpc/nicknamepb"
)

type CreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateLogic {
	return &CreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateLogic) Create(in *nicknamepb.CreateNicknameRequest) (*nicknamepb.CreateNicknameResponse, error) {
	if err := requirePlatform(l.ctx); err != nil {
		return nil, err
	}

	name := strings.TrimSpace(in.GetName())
	if !validName(name) {
		return nil, grpcerror.InvalidArgument(i18nkey.ValidationError)
	}
	row, err := l.svcCtx.DB.Nickname.Create().SetName(name).Save(l.ctx)
	if err != nil {
		return nil, enterror.Handle(l.Logger, err)
	}
	l.svcCtx.Nicknames.Invalidate(l.ctx)
	return &nicknamepb.CreateNicknameResponse{Nickname: info(row)}, nil
}
