// Package db 打开 PostgreSQL 并执行 ent Schema 迁移（由 svc 调用，不属于业务域包）。
package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"oa.98ent.com/p9/integration/rpc/ent"
	"oa.98ent.com/p9/integration/rpc/ent/migrate"
	_ "oa.98ent.com/p9/integration/rpc/ent/runtime"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/lib/pq"
	"github.com/zeromicro/go-zero/core/logx"
)

var dbNameRe = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`)

// Open 连接 Postgres；库不存在则创建，并 Schema.Create 建表。
func Open(dsn string) (*ent.Client, error) {
	if err := pingOrCreate(dsn); err != nil {
		return nil, err
	}
	drv, err := entsql.Open(dialect.Postgres, dsn)
	if err != nil {
		return nil, err
	}
	sqlDB := drv.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(50)
	client := ent.NewClient(ent.Driver(drv))
	// 只建表/加列；不 DropColumn，避免生产误删字段。
	if err := client.Schema.Create(context.Background(), migrate.WithDropIndex(true)); err != nil {
		_ = client.Close()
		return nil, err
	}
	return client, nil
}

func pingOrCreate(dsn string) error {
	db, err := openPing(dsn)
	if isDBMissing(err) {
		if err := ensureDatabase(dsn); err != nil {
			return err
		}
		db, err = openPing(dsn)
	}
	if err != nil {
		return err
	}
	return db.Close()
}

func openPing(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func ensureDatabase(dsn string) error {
	name := dbNameFromDSN(dsn)
	if !dbNameRe.MatchString(name) {
		return fmt.Errorf("invalid dbname %q", name)
	}
	admin, err := openAdmin(dsn)
	if err != nil {
		return err
	}
	defer admin.Close()
	_, err = admin.Exec("CREATE DATABASE " + pq.QuoteIdentifier(name))
	if err != nil && !isDBExists(err) {
		return fmt.Errorf("create database %s: %w", name, err)
	}
	logx.Infof("created database %s", name)
	return nil
}

func openAdmin(dsn string) (*sql.DB, error) {
	var last error
	for _, name := range []string{"postgres", "template1"} {
		db, err := openPing(withDBName(dsn, name))
		if err == nil {
			return db, nil
		}
		last = err
	}
	return nil, last
}

func isDBMissing(err error) bool {
	var e *pq.Error
	return errors.As(err, &e) && e.Code == "3D000"
}

func isDBExists(err error) bool {
	var e *pq.Error
	return errors.As(err, &e) && e.Code == "42P04"
}

func dbNameFromDSN(dsn string) string {
	if u, err := url.Parse(dsn); err == nil && u.Scheme != "" && u.Path != "" && u.Path != "/" {
		return strings.TrimPrefix(u.Path, "/")
	}
	for _, part := range strings.Fields(dsn) {
		k, v, ok := strings.Cut(part, "=")
		if ok && strings.EqualFold(k, "dbname") {
			return v
		}
	}
	return ""
}

func withDBName(dsn, name string) string {
	if u, err := url.Parse(dsn); err == nil && u.Scheme != "" {
		u.Path = "/" + name
		return u.String()
	}
	parts := strings.Fields(dsn)
	for i, part := range parts {
		k, _, ok := strings.Cut(part, "=")
		if ok && strings.EqualFold(k, "dbname") {
			parts[i] = "dbname=" + name
			return strings.Join(parts, " ")
		}
	}
	return strings.TrimSpace(dsn + " dbname=" + name)
}
