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

type GetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLogic {
	return &GetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetLogic) Get(in *nicknamepb.GetNicknameRequest) (*nicknamepb.GetNicknameResponse, error) {
	if in.GetId() <= 0 {
		return nil, grpcerror.InvalidArgument(i18nkey.ValidationError)
	}
	row, err := l.svcCtx.DB.Nickname.Get(l.ctx, in.Id)
	if err != nil {
		return nil, enterror.Handle(l.Logger, err)
	}
	return &nicknamepb.GetNicknameResponse{Nickname: info(row)}, nil
}
