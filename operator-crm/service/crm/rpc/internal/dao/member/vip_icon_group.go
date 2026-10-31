package member

import (
	"context"

	"entgo.io/ent/dialect/sql"

	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/ent/membervipicongroup"
)

// VIPIconGroup 会员VIP图标组数据访问
type VIPIconGroup struct {
	db *ent.Client // Ent数据库客户端
}

// NewVIPIconGroup 创建VIP图标组数据访问
func NewVIPIconGroup(db *ent.Client) *VIPIconGroup {
	return &VIPIconGroup{db: db}
}

// List 获取VIP图标组列表
func (d *VIPIconGroup) List(ctx context.Context) ([]*ent.MemberVipIconGroup, error) {
	return d.db.MemberVipIconGroup.Query().
		Order(membervipicongroup.ByGroupNo(sql.OrderAsc())). // 按图标组编号升序
		All(ctx)
}

// Get 根据图标组编号获取图标组
func (d *VIPIconGroup) Get(ctx context.Context, groupNo int64) (*ent.MemberVipIconGroup, error) {
	return d.db.MemberVipIconGroup.Query().
		Where(membervipicongroup.GroupNo(groupNo)).
		Only(ctx)
}

// Update 修改当前使用的VIP图标
func (d *VIPIconGroup) Update(ctx context.Context, groupNo int64, iconURL string) error {
	group, err := d.Get(ctx, groupNo)
	if err != nil {
		return err
	}

	return d.db.MemberVipIconGroup.UpdateOneID(group.ID).
		SetIconURL(iconURL). // 当前使用的图标URL
		Exec(ctx)
}

// Reset 恢复VIP图标组的系统默认图标
func (d *VIPIconGroup) Reset(ctx context.Context, groupNo int64) error {
	group, err := d.Get(ctx, groupNo)
	if err != nil {
		return err
	}

	return d.db.MemberVipIconGroup.UpdateOneID(group.ID).
		SetIconURL(group.DefaultIconURL). // 恢复系统默认图标
		Exec(ctx)
}
