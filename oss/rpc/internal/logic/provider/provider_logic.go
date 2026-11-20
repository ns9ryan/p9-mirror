package provider

import (
	"context"
	"net/http"
	"strings"

	shared "oa.98ent.com/p9/oss/common/config"
	"oa.98ent.com/p9/oss/rpc/ent"
	"oa.98ent.com/p9/oss/rpc/internal/logic"
	"oa.98ent.com/p9/oss/rpc/internal/store"
	"oa.98ent.com/p9/oss/rpc/internal/svc"
	pb "oa.98ent.com/p9/oss/rpc/types/oss"

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

func (l *Logic) needDB() error {
	if l.svcCtx.DB == nil {
		return logic.HTTP(http.StatusServiceUnavailable, "database not configured")
	}
	return nil
}

func (l *Logic) ListProviders(_ *pb.Empty) (*pb.ProviderList, error) {
	if err := l.needDB(); err != nil {
		return nil, err
	}
	rows, err := store.ListProviders(l.ctx, l.svcCtx.DB)
	if err != nil {
		return nil, logic.HTTP(http.StatusInternalServerError, err.Error())
	}
	out := &pb.ProviderList{List: make([]*pb.Provider, 0, len(rows)), Total: int64(len(rows))}
	for i := range rows {
		out.List = append(out.List, logic.ProviderFromEnt(rows[i], true))
	}
	return out, nil
}

func (l *Logic) GetProvider(in *pb.IDReq) (*pb.Provider, error) {
	if err := l.needDB(); err != nil {
		return nil, err
	}
	if in == nil || in.GetId() == 0 {
		return nil, logic.HTTP(http.StatusBadRequest, "invalid id")
	}
	row, err := store.GetProvider(l.ctx, l.svcCtx.DB, in.GetId())
	if ent.IsNotFound(err) {
		return nil, logic.HTTP(http.StatusNotFound, "provider not found")
	}
	if err != nil {
		return nil, logic.HTTP(http.StatusInternalServerError, err.Error())
	}
	return logic.ProviderFromEnt(row, true), nil
}

func (l *Logic) GetActiveProvider(_ *pb.Empty) (*pb.Provider, error) {
	return l.activeProvider(true), nil
}

func (l *Logic) GetStorageConfig(_ *pb.Empty) (*pb.Provider, error) {
	if p := l.defaultProvider(false); p != nil {
		return p, nil
	}
	return nil, logic.HTTP(http.StatusNotFound, "no active storage provider")
}

func (l *Logic) activeProvider(maskSecret bool) *pb.Provider {
	if p := l.defaultProvider(maskSecret); p != nil {
		return p
	}
	return &pb.Provider{Name: "(none)", Source: l.svcCtx.Source()}
}

func (l *Logic) defaultProvider(maskSecret bool) *pb.Provider {
	if l.svcCtx.DB == nil {
		return nil
	}
	p, err := store.FindDefault(l.ctx, l.svcCtx.DB)
	if err != nil || p == nil {
		return nil
	}
	v := logic.ProviderFromEnt(p, maskSecret)
	v.Source = svc.SourceProvider + ":" + p.Name
	return v
}

func (l *Logic) CreateProvider(in *pb.Provider) (*pb.Provider, error) {
	if err := l.needDB(); err != nil {
		return nil, err
	}
	if in == nil {
		return nil, logic.HTTP(http.StatusBadRequest, "empty request")
	}
	name := strings.TrimSpace(in.GetName())
	if name == "" {
		return nil, logic.HTTP(http.StatusBadRequest, "name is required")
	}
	row := &ent.Provider{Name: name, IsDefault: in.GetIsDefault()}
	spec := logic.SpecFromProvider(in)
	if err := l.validateSpec(spec); err != nil {
		return nil, err
	}
	if err := store.ApplySpec(row, spec); err != nil {
		return nil, logic.HTTP(http.StatusBadRequest, err.Error())
	}
	row, err := store.CreateProvider(l.ctx, l.svcCtx.DB, row)
	if err != nil {
		return nil, logic.HTTP(http.StatusInternalServerError, err.Error())
	}
	if in.GetIsDefault() {
		if err := store.SetDefault(l.ctx, l.svcCtx.DB, row.ID); err != nil {
			return nil, logic.HTTP(http.StatusInternalServerError, err.Error())
		}
	}
	row, _ = store.GetProvider(l.ctx, l.svcCtx.DB, row.ID)
	return logic.ProviderFromEnt(row, true), nil
}

func (l *Logic) UpdateProvider(in *pb.Provider) (*pb.Provider, error) {
	if err := l.needDB(); err != nil {
		return nil, err
	}
	if in == nil || in.GetId() == 0 {
		return nil, logic.HTTP(http.StatusBadRequest, "invalid id")
	}
	row, err := store.GetProvider(l.ctx, l.svcCtx.DB, in.GetId())
	if ent.IsNotFound(err) {
		return nil, logic.HTTP(http.StatusNotFound, "provider not found")
	}
	if err != nil {
		return nil, logic.HTTP(http.StatusInternalServerError, err.Error())
	}
	if n := strings.TrimSpace(in.GetName()); n != "" {
		row.Name = n
	}
	existing, err := store.SpecOf(row)
	if err != nil {
		return nil, logic.HTTP(http.StatusBadRequest, err.Error())
	}
	spec := mergeSpec(existing, in)
	if err := l.validateSpec(spec); err != nil {
		return nil, err
	}
	if err := store.ApplySpec(row, spec); err != nil {
		return nil, logic.HTTP(http.StatusBadRequest, err.Error())
	}
	if err := store.SaveProvider(l.ctx, l.svcCtx.DB, row); err != nil {
		return nil, logic.HTTP(http.StatusInternalServerError, err.Error())
	}
	return logic.ProviderFromEnt(row, true), nil
}

func (l *Logic) DeleteProvider(in *pb.IDReq) (*pb.Empty, error) {
	if err := l.needDB(); err != nil {
		return nil, err
	}
	if in == nil || in.GetId() == 0 {
		return nil, logic.HTTP(http.StatusBadRequest, "invalid id")
	}
	_, err := store.GetProvider(l.ctx, l.svcCtx.DB, in.GetId())
	if ent.IsNotFound(err) {
		return nil, logic.HTTP(http.StatusNotFound, "provider not found")
	}
	if err != nil {
		return nil, logic.HTTP(http.StatusInternalServerError, err.Error())
	}
	if err := store.DeleteProvider(l.ctx, l.svcCtx.DB, in.GetId()); err != nil {
		return nil, logic.HTTP(http.StatusInternalServerError, err.Error())
	}
	return &pb.Empty{}, nil
}

func (l *Logic) EnableProvider(in *pb.IDReq) (*pb.Provider, error) {
	if err := l.needDB(); err != nil {
		return nil, err
	}
	if in == nil || in.GetId() == 0 {
		return nil, logic.HTTP(http.StatusBadRequest, "invalid id")
	}
	row, err := store.GetProvider(l.ctx, l.svcCtx.DB, in.GetId())
	if ent.IsNotFound(err) {
		return nil, logic.HTTP(http.StatusNotFound, "provider not found")
	}
	if err != nil {
		return nil, logic.HTTP(http.StatusInternalServerError, err.Error())
	}
	spec, err := store.SpecOf(row)
	if err != nil {
		return nil, logic.HTTP(http.StatusBadRequest, err.Error())
	}
	if err := l.validateSpec(spec); err != nil {
		return nil, err
	}
	if err := store.SetDefault(l.ctx, l.svcCtx.DB, in.GetId()); err != nil {
		return nil, logic.HTTP(http.StatusInternalServerError, err.Error())
	}
	row, _ = store.GetProvider(l.ctx, l.svcCtx.DB, in.GetId())
	v := logic.ProviderFromEnt(row, true)
	v.Source = l.svcCtx.Source()
	return v, nil
}

func (l *Logic) validateSpec(spec shared.ProviderSpec) error {
	if err := spec.OSS.Validate(); err != nil {
		return logic.HTTP(http.StatusBadRequest, err.Error())
	}
	return nil
}

func mergeSpec(base shared.ProviderSpec, v *pb.Provider) shared.ProviderSpec {
	if v == nil {
		return base
	}
	if d := strings.TrimSpace(v.GetDriver()); d != "" {
		base.OSS.Driver = d
	}
	if v.GetEndpoint() != "" {
		base.OSS.Endpoint = v.GetEndpoint()
	}
	if v.GetFileUrl() != "" {
		base.OSS.FileURL = v.GetFileUrl()
	}
	if strings.TrimSpace(v.GetAccessKey()) != "" {
		base.OSS.AccessKey = v.GetAccessKey()
	}
	if strings.TrimSpace(v.GetSecretKey()) != "" {
		base.OSS.SecretKey = v.GetSecretKey()
	}
	base.OSS.PathStyle = v.GetPathStyle()
	if v.GetRegion() != "" {
		base.OSS.Region = v.GetRegion()
	}
	if buckets := logic.BucketsFromPB(v.GetBuckets()); len(buckets) > 0 {
		base.OSS.Buckets = buckets
	}
	return base
}
