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
		// 禁止绕过厅隔离
		if ctxdata.TenantSkipped(ctx) {
			return xerr.Forbidden("common.unauthorized")
		}

		// 校验当前厅身份
		operatorCode := ctxdata.OperatorCodeFromCtx(ctx)
		if operatorCode == "" || strings.TrimSpace(operatorCode) != operatorCode {
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
