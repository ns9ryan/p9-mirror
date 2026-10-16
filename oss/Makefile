# Custom configuration | 独立配置
SERVICE=Oss
SERVICE_STYLE=oss
SERVICE_DASH=oss

VERSION=1.0.0

GO ?= go
GOFMT ?= gofmt "-s"
GOFILES := $(shell find . -name "*.go")

GOARCH=amd64
DOCKER_PLATFORM=linux/amd64
DOCKER_REPO=192.168.0.15:5000/p9

.PHONY: fmt
fmt: # Format the codes | 格式化代码
	$(GOFMT) -w $(GOFILES)

.PHONY: docker
docker: build-linux # Build the docker image | 构建 docker 镜像
	docker build --platform $(DOCKER_PLATFORM) -f Dockerfile-api -t $(DOCKER_REPO)/$(SERVICE_DASH)-api:${VERSION} .
	docker build --platform $(DOCKER_PLATFORM) -f Dockerfile-rpc -t $(DOCKER_REPO)/$(SERVICE_DASH)-rpc:${VERSION} .
	@echo "Build docker successfully"

.PHONY: publish-docker
publish-docker: # Publish docker image | 发布 docker 镜像
	docker push $(DOCKER_REPO)/$(SERVICE_DASH)-rpc:${VERSION}
	docker push $(DOCKER_REPO)/$(SERVICE_DASH)-api:${VERSION}
	@echo "Publish docker successfully"

.PHONY: gen-api
gen-api: # Generate API files | 生成 API 的代码
	goctl api go --api ./api/desc/all.api --dir ./api --style=go_zero
	@echo "Generate API files successfully"

.PHONY: gen-rpc
gen-rpc: # Generate RPC files from proto | 生成 RPC 的代码
	go generate ./rpc
	@echo "Generate RPC files successfully"

.PHONY: gen-ent
gen-ent: # Generate Ent codes | 生成 Ent 的代码
	go generate ./rpc/ent
	@echo "Generate Ent files successfully"

.PHONY: build-win
build-win: # Build project for Windows | 构建Windows下的可执行文件
	env CGO_ENABLED=0 GOOS=windows GOARCH=$(GOARCH) go build -trimpath -o $(SERVICE_STYLE)_rpc.exe ./rpc
	env CGO_ENABLED=0 GOOS=windows GOARCH=$(GOARCH) go build -trimpath -o $(SERVICE_STYLE)_api.exe ./api
	@echo "Build project for Windows successfully"

.PHONY: build-mac
build-mac: # Build project for MacOS | 构建MacOS下的可执行文件
	env CGO_ENABLED=0 GOOS=darwin GOARCH=$(GOARCH) go build -trimpath -o $(SERVICE_STYLE)_rpc ./rpc
	env CGO_ENABLED=0 GOOS=darwin GOARCH=$(GOARCH) go build -trimpath -o $(SERVICE_STYLE)_api ./api
	@echo "Build project for MacOS successfully"

.PHONY: build-linux
build-linux: # Build project for Linux | 构建Linux下的可执行文件
	env CGO_ENABLED=0 GOOS=linux GOARCH=$(GOARCH) go build -trimpath -o $(SERVICE_STYLE)_rpc ./rpc
	env CGO_ENABLED=0 GOOS=linux GOARCH=$(GOARCH) go build -trimpath -o $(SERVICE_STYLE)_api ./api
	@echo "Build project for Linux successfully"

.PHONY: help
help: # Show help | 显示帮助
	@grep -E '^[a-zA-Z0-9 -]+:.*#'  Makefile | sort | while read -r l; do printf "\033[1;32m$$(echo $$l | cut -f 1 -d':')\033[00m:$$(echo $$l | cut -f 2- -d'#')\n"; done
