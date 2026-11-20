package admin

import "oa.98ent.com/p9/operator-activity/ent"

type AdminService struct {
	db *ent.Client
}

func NewService(db *ent.Client) *AdminService {
	return &AdminService{db: db}
}
