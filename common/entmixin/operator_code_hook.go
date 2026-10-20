package entmixin

import (
	"context"
	"fmt"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
)

// RequiredOperatorCodeHook 为厅业务实体提供严格的写入隔离
func RequiredOperatorCodeHook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			// 禁止绕过厅隔离
			if ctxdata.TenantSkipped(ctx) {
				return nil, xerr.Forbidden("common.unauthorized")
			}

			// 校验当前厅身份
			operatorCode := ctxdata.OperatorCodeFromCtx(ctx)
			if operatorCode == "" || strings.TrimSpace(operatorCode) != operatorCode {
				return nil, xerr.Unauthorized("common.unauthorized")
			}

			switch m.Op() {
			case ent.OpCreate:
				// 新增时自动绑定当前厅，禁止指定其他厅
				if value, exists := m.Field("operator_code"); exists {
					if value != operatorCode {
						return nil, xerr.Forbidden("common.unauthorized")
					}
				} else {
					if err := m.SetField("operator_code", operatorCode); err != nil {
						return nil, err
					}
				}

			case ent.OpUpdate, ent.OpUpdateOne, ent.OpDelete, ent.OpDeleteOne:
				// 禁止修改厅归属
				if _, exists := m.Field("operator_code"); exists {
					return nil, xerr.Forbidden("common.unauthorized")
				}

				// 更新和删除必须限定当前厅
				query, ok := m.(interface {
					WhereP(...func(*sql.Selector))
				})
				if !ok {
					return nil, fmt.Errorf("operator code hook: mutation %T does not support WhereP", m)
				}
				query.WhereP(operatorCodeWhere(operatorCode))

			default:
				return nil, fmt.Errorf("operator code hook: unsupported mutation operation %v", m.Op())
			}

			return next.Mutate(ctx, m)
		})
	}
}
