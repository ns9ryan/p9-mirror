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

type UpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateLogic {
	return &UpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateLogic) Update(in *nicknamepb.UpdateNicknameRequest) (*nicknamepb.UpdateNicknameResponse, error) {
	if err := requirePlatform(l.ctx); err != nil {
		return nil, err
	}

	name := strings.TrimSpace(in.GetName())
	if in.GetId() <= 0 || !validName(name) {
		return nil, grpcerror.InvalidArgument(i18nkey.ValidationError)
	}
	_, err := l.svcCtx.DB.Nickname.UpdateOneID(in.Id).SetName(name).Save(l.ctx)
	if err != nil {
		return nil, enterror.Handle(l.Logger, err)
	}
	l.svcCtx.Nicknames.Invalidate(l.ctx)
	return &nicknamepb.UpdateNicknameResponse{}, nil
}
