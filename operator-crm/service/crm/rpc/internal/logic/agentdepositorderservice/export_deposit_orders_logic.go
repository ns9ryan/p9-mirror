package agentdepositorderservicelogic

import (
	"context"

	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentdepositorderpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ExportDepositOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewExportDepositOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExportDepositOrdersLogic {
	return &ExportDepositOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// group: agent_deposit_order 存款订单导出
func (l *ExportDepositOrdersLogic) ExportDepositOrders(in *agentdepositorderpb.ExportDepositOrdersReq) (*agentdepositorderpb.ExportDepositOrdersResp, error) {
	resp, err := l.svcCtx.Agent.ExportDepositOrders(l.ctx, in)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
