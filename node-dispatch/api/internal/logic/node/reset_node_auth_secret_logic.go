// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package node

import (
	"context"

	"oa.98ent.com/p9/node-dispatch/api/internal/svc"
	"oa.98ent.com/p9/node-dispatch/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ResetNodeAuthSecretLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewResetNodeAuthSecretLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetNodeAuthSecretLogic {
	return &ResetNodeAuthSecretLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ResetNodeAuthSecretLogic) ResetNodeAuthSecret(req *types.ResetNodeAuthSecretRequest) (resp *types.ResetNodeAuthSecretResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
