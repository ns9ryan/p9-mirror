package entmixin

import (
	"context"

	"oa.98ent.com/p9/common/ctxdata"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// OperatorCodeNotNullMixin 分站编码混合必须不为空
type OperatorCodeNotNullMixin struct{ mixin.Schema }

// Fields 分站编码字段
func (OperatorCodeNotNullMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("operator_code").
			MaxLen(64).
			NotEmpty().
			Immutable().
			Comment("Operator Code | 分站编码"),
	}
}

// Interceptors 分站编码拦截器， 此处拦截器不会生效，因为q不是Query接口，所以需要业务侧自行注册拦截器
// 可参考core项目rpc的service_context.go设置拦截器 c.User.Intercept(fOperatorCode)
func (OperatorCodeNotNullMixin) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		ent.TraverseFunc(func(ctx context.Context, q ent.Query) error {
			w, ok := q.(Query)
			if !ok {
				return nil
			}
			FilterOperatorCode(ctx, w)
			return nil
		}),
	}
}

// Hooks 分站编码钩子
func (OperatorCodeNotNullMixin) Hooks() []ent.Hook {
	return []ent.Hook{
		func(next ent.Mutator) ent.Mutator {
			return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
				if ctxdata.TenantSkipped(ctx) {
					return next.Mutate(ctx, m)
				}
				c := ctxdata.ClaimsFromCtx(ctx)
				if c == nil {
					return next.Mutate(ctx, m)
				}
				if m.Op().Is(ent.OpCreate) {
					if c.OperatorCode != "" {
						if _, exists := m.Field("operator_code"); !exists {
							if err := m.SetField("operator_code", c.OperatorCode); err != nil {
								return nil, err
							}
						}
					}
					return next.Mutate(ctx, m)
				}
				if w, ok := m.(interface {
					WhereP(...func(*sql.Selector))
				}); ok {
					w.WhereP(operatorCodeWhere(c.OperatorCode))
				}
				return next.Mutate(ctx, m)
			})
		},
	}
}
