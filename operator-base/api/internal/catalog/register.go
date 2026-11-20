package catalog

import (
	"context"
	"fmt"
	"time"

	"oa.98ent.com/p9/core/rpc/coreclient"
)

// Register 调用本站 Core RPC 登记网站管理接口及翻译，不负责校验登录 Token。
// Core 可据此更新权限目录；最终失败返回给启动入口，由入口决定是否继续启动。
func Register(cli coreclient.Core) error {
	req := registerRequest()
	var last error
	// 包含首次调用在内最多尝试 20 次，每次 RPC 最长等待 3 秒。
	for attempt := range 20 {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		_, last = cli.RegisterCatalog(ctx, req)
		// 每次调用结束立即释放超时上下文，避免累积到整个登记流程结束。
		cancel()
		if last == nil {
			return nil
		}
		if attempt < 19 {
			// 仅失败且仍有剩余尝试时等待 200 毫秒，不在最后一次失败后继续等待。
			time.Sleep(200 * time.Millisecond)
		}
	}
	return fmt.Errorf("register operator-base-api catalog: %w", last)
}
