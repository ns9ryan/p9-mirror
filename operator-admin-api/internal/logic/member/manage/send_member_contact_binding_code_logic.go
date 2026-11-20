// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package manage

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/zrpc"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"

	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendMemberContactBindingCodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSendMemberContactBindingCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendMemberContactBindingCodeLogic {
	return &SendMemberContactBindingCodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (logic *SendMemberContactBindingCodeLogic) SendMemberContactBindingCode(request *types.SendMemberContactBindingCodeRequest) (*types.SendMemberContactBindingCodeResponse, error) {
	response, err := logic.svcCtx.MemberRpc.SendMemberContactBindingCode(logic.ctx, &memberpb.SendMemberContactBindingCodeRequest{
		MemberId: request.MemberId, Channel: request.Channel, Target: request.Target,
	}, zrpc.WithCallTimeout(40*time.Second))
	if err != nil {
		rpcStatus := status.Convert(err)
		if rpcStatus.Code() == codes.AlreadyExists && rpcStatus.Message() == "member.contactBindingCodeTooFrequent" {
			return nil, xerr.Err(http.StatusTooManyRequests, rpcStatus.Message())
		}
		return nil, err
	}
	return &types.SendMemberContactBindingCodeResponse{
		RecordId: response.GetRecordId(), ExpiresAt: response.GetExpiresAt(),
	}, nil
}
