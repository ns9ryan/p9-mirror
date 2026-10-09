# ---- 仓库配置 ----
MODULE := oa.98ent.com/p9/operator-crm

CRM_RPC_PROTO := crm.proto
SETTLEMENT_RPC_PROTO := settlement.proto

VERSION ?= 1.0.0

# ---- 公共配置 ----
PROJECT := $(notdir $(MODULE))
GO ?= go
GOCTL ?= goctl
STYLE := go_zero

GOARCH ?= amd64
LDFLAGS := -s -w

DOCKER_PLATFORM ?= linux/$(GOARCH)
DOCKER_REPO ?= 192.168.0.15:5000/p9
CRM_IMAGE := $(PROJECT)-rpc
SETTLEMENT_IMAGE := operator-settlement-rpc

CRM_RPC_PROTO_PATH := service/crm/rpc/proto
SETTLEMENT_RPC_PROTO_PATH := service/settlement/rpc/proto
ENT_DIR := ent

.PHONY: rpc-crm
rpc-crm: # 生成 CRM RPC 代码
	cd $(CRM_RPC_PROTO_PATH) && $(GOCTL) rpc protoc $(CRM_RPC_PROTO) --go_out=../../../.. --go-grpc_out=../../../.. --zrpc_out=.. --go_opt=module=$(MODULE) --go-grpc_opt=module=$(MODULE) --module=$(MODULE) -I . --multiple --style=$(STYLE)

.PHONY: rpc-settlement
rpc-settlement: # 生成 Settlement RPC 代码
	cd $(SETTLEMENT_RPC_PROTO_PATH) && $(GOCTL) rpc protoc $(SETTLEMENT_RPC_PROTO) --go_out=../../../.. --go-grpc_out=../../../.. --zrpc_out=.. --go_opt=module=$(MODULE) --go-grpc_opt=module=$(MODULE) --module=$(MODULE) -I . --multiple --style=$(STYLE)

.PHONY: rpc
rpc: rpc-crm rpc-settlement # 生成两个 RPC

.PHONY: ent-new
ent-new: # 新建 Ent Schema，需要指定 name 参数
ifndef name
	$(error name is required, example: make ent-new name=Member)
endif
	$(GO) run entgo.io/ent/cmd/ent new $(name)

.PHONY: ent
ent: # 生成 Ent 代码，至少有一个业务 Schema 后再执行
	$(GO) generate ./$(ENT_DIR)

.PHONY: build-linux
build-linux: export CGO_ENABLED := 0
build-linux: export GOOS := linux
build-linux: export GOARCH := $(GOARCH)
build-linux: # 编译两个独立的 Linux RPC 程序
	$(GO) build -trimpath -ldflags="$(LDFLAGS)" -o crm_rpc ./service/crm/rpc
	$(GO) build -trimpath -ldflags="$(LDFLAGS)" -o settlement_rpc ./service/settlement/rpc

.PHONY: docker
docker: build-linux # 构建两个独立的 RPC 镜像
	docker build --platform $(DOCKER_PLATFORM) -f Dockerfile-crm-rpc -t $(DOCKER_REPO)/$(CRM_IMAGE):$(VERSION) .
	docker build --platform $(DOCKER_PLATFORM) -f Dockerfile-settlement-rpc -t $(DOCKER_REPO)/$(SETTLEMENT_IMAGE):$(VERSION) .

.PHONY: publish-docker
publish-docker: # 发布两个 RPC 镜像
	docker push $(DOCKER_REPO)/$(CRM_IMAGE):$(VERSION)
	docker push $(DOCKER_REPO)/$(SETTLEMENT_IMAGE):$(VERSION)
