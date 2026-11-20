package temporal

import (
	"fmt"
	"sync"

	"go.temporal.io/sdk/client"
)

var (
	temporalClient client.Client
	clientOnce     sync.Once
	clientErr      error
	address        string
	namespace      string
)

// Init 由 RPC 进程在加载配置后调用
func Init(addr, ns string) {
	address = addr
	namespace = ns
}

// GetClient 获取 Temporal 客户端（单例模式）
func GetClient() (client.Client, error) {
	clientOnce.Do(func() {
		if address == "" {
			clientErr = fmt.Errorf("temporal client not initialized")
			return
		}
		temporalClient, clientErr = client.Dial(client.Options{
			HostPort:  address,
			Namespace: namespace,
		})
	})
	return temporalClient, clientErr
}

func CloseClient() {
	if temporalClient != nil {
		temporalClient.Close()
	}
}
