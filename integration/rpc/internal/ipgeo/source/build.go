package source

import (
	"fmt"

	"oa.98ent.com/p9/integration/rpc/internal/ipgeo/geo"
)

// BuildSource 按 Provider 构造 Source（无本地表）。
func BuildSource(p geo.Provider) (geo.Source, error) {
	return Build(p, nil)
}

// Build 按 Driver 构造 Source；local 驱动需 lookup。
func Build(p geo.Provider, lookup LocalLookup) (geo.Source, error) {
	p = geo.NormalizeProvider(p)
	if err := geo.ValidateProvider(p); err != nil {
		return nil, err
	}
	switch p.Driver {
	case geo.DriverHTTP:
		return NewHTTPSource(p), nil
	case geo.DriverMMDB:
		return OpenNamedMMDB(p.Name, p.Endpoint)
	case geo.DriverIP2Region:
		return OpenXDB(p.Name, p.Endpoint)
	case geo.DriverLocal:
		return NewLocalSource(p.Name, lookup), nil
	default:
		return nil, fmt.Errorf("%w: %s", geo.ErrBadDriver, p.Driver)
	}
}
