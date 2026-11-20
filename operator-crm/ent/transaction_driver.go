package ent

import "entgo.io/ent/dialect"

// DialectDriver exposes the current transaction for tables managed by SQL migrations.
func (transaction *Tx) DialectDriver() dialect.Driver {
	return transaction.config.driver
}
