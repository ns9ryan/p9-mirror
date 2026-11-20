package upload

import (
	"context"
	"path/filepath"
	"strings"

	"oa.98ent.com/p9/oss/api/internal/types"
	shared "oa.98ent.com/p9/oss/common/config"
	"oa.98ent.com/p9/oss/common/object"
	"oa.98ent.com/p9/oss/common/oss"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
)

type keyLoc struct {
	Scene  string
	Bucket string
	Ext    string
}

func resolveBucket(up shared.UploadSettings, cli *oss.Client, raw string) (scene, bucket string, err error) {
	scene, bucket, err = up.ResolveScene(raw)
	if err != nil {
		return "", "", xerr.BadRequest(err.Error())
	}
	if cli == nil {
		return "", "", xerr.BadRequest("scene bucket is not in oss.buckets")
	}
	cfg := cli.Config()
	if !cfg.AllowBucket(bucket) {
		return "", "", xerr.BadRequest("scene bucket is not in oss.buckets")
	}
	return scene, bucket, nil
}

func resolveFromKey(up shared.UploadSettings, cli *oss.Client, key, requestedScene string) (keyLoc, error) {
	got, ext, ok := object.ParsePlannedKey(key)
	if !ok {
		return keyLoc{}, xerr.BadRequest(object.ErrInvalidKey.Error())
	}
	raw := requestedScene
	if strings.TrimSpace(raw) == "" {
		raw = got
	}
	scene, bucket, err := resolveBucket(up, cli, raw)
	if err != nil {
		return keyLoc{}, err
	}
	if scene != got {
		return keyLoc{}, xerr.BadRequest(object.ErrInvalidKey.Error())
	}
	return keyLoc{Scene: scene, Bucket: bucket, Ext: ext}, nil
}

func fileView(cli *oss.Client, scene, bucket, key, url string) *types.FileData {
	d := &types.FileData{URL: url, Key: key, Scene: scene, Visibility: oss.ModePublic}
	if cli != nil && cli.Config().IsPrivate(bucket) {
		d.Visibility = oss.ModePrivate
		d.URLExpiresIn = int(cli.Config().AccessTTL(bucket).Seconds())
	}
	return d
}

func checkExt(up shared.UploadSettings, filename string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	if !object.ExtAllowed(up.AllowedExtensions, ext) {
		return "", xerr.BadRequest("file extension not allowed")
	}
	return ext, nil
}

func operatorCode(ctx context.Context) string {
	if c := ctxdata.ClaimsFromCtx(ctx); c != nil {
		return c.OperatorCode
	}
	return ""
}
