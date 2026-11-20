package provider

import (
	"net/http"

	plogic "oa.98ent.com/p9/oss/api/internal/logic/provider"
	"oa.98ent.com/p9/oss/api/internal/pkg/body"
	"oa.98ent.com/p9/oss/api/internal/svc"
	"oa.98ent.com/p9/oss/api/internal/types"

	"oa.98ent.com/p9/common/xerr"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func writeJSON(w http.ResponseWriter, r *http.Request, resp any, err error) {
	if err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}
	httpx.OkJsonCtx(r.Context(), w, resp)
}

func parseIDReq(r *http.Request) (types.IDReq, error) {
	var req types.IDReq
	if err := body.JSON(r, &req); err != nil {
		return req, xerr.BadRequest("invalid id")
	}
	if req.FirstID() == 0 {
		return req, xerr.BadRequest("invalid id")
	}
	return req, nil
}

// ListHandler 服务商列表。secret_key 不返回。
// @Tags 服务商
// @Summary 服务商列表
// @Produce json
// @Success 200 {object} types.ProviderListReply
// @Router /provider/list [post]
func ListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := plogic.NewLogic(r.Context(), svcCtx)
		resp, err := l.List()
		writeJSON(w, r, resp, err)
	}
}

// ActiveHandler 当前生效的存储配置。
// @Tags 服务商
// @Summary 当前生效配置
// @Produce json
// @Success 200 {object} types.ProviderReply
// @Router /provider/active [get]
func ActiveHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := plogic.NewLogic(r.Context(), svcCtx)
		resp, err := l.Active()
		writeJSON(w, r, resp, err)
	}
}

// GetHandler 按 id 取服务商。
// @Tags 服务商
// @Summary 服务商详情
// @Produce json
// @Param id query int true "id"
// @Success 200 {object} types.ProviderReply
// @Router /provider/detail [get]
func GetHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.IDReq
		if err := httpx.Parse(r, &req); err != nil || req.FirstID() == 0 {
			httpx.ErrorCtx(r.Context(), w, xerr.BadRequest("invalid id"))
			return
		}
		l := plogic.NewLogic(r.Context(), svcCtx)
		resp, err := l.Get(req.FirstID())
		writeJSON(w, r, resp, err)
	}
}

// CreateHandler 新增服务商。is_default=true 时立即覆盖运行时连接。
// @Tags 服务商
// @Summary 新增服务商
// @Accept json
// @Produce json
// @Param body body types.ProviderView true "连接配置。不要传 id / source"
// @Success 200 {object} types.ProviderReply
// @Router /provider/create [post]
func CreateHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ProviderView
		if err := body.JSON(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerr.BadRequest(err.Error()))
			return
		}
		l := plogic.NewLogic(r.Context(), svcCtx)
		resp, err := l.Create(&req)
		writeJSON(w, r, resp, err)
	}
}

// UpdateHandler 更新服务商。secret_key 空则保留原值。
// @Tags 服务商
// @Summary 更新服务商
// @Accept json
// @Produce json
// @Param body body types.ProviderView true "须含 id。secret_key 空则不改"
// @Success 200 {object} types.ProviderReply
// @Router /provider/update [post]
func UpdateHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ProviderView
		if err := body.JSON(r, &req); err != nil || req.ID == 0 {
			httpx.ErrorCtx(r.Context(), w, xerr.BadRequest("invalid id"))
			return
		}
		l := plogic.NewLogic(r.Context(), svcCtx)
		resp, err := l.Update(&req)
		writeJSON(w, r, resp, err)
	}
}

// DeleteHandler 删除服务商。若删的是默认项，运行时退回 yaml/etcd。
// @Tags 服务商
// @Summary 删除服务商
// @Accept json
// @Produce json
// @Param body body types.IDReq true "id 或 ids"
// @Success 200 {object} types.ProviderReply
// @Router /provider/delete [post]
func DeleteHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := parseIDReq(r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := plogic.NewLogic(r.Context(), svcCtx)
		writeJSON(w, r, nil, l.Delete(req.FirstID()))
	}
}

// EnableHandler 设为默认并立即生效。
// @Tags 服务商
// @Summary 启用为默认
// @Accept json
// @Produce json
// @Param body body types.IDReq true "id"
// @Success 200 {object} types.ProviderReply
// @Router /provider/enable [post]
func EnableHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := parseIDReq(r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := plogic.NewLogic(r.Context(), svcCtx)
		resp, err := l.Enable(req.FirstID())
		writeJSON(w, r, resp, err)
	}
}
