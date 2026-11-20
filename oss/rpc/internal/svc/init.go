package svc

import (
	"log"

	"oa.98ent.com/p9/oss/common/config"
	"oa.98ent.com/p9/oss/rpc/ent"
	rpcconfig "oa.98ent.com/p9/oss/rpc/internal/config"
	"oa.98ent.com/p9/oss/rpc/internal/store"

	"github.com/zeromicro/go-zero/core/proc"
)

func MustInit(c rpcconfig.Config) *ServiceContext {
	settings := c.App.Settings()
	ctx := NewServiceContext()
	if c.FromFile {
		ctx.SetSource(SourceYAML)
	} else {
		ctx.SetSource(SourceEtcd)
	}
	log.Printf("档案配置来源: %s（不连对象存储）", ctx.Source())
	ctx.SetDB(mustOpenDB(settings.DatabaseDSN))
	ctx.SetInitTemplate(mustLoadInitTemplate())
	return ctx
}

func mustLoadInitTemplate() config.ProviderSpec {
	const name = "init_default_settings.yaml"
	for _, p := range []string{
		"rpc/etc/" + name,
		"etc/" + name,
		name,
	} {
		spec, err := config.LoadInitProviderTemplate(p)
		if err != nil {
			continue
		}
		log.Printf("租户初始化模板: %s", p)
		return spec
	}
	log.Fatal("缺少 init_default_settings.yaml（rpc/etc/）")
	return config.ProviderSpec{}
}

func mustOpenDB(dsn string) *ent.Client {
	db, err := store.Open(dsn)
	if err != nil {
		log.Fatalf("init database: %v", err)
	}
	if db != nil {
		log.Printf("档案库已连接，表 oss_file / oss_provider")
		proc.AddShutdownListener(func() { _ = db.Close() })
	}
	return db
}
