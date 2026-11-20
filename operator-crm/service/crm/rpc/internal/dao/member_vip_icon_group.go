package dao

import (
	"context"

	"entgo.io/ent/dialect/sql"

	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/ent/membervipicongroup"
)

// MemberVipIconGroup 会员VIP图标组数据访问
type MemberVipIconGroup struct {
	db *ent.Client // Ent数据库客户端
}

// List 查询本站现有图标组，站点条件由Ent查询拦截器补充；不补建缺失组。
func (d *MemberVipIconGroup) List(ctx context.Context) ([]*ent.MemberVipIconGroup, error) {
	return d.db.MemberVipIconGroup.Query().
		Order(membervipicongroup.ByGroupNo(sql.OrderAsc())). // 按图标组编号升序
		All(ctx)
}

// Get 按组编号查询本站记录，不存在时返回Ent不存在错误。
func (d *MemberVipIconGroup) Get(ctx context.Context, groupNo int64) (*ent.MemberVipIconGroup, error) {
	return d.db.MemberVipIconGroup.Query().
		Where(membervipicongroup.GroupNo(groupNo)). // 组编号，当前站点条件由Ent拦截器补充。
		Only(ctx)
}

// Update 补齐本站十组后保存目标图片；空字符串表示恢复默认，其他组图片保持不变。
func (d *MemberVipIconGroup) Update(ctx context.Context, groupNo int64, iconURL string) error {
	// 复用现有事务组件；补组和保存一并提交，任何一步失败均回滚，不自动重试。
	return WithTx(ctx, d.db, func(tx *ent.Tx) error {
		// 复用站点查询拦截器，只检查本站0~9组，齐全时直接保存，不重复初始化。
		count, err := tx.MemberVipIconGroup.Query().
			Where(membervipicongroup.GroupNoGTE(0), membervipicongroup.GroupNoLT(10)).
			Count(ctx)
		if err != nil {
			return err
		}
		if count < 10 {
			// 兼容尚未执行发布初始化的旧站点；此处只补图标，不初始化会员层级、VIP配置或VIP0。
			// 沿用分站初始化的十组空值及冲突忽略规则；唯一索引保证并发补组不重复。
			for index := int64(0); index < 10; index++ {
				if err := tx.MemberVipIconGroup.Create().
					SetGroupNo(index). // 补齐VIP0~9至90~99的十组。
					SetIconURL("").    // 新组为空，查询时使用总站实时默认图。
					OnConflictColumns(
						membervipicongroup.FieldOperatorCode, // 站点Hook绑定登录站点。
						membervipicongroup.FieldGroupNo,      // 按本站等级组判断冲突。
					).
					DoNothing(). // 已有图片、ID和时间全部保留。
					Exec(ctx); err != nil {
					return err
				}
			}
		}

		// 复用原保存方式，仅更新目标组；不上传、复制或删除OSS对象。
		return tx.MemberVipIconGroup.Create().
			SetGroupNo(groupNo). // 当前保存的图标组，0~9由RPC入口校验。
			SetIconURL(iconURL). // 已校验的Key或历史URL；空值表示使用总站实时默认图。
			OnConflictColumns(
				membervipicongroup.FieldOperatorCode, // 冲突范围限定为登录站点。
				membervipicongroup.FieldGroupNo,      // 同一站点的同一等级组。
			).
			UpdateIconURL().   // 仅替换目标组图片。
			UpdateUpdatedAt(). // 更新保存时间，保留原ID和创建时间。
			Exec(ctx)
	})
}
