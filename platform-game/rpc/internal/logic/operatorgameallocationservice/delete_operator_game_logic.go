package operatorgameallocationservicelogic

import (
	"context"

	"oa.98ent.com/p9/platform-game/rpc/internal/svc"
	"oa.98ent.com/p9/platform-game/rpc/internal/utils"
	"oa.98ent.com/p9/platform-game/rpc/pb/platform_game"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteOperatorGameLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteOperatorGameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteOperatorGameLogic {
	return &DeleteOperatorGameLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 根据opcode删除分站游戏分配
func (l *DeleteOperatorGameLogic) DeleteOperatorGame(in *platform_game.DeleteOperatorGameRequest) (*platform_game.DeleteOperatorGameResp, error) {
	countMap, err := l.svcCtx.DAOManager.OperatorGame.DeleteOperatorGameByOpCode(l.ctx, in.OpCode)
	if err != nil {
		return nil, err
	}
	l.Infof("deleted operator game by opCode: %s, countMap: %s", in.OpCode, utils.JSON(countMap))
	return &platform_game.DeleteOperatorGameResp{}, nil
}
