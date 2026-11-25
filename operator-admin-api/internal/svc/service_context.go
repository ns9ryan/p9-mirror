// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package svc

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"

	"oa.98ent.com/p9/common/i18n"
	"oa.98ent.com/p9/core/common/coreadapt"
	"oa.98ent.com/p9/core/common/middleware"
	"oa.98ent.com/p9/core/rpc/coreclient"
	corepb "oa.98ent.com/p9/core/rpc/types/core"
	"oa.98ent.com/p9/integration/rpc/integrationclient"
	"oa.98ent.com/p9/operator-activity/client/activityservice"
	"oa.98ent.com/p9/operator-activity/client/taskservice"
	"oa.98ent.com/p9/operator-admin-api/internal/config"
	"oa.98ent.com/p9/operator-admin-api/internal/kafka"
	"oa.98ent.com/p9/operator-admin-api/internal/locales"
	"oa.98ent.com/p9/operator-base/rpc/client/operatorservice"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/client/agentdepositorderservice"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/client/agentservice"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/client/member"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/client/memberautolevel"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/client/memberfixedlevel"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/client/turnoverservice"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/client/vip"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/memberexportpb"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/client/memberwallet"
	"oa.98ent.com/p9/platform-base/rpc/client/currencyservice"
	"oa.98ent.com/p9/platform-base/rpc/client/vipdefaulticonservice"
)

// ServiceContext 分站后台服务上下文
type ServiceContext struct {
	MemberExportRpc memberexportpb.MemberExportClient
	ExportRedis     redis.UniversalClient
	// 基础依赖
	Config config.Config    // 服务配置
	Trans  *i18n.Translator // 多语言翻译器

	// 基础服务RPC
	OperatorCoreRpc coreclient.Core                 // Core RPC
	OperatorBaseRpc operatorservice.OperatorService // 分站基础RPC

	// CRM业务RPC
	AgentRpc             agentservice.AgentService                         // 分层代理RPC
	AgentDepositOrderRpc agentdepositorderservice.AgentDepositOrderService // 代理人工存款订单RPC
	TurnoverRpc          turnoverservice.TurnoverService                   // 代理流水RPC
	MemberRpc            member.Member                                     // 会员管理RPC
	MemberAutoLevelRpc   memberautolevel.MemberAutoLevel                   // 会员自动层级RPC
	MemberFixedLevelRpc  memberfixedlevel.MemberFixedLevel                 // 会员固定层级RPC
	MemberVipRpc         vip.Vip                                           // 会员VIP配置、等级和图标RPC

	// 钱包业务RPC
	MemberWalletRpc memberwallet.MemberWallet // 会员钱包RPC

	// 活动服务 RPC
	ActivityRpc activityservice.ActivityService // 活动管理RPC

	// 任务服务 RPC
	TaskServiceRpc taskservice.TaskService // 任务配置RPC
	// 总站素材RPC，仅在VIP列表请求中查询默认图。
	VipDefaultIconRpc vipdefaulticonservice.VipDefaultIconService
	// 总站货币RPC，用于查询货币换算倍率（amount_factor）。
	CurrencyRpc currencyservice.CurrencyService

	// Kafka事件处理器
	KafkaEventHandler *kafka.EventHandler

	// 第三方服务集成 RPC
	IntegrationRpc integrationclient.Integration

	// 认证权限
	Jwt       rest.Middleware // JWT认证
	Authority rest.Middleware // 权限校验

	// 日志
	ActionLog rest.Middleware // 操作日志
	ErrorLog  rest.Middleware // 错误日志
}

