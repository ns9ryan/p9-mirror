package kafka

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/trace"
	"oa.98ent.com/p9/common/xkafka/payload"
	"oa.98ent.com/p9/common/xkafka/topic"
	"oa.98ent.com/p9/operator-activity/pb/operatoractivityrpc/taskpb"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentdepositorderpb"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"
)

// 充值时调用
func (h *EventHandler) PublishReferralRegisteredDepositEvent(
	ctx context.Context,
	memberID int64,
	operatorCode string,
	updatedAt time.Time,
	amount int64,
) {
	// 如果 Kafka 生产者未配置，则跳过事件发送
	if h.producer == nil {
		h.Errorf("[代理事件] Kafka 生产者未配置，跳过发送推荐注册存款事件: member_id=%d", memberID)
		return
	}
	// 获取任务配置
	taskConfigResp, err := h.taskRpc.GetTaskConfig(ctx, &taskpb.GetTaskConfigReq{
		OperatorCode: operatorCode,
		TaskCode:     "referral_registered_deposit",
	})
	if err != nil {
		h.Errorf("[代理事件] 获取任务配置失败: operator_code=%s, err=%v", operatorCode, err)
		return
	}
	if taskConfigResp == nil || taskConfigResp.TaskConfig == nil {
		h.Errorf("[代理事件] 任务配置不存在: operator_code=%s", operatorCode)
		return
	}
	taskConfig := taskConfigResp.TaskConfig
	if taskConfig.Status != 1 {
		h.Infof("[代理事件] 任务未启用，跳过发送事件: operator_code=%s, task_code=%s", operatorCode, taskConfig.TaskCode)
		return
	}
	// 查询会员信息
	memberData, err := h.memberRpc.GetDetail(ctx, &memberpb.GetMemberDetailRequest{
		MemberId: memberID,
	})
	if err != nil {
		h.Errorf("[代理事件] 查询会员信息失败: member_id=%d, err=%v", memberID, err)
		return
	}
	if memberData == nil {
		h.Errorf("[代理事件] 会员信息不存在: member_id=%d", memberID)
		return
	}

	// 获取任务时区
	taskTimezone := taskConfig.Timezone
	// 根据时区判断会员创建时间是不是在本周
	memberCreatedAt := time.UnixMilli(memberData.CreatedAt)
	// 根据任务时区转换会员创建时间
	loc, err := time.LoadLocation(taskTimezone)
	if err != nil {
		h.Errorf("[代理事件] 加载任务时区失败: operator_code=%s, task_code=%s, err=%v", operatorCode, taskConfig.TaskCode, err)
		return
	}
	memberCreatedAtInTaskTimezone := memberCreatedAt.In(loc)
	// 判断会员创建时间是否在本周
	year, currentWeek := time.Now().In(loc).ISOWeek()
	memberYear, memberWeek := memberCreatedAtInTaskTimezone.ISOWeek()
	if year != memberYear || currentWeek != memberWeek {
		h.Infof("[代理事件] 会员创建时间不在本周，跳过发送事件: member_id=%d, operator_code=%s, current_week=%d, register_week=%d", memberID, operatorCode, currentWeek, memberWeek)
		return
	}
	// 判断金额是否符合条件
	if amount < int64(taskConfig.CompletionAmount) {
		h.Infof("[代理事件] 存款金额不符合任务要求，跳过发送事件: member_id=%d, operator_code=%s, amount=%d, required_amount=%d", memberID, operatorCode, amount, taskConfig.CompletionAmount)
		return
	}
	// 判断会员今日累计充值是否符合条件
	todayDepositResp, err := h.agentDepositorderRpc.GetMemberTodayDeposit(ctx, &agentdepositorderpb.GetMemberTodayDepositReq{
		MemberId: memberID,
	})
	if err != nil {
		h.Errorf("[代理事件] 查询会员今日累计充值失败: member_id=%d, err=%v", memberID, err)
		return
	}
	todayDeposit := todayDepositResp.TodayDepositAmount
	// 如果今日累计充值金额不符合条件，跳过发送事件
	if todayDeposit < int64(taskConfig.CompletionAmount) {
		h.Infof("[代理事件] 会员今日累计充值金额不符合任务要求，跳过发送事件: member_id=%d, operator_code=%s, today_deposit=%d, required_amount=%d", memberID, operatorCode, todayDeposit, taskConfig.CompletionAmount)
		return
	}

	// 查询上级代理
	agentData, err := h.memberRpc.GetParentAgentMember(ctx, &memberpb.GetParentAgentMemberRequest{
		MemberId:     memberID,
		OperatorCode: operatorCode,
	})
	if err != nil {
		h.Errorf("[代理事件] 查询上级代理失败: member_id=%d, err=%v", memberID, err)
		return
	}
	if agentData == nil || agentData.ParentAgentId == 0 {
		h.Errorf("[代理事件] 上级代理不存在: member_id=%d", memberID)
		return
	}

	// 获取上级代理的会员ID
	agentMemberID := agentData.ParentAgentMemberId

	// 构造事件对象
	event := payload.SingleMemberEvent{
		MemberGlobalID: agentData.ParentAgentMemberGlobalId,
		EventType:      topic.ReferralRegisteredDeposit,
		OperatorCode:   operatorCode,
		MemberID:       agentMemberID,
		OccurredAt:     updatedAt,
		PublishedAt:    time.Now(),
		TraceID:        trace.TraceIDFromContext(ctx),
	}

	// 发送事件到 Kafka
	err = h.producer.Publish(ctx, topic.ReferralRegisteredDeposit, event)
	if err != nil {
		h.Errorf("[代理事件] 发送推荐注册存款事件失败: member_id=%d, err=%v", memberID, err)
		return
	}

	h.Infof("[代理事件] 成功发送推荐注册存款事件: member_id=%d, operator_code=%s", memberID, operatorCode)
}
