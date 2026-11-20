package gamewithdrawservicelogic

import (
	"context"

	"oa.98ent.com/p9/operator-game/rpc/internal/svc"
	"oa.98ent.com/p9/operator-game/rpc/pb/operator_game"

	"github.com/zeromicro/go-zero/core/logx"
)

type WriteGameWithdrawBatchLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewWriteGameWithdrawBatchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WriteGameWithdrawBatchLogic {
	return &WriteGameWithdrawBatchLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 批量写入提现记录
func (l *WriteGameWithdrawBatchLogic) WriteGameWithdrawBatch(in *operator_game.WriteGameWithdrawBatchRequest) (*operator_game.WriteGameWithdrawBatchResponse, error) {
	// 参数验证
	if len(in.Withdraws) == 0 {
		l.Errorf("WriteGameWithdrawBatch: empty withdraws list")
		return &operator_game.WriteGameWithdrawBatchResponse{
			Code:         400,
			Message:      "withdraws list is empty",
			SuccessCount: 0,
			FailedCount:  0,
		}, nil
	}

	successCount := int32(0)
	failedCount := int32(0)
	failedWithdrawIds := make([]int64, 0)

	// 逐个创建提现记录
	for idx, withdrawReq := range in.Withdraws {
		// 基础参数验证
		if withdrawReq.OperatorCode == "" {
			l.Errorf("WriteGameWithdrawBatch: withdraw[%d] operator_code is empty", idx)
			failedCount++
			failedWithdrawIds = append(failedWithdrawIds, int64(idx))
			continue
		}

		if withdrawReq.MemberId <= 0 {
			l.Errorf("WriteGameWithdrawBatch: withdraw[%d] member_id is invalid, member_id=%d", idx, withdrawReq.MemberId)
			failedCount++
			failedWithdrawIds = append(failedWithdrawIds, int64(idx))
			continue
		}

		if withdrawReq.WithdrawAt <= 0 {
			l.Errorf("WriteGameWithdrawBatch: withdraw[%d] withdraw_at is invalid", idx)
			failedCount++
			failedWithdrawIds = append(failedWithdrawIds, int64(idx))
			continue
		}

		// 创建提现记录
		withdraw, err := l.svcCtx.DB.GameWithdraw.Create().
			SetOperatorCode(withdrawReq.OperatorCode).
			SetMemberID(withdrawReq.MemberId).
			// SetWithdrawAt(withdrawReq.WithdrawAt).
			Save(l.ctx)

		if err != nil {
			l.Errorf("WriteGameWithdrawBatch: withdraw[%d] save failed, operator_code=%s, member_id=%d, err=%v",
				idx, withdrawReq.OperatorCode, withdrawReq.MemberId, err)
			failedCount++
			failedWithdrawIds = append(failedWithdrawIds, int64(idx))
			continue
		}

		successCount++
		l.Debugf("WriteGameWithdrawBatch: withdraw[%d] created, withdraw_id=%d", idx, withdraw.ID)
	}

	l.Infof("WriteGameWithdrawBatch: completed, total=%d, success=%d, failed=%d",
		len(in.Withdraws), successCount, failedCount)

	return &operator_game.WriteGameWithdrawBatchResponse{
		Code:              0,
		Message:           "success",
		SuccessCount:      successCount,
		FailedCount:       failedCount,
		FailedWithdrawIds: failedWithdrawIds,
	}, nil
}
