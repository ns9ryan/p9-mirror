// Package kafka 实现 CDC 事件工厂，将 Kafka 消息转换为业务事件
package kafka

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/operator-activity/task/constants"
	"oa.98ent.com/p9/operator-activity/task/types"
)

// BuildEventsFromCDC 工厂函数：将 CDC 消息转换为业务事件
// 入参：msgValue - CDC 事件的 JSON 字符串
// 返回：生成的事件数组，一条消息可能生成多个事件
func BuildEventsFromCDC(kafkaID string, opcode string, table string, msgValue string) ([]types.Event, error) {
	// 第一步：解析 CDCEnvelope（包含 before/after 结构）
	var envelope types.CDCEnvelope
	if err := json.Unmarshal([]byte(msgValue), &envelope); err != nil {
		logx.Errorf("解析 CDC 信封失败: kafka_id=%s, err=%v", kafkaID, err)
		return nil, fmt.Errorf("解析 CDC 信封失败: %w", err)
	}

	events := make([]types.Event, 0)

	// 从 after 数据中提取表名（判断操作类型）
	// delete: before 有值，after 为 null；insert: before 为 null，after 有值；update: both 有值
	var dataSource json.RawMessage
	// table := ""

	if len(envelope.After) > 0 && envelope.After[0] != 'n' { // after 不是 null
		dataSource = envelope.After
	} else if len(envelope.Before) > 0 && envelope.Before[0] != 'n' { // before 不是 null
		dataSource = envelope.Before
	} else {
		logx.Infof("before 和 after 都为空，跳过: kafka_id=%s, opcode=%s, op=%s", kafkaID, opcode, envelope.Op)
		return events, nil
	}

	// 第二步：根据表名选择处理逻辑
	switch table {
	case constants.TableMember:
		evts, err := handleMemberTable(kafkaID, &envelope, dataSource)
		if err != nil {
			logx.Errorf("处理 member 表失败: kafka_id=%s, err=%v", kafkaID, err)
			return nil, err
		}
		events = append(events, evts...)

	case constants.TableMemberProfile:
		evts, err := handleMemberProfileTable(kafkaID, &envelope, envelope.Before, envelope.After)
		if err != nil {
			logx.Errorf("处理 member_profile 表失败: kafka_id=%s, err=%v", kafkaID, err)
			return nil, err
		}
		events = append(events, evts...)

	case constants.TableMemberAuth:
		evts, err := handleMemberAuthTable(kafkaID, &envelope, envelope.Before, envelope.After)
		if err != nil {
			logx.Errorf("处理 member_auth 表失败: kafka_id=%s, err=%v", kafkaID, err)
			return nil, err
		}
		events = append(events, evts...)

	case constants.TableMemberOAuthIdentity:
		evts, err := handleMemberOAuthIdentityTable(kafkaID, &envelope, envelope.Before, envelope.After)
		if err != nil {
			logx.Errorf("处理 member_oauth_identity 表失败: kafka_id=%s, err=%v", kafkaID, err)
			return nil, err
		}
		events = append(events, evts...)

	case constants.TableGameWithdraw:
		evts, err := handleGameWithdrawTable(kafkaID, &envelope, envelope.Before, envelope.After)
		if err != nil {
			logx.Errorf("处理 game_withdraw 表失败: kafka_id=%s, err=%v", kafkaID, err)
			return nil, err
		}
		events = append(events, evts...)

	case constants.TableGameDeposit:
		evts, err := handleGameDepositTable(kafkaID, &envelope, envelope.Before, envelope.After)
		if err != nil {
			logx.Errorf("处理 game_deposit 表失败: kafka_id=%s, err=%v", kafkaID, err)
			return nil, err
		}
		events = append(events, evts...)

	case constants.TableAgentApply:
		evts, err := handleAgentApplyTable(kafkaID, &envelope, envelope.Before, envelope.After)
		if err != nil {
			logx.Errorf("处理 agent_apply 表失败: kafka_id=%s, err=%v", kafkaID, err)
			return nil, err
		}
		events = append(events, evts...)

	case constants.TableGameChatLog:
		evts, err := handleGameChatLogTable(kafkaID, &envelope, envelope.Before, envelope.After)
		if err != nil {
			logx.Errorf("处理 game_chat_log 表失败: kafka_id=%s, err=%v", kafkaID, err)
			return nil, err
		}
		events = append(events, evts...)

	default:
		logx.Infof("不支持的表: kafka_id=%s, table=%s", kafkaID, table)
		return events, nil
	}
	// 过滤opcode
	filteredEvents := make([]types.Event, 0)
	for _, evt := range events {
		if evt.OperatorCode == opcode {
			filteredEvents = append(filteredEvents, evt)
		} else {
			logx.Infof("过滤掉不匹配的操作码事件: kafka_id=%s, expected_opcode=%s, actual_opcode=%s", kafkaID, opcode, evt.OperatorCode)
		}
	}
	events = filteredEvents

	return events, nil
}

