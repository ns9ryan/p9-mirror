package provider

import (
	"net/http"
	"strings"

	"oa.98ent.com/p9/oss/rpc/ent"
	"oa.98ent.com/p9/oss/rpc/internal/logic"
	"oa.98ent.com/p9/oss/rpc/internal/store"
	pb "oa.98ent.com/p9/oss/rpc/types/oss"

	"oa.98ent.com/p9/common/ctxdata"
)

const defaultTemplateName = "default"

// InitTenant 按请求里的 operator_code 写入 init_default_settings.yaml。已有服务商则不覆盖。
func (l *Logic) InitTenant(in *pb.InitTenantReq) (*pb.Provider, error) {
	if err := l.needDB(); err != nil {
		return nil, err
	}
	code := ""
	if in != nil {
		code = strings.TrimSpace(in.GetOperatorCode())
	}
	if code == "" {
		return nil, logic.HTTP(http.StatusBadRequest, "operator_code is required")
	}
	l.ctx = ctxdata.WithClaims(l.ctx, &ctxdata.Claims{OperatorCode: code})
	existing, err := l.existingTenantProvider()
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return logic.ProviderFromEnt(existing, true), nil
	}
	return l.cloneInitTemplate()
}

func (l *Logic) existingTenantProvider() (*ent.Provider, error) {
	p, err := store.FindDefault(l.ctx, l.svcCtx.DB)
	if err != nil {
		return nil, logic.HTTP(http.StatusInternalServerError, err.Error())
	}
	if p != nil {
		return p, nil
	}
	rows, err := store.ListProviders(l.ctx, l.svcCtx.DB)
	if err != nil {
		return nil, logic.HTTP(http.StatusInternalServerError, err.Error())
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return rows[0], nil
}

func (l *Logic) cloneInitTemplate() (*pb.Provider, error) {
	spec := l.svcCtx.InitTemplate()
	if err := l.validateSpec(spec); err != nil {
		return nil, err
	}
	row := &ent.Provider{Name: defaultTemplateName, IsDefault: true}
	if err := store.ApplySpec(row, spec); err != nil {
		return nil, logic.HTTP(http.StatusBadRequest, err.Error())
	}
	created, err := store.CreateProvider(l.ctx, l.svcCtx.DB, row)
	if err != nil {
		return nil, logic.HTTP(http.StatusInternalServerError, err.Error())
	}
	if err := store.SetDefault(l.ctx, l.svcCtx.DB, created.ID); err != nil {
		return nil, logic.HTTP(http.StatusInternalServerError, err.Error())
	}
	created, _ = store.GetProvider(l.ctx, l.svcCtx.DB, created.ID)
	return logic.ProviderFromEnt(created, true), nil
}
