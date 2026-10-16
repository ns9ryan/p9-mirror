package store

import (
	"context"

	"oa.98ent.com/p9/oss/rpc/ent"
	"oa.98ent.com/p9/oss/rpc/ent/intercept"

	"oa.98ent.com/p9/core/common/entmixin"
)

func attachTenant(c *ent.Client) {
	f := intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
		entmixin.FilterOperatorCode(ctx, q)
		return nil
	})
	c.File.Intercept(f)
	c.Provider.Intercept(f)
}
