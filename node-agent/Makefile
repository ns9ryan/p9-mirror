# ---- 仓库独立配置，新建仓库时需要修改 ----

MODULE := oa.98ent.com/p9/node-agent
BUILD_NAME := node_agent

# 镜像版本
VERSION ?= 1.0.0

# ---- 下面的配置通常不需要修改 ----

PROJECT := $(notdir $(MODULE))

GO ?= go

# 构建
GOARCH ?= amd64
LDFLAGS := -s -w

# Docker
DOCKER_PLATFORM ?= linux/$(GOARCH)
DOCKER_REPO ?= 192.168.0.15:5000/p9

.PHONY: build-linux
build-linux: export CGO_ENABLED := 0
build-linux: export GOOS := linux
build-linux: export GOARCH := $(GOARCH)
build-linux: # 编译 Linux 可执行文件
	$(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(BUILD_NAME) .

.PHONY: docker
docker: build-linux # 编译 Linux 可执行文件并构建 Docker 镜像
	docker build --platform $(DOCKER_PLATFORM) -f Dockerfile -t $(DOCKER_REPO)/$(PROJECT):$(VERSION) .

.PHONY: publish-docker
publish-docker: # 发布 Docker 镜像
	docker push $(DOCKER_REPO)/$(PROJECT):$(VERSION)