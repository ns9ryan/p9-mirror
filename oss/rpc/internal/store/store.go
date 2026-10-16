package store

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"oa.98ent.com/p9/oss/common/config"
	"oa.98ent.com/p9/oss/rpc/ent"
	entfile "oa.98ent.com/p9/oss/rpc/ent/file"
	"oa.98ent.com/p9/oss/rpc/ent/migrate"
	entprovider "oa.98ent.com/p9/oss/rpc/ent/provider"
	_ "oa.98ent.com/p9/oss/rpc/ent/runtime"

	"oa.98ent.com/p9/oss/common/oss"

	"oa.98ent.com/p9/common/ctxdata"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/lib/pq"
)

var ErrNoDatabase = errors.New("database not configured")

// Open 打开 PostgreSQL 并用 Ent 迁移表。dsn 为空返回 (nil, nil)，服务退化为无档案网关。
func Open(dsn string) (*ent.Client, error) {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		return nil, nil
	}
	drv, err := entsql.Open(dialect.Postgres, dsn)
	if err != nil {
		return nil, err
	}
	sqlDB := drv.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(50)
	if err := sqlDB.Ping(); err != nil {
		_ = drv.Close()
		return nil, err
	}
	client := ent.NewClient(ent.Driver(drv))
	attachTenant(client)
	if err := client.Schema.Create(context.Background(), migrate.WithDropIndex(true), migrate.WithDropColumn(true)); err != nil {
		_ = client.Close()
		return nil, err
	}
	return client, nil
}

func SpecOf(p *ent.Provider) (config.ProviderSpec, error) {
	spec := config.ProviderSpec{
		OSS: oss.Config{
			Driver:    p.Driver,
			Endpoint:  p.Endpoint,
			FileURL:   p.FileURL,
			AccessKey: p.AccessKey,
			SecretKey: p.SecretKey,
			PathStyle: p.PathStyle,
			Region:    p.Region,
		},
	}
	if strings.TrimSpace(p.BucketsJSON) != "" {
		if err := json.Unmarshal([]byte(p.BucketsJSON), &spec.OSS.Buckets); err != nil {
			return spec, err
		}
	}
	return spec, nil
}

func ApplySpec(p *ent.Provider, spec config.ProviderSpec) error {
	p.Driver = spec.OSS.Driver
	p.Endpoint = spec.OSS.Endpoint
	p.FileURL = spec.OSS.FileURL
	p.AccessKey = spec.OSS.AccessKey
	p.SecretKey = spec.OSS.SecretKey
	p.PathStyle = spec.OSS.PathStyle
	p.Region = spec.OSS.Region
	b, err := json.Marshal(spec.OSS.Buckets)
	if err != nil {
		return err
	}
	p.BucketsJSON = string(b)
	return nil
}

type FileRecord struct {
	ObjectKey   string
	Scene       string
	Bucket      string
	Name        string
	ContentType string
	Size        int64
}

func UpsertFile(ctx context.Context, db *ent.Client, rec FileRecord) error {
	if db == nil {
		return nil
	}
	rec.ObjectKey = strings.TrimPrefix(strings.TrimSpace(rec.ObjectKey), "/")
	if rec.ObjectKey == "" {
		return nil
	}
	if len(rec.Name) > 255 {
		rec.Name = rec.Name[:255]
	}
	return db.File.Create().
		SetObjectKey(rec.ObjectKey).
		SetScene(rec.Scene).
		SetBucket(rec.Bucket).
		SetName(rec.Name).
		SetSize(rec.Size).
		SetContentType(rec.ContentType).
		OnConflictColumns(entfile.FieldObjectKey).
		UpdateNewValues().
		Exec(ctx)
}