// NewServiceContext 创建分站后台服务上下文
func NewServiceContext(c config.Config) *ServiceContext {
	zrpc.DontLogClientContentForMethod(corepb.Core_CheckToken_FullMethodName)
	zrpc.DontLogClientContentForMethod(crmrpc.Member_UpdateContact_FullMethodName)
	zrpc.DontLogClientContentForMethod(crmrpc.Member_ListMemberContactBindingCodes_FullMethodName)
	zrpc.DontLogClientContentForMethod(crmrpc.Member_SendMemberContactBindingCode_FullMethodName)

	// 初始化API翻译器
	trans, err := i18n.New(c.I18n, locales.FS)
	logx.Must(err)

	// 创建Core RPC客户端
	coreClient := zrpc.MustNewClient(c.OperatorCoreRpc)
	coreRpc := coreclient.NewCore(coreClient)

	// 设置Core字典加载器和认证适配器
	coreadapt.SetDictLoader(coreRpc)
	auth := coreadapt.Auth(coreRpc)

	// 创建业务RPC客户端
	baseClient := zrpc.MustNewClient(c.OperatorBaseRpc)
	crmClient := zrpc.MustNewClient(c.CrmRpc)
	walletClient := zrpc.MustNewClient(c.MemberWalletRpc)
	activityClient := zrpc.MustNewClient(c.ActivityRpc)

	// 创建第三方服务集成客户端
	integrationClient := zrpc.MustNewClient(c.IntegrationRpc)

	var exportRedis redis.UniversalClient
	if c.MemberExport.RedisAddr != "" {
		exportRedis = redis.NewClient(&redis.Options{Addr: c.MemberExport.RedisAddr, Password: c.MemberExport.RedisPassword, DB: c.MemberExport.RedisDB, MaxRetries: -1, DialerRetries: 1, ContextTimeoutEnabled: true, PoolTimeout: 2 * time.Second, DialTimeout: 2 * time.Second, ReadTimeout: 2 * time.Second, WriteTimeout: 2 * time.Second})
	}
	// 创建 Kafka 生产者
	var kafkaProducer *kafka.Producer
	if len(c.Kafka.Brokers) > 0 {
		kafkaProducer = kafka.NewProducer(c.Kafka.Brokers)
	}
	eventHandler := kafka.NewEventHandler(
		kafkaProducer,
		member.NewMember(crmClient),
		agentdepositorderservice.NewAgentDepositOrderService(crmClient),
		taskservice.NewTaskService(activityClient),
	)

	// 创建可选的总站基础 RPC 客户端（VIP默认图标、货币倍率共用同一连接）
	platformBaseClient := newPlatformBaseClient(c.PlatformBaseRpc)

	var vipIconRpc vipdefaulticonservice.VipDefaultIconService
	var currencyRpc currencyservice.CurrencyService
	if platformBaseClient != nil {
		vipIconRpc = vipdefaulticonservice.NewVipDefaultIconService(platformBaseClient)
		currencyRpc = currencyservice.NewCurrencyService(platformBaseClient)
	}

	return &ServiceContext{
		KafkaEventHandler: eventHandler,
		MemberExportRpc:   memberexportpb.NewMemberExportClient(crmClient.Conn()),
		ExportRedis:       exportRedis,
		Config:            c,
		Trans:             trans,

		OperatorCoreRpc: coreRpc,
		OperatorBaseRpc: operatorservice.NewOperatorService(baseClient),

		AgentRpc:             agentservice.NewAgentService(crmClient),
		AgentDepositOrderRpc: agentdepositorderservice.NewAgentDepositOrderService(crmClient),
		TurnoverRpc:          turnoverservice.NewTurnoverService(crmClient),
		MemberRpc:            member.NewMember(crmClient),
		MemberAutoLevelRpc:   memberautolevel.NewMemberAutoLevel(crmClient),
		MemberFixedLevelRpc:  memberfixedlevel.NewMemberFixedLevel(crmClient),
		MemberVipRpc:         vip.NewVip(crmClient),

		MemberWalletRpc: memberwallet.NewMemberWallet(walletClient),

		ActivityRpc:    activityservice.NewActivityService(activityClient),
		TaskServiceRpc: taskservice.NewTaskService(activityClient),
		VipDefaultIconRpc: vipIconRpc, // 可选直连，不查询总站。
		CurrencyRpc:       currencyRpc, // 可选直连，查询货币换算倍率。

		IntegrationRpc: integrationclient.NewIntegration(integrationClient),

		Jwt:       middleware.JWT(auth),
		Authority: middleware.Authority(auth),

		ActionLog: middleware.ActionLog(contactActionRecorder{coreadapt.ActionRecorder(coreRpc)}),
		ErrorLog:  middleware.ErrorLog(c.Name, exportErrorRecorder{coreadapt.ErrorRecorder(coreRpc)}),
	}
}

// newPlatformBaseClient 创建可选的总站基础 RPC 客户端；缺配置或创建失败返回 nil。
// VIP默认图标、货币倍率等总站基础查询共用此连接。
func newPlatformBaseClient(conf zrpc.RpcClientConf) zrpc.Client {
	// 两种连接方式都没配置时跳过，不影响服务启动。
	if len(conf.Endpoints) == 0 && conf.Etcd.Key == "" {
		return nil
	}
	conf.NonBlock = true // 总站暂时不可达时继续启动，在请求阶段返回错误。
	client, err := zrpc.NewClient(conf)
	if err != nil {
		logx.Errorw("创建总站基础RPC客户端失败", logx.Field("error", err))
		return nil
	}
	return client
}

type contactActionRecorder struct {
	middleware.ActionRecorder
}

func (recorder contactActionRecorder) RecordAction(ctx context.Context, record middleware.ActionRecord) {
	switch record.RequestPath {
	case "/member/export/v2":
		record.RequestQuery = ""
		record.RequestBody = ""
		record.ResponseBody = ""
	case "/member/contact/update", "/member/contact/binding-codes/send":
		record.RequestBody = ""
	}
	recorder.ActionRecorder.RecordAction(ctx, record)
}