// handleMemberTable 处理 member 表事件
// 只有 insert 操作（op='c'）才生成 EventTypeMemberRegistered 事件
func handleMemberTable(kafkaID string, envelope *types.CDCEnvelope, dataSource json.RawMessage) ([]types.Event, error) {
	events := make([]types.Event, 0)

	// 只处理 insert 操作（op='c'）
	if envelope.Op != "c" {
		return events, nil
	}

	var payload types.MemberPayload
	if err := json.Unmarshal(dataSource, &payload); err != nil {
		logx.Errorf("解析 member 数据失败: kafka_id=%s, err=%v", kafkaID, err)
		return events, fmt.Errorf("解析 member 数据失败: %w", err)
	}

	event := createEventTemplate(kafkaID, envelope, &payload, constants.EventTypeMemberRegistered)
	events = append(events, event)

	logx.Infof("生成会员注册事件: member_id=%d", payload.ID)
	return events, nil
}

// ========== member_profile 表处理 ==========

// handleMemberProfileTable 处理 member_profile 表事件
// 监听多个字段的变更，每个变更生成一个事件
func handleMemberProfileTable(kafkaID string, envelope *types.CDCEnvelope, before, after json.RawMessage) ([]types.Event, error) {
	events := make([]types.Event, 0)

	// 不处理 delete 操作
	if envelope.Op == "d" {
		return events, nil
	}

	// 解析 before 和 after
	var beforePayload, afterPayload types.MemberProfilePayload
	var beforeExists, afterExists bool

	if len(before) > 0 && before[0] != 'n' {
		if err := json.Unmarshal(before, &beforePayload); err != nil {
			logx.Errorf("解析 before 数据失败: kafka_id=%s, err=%v", kafkaID, err)
		} else {
			beforeExists = true
		}
	}

	if len(after) > 0 && after[0] != 'n' {
		if err := json.Unmarshal(after, &afterPayload); err != nil {
			logx.Errorf("解析 after 数据失败: kafka_id=%s, err=%v", kafkaID, err)
			return events, fmt.Errorf("解析 after 数据失败: %w", err)
		}
		afterExists = true
	}

	if !afterExists {
		return events, nil
	}

	// 1. 检查 avatar_id 或 nickname 变更（op='u' 且新值不为空）
	if envelope.Op == "u" {
		avatarChanged := beforeExists && (beforePayload.AvatarID != afterPayload.AvatarID)
		nicknameChanged := beforeExists && (beforePayload.Nickname != afterPayload.Nickname)

		if (avatarChanged || nicknameChanged) && (afterPayload.AvatarID > 0 || afterPayload.Nickname != "") {
			event := createEventTemplate(kafkaID, envelope, &afterPayload, constants.EventTypeAvatarNicknameUpdated)
			events = append(events, event)
			logx.Infof("生成头像昵称更新事件: member_id=%d", afterPayload.MemberID)
		}
	} else if envelope.Op == "c" {
		// // 初始化时会默认插入系统头像和昵称，不需要生成事件，只有 avatar_id 或 nickname 不为空时才生成事件
		// // insert 时，只要 avatar_id 或 nickname 不为空就生成事件
		// if afterPayload.AvatarID > 0 || afterPayload.Nickname != "" {
		// 	event := createEventTemplate(envelope, &afterPayload, constants.EventTypeAvatarNicknameUpdated)
		// 	events = append(events, event)
		// 	logx.Infof("插入时生成头像昵称更新事件: member_id=%d", afterPayload.MemberID)
		// }
	}

	// 2. 检查 real_name_ciphertext 不为空
	if afterPayload.RealNameCiphertext != "" {
		// 如果是 update，需要检查是否真的变更了
		if envelope.Op == "u" {
			if !beforeExists || beforePayload.RealNameCiphertext != afterPayload.RealNameCiphertext {
				event := createEventTemplate(kafkaID, envelope, &afterPayload, constants.EventTypeRealnameVerified)
				events = append(events, event)
				logx.Infof("生成实名认证事件: member_id=%d", afterPayload.MemberID)
			}
		} else if envelope.Op == "c" {
			// insert 时，只要 real_name_ciphertext 不为空就生成事件
			event := createEventTemplate(kafkaID, envelope, &afterPayload, constants.EventTypeRealnameVerified)
			events = append(events, event)
			logx.Infof("插入时生成实名认证事件: member_id=%d", afterPayload.MemberID)
		}
	}

	// 3. 检查 identity_number_ciphertext 不为空
	if afterPayload.IdentityNumberCiphertext != "" {
		// 如果是 update，需要检查是否真的变更了
		if envelope.Op == "u" {
			if !beforeExists || beforePayload.IdentityNumberCiphertext != afterPayload.IdentityNumberCiphertext {
				event := createEventTemplate(kafkaID, envelope, &afterPayload, constants.EventTypeIDCPFVerified)
				events = append(events, event)
				logx.Infof("生成身份证认证事件: member_id=%d", afterPayload.MemberID)
			}
		} else if envelope.Op == "c" {
			// insert 时，只要 identity_number_ciphertext 不为空就生成事件
			event := createEventTemplate(kafkaID, envelope, &afterPayload, constants.EventTypeIDCPFVerified)
			events = append(events, event)
			logx.Infof("插入时生成身份证认证事件: member_id=%d", afterPayload.MemberID)
		}
	}

	// 4. 检查 birthday 变更（op='u' 且值不为 0，或 insert 时值不为 0）
	if afterPayload.Birthday != 0 {
		if envelope.Op == "u" {
			if !beforeExists || beforePayload.Birthday != afterPayload.Birthday {
				event := createEventTemplate(kafkaID, envelope, &afterPayload, constants.EventTypeBirthdateUpdated)
				events = append(events, event)
				logx.Infof("生成生日更新事件: member_id=%d, birthday=%d", afterPayload.MemberID, afterPayload.Birthday)
			}
		} else if envelope.Op == "c" {
			// insert 时，只要 birthday 不为 0 就生成事件
			event := createEventTemplate(kafkaID, envelope, &afterPayload, constants.EventTypeBirthdateUpdated)
			events = append(events, event)
			logx.Infof("插入时生成生日更新事件: member_id=%d, birthday=%d", afterPayload.MemberID, afterPayload.Birthday)
		}
	}

	return events, nil
}

