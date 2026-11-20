// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package manage

import (
	"context"

	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"

	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMemberContactBindingCodesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListMemberContactBindingCodesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMemberContactBindingCodesLogic {
	return &ListMemberContactBindingCodesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (logic *ListMemberContactBindingCodesLogic) ListMemberContactBindingCodes(request *types.ListMemberContactBindingCodesRequest) (*types.ListMemberContactBindingCodesResponse, error) {
	response, err := logic.svcCtx.MemberRpc.ListMemberContactBindingCodes(logic.ctx, &memberpb.ListMemberContactBindingCodesRequest{
		Page: request.Page, PageSize: request.PageSize, Id: request.Id, MemberId: request.MemberId,
		Channel: request.Channel, Status: request.Status,
		CreatedFrom: request.CreatedFrom, CreatedBefore: request.CreatedBefore,
	})
	if err != nil {
		return nil, err
	}

	list := make([]types.MemberContactBindingCodeItem, 0, len(response.GetList()))
	for _, record := range response.GetList() {
		list = append(list, types.MemberContactBindingCodeItem{
			VerificationCode: record.GetVerificationCode(), Id: record.GetId(), MemberId: record.GetMemberId(), Channel: record.GetChannel(),
			Status: record.GetStatus(), FailedAttempts: record.GetFailedAttempts(),
			CreatedAt: record.GetCreatedAt(), SentAt: record.GetSentAt(),
			ExpiresAt: record.GetExpiresAt(), UsedAt: record.GetUsedAt(),
		})
	}
	return &types.ListMemberContactBindingCodesResponse{Total: response.GetTotal(), List: list}, nil
}
