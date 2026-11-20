package schema

// ExtraJSON 存 ip_geo_provider.extra_json（field_map / headers）。
type ExtraJSON struct {
	FieldMap map[string]string `json:"field_map,omitempty"`
	Headers  map[string]string `json:"headers,omitempty"`
}
