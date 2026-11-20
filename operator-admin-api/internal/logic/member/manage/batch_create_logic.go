package manage

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"

	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"
)

type BatchCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBatchCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchCreateLogic {
	return &BatchCreateLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *BatchCreateLogic) BatchCreate(req *types.BatchCreateMembersRequest) (*types.BatchCreateMembersResponse, error) {
	// 转发批量创建请求，并为大批量处理单独设置 RPC 客户端超时。
	resp, err := l.svcCtx.MemberRpc.BatchCreate(l.ctx, &memberpb.BatchCreateMembersRequest{
		AccountPrefix: req.AccountPrefix,
		AccountSuffix: req.AccountSuffix,
		Count:         req.Count,
		AgentId:       req.AgentId,
	}, zrpc.WithCallTimeout(10*time.Minute))
	if err != nil {
		return nil, err
	}
	// 将 CRM 返回的会员标识转换为前端响应结构。
	members := make([]types.BatchCreatedMember, 0, len(resp.GetMembers()))
	for _, member := range resp.GetMembers() {
		members = append(members, types.BatchCreatedMember{
			Id: member.GetId(), Code: member.GetCode(), Account: member.GetAccount(),
		})
	}
	return &types.BatchCreateMembersResponse{Members: members}, nil
}
