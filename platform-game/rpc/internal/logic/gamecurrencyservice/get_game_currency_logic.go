package gamecurrencyservicelogic

import (
	"context"
	"fmt"

	"oa.98ent.com/p9/platform-game/rpc/ent"

	"oa.98ent.com/p9/platform-game/rpc/internal/constant"
	"oa.98ent.com/p9/platform-game/rpc/internal/logic"
	"oa.98ent.com/p9/platform-game/rpc/internal/svc"
	platformgame "oa.98ent.com/p9/platform-game/rpc/pb/platform_game"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetGameCurrencyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetGameCurrencyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGameCurrencyLogic {
	return &GetGameCurrencyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取单个游戏货币
func (l *GetGameCurrencyLogic) GetGameCurrency(in *platformgame.GetGameCurrencyRequest) (*platformgame.GetGameCurrencyResp, error) {
	l.Infof("[RPC GetGameCurrency] received request: id=%d", in.Id)

	if l.svcCtx == nil || l.svcCtx.DAOManager == nil {
		l.Errorf("[RPC GetGameCurrency] DAO Manager not available")
		return &platformgame.GetGameCurrencyResp{
			Code:    constant.CodeInternalError,
			Message: "DAO Manager not available",
		}, nil
	}

	currency, err := l.svcCtx.DAOManager.GameCurrency.GetGameCurrencyByID(l.ctx, in.Id)
	if err != nil {
		l.Errorf("[RPC GetGameCurrency] query failed: %v", err)
		return &platformgame.GetGameCurrencyResp{
			Code:    constant.CodeInternalError,
			Message: "failed to get currency: " + err.Error(),
		}, nil
	}
	l.Infof("[RPC GetGameCurrency] query result: id=%d, game_code=%s, currency_code=%s, status=%d, deleted_at=%v",
		currency.ID, currency.GameCode, currency.CurrencyCode, currency.Status, currency.DeletedAt)

	return &platformgame.GetGameCurrencyResp{
		Code:    constant.CodeSuccess,
		Message: "ok",
		Data:    logic.CurrencyModelToProto(currency),
	}, nil
}

func GetGameRecordByCode(ctx context.Context, svcCtx *svc.ServiceContext, gameCode string) (*ent.Game, error) {
	if svcCtx == nil || svcCtx.DAOManager == nil {
		return nil, fmt.Errorf("DAO Manager not available")
	}
	return svcCtx.DAOManager.Game.GetGameByCode(ctx, gameCode)
}