// ========== member_auth 表处理 ==========

// handleMemberAuthTable 处理 member_auth 表事件
// 监听 phone_ciphertext 和 email_ciphertext 字段，新值不为空就生成事件
func handleMemberAuthTable(kafkaID string, envelope *types.CDCEnvelope, before, after json.RawMessage) ([]types.Event, error) {
	events := make([]types.Event, 0)

	// 不处理 delete 操作
	if envelope.Op == "d" {
		return events, nil
	}

	// 解析 before 和 after
	var beforePayload, afterPayload types.MemberAuthPayload
	var beforeExists bool

	if len(before) > 0 && before[0] != 'n' {
		if err := json.Unmarshal(before, &beforePayload); err != nil {
			logx.Errorf("解析 before 数据失败: %v", err)
		} else {
			beforeExists = true
		}
	}

	if len(after) > 0 && after[0] != 'n' {
		if err := json.Unmarshal(after, &afterPayload); err != nil {
			logx.Errorf("解析 after 数据失败: %v", err)
			return events, fmt.Errorf("解析 after 数据失败: %w", err)
		}
	}

	// 1. 检查 phone_ciphertext 不为空
	if afterPayload.PhoneCiphertext != "" {
		// 如果是 update，需要检查是否真的变更了
		if envelope.Op == "u" {
			if !beforeExists || beforePayload.PhoneCiphertext != afterPayload.PhoneCiphertext {
				event := createEventTemplate(kafkaID, envelope, &afterPayload, constants.EventTypeMobileBound)
				events = append(events, event)
				logx.Infof("生成手机绑定事件: member_id=%d", afterPayload.MemberID)
			}
		} else if envelope.Op == "c" {
			// insert 时，只要 phone_ciphertext 不为空就生成事件
			event := createEventTemplate(kafkaID, envelope, &afterPayload, constants.EventTypeMobileBound)
			events = append(events, event)
			logx.Infof("插入时生成手机绑定事件: member_id=%d", afterPayload.MemberID)
		}
	}

	// 2. 检查 email_ciphertext 不为空
	if afterPayload.EmailCiphertext != "" {
		// 如果是 update，需要检查是否真的变更了
		if envelope.Op == "u" {
			if !beforeExists || beforePayload.EmailCiphertext != afterPayload.EmailCiphertext {
				event := createEventTemplate(kafkaID, envelope, &afterPayload, constants.EventTypeEmailVerified)
				events = append(events, event)
				logx.Infof("生成邮箱验证事件: member_id=%d", afterPayload.MemberID)
			}
		} else if envelope.Op == "c" {
			// insert 时，只要 email_ciphertext 不为空就生成事件
			event := createEventTemplate(kafkaID, envelope, &afterPayload, constants.EventTypeEmailVerified)
			events = append(events, event)
			logx.Infof("插入时生成邮箱验证事件: member_id=%d", afterPayload.MemberID)
		}
	}

	return events, nil
}

