package ipgeo

import (
	"context"

	"oa.98ent.com/p9/integration/rpc/ent"
	"oa.98ent.com/p9/integration/rpc/ent/providerpolicy"
	"oa.98ent.com/p9/integration/rpc/internal/ipgeo/geo"
)

// --- 租户 / 全局定位源策略（ip_geo_provider_policy）---

// ListPolicies 当前租户视角的服务商列表（无租户策略则用 Provider 全局默认）。
func (s *Service) ListPolicies(ctx context.Context, operatorCode string) ([]geo.Provider, error) {
	return s.ListProviders(ctx, operatorCode)
}

// SavePolicies 保存租户策略。
func (s *Service) SavePolicies(ctx context.Context, operatorCode string, items []geo.PolicyItem) ([]geo.Provider, error) {
	if s == nil || s.Client == nil {
		return nil, geo.ErrMissingDSN
	}
	items, err := normalizePolicyItems(items)
	if err != nil {
		return nil, err
	}
	operatorCode = geo.NormalizeOperatorCode(operatorCode)
	for _, item := range items {
		if err := s.providerExists(ctx, item.ProviderID); err != nil {
			return nil, err
		}
	}
	if err := s.upsertPolicies(ctx, operatorCode, items); err != nil {
		return nil, err
	}
	return s.ListProviders(ctx, operatorCode)
}

// SaveDefaultPolicies 保存全局默认启用/优先级。
func (s *Service) SaveDefaultPolicies(ctx context.Context, items []geo.PolicyItem) ([]geo.Provider, error) {
	if s == nil || s.Client == nil {
		return nil, geo.ErrMissingDSN
	}
	items, err := normalizePolicyItems(items)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		if err := s.providerExists(ctx, item.ProviderID); err != nil {
			return nil, err
		}
	}
	if err := s.updateDefaults(ctx, items); err != nil {
		return nil, err
	}
	return s.listAllProviders(ctx)
}

func (s *Service) upsertPolicies(ctx context.Context, operatorCode string, items []geo.PolicyItem) error {
	tx, err := s.Client.Tx(ctx)
	if err != nil {
		return err
	}
	for _, item := range items {
		err := tx.ProviderPolicy.Create().
			SetOperatorCode(operatorCode).
			SetProviderID(item.ProviderID).
			SetIsEnabled(item.IsEnabled).
			SetPriority(item.Priority).
			OnConflictColumns(providerpolicy.FieldOperatorCode, providerpolicy.FieldProviderID).
			UpdateNewValues().
			Exec(ctx)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (s *Service) updateDefaults(ctx context.Context, items []geo.PolicyItem) error {
	tx, err := s.Client.Tx(ctx)
	if err != nil {
		return err
	}
	for _, item := range items {
		err := tx.Provider.UpdateOneID(item.ProviderID).
			SetIsEnabled(item.IsEnabled).
			SetPriority(item.Priority).
			Exec(ctx)
		if err != nil {
			_ = tx.Rollback()
			if ent.IsNotFound(err) {
				return geo.ErrNotFound
			}
			return err
		}
	}
	return tx.Commit()
}
