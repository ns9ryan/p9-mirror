package turnoverservicelogic

import (
	"context"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/turnoverpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTurnoversLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListTurnoversLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTurnoversLogic {
	return &ListTurnoversLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// group: turnover 流水列表
func (l *ListTurnoversLogic) ListTurnovers(in *turnoverpb.ListTurnoversReq) (*turnoverpb.ListTurnoversResp, error) {
	resp, err := l.svcCtx.Agent.ListTurnovers(l.ctx, in)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}
	return resp, nil
}
