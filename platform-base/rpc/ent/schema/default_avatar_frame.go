package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"oa.98ent.com/p9/platform-base/rpc/ent/schema/mixins"
)

// DefaultAvatarFrame 定义总站提供给分站的固定六个头像框。
type DefaultAvatarFrame struct {
	ent.Schema
}

// Fields 定义图片对象 Key 和最后修改人，沿用原列名以兼容已有表，不执行数据转换。
func (DefaultAvatarFrame) Fields() []ent.Field {
	return []ent.Field{
		field.String("image_url").
			Default("").
			MaxLen(1024).
			Comment("头像框的 OSS 对象 Key，兼容已有完整公开 URL；空字符串表示尚未配置图片"),

		field.Int64("updated_by").
			Default(0).
			NonNegative().
			Comment("最后修改人的总站后台用户 ID，由后端登录身份取得；0 表示系统初始化"),
	}
}

// Edges 定义头像框关联关系。
func (DefaultAvatarFrame) Edges() []ent.Edge {
	return nil
}

// Mixin 复用项目现有主键和毫秒精度时间字段。
func (DefaultAvatarFrame) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.TimeMixin{},
	}
}

// Annotations 指定表名、中文注释及修改人约束，固定六张头像框由初始化和业务接口保证。
func (DefaultAvatarFrame) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("总站默认头像框表，提供六张可修改的头像框，供分站读取和选择图片"),
		entsql.Annotation{
			Table: "default_avatar_frame",
			Checks: map[string]string{
				"ck_default_avatar_frame_updated_by": "updated_by >= 0",
			},
		},
	}
}
