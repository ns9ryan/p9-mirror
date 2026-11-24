package memberlogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/ent/agentapply"
	"oa.98ent.com/p9/operator-crm/ent/member"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"
)

// GetParentAgentMemberLogic 获取会员上级代理会员ID逻辑
type GetParentAgentMemberLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewGetParentAgentMemberLogic 创建获取会员上级代理会员ID逻辑
func NewGetParentAgentMemberLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetParentAgentMemberLogic {
	return &GetParentAgentMemberLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetParentAgentMember 获取会员上级代理会员ID
// 返回上级代理的ID和对应的会员ID（如果上级代理在Member表中有记录）
func (l *GetParentAgentMemberLogic) GetParentAgentMember(in *memberpb.GetParentAgentMemberRequest) (*memberpb.GetParentAgentMemberResponse, error) {
	// 校验分站运营身份
	if _, ok := ctxdata.OperatorClaimsFromCtx(l.ctx); !ok {
		return nil, xerr.RpcErr(xerr.Unauthorized(i18n.Unauthorized))
	}

	// 校验会员ID
	if in.GetMemberId() <= 0 {
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
	}
	// 检查分站
	if in.GetOperatorCode() == "" {
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
	}

	// 1. 查询会员及其所属代理信息
	m, err := l.svcCtx.DB.Member.Query().
		Where(member.IDEQ(in.GetMemberId())).
		Where(member.OperatorCode(in.GetOperatorCode())).
		WithAgent().
		Only(l.ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, xerr.RpcErr(xerr.EntNotFound(i18n.InvalidParam, err))
		}
		return nil, xerr.RpcErr(err)
	}

	// 获取会员所属代理
	currentAgent := m.Edges.Agent
	if currentAgent == nil {
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
	}

	// 2. 检查是否有上级代理
	if currentAgent.ParentID == nil {
		// 当前代理是顶层代理，无上级代理
		return &memberpb.GetParentAgentMemberResponse{
			ParentAgentMemberId: 0,
			ParentAgentId:       0,
		}, nil
	}

	parentAgentID := *currentAgent.ParentID

	// 3. 查询上级代理在AgentApply表中是否有对应的会员记录
	parentAgentApply, err := l.svcCtx.DB.AgentApply.Query().
		Select(agentapply.FieldMemberID).
		Where(agentapply.AgentIDEQ(parentAgentID)).
		Only(l.ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			// 上级代理没有在AgentApply表中对应的会员记录
			return &memberpb.GetParentAgentMemberResponse{
				ParentAgentMemberId:       0,
				ParentAgentId:             parentAgentID,
				ParentAgentMemberGlobalId: "",
			}, nil
		}
		return nil, xerr.RpcErr(err)
	}
	// 4. 查询上级代理 member_global_id
	memberGlobalID, err := l.svcCtx.Member.GetMemberGlobalID(l.ctx, in.GetOperatorCode(), parentAgentApply.MemberID)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}

	// 4. 返回上级代理的ID和对应的会员ID
	return &memberpb.GetParentAgentMemberResponse{
		ParentAgentMemberId:       parentAgentApply.MemberID,
		ParentAgentId:             parentAgentID,
		ParentAgentMemberGlobalId: memberGlobalID,
	}, nil
}
