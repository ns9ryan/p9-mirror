package memberlogic

import (
	"context"

	memberdao "oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/dao/member"
	"sort"

	"oa.98ent.com/p9/common/xerr"

	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBasicInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetBasicInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBasicInfoLogic {
	return &GetBasicInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询会员当前自动层级及注册来源
func (l *GetBasicInfoLogic) GetBasicInfo(in *memberpb.GetMemberContextRequest) (*memberpb.GetMemberBasicInfoResponse, error) {
	m, err := l.svcCtx.Member.GetMemberWithContextData(l.ctx, in.GetMemberId(), memberdao.ContextKindBasic)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}
	resp := &memberpb.GetMemberBasicInfoResponse{
		AutoLevel:    m.AutoLevelID,
		OperatorCode: m.OperatorCode,
		RegisterIp:   m.RegisterIP, RegisterDeviceCode: m.RegisterDeviceCode,
		AgentId: m.AgentID, AgentAccountId: m.AgentAccountID, InviterMemberId: m.InviterMemberID,
		FixedLevels: make([]int64, 0),
	}
	for _, assignment := range m.Edges.FixedLevelAssignments {
		resp.FixedLevels = append(resp.FixedLevels, assignment.FixedLevelID)
	}
	sort.Slice(resp.FixedLevels, func(i, j int) bool { return resp.FixedLevels[i] < resp.FixedLevels[j] })
	return resp, nil
}
