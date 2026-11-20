package svc

import (
	"context"
	"path"
	"sync"

	"oa.98ent.com/p9/oss/common/config"
	"oa.98ent.com/p9/oss/rpc/ent"
	"oa.98ent.com/p9/oss/rpc/internal/store"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	SourceYAML     = "yaml"
	SourceEtcd     = "etcd"
	SourceProvider = "provider"
)

// ServiceContext 档案进程：库 + InitTenant 模板。不连对象存储，不读 upload/oss。
type ServiceContext struct {
	DB *ent.Client

	mu      sync.RWMutex
	source  string
	initTpl config.ProviderSpec
}

func NewServiceContext() *ServiceContext {
	return &ServiceContext{source: SourceYAML}
}

func (s *ServiceContext) SetDB(db *ent.Client) {
	s.DB = db
}

func (s *ServiceContext) SetSource(source string) {
	s.mu.Lock()
	s.source = source
	s.mu.Unlock()
}

func (s *ServiceContext) SetInitTemplate(tpl config.ProviderSpec) {
	s.mu.Lock()
	s.initTpl = config.CloneProviderSpec(tpl)
	s.mu.Unlock()
}

func (s *ServiceContext) InitTemplate() config.ProviderSpec {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return config.CloneProviderSpec(s.initTpl)
}

func (s *ServiceContext) Source() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.source
}

func (s *ServiceContext) RememberFile(ctx context.Context, rec store.FileRecord) {
	if s.DB == nil {
		return
	}
	if rec.Name == "" {
		rec.Name = path.Base(rec.ObjectKey)
	}
	if err := store.UpsertFile(ctx, s.DB, rec); err != nil {
		logx.Errorf("archive file %s: %v", rec.ObjectKey, err)
	}
}
