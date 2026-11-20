package memberlogic

import (
	"context"
	"fmt"
	"time"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/integration/rpc/integrationclient"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/dao/member"

	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
)

type SendMemberContactBindingCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendMemberContactBindingCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendMemberContactBindingCodeLogic {
	return &SendMemberContactBindingCodeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (logic *SendMemberContactBindingCodeLogic) SendMemberContactBindingCode(request *memberpb.SendMemberContactBindingCodeRequest) (*memberpb.SendMemberContactBindingCodeResponse, error) {
	if logic.svcCtx.IntegrationRpc == nil {
		return nil, xerr.RpcErr(xerr.InternalServerError("member.contactBindingCodeSendFailed"))
	}
	pending, err := logic.svcCtx.Member.CreateContactBindingCode(logic.ctx, request.GetMemberId(), request.GetChannel(), request.GetTarget(), logic.svcCtx.Cryptor)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}

	content := fmt.Sprintf("Your verification code is %s. It expires in 5 minutes.", pending.Code)
	var accepted bool
	switch request.GetChannel() {
	case member.ContactBindingCodeEmail:
		response, sendErr := logic.svcCtx.IntegrationRpc.SendEmail(logic.ctx, &integrationclient.SendEmailReq{
			Email: []string{pending.Target}, Subject: "Verification code", Content: content,
		}, zrpc.WithCallTimeout(30*time.Second))
		accepted = sendErr == nil && response.GetAccepted() == 1
	case member.ContactBindingCodeSMS:
		response, sendErr := logic.svcCtx.IntegrationRpc.SendSms(logic.ctx, &integrationclient.SendSmsReq{
			Phone: []string{"+" + pending.Target}, Content: content,
		}, zrpc.WithCallTimeout(30*time.Second))
		results := response.GetResults()
		accepted = sendErr == nil && response.GetErrorCode() == 0 && response.GetAccepted() == 1 &&
			len(results) == 1 && results[0].GetStatus() == "sent"
	}

	// Persist the delivery outcome even if the caller disconnected during the provider request.
	finishContext, cancel := context.WithTimeout(context.WithoutCancel(logic.ctx), 3*time.Second)
	defer cancel()
	if err := logic.svcCtx.Member.FinishContactBindingCodeDelivery(finishContext, pending.RecordID, accepted); err != nil {
		logic.Errorf("finish contact binding code delivery for record %d: %v", pending.RecordID, err)
		return nil, xerr.RpcErr(xerr.InternalServerError("member.contactBindingCodeSendFailed"))
	}
	if !accepted {
		return nil, xerr.RpcErr(xerr.InternalServerError("member.contactBindingCodeSendFailed"))
	}
	return &memberpb.SendMemberContactBindingCodeResponse{
		RecordId: pending.RecordID, ExpiresAt: pending.ExpiresAt.UnixMilli(),
	}, nil
}
