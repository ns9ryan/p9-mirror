package geo

import (
	"fmt"
	"strings"
)

// ValidateProvider 校验 Provider 必填与 driver/tag。
func ValidateProvider(p Provider) error {
	if strings.TrimSpace(p.Name) == "" {
		return ErrEmptyName
	}
	switch p.Driver {
	case DriverHTTP, DriverMMDB, DriverIP2Region, DriverLocal:
	default:
		return fmt.Errorf("%w: %s", ErrBadDriver, p.Driver)
	}
	if p.Driver != DriverLocal && strings.TrimSpace(p.Endpoint) == "" {
		return ErrEmptyEndpoint
	}
	if p.Driver == DriverHTTP && p.AuthType != "" {
		switch p.AuthType {
		case AuthNone, AuthBasic, AuthBearer, AuthHeader:
		default:
			return fmt.Errorf("%w: auth_type %s", ErrBadProvider, p.AuthType)
		}
	}
	if p.Tag == "" {
		p.Tag = TagFromDriver(p.Driver)
	}
	switch p.Tag {
	case TagOnline, TagLocal, TagOffline:
	default:
		return fmt.Errorf("%w: %s", ErrBadTag, p.Tag)
	}
	return nil
}

// NormalizeProvider 规范化并填默认 tag/timeout。
func NormalizeProvider(p Provider) Provider {
	p.Name = strings.TrimSpace(p.Name)
	p.Driver = strings.TrimSpace(p.Driver)
	p.Endpoint = strings.TrimSpace(p.Endpoint)
	p.AuthType = strings.TrimSpace(p.AuthType)
	p.OperatorCode = NormalizeOperatorCode(p.OperatorCode)
	if p.AuthType == "" {
		p.AuthType = AuthNone
	}
	p.Tag = strings.TrimSpace(p.Tag)
	if p.Tag == "" {
		p.Tag = TagFromDriver(p.Driver)
	}
	if p.TimeoutSec <= 0 {
		p.TimeoutSec = 5
	}
	if p.FieldMap == nil {
		p.FieldMap = map[string]string{}
	}
	if p.Headers == nil {
		p.Headers = map[string]string{}
	}
	return p
}

// TagFromDriver 按 driver 推断默认标签。
func TagFromDriver(driver string) string {
	switch driver {
	case DriverLocal:
		return TagLocal
	case DriverHTTP:
		return TagOnline
	case DriverMMDB, DriverIP2Region:
		return TagOffline
	default:
		return ""
	}
}
