// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package channel

import (
	"context"

	"oa.98ent.com/p9/operator-game/api/internal/logic"
	"oa.98ent.com/p9/operator-game/api/internal/svc"
	"oa.98ent.com/p9/operator-game/api/internal/types"
	pb "oa.98ent.com/p9/operator-game/rpc/pb/operator_game"

	"github.com/zeromicro/go-zero/core/logx"
)

type GameChannelListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGameChannelListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GameChannelListLogic {
	return &GameChannelListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GameChannelListLogic) GameChannelList(req *types.GameChannelListReq) (resp *types.GameChannelListResp, err error) {
	pbReq := &pb.GetGameChannelListRequest{
		Page:        int32(req.Page),
		PageSize:    int32(req.PageSize),
		ChannelCode: req.ChannelCode,
		Status:      int32(req.Status),
	}

	pbResp, err := l.svcCtx.OperatorGameGrpcClient.GetGameChannelServiceClient().GetGameChannelList(l.ctx, pbReq)
	if err != nil {
		l.Errorf("GetGameChannelList failed: %v", err)
		return nil, err
	}

	var respList []types.GameChannelResp
	for _, item := range pbResp.Items {
		respList = append(respList, *logic.ChannelProtoToResponse(l.ctx, item))
	}

	return &types.GameChannelListResp{
		List:  respList,
		Total: pbResp.Total,
	}, nil
}
