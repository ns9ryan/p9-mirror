# ---- 仓库独立配置，新建仓库时需要修改 ----

MODULE := oa.98ent.com/p9/operator-admin-api
BUILD_NAME := $(notdir $(MODULE))
PROJECT_NAME := operator_admin_api


GO ?= go
GOCTL ?= goctl

STYLE := go_zero

# 构建
GOARCH ?= amd64
LDFLAGS := -s -w

# Docker
DOCKER_PLATFORM ?= linux/$(GOARCH)
DOCKER_REPO ?= 192.168.0.15:5000/p9
VERSION ?= 1.0.0

# API
API_FILE := desc/main.api
API_DIR := .

.PHONY: api
api: # 生成 API 代码
	$(GOCTL) api go --api $(API_FILE) --dir $(API_DIR) --style=$(STYLE)

.PHONY: gen
gen: api # 生成 API 代码

.PHONY: build-linux
build-linux: export CGO_ENABLED := 0
build-linux: export GOOS := linux
build-linux: export GOARCH := $(GOARCH)
build-linux: # 编译 Linux 可执行文件
	$(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(PROJECT_NAME) .

.PHONY: docker
docker: build-linux # 编译 Linux 可执行文件并构建 Docker 镜像
	docker build --platform $(DOCKER_PLATFORM) -f Dockerfile -t $(DOCKER_REPO)/$(BUILD_NAME):$(VERSION) .

.PHONY: publish-docker
publish-docker: # 发布 Docker 镜像
	docker push $(DOCKER_REPO)/$(BUILD_NAME):$(VERSION)
