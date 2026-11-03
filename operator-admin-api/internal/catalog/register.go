package catalog

import (
	"context"
	"fmt"
	"time"

	"github.com/duke-git/lancet/v2/retry"
	"oa.98ent.com/p9/common/i18n"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
)

// 注册菜单、API目录、多语言数据
func Register(serverCtx *svc.ServiceContext) error {
	// 注册菜单、API目录、多语言数据Req
	req := catalogReq(i18n.CodeOperator)
	// Core Rpc可能还未就绪，失败时短暂重试5次，每次间隔200毫秒
	err := retry.Retry(
		func() error {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			_, err := serverCtx.OperatorCoreRpc.RegisterCatalog(ctx, req)
			return err
		},
		retry.RetryTimes(5),
		retry.RetryWithLinearBackoff(200*time.Millisecond),
	)
	if err != nil {
		return fmt.Errorf("register core-api catalog: %w", err)
	}

	return nil
}
