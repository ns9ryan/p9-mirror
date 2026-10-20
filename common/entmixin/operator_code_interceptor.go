package entmixin

import (
	"context"
	"strings"

	"entgo.io/ent"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
)

// RequiredOperatorCodeInterceptor 为厅业务实体提供严格的查询隔离
// newQuery 用于适配各仓库生成的 Ent Query
func RequiredOperatorCodeInterceptor[T Query](newQuery func(ent.Query) (T, error)) ent.Interceptor {
	return ent.TraverseFunc(func(ctx context.Context, q ent.Query) error {
		// 严格隔离不允许跳过厅过滤
		if ctxdata.TenantSkipped(ctx) {
			return xerr.Forbidden("common.unauthorized")
		}

		// 缺少厅身份时拒绝查询，避免无条件访问数据
		if strings.TrimSpace(ctxdata.OperatorCodeFromCtx(ctx)) == "" {
			return xerr.Unauthorized("common.unauthorized")
		}

		// 将原始 Ent Query 转换为支持 WhereP 的查询对象
		query, err := newQuery(q)
		if err != nil {
			return err
		}

		// 按当前厅编码过滤查询数据
		FilterOperatorCode(ctx, query)
		return nil
	})
}
