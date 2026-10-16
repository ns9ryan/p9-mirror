package main

//go:generate protoc --go_out=./types/oss --go_opt=paths=source_relative --go-grpc_out=./types/oss --go-grpc_opt=paths=source_relative oss.proto

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	shared "oa.98ent.com/p9/oss/common/config"
	"oa.98ent.com/p9/oss/rpc/internal/config"
	"oa.98ent.com/p9/oss/rpc/internal/server"
	"oa.98ent.com/p9/oss/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/proc"
)

var configFile = flag.String("f", "rpc/etc/oss.yaml", "RPC 配置；空或已设 ETCD_CONFIG_KEY 则读 etcd")

func main() {
	flag.Parse()
	path := strings.TrimSpace(*configFile)
	if os.Getenv("ETCD_CONFIG_KEY") != "" || path == "" {
		path = ""
	}

	c := config.MustLoad(path)
	ctx := svc.MustInit(c)
	proc.AddShutdownListener(shared.Close)

	if c.ListenOn == "" {
		log.Fatal("rpc.listen_on is empty")
	}
	s := server.NewZrpcServer(c, ctx)
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
