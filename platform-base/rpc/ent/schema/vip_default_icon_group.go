package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"oa.98ent.com/p9/platform-base/rpc/ent/schema/mixins"
)

// VipDefaultIconGroup 定义总网提供给分站的默认 VIP 图标组，不保存分站自定义配置。
type VipDefaultIconGroup struct {
	ent.Schema
}

// Fields 定义固定等级范围、图标对象 Key 和修改人，沿用原列名以兼容已有表。
func (VipDefaultIconGroup) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("min_level").
			Range(0, 90).
			Immutable().
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("VIP 等级下限，含本级，固定为 0、10、20、30、40、50、60、70、80、90"),

		field.Int64("max_level").
			Range(9, 99).
			Immutable().
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("VIP 等级上限，含本级，必须等于 min_level 加 9"),

		field.String("icon_url").
			Default("").
			MaxLen(1024).
			Comment("本组默认图标的 OSS 对象 Key，兼容已有完整公开 URL；空字符串表示尚未配置图片"),

		field.Int64("updated_by").
			Default(0).
			NonNegative().
			Comment("最后修改人的总网后台用户 ID，由后端登录身份取得；0 表示系统初始化"),
	}
}

// Edges 定义默认图标组关联关系。
func (VipDefaultIconGroup) Edges() []ent.Edge {
	return nil
}

// Indexes 保证每个固定等级范围仅有一条默认图标配置。
func (VipDefaultIconGroup) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("min_level").
			Unique().
			StorageKey("uk_vip_default_icon_group_min_level"),
	}
}

// Mixin 复用项目现有主键和毫秒精度时间字段。
func (VipDefaultIconGroup) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.TimeMixin{},
	}
}

// Annotations 指定表名、中文注释及固定十组等级范围的数据库约束。
func (VipDefaultIconGroup) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("总网 VIP 默认图标组表，固定十组，供分站读取默认图片"),
		entsql.Annotation{
			Table: "vip_default_icon_group",
			Checks: map[string]string{
				"ck_vip_default_icon_group_level_range": "min_level BETWEEN 0 AND 90 AND min_level % 10 = 0 AND max_level = min_level + 9",
				"ck_vip_default_icon_group_updated_by":  "updated_by >= 0",
			},
		},
	}
}
