package catalog

import (
	"context"
	"net/http"
	"time"

	"oa.98ent.com/p9/core/rpc/coreclient"
)

const operatorCatalogRegistrationTimeout = 10 * time.Second

// RegisterOperatorUpload 仅向分站注册上传权限，保持原总站目录注册逻辑不变。
func RegisterOperatorUpload(cli coreclient.Core) error {
	ctx, cancel := context.WithTimeout(context.Background(), operatorCatalogRegistrationTimeout)
	defer cancel()
	_, err := cli.RegisterCatalog(ctx, &coreclient.RegisterCatalogReq{
		Apis: []*coreclient.CreateApiReq{api("api.group.oss", "api.ossFileUpload", http.MethodPost, "/file/upload")},
	})
	return err
}
