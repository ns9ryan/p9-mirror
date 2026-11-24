package kafka

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/trace"
	"oa.98ent.com/p9/common/xkafka/payload"
	"oa.98ent.com/p9/common/xkafka/topic"
	"oa.98ent.com/p9/operator-crm/ent"
)

// PublishMemberRegisteredEvent 发送会员注册事件到 Kafka
func (h *EventHandler) PublishMemberRegisteredEvent(ctx context.Context, member *ent.Member) {
	// 如果 Kafka 生产者未配置，则跳过事件发送
	if h.producer == nil {
		h.Errorf("[会员事件] Kafka 生产者未配置，跳过发送会员注册事件: member_id=%d", member.ID)
		return
	}

	// 构造事件对象
	event := payload.SingleMemberEvent{
		MemberGlobalID: member.GlobalID.String(),      // 会员全局ID
		EventType:      topic.GameUserRegister,        // 事件类型：game-user.register
		OperatorCode:   member.OperatorCode,           // 分站标识
		MemberID:       member.ID,                     // 会员ID
		OccurredAt:     member.CreatedAt,              // 事件发生时间（会员创建时间）
		PublishedAt:    time.Now(),                    // 事件发布时间（当前时间）
		TraceID:        trace.TraceIDFromContext(ctx), // 链路追踪ID
	}

	// 发送事件到 Kafka
	err := h.producer.Publish(ctx, topic.GameUserRegister, event)
	if err != nil {
		h.Errorf("[会员事件] 发送会员注册事件失败: member_id=%d, err=%v", member.ID, err)
		// 此处不返回错误，避免影响会员创建流程
		return
	}

	h.Infof("[会员事件] 成功发送会员注册事件: member_id=%d, operator_code=%s", member.ID, member.OperatorCode)
}

func (h *EventHandler) PublishMemberProfileUpdatedEvent(
	ctx context.Context,
	beforeMemberProfile *ent.MemberProfile,
	afterMemberProfile *ent.MemberProfile,
) {
	// 如果 Kafka 生产者未配置，则跳过事件发送
	if h.producer == nil {
		h.Errorf("[会员事件] Kafka 生产者未配置，跳过发送会员资料更新事件: member_id=%d", beforeMemberProfile.MemberID)
		return
	}
	if beforeMemberProfile.Nickname != afterMemberProfile.Nickname ||
		beforeMemberProfile.AvatarID != afterMemberProfile.AvatarID {
		h.PublishMemberAvatarNicknameUpdatedEvent(
			ctx,
			beforeMemberProfile.MemberID,
			beforeMemberProfile.OperatorCode,
			afterMemberProfile.UpdatedAt,
			beforeMemberProfile.AvatarID,
			beforeMemberProfile.Nickname,
			afterMemberProfile.AvatarID,
			afterMemberProfile.Nickname,
		)
	}
	if beforeMemberProfile.Birthday != afterMemberProfile.Birthday {
		h.PublishMemberBirthdayEvent(
			ctx,
			beforeMemberProfile.MemberID,
			beforeMemberProfile.OperatorCode,
			afterMemberProfile.UpdatedAt,
			beforeMemberProfile.Birthday,
			afterMemberProfile.Birthday,
		)
	}
	if beforeMemberProfile.RealNameCiphertext != afterMemberProfile.RealNameCiphertext {
		h.PublishMemberRealnameVerifiedEvent(
			ctx,
			beforeMemberProfile.MemberID,
			beforeMemberProfile.OperatorCode,
			afterMemberProfile.UpdatedAt,
			beforeMemberProfile.RealNameCiphertext,
			afterMemberProfile.RealNameCiphertext,
		)
	}
	if beforeMemberProfile.IdentityNumberCiphertext != afterMemberProfile.IdentityNumberCiphertext {
		h.PublishMemberIDCPFVerifiedEvent(
			ctx,
			beforeMemberProfile.MemberID,
			beforeMemberProfile.OperatorCode,
			afterMemberProfile.UpdatedAt,
			beforeMemberProfile.IdentityNumberCiphertext,
			afterMemberProfile.IdentityNumberCiphertext,
		)
	}
}

