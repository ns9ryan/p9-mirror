package ipgeo

import (
	"sort"

	"oa.98ent.com/p9/integration/rpc/internal/ipgeo/geo"
)

// tenantPolicy 租户侧对某 Provider 的启用与优先级覆盖。
type tenantPolicy struct {
	IsEnabled bool // 覆盖全局 IsEnabled
	Priority  int  // 覆盖全局 Priority
}

// normalizePolicyItems 校验策略列表：非空、无重复、未超限。
func normalizePolicyItems(items []geo.PolicyItem) ([]geo.PolicyItem, error) {
	if len(items) == 0 {
		return nil, geo.ErrEmptyPolicy
	}
	if len(items) > geo.MaxPolicies {
		return nil, geo.ErrTooManyPolicies
	}
	seen := make(map[int64]struct{}, len(items))
	out := make([]geo.PolicyItem, 0, len(items))
	for _, item := range items {
		if item.ProviderID <= 0 {
			return nil, geo.ErrBadProvider
		}
		if _, ok := seen[item.ProviderID]; ok {
			return nil, geo.ErrDuplicateProvider
		}
		seen[item.ProviderID] = struct{}{}
		out = append(out, item)
	}
	return out, nil
}

// applyTenantPolicy 把租户策略叠到 Provider 上（无策略则保留全局默认）。
func applyTenantPolicy(p geo.Provider, operatorCode string, pol *tenantPolicy) geo.Provider {
	p.OperatorCode = geo.NormalizeOperatorCode(operatorCode)
	if pol != nil {
		p.IsEnabled = pol.IsEnabled
		p.Priority = pol.Priority
	}
	return p
}

func sortProviders(list []geo.Provider) {
	sort.Slice(list, func(i, j int) bool {
		if list[i].Priority != list[j].Priority {
			return list[i].Priority < list[j].Priority
		}
		return list[i].ID < list[j].ID
	})
}

// filterEnabled 仅保留 IsEnabled=true 的源（Search 选源用）。
func filterEnabled(list []geo.Provider) []geo.Provider {
	out := make([]geo.Provider, 0, len(list))
	for _, p := range list {
		if p.IsEnabled {
			out = append(out, p)
		}
	}
	return out
}

// overlayList 对全局 Provider 列表逐条叠租户策略，并按优先级排序。
func overlayList(list []geo.Provider, operatorCode string, policies map[int64]tenantPolicy) []geo.Provider {
	out := make([]geo.Provider, 0, len(list))
	for _, p := range list {
		var pol *tenantPolicy
		if x, ok := policies[p.ID]; ok {
			cp := x
			pol = &cp
		}
		out = append(out, applyTenantPolicy(p, operatorCode, pol))
	}
	sortProviders(out)
	return out
}
