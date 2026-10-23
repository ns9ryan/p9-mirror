package agentservicelogic

import (
	"context"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAgentOpLogsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListAgentOpLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAgentOpLogsLogic {
	return &ListAgentOpLogsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListAgentOpLogs 代理资料变更操作日志列表
func (l *ListAgentOpLogsLogic) ListAgentOpLogs(in *agentpb.ListAgentOpLogsReq) (*agentpb.ListAgentOpLogsResp, error) {
	resp, err := l.svcCtx.Agent.ListAgentOpLogs(l.ctx, in)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}
	return resp, nil
}
