// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package category

import (
	"context"

	"oa.98ent.com/p9/operator-game/api/internal/logic"
	"oa.98ent.com/p9/operator-game/api/internal/svc"
	"oa.98ent.com/p9/operator-game/api/internal/types"
	pb "oa.98ent.com/p9/operator-game/rpc/pb/operator_game"

	"github.com/zeromicro/go-zero/core/logx"
)

type GameCategoryListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGameCategoryListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GameCategoryListLogic {
	return &GameCategoryListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GameCategoryListLogic) GameCategoryList(req *types.GameCategoryListReq) (resp *types.GameCategoryListResp, err error) {
	// 构建 RPC 请求
	pbReq := &pb.GetGameCategoryListRequest{
		Page:         int32(req.Page),
		PageSize:     int32(req.PageSize),
		CategoryCode: req.CategoryCode,
		Status:       int32(req.Status),
	}

	// 调用 RPC 服务
	pbResp, err := l.svcCtx.OperatorGameGrpcClient.GetGameCategoryServiceClient().GetGameCategoryList(l.ctx, pbReq)
	if err != nil {
		l.Errorf("GetGameCategoryList failed: %v", err)
		return nil, err
	}

	// 转换响应数据
	var respList []types.GameCategoryResp
	for _, item := range pbResp.Items {
		respList = append(respList, *logic.CategoryProtoToResponse(l.ctx, item))
	}

	return &types.GameCategoryListResp{
		List:  respList,
		Total: pbResp.Total,
	}, nil
}