// ========== member_oauth_identity 表处理 ==========

// handleMemberOAuthIdentityTable 处理 member_oauth_identity 表事件
// 监听 provider 字段变更，所有 provider 都生成 EventTypeSocialMediaUpdated 事件
func handleMemberOAuthIdentityTable(kafkaID string, envelope *types.CDCEnvelope, before, after json.RawMessage) ([]types.Event, error) {
	events := make([]types.Event, 0)

	// 不处理 delete 操作
	if envelope.Op == "d" {
		return events, nil
	}

	// 解析 before 和 after
	var beforePayload, afterPayload types.MemberOAuthIdentityPayload
	var beforeExists bool

	if len(before) > 0 && before[0] != 'n' {
		if err := json.Unmarshal(before, &beforePayload); err != nil {
			logx.Errorf("Failed to unmarshal before payload: %v", err)
		} else {
			beforeExists = true
		}
	}

	if len(after) > 0 && after[0] != 'n' {
		if err := json.Unmarshal(after, &afterPayload); err != nil {
			logx.Errorf("Failed to unmarshal after payload: %v", err)
			return events, fmt.Errorf("failed to unmarshal after payload: %w", err)
		}
	}

	// 检查 provider 不为空（任何 provider 变更都生成事件）
	if afterPayload.Provider != "" {
		// 如果是 update，需要检查是否真的变更了
		if envelope.Op == "u" {
			if !beforeExists || beforePayload.Provider != afterPayload.Provider {
				event := createEventTemplate(kafkaID, envelope, &afterPayload, constants.EventTypeSocialMediaUpdated)
				events = append(events, event)
				logx.Infof("生成社交媒体更新事件: member_id=%d, provider=%s", afterPayload.MemberID, afterPayload.Provider)
			}
		} else if envelope.Op == "c" {
			// insert 时，只要 provider 不为空就生成事件
			event := createEventTemplate(kafkaID, envelope, &afterPayload, constants.EventTypeSocialMediaUpdated)
			events = append(events, event)
			logx.Infof("插入时生成社交媒体更新事件: member_id=%d, provider=%s", afterPayload.MemberID, afterPayload.Provider)
		}
	}

	return events, nil
}

