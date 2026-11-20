package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"oa.98ent.com/p9/platform-base/rpc/ent/schema/mixins"
)

// Nickname 总站共用的默认会员昵称库，允许不同会员使用同一昵称。
type Nickname struct{ ent.Schema }

func (Nickname) Fields() []ent.Field {
	return []ent.Field{field.String("name").NotEmpty().MaxRuneLen(100).Unique().Comment("会员昵称")}
}
func (Nickname) Mixin() []ent.Mixin { return []ent.Mixin{mixins.IDMixin{}, mixins.TimeMixin{}} }
func (Nickname) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "nickname"}}
}
