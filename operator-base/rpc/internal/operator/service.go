package operator

import (
	"oa.98ent.com/p9/operator-base/rpc/ent"
	"oa.98ent.com/p9/operator-base/rpc/internal/config"
)

// Service operator 业务服务
type Service struct {
	db     *ent.Client
	config config.Config
}

// NewService 创建 operator 业务服务
func NewService(db *ent.Client, c config.Config) *Service {
	return &Service{
		db:     db,
		config: c,
	}
}
