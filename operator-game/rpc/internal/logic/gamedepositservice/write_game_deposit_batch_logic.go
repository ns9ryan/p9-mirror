package gamedepositservicelogic

import (
	"context"

	"oa.98ent.com/p9/operator-game/rpc/internal/svc"
	"oa.98ent.com/p9/operator-game/rpc/pb/operator_game"

	"github.com/zeromicro/go-zero/core/logx"
)

type WriteGameDepositBatchLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewWriteGameDepositBatchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WriteGameDepositBatchLogic {
	return &WriteGameDepositBatchLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 批量写入充值记录
func (l *WriteGameDepositBatchLogic) WriteGameDepositBatch(in *operator_game.WriteGameDepositBatchRequest) (*operator_game.WriteGameDepositBatchResponse, error) {
	// 参数验证
	if len(in.Deposits) == 0 {
		l.Errorf("WriteGameDepositBatch: empty deposits list")
		return &operator_game.WriteGameDepositBatchResponse{
			Code:         400,
			Message:      "deposits list is empty",
			SuccessCount: 0,
			FailedCount:  0,
		}, nil
	}

	successCount := int32(0)
	failedCount := int32(0)
	failedDepositIds := make([]int64, 0)

	// 逐个创建充值记录
	for idx, depositReq := range in.Deposits {
		// 基础参数验证
		if depositReq.OperatorCode == "" {
			l.Errorf("WriteGameDepositBatch: deposit[%d] operator_code is empty", idx)
			failedCount++
			failedDepositIds = append(failedDepositIds, int64(idx))
			continue
		}

		if depositReq.MemberId <= 0 {
			l.Errorf("WriteGameDepositBatch: deposit[%d] member_id is invalid, member_id=%d", idx, depositReq.MemberId)
			failedCount++
			failedDepositIds = append(failedDepositIds, int64(idx))
			continue
		}

		if depositReq.DepositAt <= 0 {
			l.Errorf("WriteGameDepositBatch: deposit[%d] deposit_at is invalid", idx)
			failedCount++
			failedDepositIds = append(failedDepositIds, int64(idx))
			continue
		}

		// 创建充值记录
		deposit, err := l.svcCtx.DB.GameDeposit.Create().
			SetOperatorCode(depositReq.OperatorCode).
			SetMemberID(depositReq.MemberId).
			// SetDepositAt(depositReq.DepositAt).
			Save(l.ctx)

		if err != nil {
			l.Errorf("WriteGameDepositBatch: deposit[%d] save failed, operator_code=%s, member_id=%d, err=%v",
				idx, depositReq.OperatorCode, depositReq.MemberId, err)
			failedCount++
			failedDepositIds = append(failedDepositIds, int64(idx))
			continue
		}

		successCount++
		l.Debugf("WriteGameDepositBatch: deposit[%d] created, deposit_id=%d", idx, deposit.ID)
	}

	l.Infof("WriteGameDepositBatch: completed, total=%d, success=%d, failed=%d",
		len(in.Deposits), successCount, failedCount)

	return &operator_game.WriteGameDepositBatchResponse{
		Code:             0,
		Message:          "success",
		SuccessCount:     successCount,
		FailedCount:      failedCount,
		FailedDepositIds: failedDepositIds,
	}, nil
}
