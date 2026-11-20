package crm

import (
	"sync"

	"oa.98ent.com/p9/operator-activity/internal/svc"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"
	"oa.98ent.com/p9/operator-base/rpc/client/operatorservice"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/client/member"
)

// CrmServ CRM服务
type CrmServ struct {
	// 会员管理RPC
	memberRpc member.Member
	// 分站基础RPC
	opBaseRpc operatorservice.OperatorService

	opBaseMu    sync.RWMutex
	opBaseCache map[string]*types.OpBaseInfo
}

// NewMemberServ 创建会员服务
func NewCrmServ(ctx *svc.ServiceContext) *CrmServ {
	return &CrmServ{
		memberRpc:   ctx.MemberRpc,
		opBaseRpc:   ctx.OperatorBaseRpc,
		opBaseCache: make(map[string]*types.OpBaseInfo),
	}
}
