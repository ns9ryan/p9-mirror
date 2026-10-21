package memberautolevel

import "oa.98ent.com/p9/operator-crm/ent"

// Service 会员自动层级数据访问
type Service struct {
	DB *ent.Client // Ent数据库客户端
}

// New 创建自动层级数据访问服务
func New(db *ent.Client) *Service {
	return &Service{DB: db}
}
