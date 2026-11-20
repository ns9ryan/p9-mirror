package gameorderservicelogic

import (
	"context"

	"oa.98ent.com/p9/operator-game/rpc/ent/gameorder"
	"oa.98ent.com/p9/operator-game/rpc/internal/svc"
	"oa.98ent.com/p9/operator-game/rpc/pb/operator_game"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryBetMembersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryBetMembersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryBetMembersLogic {
	return &QueryBetMembersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询指定时间范围内下注的会员
func (l *QueryBetMembersLogic) QueryBetMembers(in *operator_game.QueryBetMembersRequest) (*operator_game.QueryBetMembersResponse, error) {
	// 参数验证
	if in.OperatorCode == "" {
		l.Errorf("QueryBetMembers: operator_code is empty")
		return &operator_game.QueryBetMembersResponse{
			Code:    400,
			Message: "operator_code is required",
		}, nil
	}

	if in.StartTime >= in.EndTime {
		l.Errorf("QueryBetMembers: invalid time range, start_time=%d >= end_time=%d", in.StartTime, in.EndTime)
		return &operator_game.QueryBetMembersResponse{
			Code:    400,
			Message: "start_time must be less than end_time",
		}, nil
	}

	// 查询满足条件的所有 game_order
	// WHERE operator_code = ? AND settle_at BETWEEN ? AND ?
	orders, err := l.svcCtx.DB.GameOrder.Query().
		Where(
			gameorder.OperatorCode(in.OperatorCode),
			gameorder.SettleAtGTE(in.StartTime),
			gameorder.SettleAtLTE(in.EndTime),
		).
		All(l.ctx)

	if err != nil {
		l.Errorf("QueryBetMembers: query game_order failed, operator_code=%s, start_time=%d, end_time=%d, err=%v",
			in.OperatorCode, in.StartTime, in.EndTime, err)
		return &operator_game.QueryBetMembersResponse{
			Code:    500,
			Message: "query database failed",
		}, nil
	}

	// 去重并收集 member_ids
	memberIDMap := make(map[int64]bool)
	for _, order := range orders {
		memberIDMap[order.MemberID] = true
	}

	// 转换为切片
	memberIDs := make([]int64, 0, len(memberIDMap))
	for memberID := range memberIDMap {
		memberIDs = append(memberIDs, memberID)
	}

	l.Infof("QueryBetMembers: success, operator_code=%s, time_range=[%d, %d], member_count=%d, total_orders=%d",
		in.OperatorCode, in.StartTime, in.EndTime, len(memberIDs), len(orders))

	return &operator_game.QueryBetMembersResponse{
		Code:       0,
		Message:    "success",
		MemberIds:  memberIDs,
		TotalCount: int32(len(orders)),
	}, nil
}
