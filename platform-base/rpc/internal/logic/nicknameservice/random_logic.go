package nicknameservicelogic

import (
	"context"
	"math/rand/v2"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/platform-base/pkg/i18nkey"
	"oa.98ent.com/p9/platform-base/pkg/rpc/grpcerror"
	"oa.98ent.com/p9/platform-base/rpc/internal/enterror"
	"oa.98ent.com/p9/platform-base/rpc/internal/svc"
	"oa.98ent.com/p9/platform-base/rpc/pb/platformbaserpc/nicknamepb"
)

type RandomLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRandomLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RandomLogic {
	return &RandomLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RandomLogic) Random(in *nicknamepb.RandomNicknameRequest) (*nicknamepb.RandomNicknameResponse, error) {
	rows, err := l.svcCtx.Nicknames.All(l.ctx)
	if err != nil {
		return nil, enterror.Handle(l.Logger, err)
	}
	if len(rows) == 0 {
		return nil, grpcerror.NotFound(i18nkey.DataNotFound)
	}
	row := rows[rand.IntN(len(rows))]
	return &nicknamepb.RandomNicknameResponse{Nickname: &nicknamepb.NicknameInfo{Id: row.ID, Name: row.Name}}, nil
}