func (h *EventHandler) PublishMemberAuthUpdatedEvent(
	ctx context.Context,
	memberID int64,
	operatorCode string,
	updatedAt time.Time,
	beforeMemberAuth *ent.MemberAuth,
	afterMemberAuth *ent.MemberAuth,
) {
	// 如果 Kafka 生产者未配置，则跳过事件发送
	if h.producer == nil {
		h.Errorf("[会员事件] Kafka 生产者未配置，跳过发送会员认证状态更新事件: member_id=%d", memberID)
		return
	}
	if beforeMemberAuth.EmailLookupHash != afterMemberAuth.EmailLookupHash {
		h.PublishMemberEmailVerifiedEvent(
			ctx,
			memberID,
			operatorCode,
			updatedAt,
			beforeMemberAuth.EmailLookupHash,
			afterMemberAuth.EmailLookupHash,
		)
	}
	if beforeMemberAuth.PhoneLookupHash != afterMemberAuth.PhoneLookupHash {
		h.PublishMemberMobileBoundEvent(
			ctx,
			memberID,
			operatorCode,
			updatedAt,
			beforeMemberAuth.PhoneLookupHash,
			afterMemberAuth.PhoneLookupHash,
		)
	}
}

func (h *EventHandler) PublishMemberAvatarNicknameUpdatedEvent(
	ctx context.Context,
	memberID int64,
	operatorCode string,
	updatedAt time.Time,
	beforeAvatarID int64,
	beforeNickname string,
	afterAvatarID int64,
	afterNickname string,
) {
	// 如果 Kafka 生产者未配置，则跳过事件发送
	if h.producer == nil {
		h.Errorf("[会员事件] Kafka 生产者未配置，跳过发送会员头像昵称更新事件: member_id=%d", memberID)
		return
	}
	if beforeAvatarID == afterAvatarID && beforeNickname == afterNickname {
		h.Infof("[会员事件] 会员头像昵称未发生变化，跳过发送事件: member_id=%d, operator_code=%s", memberID, operatorCode)
		return
	}
	memberGlobalID, err := h.memberDao.GetMemberGlobalID(ctx, operatorCode, memberID)
	if err != nil {
		h.Errorf("[代理事件] 获取会员全局ID失败: member_id=%d, operator_code=%s, err=%v", memberID, operatorCode, err)
		return
	}

	// 构造事件对象
	event := payload.SingleMemberEvent{
		MemberGlobalID: memberGlobalID, // 会员全局ID
		EventType:      topic.AvatarNicknameUpdated,
		OperatorCode:   operatorCode,
		MemberID:       memberID,
		OccurredAt:     updatedAt,
		PublishedAt:    time.Now(),
		TraceID:        trace.TraceIDFromContext(ctx), // 链路追踪ID
	}

	// 发送事件到 Kafka
	err = h.producer.Publish(ctx, topic.AvatarNicknameUpdated, event)
	if err != nil {
		h.Errorf("[会员事件] 发送会员头像昵称更新事件失败: member_id=%d, operator_code=%s, err=%v", memberID, operatorCode, err)
		return
	}

	h.Infof("[会员事件] 成功发送会员头像昵称更新事件: member_id=%d, operator_code=%s", memberID, operatorCode)
}

func (h *EventHandler) PublishMemberBirthdayEvent(
	ctx context.Context,
	memberID int64,
	operatorCode string,
	updatedAt time.Time,
	beforeBirthday *time.Time,
	afterBirthday *time.Time,
) {
	// 如果 Kafka 生产者未配置，则跳过事件发送
	if h.producer == nil {
		h.Errorf("[会员事件] Kafka 生产者未配置，跳过发送会员生日事件: member_id=%d", memberID)
		return
	}
	if (beforeBirthday == nil && afterBirthday == nil) || (beforeBirthday != nil && afterBirthday != nil && beforeBirthday.Equal(*afterBirthday)) {
		h.Infof("[会员事件] 会员生日未发生变化，跳过发送事件: member_id=%d, operator_code=%s", memberID, operatorCode)
		return
	}
	memberGlobalID, err := h.memberDao.GetMemberGlobalID(ctx, operatorCode, memberID)
	if err != nil {
		h.Errorf("[代理事件] 获取会员全局ID失败: member_id=%d, operator_code=%s, err=%v", memberID, operatorCode, err)
		return
	}

	// 构造事件对象
	event := payload.SingleMemberEvent{
		MemberGlobalID: memberGlobalID, // 会员全局ID
		EventType:      topic.BirthdateUpdated,
		OperatorCode:   operatorCode,
		MemberID:       memberID,
		OccurredAt:     updatedAt,
		PublishedAt:    time.Now(),
		TraceID:        trace.TraceIDFromContext(ctx), // 链路追踪ID
	}

	// 发送事件到 Kafka
	err = h.producer.Publish(ctx, topic.BirthdateUpdated, event)
	if err != nil {
		h.Errorf("[会员事件] 发送会员生日事件失败: member_id=%d, operator_code=%s, err=%v", memberID, operatorCode, err)
		return
	}

	h.Infof("[会员事件] 成功发送会员生日事件: member_id=%d, operator_code=%s", memberID, operatorCode)
}

