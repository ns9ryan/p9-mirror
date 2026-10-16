package handler

import (
	"net/http"

	"oa.98ent.com/p9/oss/api/internal/handler/file"
	"oa.98ent.com/p9/oss/api/internal/handler/provider"
	"oa.98ent.com/p9/oss/api/internal/handler/upload"
	"oa.98ent.com/p9/oss/api/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

var ossRoutes = func(ctx *svc.ServiceContext) []rest.Route {
	return []rest.Route{
		{Method: http.MethodGet, Path: "/url/detail", Handler: upload.ResolveHandler(ctx)},
		{Method: http.MethodPost, Path: "/file/upload", Handler: upload.UploadHandler(ctx)},
		{Method: http.MethodPost, Path: "/upload/presign", Handler: upload.PresignHandler(ctx)},
		{Method: http.MethodPost, Path: "/upload/complete", Handler: upload.CompleteHandler(ctx)},
		{Method: http.MethodPost, Path: "/file/list", Handler: file.ListHandler(ctx)},
		{Method: http.MethodPost, Path: "/file/delete", Handler: file.DeleteHandler(ctx)},
		{Method: http.MethodPost, Path: "/provider/list", Handler: provider.ListHandler(ctx)},
		{Method: http.MethodPost, Path: "/provider/create", Handler: provider.CreateHandler(ctx)},
		{Method: http.MethodGet, Path: "/provider/active", Handler: provider.ActiveHandler(ctx)},
		{Method: http.MethodGet, Path: "/provider/detail", Handler: provider.GetHandler(ctx)},
		{Method: http.MethodPost, Path: "/provider/update", Handler: provider.UpdateHandler(ctx)},
		{Method: http.MethodPost, Path: "/provider/delete", Handler: provider.DeleteHandler(ctx)},
		{Method: http.MethodPost, Path: "/provider/enable", Handler: provider.EnableHandler(ctx)},
	}
}

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes(
		rest.WithMiddlewares(
			[]rest.Middleware{serverCtx.Jwt, serverCtx.ActionLog, serverCtx.Authority},
			ossRoutes(serverCtx)...,
		),
		rest.WithPrefix("/p9/oss"),
	)
}
