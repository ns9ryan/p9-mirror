package cache

import (
	"context"
	"encoding/json"
	"time"

	"entgo.io/ent/dialect/sql"
	"oa.98ent.com/p9/integration/rpc/ent"
	"oa.98ent.com/p9/integration/rpc/ent/ipgeocache"
	"oa.98ent.com/p9/integration/rpc/ent/predicate"
	"oa.98ent.com/p9/integration/rpc/internal/ipgeo/geo"
)

// TableCache Postgres 表驱动缓存（ip_geo_cache）。
type TableCache struct {
	db *ent.Client // Ent 客户端，读写 ip_geo_cache
}

// NewTableCache 用 ent Client 建表缓存。
func NewTableCache(db *ent.Client) *TableCache {
	return &TableCache{db: db}
}

func (c *TableCache) Get(ctx context.Context, key string) (*geo.Record, error) {
	row, err := c.db.IpGeoCache.Get(ctx, key)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return fromTableCache(row)
}

func (c *TableCache) Set(ctx context.Context, rec *geo.Record) error {
	if rec == nil || rec.IP == "" {
		return nil
	}
	id := recordCacheID(rec)
	if id == "" {
		return nil
	}
	if rec.UpdatedAt.IsZero() {
		rec.UpdatedAt = time.Now()
	}
	raw, err := encodeRecord(rec)
	if err != nil {
		return err
	}
	b := c.db.IpGeoCache.Create().SetID(id).SetPayload(json.RawMessage(raw))
	if !rec.ExpiresAt.IsZero() {
		b = b.SetExpiresAt(rec.ExpiresAt)
	}
	u := b.OnConflictColumns(ipgeocache.FieldID).UpdateNewValues()
	if rec.ExpiresAt.IsZero() {
		u = u.ClearExpiresAt()
	}
	return u.Exec(ctx)
}

func (c *TableCache) Touch(ctx context.Context, key string, expiresAt time.Time) error {
	err := c.db.IpGeoCache.UpdateOneID(key).SetExpiresAt(expiresAt).Exec(ctx)
	if ent.IsNotFound(err) {
		return nil
	}
	return err
}

func (c *TableCache) DeleteByIP(ctx context.Context, ip string) error {
	if ip == "" {
		return nil
	}
	_, err := c.db.IpGeoCache.Delete().Where(cacheKeyByIP(ip)).Exec(ctx)
	return err
}

func cacheKeyByIP(ip string) predicate.IpGeoCache {
	return predicate.IpGeoCache(sql.FieldHasSuffix(ipgeocache.FieldID, "|"+ip))
}

func fromTableCache(row *ent.IpGeoCache) (*geo.Record, error) {
	rec, err := decodeRecord(row.Payload)
	if err != nil || rec == nil {
		return rec, err
	}
	if rec.IP == "" {
		_, rec.IP = parseCacheID(row.ID)
	}
	if row.ExpiresAt != nil {
		rec.ExpiresAt = *row.ExpiresAt
	}
	if rec.UpdatedAt.IsZero() {
		rec.UpdatedAt = row.UpdatedAt
	}
	return rec, nil
}
