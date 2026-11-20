package memberlogic

import (
	"context"
	"time"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/paging"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/dao/member"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"

	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMemberContactBindingCodesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListMemberContactBindingCodesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMemberContactBindingCodesLogic {
	return &ListMemberContactBindingCodesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (logic *ListMemberContactBindingCodesLogic) ListMemberContactBindingCodes(request *memberpb.ListMemberContactBindingCodesRequest) (*memberpb.ListMemberContactBindingCodesResponse, error) {
	operatorCode := ctxdata.OperatorCodeFromCtx(logic.ctx)
	if _, authorized := ctxdata.OperatorClaimsFromCtx(logic.ctx); !authorized || operatorCode == "" {
		return nil, xerr.RpcErr(xerr.Unauthorized(i18n.Unauthorized))
	}

	offset, limit, valid := paging.OffsetLimit(request.GetPage(), request.GetPageSize())
	if !valid || request.GetPage() < 0 || request.GetPageSize() < 0 ||
		request.GetId() < 0 || request.GetMemberId() < 0 ||
		request.GetChannel() < 0 || request.GetChannel() > 2 ||
		request.GetStatus() < 0 || request.GetStatus() > 4 ||
		request.GetCreatedFrom() < 0 || request.GetCreatedBefore() < 0 ||
		(request.GetCreatedBefore() > 0 && request.GetCreatedFrom() >= request.GetCreatedBefore()) {
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
	}

	records, total, err := logic.svcCtx.Member.ListContactBindingCodes(logic.ctx, operatorCode, member.ContactBindingCodeFilters{
		RecordID: request.GetId(), MemberID: request.GetMemberId(), Channel: request.GetChannel(),
		Status: request.Status, CreatedFrom: request.GetCreatedFrom(), CreatedBefore: request.GetCreatedBefore(),
	}, offset, limit)
	if err != nil {
		logic.Errorf("list member contact binding codes: %v", err)
		return nil, xerr.RpcErr(xerr.InternalServerError(i18n.InternalError))
	}

	list := make([]*memberpb.MemberContactBindingCodeItem, 0, len(records))
	for _, record := range records {
		list = append(list, &memberpb.MemberContactBindingCodeItem{
			VerificationCode: record.VerificationCode, Id: record.ID, MemberId: record.MemberID, Channel: record.Channel, Status: record.Status,
			FailedAttempts: record.FailedAttempts, CreatedAt: record.CreatedAt.UnixMilli(),
			SentAt: contactBindingCodeTimestamp(record.SentAt), ExpiresAt: record.ExpiresAt.UnixMilli(),
			UsedAt: contactBindingCodeTimestamp(record.UsedAt),
		})
	}
	return &memberpb.ListMemberContactBindingCodesResponse{Total: total, List: list}, nil
}

func contactBindingCodeTimestamp(timestamp *time.Time) int64 {
	if timestamp == nil {
		return 0
	}
	return timestamp.UnixMilli()
}
