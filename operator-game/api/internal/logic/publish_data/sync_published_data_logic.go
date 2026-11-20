// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package publish_data

import (
	"context"

	"oa.98ent.com/p9/operator-game/rpc/pb/operator_game"

	"oa.98ent.com/p9/operator-game/api/internal/svc"
	"oa.98ent.com/p9/operator-game/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SyncPublishedDataLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSyncPublishedDataLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncPublishedDataLogic {
	return &SyncPublishedDataLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SyncPublishedDataLogic) SyncPublishedData(req *types.SyncPublishedDataReq) (resp *types.SyncPublishedDataResp, err error) {
	resp = &types.SyncPublishedDataResp{}

	// rate limit check
	if !l.svcCtx.SyncRateLimiter.CheckRateLimit("publish_data_sync") {
		resp.Success = false
		resp.Message = "操作过于频繁，请稍后再试"
		return resp, nil
	}

	// call RPC
	if l.svcCtx.OperatorGameGrpcClient == nil {
		resp.Success = false
		resp.Message = "RPC 客户端未初始化"
		return resp, nil
	}

	rpcReq := &operator_game.SyncPublishedDataRequest{OpCode: req.OpCode}
	rpcResp, rpcErr := l.svcCtx.OperatorGameGrpcClient.GetPublishDataServiceClient().SyncPublishedData(l.ctx, rpcReq)
	if rpcErr != nil {
		l.Errorf("call rpc SyncPublishedData failed: %v", rpcErr)
		resp.Success = false
		resp.Message = rpcErr.Error()
		return resp, nil
	}

	// map RPC response to API response
	resp.Success = rpcResp.GetSuccess()
	resp.Message = rpcResp.GetMessage()

	if s := rpcResp.GetCategoryStat(); s != nil {
		resp.CategoryStat = types.SyncStatistic{
			Total:        s.GetTotal(),
			Success:      s.GetSuccess(),
			Failed:       s.GetFailed(),
			FailedReason: s.GetFailedReason(),
		}
	}
	if s := rpcResp.GetProviderStat(); s != nil {
		resp.ProviderStat = types.SyncStatistic{
			Total:        s.GetTotal(),
			Success:      s.GetSuccess(),
			Failed:       s.GetFailed(),
			FailedReason: s.GetFailedReason(),
		}
	}
	if s := rpcResp.GetChannelStat(); s != nil {
		resp.ChannelStat = types.SyncStatistic{
			Total:        s.GetTotal(),
			Success:      s.GetSuccess(),
			Failed:       s.GetFailed(),
			FailedReason: s.GetFailedReason(),
		}
	}
	if s := rpcResp.GetGameStat(); s != nil {
		resp.GameStat = types.SyncStatistic{
			Total:        s.GetTotal(),
			Success:      s.GetSuccess(),
			Failed:       s.GetFailed(),
			FailedReason: s.GetFailedReason(),
		}
	}

	return resp, nil
}
