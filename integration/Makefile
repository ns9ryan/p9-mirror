# Custom configuration | 独立配置
SERVICE=Integration
SERVICE_STYLE=integration
SERVICE_LOWER=integration
SERVICE_SNAKE=integration
SERVICE_DASH=integration

VERSION=1.0.0
PROJECT_STYLE=go_zero

GO ?= go
GOFMT ?= gofmt "-s"
GOFILES := $(shell find . -name "*.go" -not -path "./rpc/ent/*")

GOARCH=amd64
DOCKER_PLATFORM=linux/amd64
DOCKER_REPO=192.168.0.15:5000/p9

GOCTL ?= goctls
STYLE := $(PROJECT_STYLE)
API_FILE := api/desc/main.api
API_DIR := api

.PHONY: fmt
fmt: # Format the codes | 格式化代码
	$(GOFMT) -w $(GOFILES)

.PHONY: api
api: # Generate API files | 生成 API 代码
	$(GOCTL) api go --api $(API_FILE) --dir $(API_DIR) --style=$(STYLE)

# 从 api/desc 生成 docs/swagger.*，并包一层真实 HTTP 外壳 {code,msg,data}
.PHONY: swagger
swagger:
	$(GOCTL) api swagger --api=$(API_FILE) --filename=swagger --dir=./docs
	$(GOCTL) api swagger --api=$(API_FILE) --filename=swagger --dir=./docs --yaml
	$(GO) run ./tools/wrapswagger .
	@echo "Generate swagger successfully"

# 先改 rpc/desc/*.proto，goctls 会合并进 rpc/integration.proto 再生成。
# RPC 方法名勿用 LocalIP 这类连续大写（会生成 create_local_i_p_logic.go），用 LocalIp。
.PHONY: gen-rpc
gen-rpc: # Generate RPC files from proto | 生成 RPC 代码
	$(GOCTL) rpc protoc ./rpc/$(SERVICE_STYLE).proto --style=$(STYLE) --go_out=./rpc/types --go-grpc_out=./rpc/types --zrpc_out=./rpc

.PHONY: gen-ent
gen-ent: # Generate Ent codes | 生成 Ent 代码
	go generate ./rpc/ent

.PHONY: gen
gen: api gen-rpc gen-ent # Generate all | 一键生成

.PHONY: build-win
build-win: # Build project for Windows | 构建 Windows 可执行文件
	env CGO_ENABLED=0 GOOS=windows GOARCH=$(GOARCH) go build -ldflags "$(LDFLAGS)" -trimpath -o $(SERVICE_STYLE)_rpc.exe ./rpc/$(SERVICE_STYLE).go
	env CGO_ENABLED=0 GOOS=windows GOARCH=$(GOARCH) go build -ldflags "$(LDFLAGS)" -trimpath -o $(SERVICE_STYLE)_api.exe ./api/$(SERVICE_STYLE).go
	@echo "Build project for Windows successfully"

.PHONY: build-mac
build-mac: # Build project for MacOS | 构建 MacOS 可执行文件
	env CGO_ENABLED=0 GOOS=darwin GOARCH=$(GOARCH) go build -ldflags "$(LDFLAGS)" -trimpath -o $(SERVICE_STYLE)_rpc ./rpc/$(SERVICE_STYLE).go
	env CGO_ENABLED=0 GOOS=darwin GOARCH=$(GOARCH) go build -ldflags "$(LDFLAGS)" -trimpath -o $(SERVICE_STYLE)_api ./api/$(SERVICE_STYLE).go
	@echo "Build project for MacOS successfully"

.PHONY: build-linux
build-linux: # Build project for Linux | 构建 Linux 可执行文件
	env CGO_ENABLED=0 GOOS=linux GOARCH=$(GOARCH) go build -ldflags "$(LDFLAGS)" -trimpath -o $(SERVICE_STYLE)_rpc ./rpc/$(SERVICE_STYLE).go
	env CGO_ENABLED=0 GOOS=linux GOARCH=$(GOARCH) go build -ldflags "$(LDFLAGS)" -trimpath -o $(SERVICE_STYLE)_api ./api/$(SERVICE_STYLE).go
	@echo "Build project for Linux successfully"

.PHONY: docker
docker: build-linux # Build the docker image | 构建 docker 镜像
	docker build --platform $(DOCKER_PLATFORM) -f Dockerfile-api -t $(DOCKER_REPO)/$(SERVICE_DASH)-api:$(VERSION) .
	docker build --platform $(DOCKER_PLATFORM) -f Dockerfile-rpc -t $(DOCKER_REPO)/$(SERVICE_DASH)-rpc:$(VERSION) .
	@echo "Build docker successfully"

.PHONY: publish-docker
publish-docker: # Publish docker image | 发布 docker 镜像
	docker push $(DOCKER_REPO)/$(SERVICE_DASH)-rpc:$(VERSION)
	docker push $(DOCKER_REPO)/$(SERVICE_DASH)-api:$(VERSION)
	@echo "Publish docker successfully"

.PHONY: help
help: # Show help | 显示帮助
	@grep -E '^[a-zA-Z0-9 -]+:.*#' Makefile | sort | while read -r l; do printf "\033[1;32m$$(echo $$l | cut -f 1 -d':')\033[00m:$$(echo $$l | cut -f 2- -d'#')\n"; done
