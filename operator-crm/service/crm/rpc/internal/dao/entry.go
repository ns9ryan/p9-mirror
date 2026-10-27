package dao

import "oa.98ent.com/p9/operator-crm/ent"

// Group 统一管理各业务数据访问
type Group struct {
	MemberAutoLevel  MemberAutoLevel  // 会员自动层级
	MemberFixedLevel MemberFixedLevel // 会员固定层级

	MemberVipConfig    MemberVipConfig    // 会员VIP配置
	MemberVipLevel     MemberVipLevel     // 会员VIP等级
	MemberVipIconGroup MemberVipIconGroup // 会员VIP图标组
}

// New 创建数据访问入口
func New(db *ent.Client) *Group {
	return &Group{
		MemberAutoLevel:    MemberAutoLevel{db: db},
		MemberFixedLevel:   MemberFixedLevel{db: db},
		MemberVipConfig:    MemberVipConfig{db: db},
		MemberVipLevel:     MemberVipLevel{db: db},
		MemberVipIconGroup: MemberVipIconGroup{db: db},
	}
}
