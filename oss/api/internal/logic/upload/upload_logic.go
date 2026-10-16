package upload

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"oa.98ent.com/p9/oss/api/internal/pkg/body"
	"oa.98ent.com/p9/oss/api/internal/pkg/sniff"
	"oa.98ent.com/p9/oss/api/internal/svc"
	"oa.98ent.com/p9/oss/api/internal/types"
	shared "oa.98ent.com/p9/oss/common/config"
	"oa.98ent.com/p9/oss/common/object"
	"oa.98ent.com/p9/oss/common/oss"
	"oa.98ent.com/p9/oss/rpc/ossclient"

	"oa.98ent.com/p9/common/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type Logic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *Logic {
	return &Logic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *Logic) Upload(scene, filename string, content []byte) (*types.FileData, error) {
	if len(content) == 0 {
		return nil, xerr.BadRequest("no file provided")
	}
	cli, up, err := l.svcCtx.Store(l.ctx)
	if err != nil {
		return nil, err
	}
	if int64(len(content)) > up.MaxSize {
		return nil, xerr.BadRequest("file too large")
	}
	scene, bucket, err := resolveBucket(up, cli, scene)
	if err != nil {
		return nil, err
	}
	filename = strings.TrimSpace(filename)
	if filename == "" {
		return nil, xerr.BadRequest("filename is required")
	}
	ext, err := checkExt(up, filename)
	if err != nil {
		return nil, err
	}
	if err := sniffContent(content, ext); err != nil {
		return nil, err
	}
	key, err := object.PlanObjectKey(scene, ext, operatorCode(l.ctx))
	if err != nil {
		return nil, xerr.InternalServerError(fmt.Sprintf("Failed to read file: %v", err))
	}
	return l.putBytes(putArgs{
		cli: cli, bucket: bucket, scene: scene, key: key,
		name: filename, contentType: mime.TypeByExtension(ext), content: content,
	})
}

type putArgs struct {
	cli         *oss.Client
	bucket      string
	scene       string
	key         string
	name        string
	contentType string
	content     []byte
}

func (l *Logic) putBytes(in putArgs) (*types.FileData, error) {
	accessURL, err := in.cli.Put(l.ctx, oss.PutInput{
		Bucket: in.bucket, Key: in.key, Body: bytes.NewReader(in.content),
		ContentLength: int64(len(in.content)), ContentType: in.contentType,
	})
	if err != nil {
		return nil, xerr.InternalServerError(fmt.Sprintf("upload failed: %v", err))
	}
	l.svcCtx.Remember(l.ctx, &ossclient.FileRecordReq{
		Key: in.key, Scene: in.scene, Bucket: in.bucket, Name: in.name,
		ContentType: in.contentType, Size: int64(len(in.content)),
	})
	return fileView(in.cli, in.scene, in.bucket, in.key, accessURL), nil
}

func sniffContent(content []byte, ext string) error {
	header, err := sniff.PeekHeader(bytes.NewReader(content), sniff.HeaderSize)
	if err != nil {
		return xerr.InternalServerError(fmt.Sprintf("Failed to read file: %v", err))
	}
	if err := sniff.HeaderMatchesExt(ext, header); err != nil {
		return xerr.BadRequest(err.Error())
	}
	return nil
}

func FormScene(r *http.Request) string {
	if r == nil {
		return ""
	}
	return strings.TrimSpace(r.FormValue("scene"))
}

func ReadMultipart(r *http.Request, w http.ResponseWriter, maxSize, memory int64) (string, []byte, error) {
	r.Body = http.MaxBytesReader(w, r.Body, maxSize+shared.MaxRequestOverhead)
	if err := body.ParseForm(r, memory); err != nil {
		if body.IsTooLarge(err) {
			return "", nil, xerr.BadRequest("file too large")
		}
		return "", nil, xerr.BadRequest(err.Error())
	}
	src, file, err := body.File(r, "file")
	if err != nil {
		return "", nil, xerr.BadRequest(err.Error())
	}
	defer src.Close()
	prepared, err := body.Prepare(src, file.Size, maxSize)
	defer prepared.Close()
	if err != nil {
		if body.IsTooLarge(err) {
			return "", nil, xerr.BadRequest("file too large")
		}
		return "", nil, xerr.InternalServerError(err.Error())
	}
	raw, err := io.ReadAll(prepared.Body)
	if err != nil {
		return "", nil, xerr.InternalServerError(err.Error())
	}
	return filepath.Base(file.Filename), raw, nil
}
