package gameorderservicelogic

import (
	"context"

	"oa.98ent.com/p9/operator-game/rpc/internal/svc"
	"oa.98ent.com/p9/operator-game/rpc/pb/operator_game"

	"github.com/zeromicro/go-zero/core/logx"
)

type WriteGameOrderBatchLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewWriteGameOrderBatchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WriteGameOrderBatchLogic {
	return &WriteGameOrderBatchLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 批量写入游戏订单
func (l *WriteGameOrderBatchLogic) WriteGameOrderBatch(in *operator_game.WriteGameOrderBatchRequest) (*operator_game.WriteGameOrderBatchResponse, error) {
	// 参数验证
	if len(in.Orders) == 0 {
		l.Errorf("WriteGameOrderBatch: empty orders list")
		return &operator_game.WriteGameOrderBatchResponse{
			Code:         400,
			Message:      "orders list is empty",
			SuccessCount: 0,
			FailedCount:  0,
		}, nil
	}

	successCount := int32(0)
	failedCount := int32(0)
	failedOrderIds := make([]int64, 0)

	// 逐个创建订单
	for idx, orderReq := range in.Orders {
		// 基础参数验证
		if orderReq.OperatorCode == "" {
			l.Errorf("WriteGameOrderBatch: order[%d] operator_code is empty", idx)
			failedCount++
			failedOrderIds = append(failedOrderIds, int64(idx))
			continue
		}

		if orderReq.MemberId <= 0 {
			l.Errorf("WriteGameOrderBatch: order[%d] member_id is invalid, member_id=%d", idx, orderReq.MemberId)
			failedCount++
			failedOrderIds = append(failedOrderIds, int64(idx))
			continue
		}

		if orderReq.SettleAt <= 0 {
			l.Errorf("WriteGameOrderBatch: order[%d] settle_at is invalid", idx)
			failedCount++
			failedOrderIds = append(failedOrderIds, int64(idx))
			continue
		}

		// 创建订单
		order, err := l.svcCtx.DB.GameOrder.Create().
			SetOperatorCode(orderReq.OperatorCode).
			SetMemberID(orderReq.MemberId).
			SetSettleAt(orderReq.SettleAt).
			Save(l.ctx)

		if err != nil {
			l.Errorf("WriteGameOrderBatch: order[%d] save failed, operator_code=%s, member_id=%d, err=%v",
				idx, orderReq.OperatorCode, orderReq.MemberId, err)
			failedCount++
			failedOrderIds = append(failedOrderIds, int64(idx))
			continue
		}

		successCount++
		l.Debugf("WriteGameOrderBatch: order[%d] created, order_id=%d", idx, order.ID)
	}

	l.Infof("WriteGameOrderBatch: completed, total=%d, success=%d, failed=%d",
		len(in.Orders), successCount, failedCount)

	return &operator_game.WriteGameOrderBatchResponse{
		Code:           0,
		Message:        "success",
		SuccessCount:   successCount,
		FailedCount:    failedCount,
		FailedOrderIds: failedOrderIds,
	}, nil
}