// ========== game_withdraw 表处理 ==========

// handleGameWithdrawTable 处理 game_withdraw 表事件
// 只有 insert 操作（op='c'）才生成 EventTypeWithdrawalSucceeded 事件
func handleGameWithdrawTable(kafkaID string, envelope *types.CDCEnvelope, before, after json.RawMessage) ([]types.Event, error) {
	events := make([]types.Event, 0)

	// 只处理 insert 操作（op='c'）
	if envelope.Op != "c" {
		return events, nil
	}

	var payload types.GameWithdrawPayload
	if err := json.Unmarshal(after, &payload); err != nil {
		logx.Errorf("解析 game_withdraw 数据失败: kafka_id=%s, err=%v", kafkaID, err)
		return events, fmt.Errorf("解析 game_withdraw 数据失败: %w", err)
	}

	event := createEventTemplate(kafkaID, envelope, &payload, constants.EventTypeWithdrawalSucceeded)
	events = append(events, event)

	logx.Infof("生成提现成功事件: member_id=%d, withdraw_at=%d", payload.MemberID, payload.UpdatedAt.UnixMilli())
	return events, nil
}

// ========== game_deposit 表处理 ==========

// handleGameDepositTable 处理 game_deposit 表事件
// 只有 insert 操作（op='c'）才生成 EventTypeDepositSucceeded 事件
func handleGameDepositTable(kafkaID string, envelope *types.CDCEnvelope, before, after json.RawMessage) ([]types.Event, error) {
	events := make([]types.Event, 0)

	// 只处理 insert 操作（op='c'）
	if envelope.Op != "c" {
		return events, nil
	}

	var payload types.GameDepositPayload
	if err := json.Unmarshal(after, &payload); err != nil {
		logx.Errorf("解析 game_deposit 数据失败: kafka_id=%s, err=%v", kafkaID, err)
		return events, fmt.Errorf("解析 game_deposit 数据失败: %w", err)
	}

	event := createEventTemplate(kafkaID, envelope, &payload, constants.EventTypeDepositSucceeded)
	events = append(events, event)

	logx.Infof("生成充值成功事件: member_id=%d, deposit_at=%d", payload.MemberID, payload.UpdatedAt.UnixMilli())
	return events, nil
}

// ========== agent_apply 表处理 ==========

// handleAgentApplyTable 处理 agent_apply 表事件
// 只有 update 操作（op='u'）并且register_source = 4 且 audit_status = 2 才生成，判断before与after的状态是否不符合条件，只有首次进入符合条件的状态时才推送， EventTypeAgentApplicationApproved 事件
func handleAgentApplyTable(kafkaID string, envelope *types.CDCEnvelope, before, after json.RawMessage) ([]types.Event, error) {
	events := make([]types.Event, 0)
	// 只处理 update 操作（op='u'）
	if envelope.Op != "u" {
		return events, nil
	}

	var beforePayload, afterPayload types.AgentApplyPayload
	if err := json.Unmarshal(before, &beforePayload); err != nil {
		logx.Errorf("解析 agent_apply before 数据失败: kafka_id=%s, err=%v", kafkaID, err)
		return events, fmt.Errorf("解析 agent_apply before 数据失败: %w", err)
	}
	if err := json.Unmarshal(after, &afterPayload); err != nil {
		logx.Errorf("解析 agent_apply after 数据失败: kafka_id=%s, err=%v", kafkaID, err)
		return events, fmt.Errorf("解析 agent_apply after 数据失败: %w", err)
	}

	// 只有 register_source = 4 且 audit_status = 2 且 before 不符合条件时才生成事件
	if afterPayload.RegisterSource == 4 && afterPayload.AuditStatus == 2 &&
		!(beforePayload.RegisterSource == 4 && beforePayload.AuditStatus == 2) {
		event := createEventTemplate(kafkaID, envelope, &afterPayload, constants.EventTypeAgentApplicationApproved)
		events = append(events, event)

		logx.Infof("生成代理申请通过事件: member_id=%d, audited_at=%d", afterPayload.MemberID, afterPayload.UpdatedAt.UnixMilli())
	}
	return events, nil
}

