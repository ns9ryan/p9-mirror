package member

import (
	"entgo.io/ent/dialect"

	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/dao"
)

// Service 会员数据访问
type Service struct {
	exportDB                 *ent.Client // same driver without SQL debug logging of sensitive filters
	contactBindingCodeDriver dialect.Driver
	db                       *ent.Client          // Ent数据库客户端
	autoLevel                *dao.MemberAutoLevel // 会员自动层级DAO
	vipLevel                 *dao.MemberVipLevel  // 会员VIP等级DAO
}

// New 创建会员数据访问服务
func New(db *ent.Client, autoLevel *dao.MemberAutoLevel, vipLevel *dao.MemberVipLevel, contactBindingCodeDriver dialect.Driver) *Service {
	exportDB := db
	if contactBindingCodeDriver != nil {
		exportDB = ent.NewClient(ent.Driver(contactBindingCodeDriver))
	}
	return &Service{
		exportDB:                 exportDB,
		contactBindingCodeDriver: contactBindingCodeDriver,
		db:                       db,
		autoLevel:                autoLevel,
		vipLevel:                 vipLevel,
	}
}
