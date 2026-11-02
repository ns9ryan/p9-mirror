package member

import (
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/dao"
)

// Service 会员数据访问
type Service struct {
	db        *ent.Client          // Ent数据库客户端
	autoLevel *dao.MemberAutoLevel // 会员自动层级DAO
	vipLevel  *dao.MemberVipLevel  // 会员VIP等级DAO
}

// New 创建会员数据访问服务
func New(db *ent.Client, autoLevel *dao.MemberAutoLevel, vipLevel *dao.MemberVipLevel) *Service {
	return &Service{
		db:        db,
		autoLevel: autoLevel,
		vipLevel:  vipLevel,
	}
}
