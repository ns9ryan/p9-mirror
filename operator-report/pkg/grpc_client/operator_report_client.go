package grpc_client

import (
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"oa.98ent.com/p9/operator-report/rpc/client/statisticsservice"
)

// ClientManager gRPC客户端管理器
type OperatorReportClientManager struct {
	statisticsServiceClient statisticsservice.StatisticsService
}

// 创建新的OperatorReport gRPC客户端管理器
func NewOperatorReportClientManager(cfg zrpc.RpcClientConf) (*OperatorReportClientManager, error) {
	logx.Infof("create version 1.0.0 operator report client manager")
	client := zrpc.MustNewClient(
		cfg,
	)

	return &OperatorReportClientManager{
		statisticsServiceClient: statisticsservice.NewStatisticsService(client),
	}, nil
}

// GetStatisticsServiceClient 获取统计服务客户端
func (cm *OperatorReportClientManager) GetStatisticsServiceClient() statisticsservice.StatisticsService {
	return cm.statisticsServiceClient
}
