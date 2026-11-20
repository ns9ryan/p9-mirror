package message

import (
	"context"
	"errors"

	"oa.98ent.com/p9/integration/rpc/internal/message/sms"
	"oa.98ent.com/p9/integration/rpc/internal/svc"
	"oa.98ent.com/p9/integration/rpc/types/integration"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SendSmsLogic 负责短信 RPC 参数及响应转换，不直接调用短信厂商。
type SendSmsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewSendSmsLogic 绑定本次 RPC 上下文和日志。
func NewSendSmsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendSmsLogic {
	return &SendSmsLogic{
		ctx:    ctx,                   // 本次请求的取消和超时上下文
		svcCtx: svcCtx,                // 复用已初始化的短信业务服务
		Logger: logx.WithContext(ctx), // 关联本次 RPC 请求的日志
	}
}

// SendSms 将批次结果和逐号码结果一起返回。
// RPC 无错误不代表全部发送成功，调用方仍须检查 error_code 和 results 的状态。
func (l *SendSmsLogic) SendSms(in *integration.SendSmsReq) (*integration.SendSmsResp, error) {
	if in == nil {
		return nil, status.Error(codes.InvalidArgument, "sms request is required")
	}
	if l.svcCtx == nil || l.svcCtx.Deps == nil || l.svcCtx.Deps.Message == nil {
		return nil, status.Error(codes.Unavailable, sms.ErrNotConfigured.Error())
	}
	// 转换为业务请求；整批校验后逐号码发送，厂商失败通过逐号码结果返回。
	result, err := l.svcCtx.Deps.Message.SendSms(l.ctx, sms.Request{
		Phones:   in.GetPhone(),    // 国际号码列表，发送前统一校验并去重
		Content:  in.GetContent(),  // 本批共用的短信正文
		Provider: in.GetProvider(), // 指定单通道；空值使用配置中的加权路由
	})
	if err != nil {
		return nil, smsStatus(err)
	}
	// 保留每个号码的最终状态，调用方不能仅以 RPC 无错误判断全部成功。
	items := make([]*integration.SmsRecipientResult, 0, len(result.Results))
	for _, item := range result.Results {
		items = append(items, &integration.SmsRecipientResult{
			Phone:     item.Phone,     // 规范化、去重后的国际号码
			Provider:  item.Provider,  // 最后尝试的厂商；未提交时可为空
			Status:    item.Status,    // sent 已受理、failed 明确失败、unknown 结果不确定
			MessageId: item.MessageID, // 厂商返回的受理编号，部分通道允许为空
			Message:   item.Message,   // 失败或结果不确定的说明
			Data:      item.Data,      // 厂商白名单附加数据，不透传完整响应
		})
	}
	return &integration.SendSmsResp{
		ErrorCode:  result.ErrorCode,       // 0 全部受理、2 失败或部分完成、3 存在不确定结果
		Msg:        result.Message,         // 整批结果说明
		Data:       result.Data,            // 逐号码结果的 JSON 字符串，兼容旧调用方
		Errors:     result.Errors,          // 仅保留最终失败或不确定号码的原因
		Accepted:   int32(result.Accepted), // 厂商受理的唯一号码数，不是最终送达数
		DeliveryId: result.DeliveryID,      // 本批调用编号，不是幂等键或送达查询凭据
		Results:    items,                  // 结构化逐号码结果，调用方应优先读取
	}, nil
}

// smsStatus 映射请求级错误；已开始发送后的部分失败由响应结果表达。
func smsStatus(err error) error {
	switch {
	case errors.Is(err, sms.ErrInvalidRequest):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, sms.ErrNotConfigured):
		return status.Error(codes.Unavailable, err.Error())
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, err.Error())
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, err.Error())
	default:
		return status.Error(codes.Internal, "sms request failed")
	}
}
