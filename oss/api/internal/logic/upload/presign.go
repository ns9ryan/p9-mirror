package upload

import (
	"errors"
	"fmt"
	"mime"
	"path/filepath"
	"strings"

	"oa.98ent.com/p9/oss/api/internal/types"
	"oa.98ent.com/p9/oss/common/object"
	"oa.98ent.com/p9/oss/common/oss"
	"oa.98ent.com/p9/oss/rpc/ossclient"

	"oa.98ent.com/p9/common/xerr"
)

func (l *Logic) Presign(req *types.PresignRequest) (*types.PresignData, error) {
	if req == nil {
		req = &types.PresignRequest{}
	}
	cli, up, err := l.svcCtx.Store(l.ctx)
	if err != nil {
		return nil, err
	}
	scene, bucket, err := resolveBucket(up, cli, req.Scene)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Filename) == "" {
		return nil, xerr.BadRequest("filename is required")
	}
	if req.Size <= 0 {
		return nil, xerr.BadRequest("size is required")
	}
	if req.Size > up.PresignMaxSize {
		return nil, xerr.BadRequest("file too large")
	}
	ext, err := checkExt(up, req.Filename)
	if err != nil {
		return nil, err
	}
	key, err := object.PlanObjectKey(scene, ext, operatorCode(l.ctx))
	if err != nil {
		return nil, xerr.InternalServerError(fmt.Sprintf("Failed to plan key: %v", err))
	}
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	signed, err := cli.PresignPut(l.ctx, oss.PresignPutInput{
		Bucket: bucket, Key: key, ContentType: contentType,
		ContentLength: req.Size, TTL: up.PresignTTL,
	})
	if err != nil {
		return nil, xerr.InternalServerError(fmt.Sprintf("presign failed: %v", err))
	}
	accessURL, err := cli.URL(l.ctx, bucket, key)
	if err != nil {
		return nil, xerr.InternalServerError(fmt.Sprintf("url failed: %v", err))
	}
	out := &types.PresignData{
		PutURL: signed.URL, Method: signed.Method, Headers: signed.Headers,
		Key: key, URL: accessURL, ExpiresIn: int(up.PresignTTL.Seconds()), Scene: scene,
	}
	vis := fileView(cli, scene, bucket, key, accessURL)
	out.Visibility, out.URLExpiresIn = vis.Visibility, vis.URLExpiresIn
	return out, nil
}

func (l *Logic) Complete(req *types.CompleteRequest) (*types.FileData, error) {
	if req == nil {
		req = &types.CompleteRequest{}
	}
	cli, up, err := l.svcCtx.Store(l.ctx)
	if err != nil {
		return nil, err
	}
	key := strings.TrimPrefix(strings.TrimSpace(req.Key), "/")
	if key == "" {
		return nil, xerr.BadRequest("key is required")
	}
	loc, err := resolveFromKey(up, cli, key, req.Scene)
	if err != nil {
		return nil, err
	}
	if loc.Ext != "" {
		if _, err := checkExt(up, "file"+loc.Ext); err != nil {
			return nil, err
		}
	}
	head, err := cli.Head(l.ctx, loc.Bucket, key)
	if err != nil {
		if errors.Is(err, oss.ErrNotFound) {
			return nil, xerr.BadRequest("object not found: PUT put_url first, then complete")
		}
		return nil, xerr.InternalServerError(fmt.Sprintf("head object: %v", err))
	}
	if head.Size > up.PresignMaxSize {
		_ = cli.Delete(l.ctx, loc.Bucket, key)
		return nil, xerr.BadRequest("file too large")
	}
	accessURL, err := cli.URL(l.ctx, loc.Bucket, key)
	if err != nil {
		return nil, xerr.InternalServerError(fmt.Sprintf("url failed: %v", err))
	}
	name := filepath.Base(key)
	l.svcCtx.Remember(l.ctx, &ossclient.FileRecordReq{
		Key: key, Scene: loc.Scene, Bucket: loc.Bucket, Name: name,
		ContentType: head.ContentType, Size: head.Size,
	})
	return fileView(cli, loc.Scene, loc.Bucket, key, accessURL), nil
}
