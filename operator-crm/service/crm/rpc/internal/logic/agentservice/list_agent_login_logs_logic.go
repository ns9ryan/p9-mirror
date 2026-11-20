package agentservicelogic

import (
	"context"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAgentLoginLogsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListAgentLoginLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAgentLoginLogsLogic {
	return &ListAgentLoginLogsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListAgentLoginLogs 代理登录日志列表
func (l *ListAgentLoginLogsLogic) ListAgentLoginLogs(in *agentpb.ListAgentLoginLogsReq) (*agentpb.ListAgentLoginLogsResp, error) {
	resp, err := l.svcCtx.Agent.ListAgentLoginLogs(l.ctx, in)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}
	return resp, nil
}