// ========== game_chat_log 表处理 ==========

// handleGameChatLogTable 处理 game_chat_log 表事件
// 只有 insert 操作（op='c'）才生成， EventTypeChatMessageAccepted 事件
func handleGameChatLogTable(kafkaID string, envelope *types.CDCEnvelope, before, after json.RawMessage) ([]types.Event, error) {
	events := make([]types.Event, 0)
	// 只处理 insert 操作（op='c'）
	if envelope.Op != "c" {
		return events, nil
	}

	var payload types.GameChatLogPayload
	if err := json.Unmarshal(after, &payload); err != nil {
		logx.Errorf("解析 game_chat_log 数据失败: kafka_id=%s, err=%v", kafkaID, err)
		return events, fmt.Errorf("解析 game_chat_log 数据失败: %w", err)
	}

	event := createEventTemplate(kafkaID, envelope, &payload, constants.EventTypeChatMessageAccepted)
	events = append(events, event)

	logx.Infof("生成聊天消息被接受事件: member_id=%d, updated_at=%d", payload.MemberID, payload.UpdatedAt.UnixMilli())
	return events, nil
}

// ========== 辅助函数 ==========

// createEventTemplate 创建事件模板
// 支持从不同的 Payload 类型提取元数据
func createEventTemplate(kafkaID string, envelope *types.CDCEnvelope, payload interface{}, eventType string) types.Event {
	var operatorCode string
	var memberID int64
	var occurredAt time.Time

	switch p := payload.(type) {
	case *types.MemberPayload:
		operatorCode = p.OperatorCode
		memberID = int64(p.ID)
		occurredAt = p.UpdatedAt
	case *types.MemberOAuthIdentityPayload:
		operatorCode = p.OperatorCode
		memberID = p.MemberID
		occurredAt = p.UpdatedAt
	case *types.MemberAuthPayload:
		operatorCode = p.OperatorCode
		memberID = p.MemberID
		occurredAt = p.UpdatedAt
	case *types.MemberProfilePayload:
		operatorCode = p.OperatorCode
		memberID = p.MemberID
		occurredAt = p.UpdatedAt
	case *types.GameWithdrawPayload:
		operatorCode = p.OperatorCode
		memberID = p.MemberID
		occurredAt = p.UpdatedAt
	case *types.GameDepositPayload:
		operatorCode = p.OperatorCode
		memberID = p.MemberID
		occurredAt = p.UpdatedAt
	case *types.AgentApplyPayload:
		operatorCode = p.OperatorCode
		memberID = p.MemberID
		occurredAt = p.UpdatedAt
	case *types.GameChatLogPayload:
		operatorCode = p.OperatorCode
		memberID = p.MemberID
		occurredAt = p.UpdatedAt
	default:
		logx.Errorf("未知的负载类型: %T", p)
	}

	return types.Event{
		EventType:    eventType,
		KafkaID:      kafkaID,
		OperatorCode: operatorCode,
		MemberID:     memberID,
		OccurredAt:   occurredAt,
		PublishedAt:  time.UnixMilli(envelope.TSMs),
		MaxRetry:     5,
	}
}