func (h *EventHandler) PublishMemberEmailVerifiedEvent(
	ctx context.Context,
	memberID int64,
	operatorCode string,
	updatedAt time.Time,
	beforeEmail *string,
	afterEmail *string,
) {
	// 如果 Kafka 生产者未配置，则跳过事件发送
	if h.producer == nil {
		h.Errorf("[会员事件] Kafka 生产者未配置，跳过发送会员邮箱验证事件: member_id=%d", memberID)
		return
	}
	if beforeEmail != nil && afterEmail != nil && *beforeEmail == *afterEmail {
		h.Infof("[会员事件] 会员邮箱未发生变化，跳过发送事件: member_id=%d, operator_code=%s", memberID, operatorCode)
		return
	}
	if afterEmail == nil || *afterEmail == "" {
		h.Infof("[会员事件] 会员邮箱为空，跳过发送事件: member_id=%d, operator_code=%s", memberID, operatorCode)
		return
	}
	memberGlobalID, err := h.memberDao.GetMemberGlobalID(ctx, operatorCode, memberID)
	if err != nil {
		h.Errorf("[代理事件] 获取会员全局ID失败: member_id=%d, operator_code=%s, err=%v", memberID, operatorCode, err)
		return
	}

	// 构造事件对象
	event := payload.SingleMemberEvent{
		MemberGlobalID: memberGlobalID, // 会员全局ID
		EventType:      topic.EmailVerified,
		OperatorCode:   operatorCode,
		MemberID:       memberID,
		OccurredAt:     updatedAt,
		PublishedAt:    time.Now(),
		TraceID:        trace.TraceIDFromContext(ctx), // 链路追踪ID
	}

	// 发送事件到 Kafka
	err = h.producer.Publish(ctx, topic.EmailVerified, event)
	if err != nil {
		h.Errorf("[会员事件] 发送会员邮箱验证事件失败: member_id=%d, operator_code=%s, err=%v", memberID, operatorCode, err)
		return
	}

	h.Infof("[会员事件] 成功发送会员邮箱验证事件: member_id=%d, operator_code=%s", memberID, operatorCode)
}

func (h *EventHandler) PublishMemberMobileBoundEvent(
	ctx context.Context,
	memberID int64,
	operatorCode string,
	updatedAt time.Time,
	beforeMobile *string,
	afterMobile *string,
) {
	// 如果 Kafka 生产者未配置，则跳过事件发送
	if h.producer == nil {
		h.Errorf("[会员事件] Kafka 生产者未配置，跳过发送会员手机绑定事件: member_id=%d", memberID)
		return
	}
	if beforeMobile != nil && afterMobile != nil && *beforeMobile == *afterMobile {
		h.Infof("[会员事件] 会员手机未发生变化，跳过发送事件: member_id=%d, operator_code=%s", memberID, operatorCode)
		return
	}
	if afterMobile == nil || *afterMobile == "" {
		h.Infof("[会员事件] 会员手机为空，跳过发送事件: member_id=%d, operator_code=%s", memberID, operatorCode)
		return
	}

	memberGlobalID, err := h.memberDao.GetMemberGlobalID(ctx, operatorCode, memberID)
	if err != nil {
		h.Errorf("[代理事件] 获取会员全局ID失败: member_id=%d, operator_code=%s, err=%v", memberID, operatorCode, err)
		return
	}
	// 构造事件对象
	event := payload.SingleMemberEvent{
		MemberGlobalID: memberGlobalID, // 会员全局ID
		EventType:      topic.MobileBound,
		OperatorCode:   operatorCode,
		MemberID:       memberID,
		OccurredAt:     updatedAt,
		PublishedAt:    time.Now(),
		TraceID:        trace.TraceIDFromContext(ctx), // 链路追踪ID
	}

	// 发送事件到 Kafka
	err = h.producer.Publish(ctx, topic.MobileBound, event)
	if err != nil {
		h.Errorf("[会员事件] 发送会员手机绑定事件失败: member_id=%d, operator_code=%s, err=%v", memberID, operatorCode, err)
		return
	}

	h.Infof("[会员事件] 成功发送会员手机绑定事件: member_id=%d, operator_code=%s", memberID, operatorCode)
}

