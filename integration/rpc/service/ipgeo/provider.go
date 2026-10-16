package ipgeo

import (
	"context"

	"oa.98ent.com/p9/integration/rpc/ent"
	entprovider "oa.98ent.com/p9/integration/rpc/ent/provider"
	"oa.98ent.com/p9/integration/rpc/ent/providerpolicy"
	"oa.98ent.com/p9/integration/rpc/ent/schema"
	"oa.98ent.com/p9/integration/rpc/internal/ipgeo/geo"
)

// --- 全局定位源 ip_geo_provider（连接配置无管理接口，策略可覆盖启用/优先级）---

// ListProviders 当前租户视角的服务商列表（叠策略后排序）。
func (s *Service) ListProviders(ctx context.Context, operatorCode string) ([]geo.Provider, error) {
	if s == nil || s.Client == nil {
		return nil, geo.ErrMissingDSN
	}
	list, err := s.listAllProviders(ctx)
	if err != nil {
		return nil, err
	}
	return s.overlayByOperator(ctx, list, operatorCode)
}

// ListEnabledProviders Search 用：仅启用源，已按优先级排序。
func (s *Service) ListEnabledProviders(ctx context.Context, operatorCode string) ([]geo.Provider, error) {
	list, err := s.ListProviders(ctx, operatorCode)
	if err != nil {
		return nil, err
	}
	return filterEnabled(list), nil
}

// SeedProvidersIfEmpty 表空时写入 yaml 种子；已有数据不覆盖。
func (s *Service) SeedProvidersIfEmpty(ctx context.Context, seed []geo.Provider) error {
	if s == nil || s.Client == nil {
		return geo.ErrMissingDSN
	}
	list, err := s.listAllProviders(ctx)
	if err != nil || len(list) > 0 || len(seed) == 0 {
		return err
	}
	for _, p := range seed {
		if _, err := s.createProvider(ctx, p); err != nil {
			return err
		}
	}
	return nil
}

// FillEmptyProviderTags 按 driver 回填空 tag。
func (s *Service) FillEmptyProviderTags(ctx context.Context) error {
	if s == nil || s.Client == nil {
		return geo.ErrMissingDSN
	}
	rows, err := s.Client.Provider.Query().Where(entprovider.TagEQ("")).All(ctx)
	if err != nil {
		return err
	}
	for _, row := range rows {
		tag := geo.TagFromDriver(row.Driver)
		if tag == "" {
			continue
		}
		if err := s.Client.Provider.UpdateOneID(row.ID).SetTag(tag).Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) listAllProviders(ctx context.Context) ([]geo.Provider, error) {
	rows, err := s.Client.Provider.Query().
		Order(ent.Asc(entprovider.FieldPriority), ent.Asc(entprovider.FieldID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return fromEntProviders(rows), nil
}

func (s *Service) createProvider(ctx context.Context, p geo.Provider) (*geo.Provider, error) {
	p = geo.NormalizeProvider(p)
	row, err := s.Client.Provider.Create().
		SetName(p.Name).
		SetDriver(p.Driver).
		SetTag(p.Tag).
		SetIsEnabled(p.IsEnabled).
		SetPriority(p.Priority).
		SetEndpoint(p.Endpoint).
		SetAuthType(p.AuthType).
		SetAccessKey(p.AccessKey).
		SetSecretKey(p.SecretKey).
		SetTimeoutSec(p.TimeoutSec).
		SetExtraJSON(schema.ExtraJSON{FieldMap: p.FieldMap, Headers: p.Headers}).
		Save(ctx)
	if ent.IsConstraintError(err) {
		return nil, geo.ErrNameTaken
	}
	if err != nil {
		return nil, err
	}
	out := fromEntProvider(row)
	return &out, nil
}

func (s *Service) overlayByOperator(ctx context.Context, list []geo.Provider, operatorCode string) ([]geo.Provider, error) {
	operatorCode = geo.NormalizeOperatorCode(operatorCode)
	pols, err := s.listTenantPolicies(ctx, operatorCode)
	if err != nil {
		return nil, err
	}
	return overlayList(list, operatorCode, pols), nil
}

func (s *Service) listTenantPolicies(ctx context.Context, operatorCode string) (map[int64]tenantPolicy, error) {
	rows, err := s.Client.ProviderPolicy.Query().
		Where(providerpolicy.OperatorCodeEQ(operatorCode)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	out := make(map[int64]tenantPolicy, len(rows))
	for _, row := range rows {
		out[row.ProviderID] = tenantPolicy{IsEnabled: row.IsEnabled, Priority: row.Priority}
	}
	return out, nil
}

func (s *Service) providerExists(ctx context.Context, id int64) error {
	_, err := s.Client.Provider.Get(ctx, id)
	if ent.IsNotFound(err) {
		return geo.ErrNotFound
	}
	return err
}

func fromEntProviders(rows []*ent.Provider) []geo.Provider {
	out := make([]geo.Provider, 0, len(rows))
	for _, row := range rows {
		out = append(out, fromEntProvider(row))
	}
	return out
}

func fromEntProvider(row *ent.Provider) geo.Provider {
	return geo.NormalizeProvider(geo.Provider{
		ID:         row.ID,
		Name:       row.Name,
		Driver:     row.Driver,
		Tag:        row.Tag,
		IsEnabled:  row.IsEnabled,
		Priority:   row.Priority,
		Endpoint:   row.Endpoint,
		AuthType:   row.AuthType,
		AccessKey:  row.AccessKey,
		SecretKey:  row.SecretKey,
		TimeoutSec: row.TimeoutSec,
		FieldMap:   row.ExtraJSON.FieldMap,
		Headers:    row.ExtraJSON.Headers,
	})
}
