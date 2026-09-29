// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package node

import (
	"context"

	"oa.98ent.com/p9/node-dispatch/api/internal/svc"
	"oa.98ent.com/p9/node-dispatch/api/internal/types"
	"oa.98ent.com/p9/node-dispatch/rpc/pb/nodedispatchrpc/nodepb"

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

// ResetNodeAuthSecret 重置节点认证密钥
func (l *ResetNodeAuthSecretLogic) ResetNodeAuthSecret(req *types.ResetNodeAuthSecretRequest) (resp *types.ResetNodeAuthSecretResponse, err error) {
	// 重置节点认证密钥
	result, err := l.svcCtx.NodeRpc.ResetAuthSecret(
		l.ctx,
		&nodepb.ResetNodeAuthSecretRequest{
			Id: req.Id, // 节点ID
		},
	)
	if err != nil {
		return nil, err
	}

	// 返回新的节点认证密钥
	return &types.ResetNodeAuthSecretResponse{
		AuthSecret: result.AuthSecret, // 新节点认证密钥, 仅本次返回
	}, nil
}
