package file

import (
	"context"
	"fmt"
	"strings"

	"oa.98ent.com/p9/oss/api/internal/convert"
	"oa.98ent.com/p9/oss/api/internal/svc"
	"oa.98ent.com/p9/oss/api/internal/types"
	"oa.98ent.com/p9/oss/common/object"
	"oa.98ent.com/p9/oss/common/oss"
	"oa.98ent.com/p9/oss/rpc/ossclient"

	"oa.98ent.com/p9/common/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLogic {
	return &ListLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *ListLogic) List(req *types.FileListReq) (*types.FileListData, error) {
	arch, err := l.svcCtx.Archive()
	if err != nil {
		return nil, err
	}
	in := &ossclient.FileListReq{}
	if req != nil {
		in.Scene, in.Page, in.PageSize = req.Scene, int32(req.Page), int32(req.PageSize)
	}
	out, err := arch.ListFiles(l.ctx, in)
	if err != nil {
		return nil, err
	}
	data := convert.FileList(out)
	cli, _, err := l.svcCtx.Store(l.ctx)
	if err != nil {
		return nil, err
	}
	for i := range data.List {
		attachURL(l.ctx, cli, &data.List[i])
	}
	return data, nil
}

type DeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteLogic {
	return &DeleteLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *DeleteLogic) Delete(key string) (*types.FileData, error) {
	key = strings.TrimPrefix(strings.TrimSpace(key), "/")
	if key == "" {
		return nil, xerr.BadRequest("key is required")
	}
	if l.svcCtx.Oss != nil {
		if _, err := l.svcCtx.Oss.DeleteFile(l.ctx, &ossclient.KeyReq{Key: key}); err != nil {
			return nil, err
		}
	}
	return l.deleteObject(key)
}

func (l *DeleteLogic) deleteObject(key string) (*types.FileData, error) {
	cli, up, err := l.svcCtx.Store(l.ctx)
	if err != nil {
		return nil, err
	}
	got, _, ok := object.ParsePlannedKey(key)
	if !ok {
		return nil, xerr.BadRequest("invalid key")
	}
	scene, bucket, err := up.ResolveScene(got)
	if err != nil {
		return nil, xerr.BadRequest(err.Error())
	}
	if err := cli.Delete(l.ctx, bucket, key); err != nil {
		return nil, xerr.InternalServerError(fmt.Sprintf("delete object: %v", err))
	}
	return fileData(cli, scene, bucket, key), nil
}

func attachURL(ctx context.Context, cli *oss.Client, item *types.FileInfo) {
	if cli == nil || item == nil || item.Bucket == "" || item.Key == "" {
		return
	}
	if u, err := cli.URL(ctx, item.Bucket, item.Key); err == nil {
		item.URL = u
	}
	file := fileData(cli, item.Scene, item.Bucket, item.Key)
	item.Visibility, item.URLExpiresIn = file.Visibility, file.URLExpiresIn
}

func fileData(cli *oss.Client, scene, bucket, key string) *types.FileData {
	d := &types.FileData{Key: key, Scene: scene, Visibility: oss.ModePublic}
	if cli != nil && cli.Config().IsPrivate(bucket) {
		d.Visibility = oss.ModePrivate
		d.URLExpiresIn = int(cli.Config().AccessTTL(bucket).Seconds())
	}
	return d
}
