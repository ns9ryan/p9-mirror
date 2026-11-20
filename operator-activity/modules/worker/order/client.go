package order

import (
	"context"
	"fmt"
	"log"
	"time"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/operator-activity/constants"
	"oa.98ent.com/p9/operator-activity/modules/activity/pipeline/payout"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"
	"oa.98ent.com/p9/operator-activity/modules/temporal"

	temporalSDK "go.temporal.io/sdk/temporal"
)

// Starter 实现 payout.WorkflowStarter，供活动引擎启动领取/派发工作流
type Starter struct{}

func NewStarter() *Starter { return &Starter{} }

var _ payout.WorkflowStarter = (*Starter)(nil)

func (s *Starter) StartOrderLifecycle(ctx context.Context, order *types.ActivityOrder) error {
	return StartCreateOrderWf(ctxdata.WithOperatorCode(ctx, order.OperatorCode), ToCreateOrderReq(order))
}

// ToCreateOrderReq 领域订单 → 工作流入参
func ToCreateOrderReq(order *types.ActivityOrder) CreateOrderReq {
	return CreateOrderReq{
		ID:           order.ID,
		OrderNo:      order.OrderNo,
		OperatorCode: order.OperatorCode,
		ActivityID:   order.ActivityID,
		MemberID:     order.MemberID,
		RewardsTotal: order.RewardsTotal,
		ClaimExp:     order.ClaimExp,
		ApplyAt:      order.AppliedAt,
		PayoutAt:     order.PayoutAt,
		PayoutMode:   types.PayoutMode(order.PayoutMode),
		PayoutTiming: types.PayoutTiming(order.PayoutTiming),
		Status:       types.OrderStatus(order.Status),
		CreatedAt:    order.CreatedAt,
		Remark:       order.Remark,
	}
}

// StartCreateOrderWf 创建订单工作流
func StartCreateOrderWf(ctx context.Context, order CreateOrderReq) error {
	workflowID := getCreateOrderWorkflowID(order.OperatorCode, order.OrderNo)

	options := temporal.WorkflowOptions{
		WorkflowID:               workflowID,
		TaskQueue:                constants.WorkflowQueueName,
		WorkflowExecutionTimeout: time.Hour * 24 * 120,
		WorkflowRunTimeout:       time.Hour * 24 * 120,
		WorkflowTaskTimeout:      time.Minute * 3,
		RetryPolicy: &temporalSDK.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute * 10,
			MaximumAttempts:    3,
		},
		SearchAttributes: map[string]interface{}{},
	}

	we, err := temporal.StartWorkflow(ctx, StartActivityOrderWorkflow, options, order)
	if err != nil {
		log.Printf("无法启动工作流: %v", err)
		return fmt.Errorf("start create order workflow: %w", err)
	}

	log.Printf("活动订单工作流已启动: WorkflowID=%s, RunID=%s", we.GetID(), we.GetRunID())
	return nil
}

// UserClaimSignal 用户领取信号
func UserClaimSignal(opCode string, orderNo string) error {
	workflowID := getCreateOrderWorkflowID(opCode, orderNo)
	c, err := temporal.GetClient()
	if err != nil {
		return fmt.Errorf("无法获取 Temporal 客户端: %w", err)
	}

	err = c.SignalWorkflow(context.Background(), workflowID, "", constants.WorkflowUserClaimSignal, orderNo)
	if err != nil {
		return fmt.Errorf("发送用户领取信号失败: %w", err)
	}

	log.Printf("用户领取信号发送成功: workflowID=%s", workflowID)
	return nil
}

// UserAuditSignalReq 用户审核信号请求
type UserAuditSignalReq struct {
	OrderNo string
	Result  string
	Reason  string
}

// UserAuditSignal 用户审核信号
func UserAuditSignal(opCode, orderNo, result, reason string) error {
	workflowID := getCreateOrderWorkflowID(opCode, orderNo)
	c, err := temporal.GetClient()
	if err != nil {
		return fmt.Errorf("无法获取 Temporal 客户端: %w", err)
	}

	req := UserAuditSignalReq{
		OrderNo: orderNo,
		Result:  result,
		Reason:  reason,
	}
	err = c.SignalWorkflow(context.Background(), workflowID, "", constants.WorkflowUserAuditSignal, req)
	if err != nil {
		return fmt.Errorf("发送用户审核信号失败: %w", err)
	}

	log.Printf("用户审核信号发送成功: workflowID=%s", workflowID)
	return nil
}