func (h *EventHandler) PublishMemberRealnameVerifiedEvent(
	ctx context.Context,
	memberID int64,
	operatorCode string,
	updatedAt time.Time,
	beforeRealname *string,
	afterRealname *string,
) {
	// 如果 Kafka 生产者未配置，则跳过事件发送
	if h.producer == nil {
		h.Errorf("[会员事件] Kafka 生产者未配置，跳过发送会员实名认证事件: member_id=%d", memberID)
		return
	}
	if beforeRealname != nil && afterRealname != nil && *beforeRealname == *afterRealname {
		h.Infof("[会员事件] 会员真实姓名未发生变化，跳过发送事件: member_id=%d, operator_code=%s", memberID, operatorCode)
		return
	}
	if afterRealname == nil || *afterRealname == "" {
		h.Infof("[会员事件] 会员真实姓名为空，跳过发送事件: member_id=%d, operator_code=%s", memberID, operatorCode)
		return
	}
	memberGlobalID, err := h.memberDao.GetMemberGlobalID(ctx, operatorCode, memberID)
	if err != nil {
		h.Errorf("[代理事件] 获取会员全局ID失败: member_id=%d, operator_code=%s, err=%v", memberID, operatorCode, err)
		return
	}

	// 构造事件对象
	event := payload.SingleMemberEvent{
		MemberGlobalID: memberGlobalID, // 会员全局ID
		EventType:      topic.RealnameVerified,
		OperatorCode:   operatorCode,
		MemberID:       memberID,
		OccurredAt:     updatedAt,
		PublishedAt:    time.Now(),
		TraceID:        trace.TraceIDFromContext(ctx), // 链路追踪ID
	}

	// 发送事件到 Kafka
	err = h.producer.Publish(ctx, topic.RealnameVerified, event)
	if err != nil {
		h.Errorf("[会员事件] 发送会员实名认证事件失败: member_id=%d, operator_code=%s, err=%v", memberID, operatorCode, err)
		return
	}

	h.Infof("[会员事件] 成功发送会员实名认证事件: member_id=%d, operator_code=%s", memberID, operatorCode)
}

func (h *EventHandler) PublishMemberIDCPFVerifiedEvent(
	ctx context.Context,
	memberID int64,
	operatorCode string,
	updatedAt time.Time,
	beforeIDCPF *string,
	afterIDCPF *string,
) {
	// 如果 Kafka 生产者未配置，则跳过事件发送
	if h.producer == nil {
		h.Errorf("[会员事件] Kafka 生产者未配置，跳过发送会员身份证/护照验证事件: member_id=%d", memberID)
		return
	}
	if beforeIDCPF != nil && afterIDCPF != nil && *beforeIDCPF == *afterIDCPF {
		h.Infof("[会员事件] 会员身份证/护照未发生变化，跳过发送事件: member_id=%d, operator_code=%s", memberID, operatorCode)
		return
	}
	if afterIDCPF == nil || *afterIDCPF == "" {
		h.Infof("[会员事件] 会员身份证/护照为空，跳过发送事件: member_id=%d, operator_code=%s", memberID, operatorCode)
		return
	}
	memberGlobalID, err := h.memberDao.GetMemberGlobalID(ctx, operatorCode, memberID)
	if err != nil {
		h.Errorf("[代理事件] 获取会员全局ID失败: member_id=%d, operator_code=%s, err=%v", memberID, operatorCode, err)
		return
	}

	// 构造事件对象
	event := payload.SingleMemberEvent{
		MemberGlobalID: memberGlobalID, // 会员全局ID
		EventType:      topic.IDCPFVerified,
		OperatorCode:   operatorCode,
		MemberID:       memberID,
		OccurredAt:     updatedAt,
		PublishedAt:    time.Now(),
		TraceID:        trace.TraceIDFromContext(ctx), // 链路追踪ID
	}

	// 发送事件到 Kafka
	err = h.producer.Publish(ctx, topic.IDCPFVerified, event)
	if err != nil {
		h.Errorf("[会员事件] 发送会员身份证/护照验证事件失败: member_id=%d, operator_code=%s, err=%v", memberID, operatorCode, err)
		return
	}

	h.Infof("[会员事件] 成功发送会员身份证/护照验证事件: member_id=%d, operator_code=%s", memberID, operatorCode)
}