func ListFiles(ctx context.Context, db *ent.Client, scene string, page, pageSize int) ([]*ent.File, int, error) {
	if db == nil {
		return nil, 0, ErrNoDatabase
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	q := db.File.Query()
	if p := strings.ToLower(strings.TrimSpace(scene)); p != "" {
		q = q.Where(entfile.SceneEQ(p))
	}
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := q.Order(ent.Desc(entfile.FieldID)).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	return rows, total, err
}

func GetFileByKey(ctx context.Context, db *ent.Client, key string) (*ent.File, error) {
	if db == nil {
		return nil, ErrNoDatabase
	}
	key = strings.TrimPrefix(strings.TrimSpace(key), "/")
	return db.File.Query().Where(entfile.ObjectKeyEQ(key)).Only(ctx)
}

func DeleteFileByKey(ctx context.Context, db *ent.Client, key string) (int, error) {
	if db == nil {
		return 0, ErrNoDatabase
	}
	key = strings.TrimPrefix(strings.TrimSpace(key), "/")
	return db.File.Delete().Where(entfile.ObjectKeyEQ(key)).Exec(ctx)
}

func ListProviders(ctx context.Context, db *ent.Client) ([]*ent.Provider, error) {
	if db == nil {
		return nil, ErrNoDatabase
	}
	return db.Provider.Query().Order(ent.Asc(entprovider.FieldID)).All(ctx)
}

func GetProvider(ctx context.Context, db *ent.Client, id uint64) (*ent.Provider, error) {
	if db == nil {
		return nil, ErrNoDatabase
	}
	return db.Provider.Get(ctx, id)
}

func FindDefault(ctx context.Context, db *ent.Client) (*ent.Provider, error) {
	if db == nil {
		return nil, nil
	}
	c := ctxdata.ClaimsFromCtx(ctx)
	if c == nil || strings.TrimSpace(c.OperatorCode) == "" {
		return nil, nil
	}
	p, err := db.Provider.Query().Where(entprovider.IsDefaultEQ(true)).First(ctx)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	return p, err
}

func CreateProvider(ctx context.Context, db *ent.Client, p *ent.Provider) (*ent.Provider, error) {
	if db == nil {
		return nil, ErrNoDatabase
	}
	return db.Provider.Create().
		SetName(p.Name).
		SetIsDefault(p.IsDefault).
		SetDriver(p.Driver).
		SetEndpoint(p.Endpoint).
		SetFileURL(p.FileURL).
		SetAccessKey(p.AccessKey).
		SetSecretKey(p.SecretKey).
		SetPathStyle(p.PathStyle).
		SetRegion(p.Region).
		SetBucketsJSON(p.BucketsJSON).
		Save(ctx)
}

func SaveProvider(ctx context.Context, db *ent.Client, p *ent.Provider) error {
	if db == nil {
		return ErrNoDatabase
	}
	_, err := db.Provider.UpdateOneID(p.ID).
		SetName(p.Name).
		SetIsDefault(p.IsDefault).
		SetDriver(p.Driver).
		SetEndpoint(p.Endpoint).
		SetFileURL(p.FileURL).
		SetAccessKey(p.AccessKey).
		SetSecretKey(p.SecretKey).
		SetPathStyle(p.PathStyle).
		SetRegion(p.Region).
		SetBucketsJSON(p.BucketsJSON).
		Save(ctx)
	return err
}

func DeleteProvider(ctx context.Context, db *ent.Client, id uint64) error {
	if db == nil {
		return ErrNoDatabase
	}
	return db.Provider.DeleteOneID(id).Exec(ctx)
}

func SetDefault(ctx context.Context, db *ent.Client, id uint64) error {
	if db == nil {
		return ErrNoDatabase
	}
	tx, err := db.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.Provider.Update().Where(entprovider.IsDefaultEQ(true)).SetIsDefault(false).Save(ctx); err != nil {
		return err
	}
	n, err := tx.Provider.Update().Where(entprovider.IDEQ(id)).SetIsDefault(true).Save(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("provider not found")
	}
	return tx.Commit()
}
