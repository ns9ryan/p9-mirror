package file

import (
	"context"
	"net/http"
	"strings"

	"oa.98ent.com/p9/oss/rpc/ent"
	"oa.98ent.com/p9/oss/rpc/internal/logic"
	"oa.98ent.com/p9/oss/rpc/internal/store"
	"oa.98ent.com/p9/oss/rpc/internal/svc"
	pb "oa.98ent.com/p9/oss/rpc/types/oss"

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

func (l *ListLogic) ListFiles(in *pb.FileListReq) (*pb.FileList, error) {
	if l.svcCtx.DB == nil {
		return nil, logic.HTTP(http.StatusServiceUnavailable, "database not configured")
	}
	page, pageSize := 1, 20
	scene := ""
	if in != nil {
		scene = in.GetScene()
		if in.GetPage() > 0 {
			page = int(in.GetPage())
		}
		if in.GetPageSize() > 0 {
			pageSize = int(in.GetPageSize())
		}
	}
	rows, total, err := store.ListFiles(l.ctx, l.svcCtx.DB, scene, page, pageSize)
	if err != nil {
		return nil, logic.HTTP(http.StatusInternalServerError, err.Error())
	}
	out := &pb.FileList{List: make([]*pb.FileInfo, 0, len(rows)), Total: int64(total)}
	for _, row := range rows {
		out.List = append(out.List, &pb.FileInfo{
			Id: row.ID, Key: row.ObjectKey, Scene: row.Scene, Bucket: row.Bucket,
			Name: row.Name, Size: row.Size, ContentType: row.ContentType,
			CreatedAt: row.CreatedAt.UnixMilli(),
		})
	}
	return out, nil
}

type DeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteLogic {
	return &DeleteLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *DeleteLogic) DeleteFile(in *pb.KeyReq) (*pb.FileData, error) {
	if l.svcCtx.DB == nil {
		return nil, logic.HTTP(http.StatusServiceUnavailable, "database not configured")
	}
	key := ""
	if in != nil {
		key = strings.TrimPrefix(strings.TrimSpace(in.GetKey()), "/")
	}
	if key == "" {
		return nil, logic.HTTP(http.StatusBadRequest, "key is required")
	}
	row, err := store.GetFileByKey(l.ctx, l.svcCtx.DB, key)
	if ent.IsNotFound(err) {
		return nil, logic.HTTP(http.StatusNotFound, "file not found")
	}
	if err != nil {
		return nil, logic.HTTP(http.StatusInternalServerError, err.Error())
	}
	n, err := store.DeleteFileByKey(l.ctx, l.svcCtx.DB, key)
	if err != nil {
		return nil, logic.HTTP(http.StatusInternalServerError, err.Error())
	}
	if n == 0 {
		return nil, logic.HTTP(http.StatusNotFound, "file not found")
	}
	return &pb.FileData{Key: key, Scene: row.Scene}, nil
}
