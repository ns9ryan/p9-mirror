package schema

import "regexp"

// trimmedNonEmpty 对应 SQL 中「非空且等于 BTRIM」的约束。
// PostgreSQL 默认 BTRIM 只去掉空格。
var trimmedNonEmpty = regexp.MustCompile(`^[^ ](.*[^ ])?$`)

// regionCodePattern 对应 ISO 3166-1 alpha-2。
var regionCodePattern = regexp.MustCompile(`^[A-Z]{2}$`)
