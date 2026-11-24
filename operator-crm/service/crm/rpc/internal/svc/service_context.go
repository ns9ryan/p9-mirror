package svc

import (
	"context"
	"strings"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"

	"oa.98ent.com/p9/common/modules/fieldcrypto"
	"oa.98ent.com/p9/core/common/coreadapt"
	"oa.98ent.com/p9/core/common/middleware"
	"oa.98ent.com/p9/core/rpc/coreclient"
	corepb "oa.98ent.com/p9/core/rpc/types/core"
	"oa.98ent.com/p9/integration/rpc/integrationclient"
	"oa.98ent.com/p9/integration/rpc/types/integration"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/ent/migrate"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/config"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/dao"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/dao/agent"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/dao/auth"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/dao/member"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/kafka"
	"oa.98ent.com/p9/platform-base/rpc/client/defaultavatarframeservice"
	"oa.98ent.com/p9/platform-base/rpc/client/defaultavatarservice"
	"oa.98ent.com/p9/platform-base/rpc/client/nicknameservice"
)

// ServiceContext CRM RPC服务上下文
type ServiceContext struct {
	ExportAuth        middleware.Client
	IntegrationRpc    integrationclient.Integration
	Nicknames         nicknameservice.NicknameService
	Config            config.Config
	Cryptor           *fieldcrypto.Cryptor
	Avatars           defaultavatarservice.DefaultAvatarService
	AvatarFrames      defaultavatarframeservice.DefaultAvatarFrameService
	DB                *ent.Client           // Ent数据库客户端
	Redis             redis.UniversalClient // 令牌黑名单
	KafkaEventHandler *kafka.EventHandler   // Kafka事件处理器

	// 业务数据访问
	DAO    *dao.Group      // 通用业务DAO
	Member *member.Service // 会员管理
	Auth   *auth.Service   // 代理鉴权
	Agent  *agent.Service  // 代理管理
}

// NewServiceContext 创建CRM RPC服务上下文
func NewServiceContext(c config.Config) *ServiceContext {
	zrpc.DontLogClientContentForMethod(corepb.Core_CheckToken_FullMethodName)
	var exportAuth middleware.Client
	if len(c.OperatorCoreRpc.Endpoints) > 0 || c.OperatorCoreRpc.Target != "" || c.OperatorCoreRpc.Etcd.Key != "" {
		exportAuth = coreadapt.Auth(coreclient.NewCore(zrpc.MustNewClient(c.OperatorCoreRpc)))
	}
	var cryptor *fieldcrypto.Cryptor
	if c.FieldCrypto.EncryptionKey != "" || c.FieldCrypto.LookupKey != "" {
		var err error
		cryptor, err = fieldcrypto.New([]byte(c.FieldCrypto.EncryptionKey), []byte(c.FieldCrypto.LookupKey))
		logx.Must(err)
	}
	var nicknames nicknameservice.NicknameService
	var avatars defaultavatarservice.DefaultAvatarService
	var frames defaultavatarframeservice.DefaultAvatarFrameService
	if len(c.PlatformBaseRpc.Endpoints) > 0 || c.PlatformBaseRpc.Target != "" || c.PlatformBaseRpc.Etcd.Key != "" {
		cli := zrpc.MustNewClient(c.PlatformBaseRpc)
		nicknames = nicknameservice.NewNicknameService(cli)
		avatars = defaultavatarservice.NewDefaultAvatarService(cli)
		frames = defaultavatarframeservice.NewDefaultAvatarFrameService(cli)
	}
	// 创建数据库驱动
	driver, err := c.DatabaseConf.NewDriver()
	logx.Must(err)

	// 创建Ent客户端配置
	entOpts := []ent.Option{
		ent.Log(func(values ...any) {
			for _, value := range values {
				if message, ok := value.(string); ok && strings.Contains(message, "member_contact_binding_code") {
					return
				}
			}
			logx.Info(values...)
		}), // 使用go-zero日志输出SQL
		ent.Driver(driver), // 设置数据库驱动
	}

	// 调试环境开启Ent调试模式
	if c.IsDebug() {
		entOpts = append(entOpts, ent.Debug())
	}

	// 创建Ent数据库客户端
	db := ent.NewClient(entOpts...)

	// 创建数据库表
	logx.Must(db.Schema.Create(
		context.Background(),
		migrate.WithDropIndex(true),    // 允许删除旧索引
		migrate.WithDropColumn(true),   // 允许删除旧字段
		migrate.WithForeignKeys(false), // 不创建数据库外键约束
	))

	var rdb redis.UniversalClient
	if c.DataRedis.Host != "" {
		rdb = redis.NewClient(&redis.Options{
			Addr:     c.DataRedis.Host,
			Password: c.DataRedis.Pass,
			DB:       c.DataRedis.DB,
		})
	}

	// 创建业务数据访问
	daoGroup := dao.New(db)
	memberServ := member.New(db, &daoGroup.MemberAutoLevel, &daoGroup.MemberVipLevel, driver)
	// 创建代理管理服务
	agentServ := agent.New(db)

	// 创建 Kafka 生产者
	var kafkaProducer *kafka.Producer
	if len(c.Kafka.Brokers) > 0 {
		kafkaProducer = kafka.NewProducer(c.Kafka.Brokers)
	}

	// 创建事件处理器
	eventHandler := kafka.NewEventHandler(kafkaProducer, memberServ)

	var integrationRpc integrationclient.Integration
	if len(c.IntegrationRpc.Endpoints) != 0 || c.IntegrationRpc.Target != "" || c.IntegrationRpc.Etcd.Key != "" {
		zrpc.DontLogClientContentForMethod(integration.Integration_SendEmail_FullMethodName)
		zrpc.DontLogClientContentForMethod(integration.Integration_SendSms_FullMethodName)
		integrationRpc = integrationclient.NewIntegration(zrpc.MustNewClient(c.IntegrationRpc))
	}
	return &ServiceContext{
		ExportAuth:     exportAuth,
		IntegrationRpc: integrationRpc,
		Config:         c, Cryptor: cryptor, Avatars: avatars, AvatarFrames: frames, Nicknames: nicknames,
		DB:                db,
		Redis:             rdb,
		KafkaEventHandler: eventHandler,
		DAO:               daoGroup,
		Member:            memberServ,
		Auth: auth.New(db, agentServ, rdb, auth.Config{
			AccessSecret:  c.AgentAdminJwt.AccessSecret,
			AccessExpire:  c.AgentAdminJwt.AccessExpire,
			RefreshSecret: c.AgentAdminJwt.RefreshSecret,
			RefreshExpire: c.AgentAdminJwt.RefreshExpire,
		}),
		Agent: agentServ,
	}
}
