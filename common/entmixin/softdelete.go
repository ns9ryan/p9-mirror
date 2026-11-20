package entmixin

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"oa.98ent.com/p9/common/ctxdata"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

type SoftDeleteMixin struct{ mixin.Schema }

func (SoftDeleteMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("deleted_at").Optional().Nillable().Comment("Deleted At | 删除时间"),
	}
}

func (SoftDeleteMixin) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		ent.TraverseFunc(func(ctx context.Context, q ent.Query) error {
			w, ok := q.(Query)
			if !ok {
				return nil
			}
			FilterSoftDelete(ctx, w)
			return nil
		}),
	}
}

func (SoftDeleteMixin) Hooks() []ent.Hook {
	return []ent.Hook{
		func(next ent.Mutator) ent.Mutator {
			return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
				if !m.Op().Is(ent.OpDelete | ent.OpDeleteOne) {
					return next.Mutate(ctx, m)
				}
				if ctxdata.SoftDeleteSkipped(ctx) {
					return next.Mutate(ctx, m)
				}
				mx, ok := m.(interface {
					SetOp(ent.Op)
					SetDeletedAt(time.Time)
					WhereP(...func(*sql.Selector))
				})
				if !ok {
					return nil, fmt.Errorf("soft delete: unexpected mutation %T", m)
				}
				mx.WhereP(func(s *sql.Selector) {
					s.Where(sql.IsNull(s.C("deleted_at")))
				})
				// 转为 Update，经 Client.Mutate 重入，避免仍走 Delete.sqlExec 物理删除。
				mx.SetOp(ent.OpUpdate)
				mx.SetDeletedAt(time.Now())
				return mutateViaClient(ctx, m)
			})
		},
	}
}

// mutateViaClient 调用各仓库生成的 mutation.Client().Mutate（返回类型因仓库而异，用反射适配）。
func mutateViaClient(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	rv := reflect.ValueOf(m)
	method := rv.MethodByName("Client")
	if !method.IsValid() {
		return nil, fmt.Errorf("soft delete: mutation %T has no Client()", m)
	}
	client := method.Call(nil)[0]
	mutate := client.MethodByName("Mutate")
	if !mutate.IsValid() {
		return nil, fmt.Errorf("soft delete: client %T has no Mutate()", client.Interface())
	}
	out := mutate.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(m)})
	if len(out) != 2 {
		return nil, fmt.Errorf("soft delete: unexpected Mutate return from %T", client.Interface())
	}
	if err, _ := out[1].Interface().(error); err != nil {
		return nil, err
	}
	if !out[0].IsValid() {
		return nil, nil
	}
	switch out[0].Kind() {
	case reflect.Interface, reflect.Pointer, reflect.Map, reflect.Slice, reflect.Chan, reflect.Func:
		if out[0].IsNil() {
			return nil, nil
		}
	}
	return out[0].Interface(), nil
}

func FilterSoftDelete(ctx context.Context, q Query) {
	if ctxdata.SoftDeleteSkipped(ctx) {
		return
	}
	q.WhereP(func(s *sql.Selector) {
		s.Where(sql.IsNull(s.C("deleted_at")))
	})
}
