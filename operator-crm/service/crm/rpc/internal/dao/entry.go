package dao

import "oa.98ent.com/p9/operator-crm/ent"

// Group 统一管理各业务数据访问
type Group struct {
	MemberAutoLevel MemberAutoLevel // 会员自动层级
}

// New 创建数据访问入口
func New(db *ent.Client) *Group {
	return &Group{
		MemberAutoLevel: MemberAutoLevel{db: db},
	}
}
