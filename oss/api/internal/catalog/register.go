package catalog

import (
	"context"
	"fmt"
	"time"

	"oa.98ent.com/p9/core/rpc/coreclient"
)

// Register 向 core-rpc 注册本服务菜单与 API（重试，等 core 就绪）。
func Register(cli coreclient.Core) error {
	if cli == nil {
		return fmt.Errorf("register oss-api catalog: core client is nil")
	}
	req := CatalogReq()
	var last error
	for range 20 {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		_, last = cli.RegisterCatalog(ctx, req)
		cancel()
		if last == nil {
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("register oss-api catalog: %w", last)
}
