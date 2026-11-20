package nickname

import (
	"context"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"oa.98ent.com/p9/platform-base/rpc/ent"
	_ "oa.98ent.com/p9/platform-base/rpc/ent/runtime"
	"testing"
	"time"
)

func setup(t *testing.T) (*Store, sqlmock.Sqlmock, *miniredis.Miniredis) {
	t.Helper()
	db, m, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	c := ent.NewClient(ent.Driver(sql.OpenDB(dialect.Postgres, db)))
	r := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: r.Addr(), MaxRetries: -1})
	t.Cleanup(func() {
		client.Close()
		c.Close()
		if e := m.ExpectationsWereMet(); e != nil {
			t.Error(e)
		}
	})
	return New(c, client, time.Hour), m, r
}
func expectPool(m sqlmock.Sqlmock, id int64, name string) {
	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"})
	if id > 0 {
		rows.AddRow(id, time.Now(), time.Now(), name)
	}
	m.ExpectQuery(`SELECT .* FROM "nickname"`).WillReturnRows(rows)
}
func TestCacheLifecycle(t *testing.T) {
	s, m, r := setup(t)
	ctx := context.Background()
	expectPool(m, 1, "旧昵称")
	rows, err := s.All(ctx)
	if err != nil || len(rows) != 1 || rows[0].Name != "旧昵称" {
		t.Fatalf("load: %v %v", rows, err)
	}
	if ttl := r.TTL(poolPrefix + "0"); ttl < time.Hour || ttl >= time.Hour+5*time.Minute {
		t.Fatalf("TTL %v", ttl)
	}
	// 第二次调用应命中 MessagePack 缓存，因此不设置数据库查询预期。
	if _, err = s.All(ctx); err != nil {
		t.Fatal(err)
	}
	s.Invalidate(ctx)
	expectPool(m, 2, "新昵称")
	rows, err = s.All(ctx)
	if err != nil || rows[0].Name != "新昵称" {
		t.Fatalf("invalidate: %v %v", rows, err)
	}
	// 缓存数据损坏时应从数据库重新加载。
	r.Set(poolPrefix+"1", "broken")
	expectPool(m, 2, "新昵称")
	if _, err = s.All(ctx); err != nil {
		t.Fatal(err)
	}
	s.Invalidate(ctx)
	expectPool(m, 0, "")
	rows, err = s.All(ctx)
	if err != nil || len(rows) != 0 {
		t.Fatalf("empty: %v %v", rows, err)
	}
	if value, _ := r.Get(poolPrefix + "2"); value != "NF" {
		t.Fatalf("negative marker %q", value)
	}
	if _, err = s.All(ctx); err != nil {
		t.Fatal(err)
	}
	r.FastForward(40 * time.Second)
	expectPool(m, 3, "新增昵称")
	rows, err = s.All(ctx)
	if err != nil || rows[0].ID != 3 {
		t.Fatalf("negative expiry: %v %v", rows, err)
	}
}
func TestRedisFailureFallsBackAndDBErrorIsNotCached(t *testing.T) {
	s, m, r := setup(t)
	r.SetError("unavailable")
	expectPool(m, 1, "昵称")
	if rows, err := s.All(context.Background()); err != nil || len(rows) != 1 {
		t.Fatalf("fallback: %v %v", rows, err)
	}
	r.SetError("")
	m.ExpectQuery(`SELECT .* FROM "nickname"`).WillReturnError(errors.New("db down"))
	if _, err := s.All(context.Background()); err == nil {
		t.Fatal("expected DB error")
	}
	expectPool(m, 1, "恢复")
	if _, err := s.All(context.Background()); err != nil {
		t.Fatal(err)
	}
}
func TestConcurrentOldLoadCannotPopulateNewRevision(t *testing.T) {
	s, m, r := setup(t)
	ctx := context.Background()
	m.ExpectQuery(`SELECT .* FROM "nickname"`).WillDelayFor(100 * time.Millisecond).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).AddRow(1, time.Now(), time.Now(), "旧昵称"))
	done := make(chan error, 1)
	go func() { _, err := s.All(ctx); done <- err }()
	// 等待回源任务开始执行 SQL 查询后再递增版本号。
	deadline := time.Now().Add(time.Second)
	for m.ExpectationsWereMet() != nil && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if err := m.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
	s.Invalidate(ctx)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if r.Exists(poolPrefix + "1") {
		t.Fatal("old load wrote new revision")
	}
	expectPool(m, 2, "新昵称")
	rows, err := s.All(ctx)
	if err != nil || rows[0].Name != "新昵称" {
		t.Fatalf("new revision: %v %v", rows, err)
	}
}
