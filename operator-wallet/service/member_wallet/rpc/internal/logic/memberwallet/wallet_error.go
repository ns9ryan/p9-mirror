package memberwalletlogic

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/wallet"
)

// mapWalletError 将钱包运行时错误转换为统一 RPC 错误
func mapWalletError(err error) error {
	if err == nil {
		return nil
	}

	// 已经转换过的 RPC 错误直接保留
	if _, ok := status.FromError(err); ok {
		return err
	}

	switch {
	// 操作已经进入 Mailbox, 但调用方等待被中断
	case errors.Is(err, wallet.ErrOperationWaitInterrupted):
		if errors.Is(err, context.DeadlineExceeded) {
			return status.Error(codes.DeadlineExceeded, "等待钱包操作结果超时，操作已被接收，可能仍在处理中，请核查操作结果")
		}

		return status.Error(codes.Canceled, "钱包操作等待已取消，操作已被接收，可能仍在处理中，请核查操作结果")

	// 操作尚未进入 Mailbox, 请求已取消
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, "钱包操作在进入处理队列前已取消")

	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, "钱包操作在进入处理队列前超时")

	// 记账参数与余额校验
	case errors.Is(err, wallet.ErrInvalidRecordParams):
		return status.Error(codes.InvalidArgument, "钱包记账参数无效")

	case errors.Is(err, wallet.ErrInsufficientBalance):
		return status.Error(codes.FailedPrecondition, "钱包余额不足")

	case errors.Is(err, wallet.ErrBalanceOverflow):
		return status.Error(codes.OutOfRange, "钱包余额超出允许范围")

	case errors.Is(err, wallet.ErrWalletVersionOverflow):
		return status.Error(codes.OutOfRange, "钱包版本超出允许范围")

	// Actor 故障
	case errors.Is(err, wallet.ErrWalletFailed):
		return status.Error(codes.FailedPrecondition, "钱包处于故障状态，暂不可操作")

	// Engine、Actor 和 Mailbox 生命周期
	case errors.Is(err, wallet.ErrEngineClosed),
		errors.Is(err, wallet.ErrActorClosed),
		errors.Is(err, wallet.ErrMailboxClosed),
		errors.Is(err, wallet.ErrWorkerPoolClosed):
		return status.Error(codes.Unavailable, "钱包服务暂不可用")

	case errors.Is(err, wallet.ErrMailboxFull):
		return status.Error(codes.ResourceExhausted, "钱包请求繁忙，请稍后重试")

	// Kafka 生产者
	case errors.Is(err, wallet.ErrKafkaProducerUnavailable):
		return status.Error(codes.Unavailable, "钱包资金事件服务暂不可用")

	case errors.Is(err, wallet.ErrKafkaWriteUncertain):
		return status.Error(codes.Unknown, "钱包操作结果不明确，钱包已进入故障状态，请勿重复提交，需人工核查")

	// 内部数据或事件异常
	case errors.Is(err, wallet.ErrWalletActorNotFound):
		return status.Error(codes.NotFound, "钱包 Actor 不存在")

	case errors.Is(err, wallet.ErrInvalidWalletState),
		errors.Is(err, wallet.ErrKafkaInvalidRecord):
		return status.Error(codes.Internal, "钱包内部状态异常")

	default:
		return status.Error(codes.Internal, "钱包操作失败")
	}
}

// handleWalletSubmitError 统一处理钱包资金操作提交错误
func handleWalletSubmitError(err error, operationName string, operationID uuid.UUID, logger logx.Logger) error {
	if err == nil {
		return nil
	}

	mappedErr := mapWalletError(err)

	switch {
	// 操作已进入 Actor, 但调用方没有收到处理结果
	case errors.Is(err, wallet.ErrOperationWaitInterrupted):
		logger.Infof("%s result not received, operation_id=%s: %v", operationName, operationID, err)

	// 未知错误或内部异常, 需要保留日志供排查
	case status.Code(mappedErr) == codes.Unknown, status.Code(mappedErr) == codes.Internal:
		logger.Errorf("%s failed, operation_id=%s: %v", operationName, operationID, err)
	}

	return mappedErr
}
