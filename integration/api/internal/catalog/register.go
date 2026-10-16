package catalog

import (
	"context"
	"fmt"
	"time"

	"oa.98ent.com/p9/core/rpc/coreclient"
)

// Register 向 core-rpc 注册本服务菜单与 API（重试，等 core 就绪）。
// 写入 sys_api / sys_menu，并为各厅 super_admin 补 Casbin 策略。
func Register(cli coreclient.Core) error {
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
	return fmt.Errorf("register integration-api catalog: %w", last)
}
