# ---- 仓库配置 ----
MODULE := oa.98ent.com/p9/operator-wallet

MEMBER_RPC_PROTO := member_wallet.proto

VERSION ?= 1.0.0

# ---- 公共配置 ----
PROJECT := $(notdir $(MODULE))
GO ?= go
GOCTL ?= goctl
STYLE := go_zero

GOARCH ?= amd64
LDFLAGS := -s -w

DOCKER_PLATFORM ?= linux/$(GOARCH)
DOCKER_REPO ?= 192.168.10.220:5000/p9
MEMBER_IMAGE := member-wallet-rpc

MEMBER_RPC_PROTO_PATH := service/member_wallet/rpc/proto
ENT_DIR := ent

.PHONY: rpc-member
rpc-member: # 生成会员钱包 RPC 代码
	cd $(MEMBER_RPC_PROTO_PATH) && $(GOCTL) rpc protoc $(MEMBER_RPC_PROTO) --go_out=../../../.. --go-grpc_out=../../../.. --zrpc_out=.. --go_opt=module=$(MODULE) --go-grpc_opt=module=$(MODULE) --module=$(MODULE) -I . --multiple --style=$(STYLE)

.PHONY: rpc
rpc: rpc-member # 生成所有 RPC 代码

.PHONY: ent-new
ent-new: # 新建 Ent Schema, 需要指定 name 参数
ifndef name
	$(error name is required, example: make ent-new name=MemberWallet)
endif
	$(GO) run entgo.io/ent/cmd/ent new $(name)

.PHONY: ent
ent: # 生成 Ent 代码, 至少有一个业务 Schema 后再执行
	$(GO) generate ./$(ENT_DIR)

.PHONY: build-linux
build-linux: export CGO_ENABLED := 0
build-linux: export GOOS := linux
build-linux: export GOARCH := $(GOARCH)
build-linux: # 编译会员钱包 Linux RPC 程序
	$(GO) build -trimpath -ldflags="$(LDFLAGS)" -o member_wallet_rpc ./service/member_wallet/rpc

.PHONY: docker
docker: build-linux # 构建会员钱包 RPC 镜像
	docker build --platform $(DOCKER_PLATFORM) -f Dockerfile-member-rpc -t $(DOCKER_REPO)/$(MEMBER_IMAGE):$(VERSION) .

.PHONY: publish-docker
publish-docker: # 发布会员钱包 RPC 镜像
	docker push $(DOCKER_REPO)/$(MEMBER_IMAGE):$(VERSION)