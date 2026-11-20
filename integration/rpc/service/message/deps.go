package message

import (
	"context"

	"oa.98ent.com/p9/integration/rpc/internal/message/email"
	"oa.98ent.com/p9/integration/rpc/internal/message/sms"
)

// Service 聚合邮件和短信业务，为 RPC Logic 提供统一调用入口。
type Service struct {
	Email *email.Service // 邮件发送实例；为空时邮件入口返回未配置错误
	SMS   *sms.Service   // 短信发送实例；为空时短信入口返回未配置错误
}

// New 注入邮件服务，短信服务在启动组装时通过 SMS 字段注入。
func New(emailService *email.Service) *Service {
	return &Service{
		Email: emailService, // 复用调用方注入的邮件实例，短信由启动组装代码另行注入
	}
}

// SendEmail 保留只返回受理人数的简化入口。
func (s *Service) SendEmail(ctx context.Context, req email.Request) (int, error) {
	result, err := s.SendEmailResult(ctx, req)
	return result.Accepted, err
}

// SendEmailResult 返回邮件受理结果和投递 ID，依赖未就绪时返回未配置错误。
func (s *Service) SendEmailResult(ctx context.Context, req email.Request) (email.SendResult, error) {
	if s == nil || s.Email == nil {
		return email.SendResult{}, email.ErrNotConfigured
	}
	// 直接复用邮件业务服务；本层不另做重试，避免重复提交。
	return s.Email.SendResult(ctx, req)
}

// SendSms 返回整批及逐号码的短信受理结果。
func (s *Service) SendSms(ctx context.Context, req sms.Request) (sms.SendResult, error) {
	if s == nil || s.SMS == nil {
		return sms.SendResult{}, sms.ErrNotConfigured
	}
	// 直接复用短信业务服务，返回批次和逐号码状态供 RPC 层映射。
	return s.SMS.Send(ctx, req)
}
