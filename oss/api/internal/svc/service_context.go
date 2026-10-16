package svc

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"oa.98ent.com/p9/oss/api/internal/config"
	"oa.98ent.com/p9/oss/api/internal/convert"
	"oa.98ent.com/p9/oss/api/internal/locales"
	"oa.98ent.com/p9/oss/api/internal/middleware"
	shared "oa.98ent.com/p9/oss/common/config"
	"oa.98ent.com/p9/oss/common/oss"
	"oa.98ent.com/p9/oss/rpc/ossclient"

	"oa.98ent.com/p9/common/i18n"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/core/rpc/coreclient"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ServiceContext struct {
	Config       config.Config
	Client       *oss.Client
	Oss          ossclient.Oss
	Core         coreclient.Core // 可空：未配 core_rpc 时不鉴权、不注册 catalog
	Jwt          rest.Middleware
	Authority    rest.Middleware
	ActionLog    rest.Middleware
	Upload       shared.UploadSettings
	Settings     shared.Settings
	ConfigSource string
	Trans        *i18n.Translator
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 创建API翻译器
	trans, err := i18n.New(c.App.I18n, locales.FS)
	logx.Must(err)
	settings := c.App.Settings()
	if err := settings.Validate(); err != nil {
		log.Fatal(err)
	}
	cli, err := oss.New(settings.OSS)
	if err != nil {
		log.Fatalf("init storage: %v", err)
	}
	coreCli, jwtMW, authMW := newAuth(c)
	s := &ServiceContext{
		Config: c, Client: cli, Upload: settings.Upload,
		Settings: shared.CloneSettings(settings),
		Core:     coreCli, Jwt: jwtMW, Authority: authMW,
		ActionLog:    middleware.ActionLogFromCore(coreCli),
		ConfigSource: configSource(),
		Trans:        trans,
	}
	if len(c.OssRpc.Endpoints) > 0 {
		s.Oss = ossclient.NewOss(zrpc.MustNewClient(c.OssRpc))
		log.Printf("档案 RPC %v", c.OssRpc.Endpoints)
	} else {
		log.Printf("未配置 oss_rpc：只上传对象，不写档案")
	}
	watchStorage(s)
	return s
}

func watchStorage(s *ServiceContext) {
	shared.OnChange(func(next shared.Settings) {
		if err := s.Client.Apply(next.OSS); err != nil {
			logx.Errorf("hot reload oss skipped: %v", err)
			return
		}
		s.Upload = next.Upload
		s.Settings = shared.CloneSettings(next)
	})
}

func (s *ServiceContext) Archive() (ossclient.Oss, error) {
	if s.Oss == nil {
		return nil, xerr.Err(http.StatusServiceUnavailable, "archive not configured")
	}
	return s.Oss, nil
}

func (s *ServiceContext) Store(ctx context.Context) (*oss.Client, shared.UploadSettings, error) {
	if s.Oss == nil {
		return s.Client, s.Upload, nil
	}
	p, err := s.Oss.GetStorageConfig(ctx, &ossclient.Empty{})
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			return s.Client, s.Upload, nil
		}
		return nil, shared.UploadSettings{}, err
	}
	if p == nil || !strings.HasPrefix(p.GetSource(), "provider:") {
		return s.Client, s.Upload, nil
	}
	over, err := shared.ApplyProvider(s.Settings, convert.SpecFromProvider(p))
	if err != nil {
		return nil, shared.UploadSettings{}, err
	}
	cli, err := oss.New(over.OSS)
	if err != nil {
		return nil, shared.UploadSettings{}, err
	}
	return cli, over.Upload, nil
}

func configSource() string {
	if strings.TrimSpace(os.Getenv("ETCD_CONFIG_KEY")) != "" {
		return "etcd"
	}
	return "yaml"
}

func (s *ServiceContext) Remember(ctx context.Context, rec *ossclient.FileRecordReq) {
	if s.Oss == nil || rec == nil {
		return
	}
	if _, err := s.Oss.RecordFile(ctx, rec); err != nil {
		logx.Errorf("archive file %s: %v", rec.GetKey(), err)
	}
}

func newAuth(c config.Config) (coreclient.Core, rest.Middleware, rest.Middleware) {
	coreCli, jwtMW, authMW := middleware.AuthFromRPC(c.CoreRpc.Endpoints, c.CoreRpc.Timeout)
	if coreCli == nil {
		log.Printf("JWT/Authority 未启用：未配置 core_rpc.endpoints")
		return nil, jwtMW, authMW
	}
	log.Printf("JWT+Authority 已启用，core RPC %v", c.CoreRpc.Endpoints)
	return coreCli, jwtMW, authMW
}
