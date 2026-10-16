package logic

import (
	"errors"
	"net/http"

	shared "oa.98ent.com/p9/oss/common/config"
	"oa.98ent.com/p9/oss/common/object"

	"oa.98ent.com/p9/common/xerr"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func RpcErr(err error) error {
	if err == nil {
		return nil
	}
	if _, ok := status.FromError(err); ok {
		return err
	}
	e := xerr.AsError(err)
	var code codes.Code
	switch e.Status {
	case http.StatusBadRequest:
		code = codes.InvalidArgument
	case http.StatusUnauthorized:
		code = codes.Unauthenticated
	case http.StatusForbidden:
		code = codes.PermissionDenied
	case http.StatusNotFound:
		code = codes.NotFound
	default:
		if e.Status >= 400 && e.Status < 600 {
			code = codes.Code(e.Status)
		} else {
			code = codes.Internal
		}
	}
	return status.Error(code, e.Message)
}

func HTTP(code int, msg string) error {
	return RpcErr(xerr.Err(code, msg))
}

func SceneErr(err error) error {
	if err == nil {
		return nil
	}
	code := http.StatusInternalServerError
	if errors.Is(err, shared.ErrInvalidScene) || errors.Is(err, object.ErrInvalidKey) {
		code = http.StatusBadRequest
	}
	return HTTP(code, err.Error())
}
