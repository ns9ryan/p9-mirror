package message

import (
	"context"
	"errors"

	"oa.98ent.com/p9/integration/rpc/internal/message/email"
	"oa.98ent.com/p9/integration/rpc/internal/svc"
	"oa.98ent.com/p9/integration/rpc/types/integration"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SendEmailLogic 负责邮件 RPC 参数转换和错误映射，发送规则由邮件业务服务处理。
type SendEmailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewSendEmailLogic 绑定本次 RPC 上下文和日志。
func NewSendEmailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendEmailLogic {
	return &SendEmailLogic{
		ctx:    ctx,                   // 本次 RPC 请求上下文，传递取消和超时
		svcCtx: svcCtx,                // 复用已初始化的消息服务
		Logger: logx.WithContext(ctx), // 关联本次请求的日志
	}
}

// SendEmail 将 Proto 请求转换为业务请求，成功时返回受理人数和本服务投递 ID。
func (l *SendEmailLogic) SendEmail(in *integration.SendEmailReq) (*integration.SendEmailResp, error) {
	if in == nil {
		return nil, status.Error(codes.InvalidArgument, "email request is required")
	}

	// 将 RPC 附件转换为业务类型；附件为空项时整次拒绝，尚未向厂商提交。
	attachments := make([]email.Attachment, 0, len(in.GetAttachment()))
	for _, item := range in.GetAttachment() {
		if item == nil {
			return nil, status.Error(codes.InvalidArgument, "email attachment is required")
		}
		attachments = append(attachments, email.Attachment{
			Filename: item.GetFilename(), // 调用方提供的附件名称
			Base64:   item.GetBase64(),   // 标准 Base64 文件内容，不带 data URL 前缀
			MIME:     item.GetMime(),     // 附件媒体类型，后续构建邮件时校验
		})
	}

	if l.svcCtx == nil || l.svcCtx.Deps == nil || l.svcCtx.Deps.Message == nil {
		return nil, status.Error(codes.Unavailable, email.ErrNotConfigured.Error())
	}
	// 调用本服务的邮件业务入口，统一校验收件人、选择通道并处理发送结果。
	result, err := l.svcCtx.Deps.Message.SendEmailResult(l.ctx, email.Request{
		Recipients:  in.GetEmail(),       // 主收件人列表，与抄送和密送一起校验、去重
		CC:          in.GetCc(),          // 抄送地址，显示在邮件头中
		BCC:         in.GetBcc(),         // 密送地址，仅用于投递，不写入邮件头
		Subject:     in.GetSubject(),     // 邮件主题，按通道要求校验
		Content:     in.GetContent(),     // 纯文本正文，与 HTML 至少填写一项
		HTMLContent: in.GetHtmlContent(), // HTML 正文，可与纯文本并存
		Attachments: attachments,         // 已转换的附件，发送前校验格式和大小
		Provider:    in.GetProvider(),    // 空值按优先级选择，否则先匹配通道名称再匹配类型
	})
	if err != nil {
		// 发送失败时保留可用的投递编号；结果不确定由错误详情提示调用方停止自动重发。
		return nil, emailErrorDetails(err, result)
	}
	return &integration.SendEmailResp{
		Accepted:   int32(result.Accepted), // 厂商已受理人数，不代表已进入收件箱
		Message:    "success",              // 本次提交成功
		DeliveryId: result.DeliveryID,      // 本进程记录编号，不是幂等键
		Provider:   result.Provider,        // 实际使用的通道名称
	}, nil
}

// emailErrorDetails 在 RPC 错误详情中保留投递 ID，便于关联日志和内存记录。
// EMAIL_DELIVERY_UNCERTAIN 表示可能已经发出，调用方不应自动重发。
func emailErrorDetails(err error, result email.SendResult) error {
	rpcError := emailStatus(err)
	if result.DeliveryID == "" {
		return rpcError
	}
	reason := "EMAIL_SEND_FAILED"
	if errors.Is(err, email.ErrDeliveryUncertain) {
		reason = "EMAIL_DELIVERY_UNCERTAIN"
	}
	detailed, detailErr := status.Convert(rpcError).WithDetails(&errdetails.ErrorInfo{
		Reason: reason,              // 区分明确失败和可能已发送的不确定结果
		Domain: "integration.email", // 邮件错误详情所属业务域
		Metadata: map[string]string{
			"delivery_id": result.DeliveryID, // 用于关联服务端投递记录和日志
			"provider":    result.Provider,   // 发生本次错误的发送通道
		},
	})
	if detailErr != nil {
		return rpcError
	}
	return detailed.Err()
}

// emailStatus 将业务错误映射为 gRPC 状态，保留取消和超时语义。
func emailStatus(err error) error {
	switch {
	case errors.Is(err, email.ErrInvalidRequest),
		errors.Is(err, email.ErrInvalidRecipient),
		errors.Is(err, email.ErrRecipientBlocked),
		errors.Is(err, email.ErrInvalidAttachment):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, email.ErrNotConfigured),
		errors.Is(err, email.ErrProviderUnavailable):
		return status.Error(codes.Unavailable, err.Error())
	case errors.Is(err, email.ErrProviderUnsupported),
		errors.Is(err, email.ErrProviderRejected),
		errors.Is(err, email.ErrDeliveryUncertain):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, err.Error())
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, err.Error())
	default:
		return err
	}
}
