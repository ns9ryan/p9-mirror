package upload

import (
	"fmt"
	"strings"

	"oa.98ent.com/p9/oss/api/internal/types"
	"oa.98ent.com/p9/oss/common/object"
	"oa.98ent.com/p9/oss/common/oss"

	"oa.98ent.com/p9/common/xerr"
)

func (l *Logic) Resolve(keyOrURL string) (*types.FileData, error) {
	if strings.TrimSpace(keyOrURL) == "" {
		return nil, xerr.BadRequest("url or key is required")
	}
	cli, up, err := l.svcCtx.Store(l.ctx)
	if err != nil {
		return nil, err
	}
	key := object.KeyFromLocator(keyOrURL, fileBase(cli))
	if key == "" {
		return nil, xerr.BadRequest("invalid key")
	}
	loc, err := resolveFromKey(up, cli, key, "")
	if err != nil {
		return nil, err
	}
	accessURL, err := cli.URL(l.ctx, loc.Bucket, key)
	if err != nil {
		return nil, xerr.InternalServerError(fmt.Sprintf("url failed: %v", err))
	}
	return fileView(cli, loc.Scene, loc.Bucket, key, accessURL), nil
}

func fileBase(cli *oss.Client) string {
	if cli == nil {
		return ""
	}
	cfg := cli.Config()
	base := strings.TrimRight(strings.TrimSpace(cfg.FileURL), "/")
	if base == "" {
		base = strings.TrimRight(strings.TrimSpace(cfg.Endpoint), "/")
	}
	return base
}
