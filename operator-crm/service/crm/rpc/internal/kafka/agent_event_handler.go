package kafka

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/trace"
	"oa.98ent.com/p9/common/xkafka/payload"
	"oa.98ent.com/p9/common/xkafka/topic"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/constants"
)

func (h *EventHandler) PublishAgentApplicationApprovedEvent(
	ctx context.Context,
	memberID int64,
	operatorCode string,
	updatedAt *time.Time,
	auditStatus int64,
) {
	// 如果 Kafka 生产者未配置，则跳过事件发送
	if h.producer == nil {
		h.Errorf("[代理事件] Kafka 生产者未配置，跳过发送代理申请通过事件: member_id=%d", memberID)
		return
	}
	if auditStatus != constants.AuditStatusApproved {
		h.Infof("[代理事件] 代理申请未通过，跳过发送事件: member_id=%d, operator_code=%s", memberID, operatorCode)
		return
	}
	memberGlobalID, err := h.memberDao.GetMemberGlobalID(ctx, operatorCode, memberID)
	if err != nil {
		h.Errorf("[代理事件] 获取会员全局ID失败: member_id=%d, operator_code=%s, err=%v", memberID, operatorCode, err)
		return
	}

	// 构造事件对象
	event := payload.SingleMemberEvent{
		EventType:      topic.AgentApplicationApproved,
		OperatorCode:   operatorCode,
		MemberID:       memberID,
		MemberGlobalID: memberGlobalID,
		OccurredAt:     *updatedAt,
		PublishedAt:    time.Now(),
		TraceID:        trace.TraceIDFromContext(ctx),
	}

	// 发送事件到 Kafka
	err = h.producer.Publish(ctx, topic.AgentApplicationApproved, event)
	if err != nil {
		h.Errorf("[代理事件] 发送代理申请通过事件失败: member_id=%d, operator_code=%s, err=%v", memberID, operatorCode, err)
		return
	}

	h.Infof("[代理事件] 成功发送代理申请通过事件: member_id=%d, operator_code=%s, audit_status=%d", memberID, operatorCode, auditStatus)
}
