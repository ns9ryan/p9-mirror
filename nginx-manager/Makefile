# Custom configuration | 独立配置
SERVICE=nginx-manager
SERVICE_STYLE=nginx_manager
SERVICE_LOWER=nginx-manager
SERVICE_SNAKE=nginx_manager
SERVICE_DASH=nginx-manager

# The project version | 项目版本
VERSION=1.0.0

# The arch of the build | 构建的架构
GOARCH=amd64
DOCKER_OPERATOR=linux/amd64

# The repository of docker | Docker 仓库地址
DOCKER_REPO=192.168.0.15:5000/p9

.PHONY: help docker publish-docker build-linux build-win build-mac clean

help: # 显示帮助信息
	@echo "Available targets:"
	@echo "  make docker           - 构建 Docker 镜像"
	@echo "  make publish-docker   - 发布镜像到仓库"
	@echo "  make clean            - 清理本地镜像"
	@echo "  make build-linux      - 构建 Linux 可执行文件"
	@echo "  make build-win        - 构建 Windows 可执行文件"
	@echo "  make build-mac        - 构建 MacOS 可执行文件"

.PHONY: docker
docker: build-linux # 构建 docker 镜像 | Build the docker image
	docker build --platform $(DOCKER_OPERATOR) -f Dockerfile -t $(DOCKER_REPO)/$(SERVICE_DASH):${VERSION} .
	@echo "Build docker successfully"

.PHONY: publish-docker
publish-docker: docker # 发布 docker 镜像 | Publish docker image
	docker push $(DOCKER_REPO)/$(SERVICE_DASH):${VERSION}
	@echo "Publish docker successfully"

.PHONY: build-linux
build-linux: # 构建 Linux 可执行文件 | Build project for Linux
	env CGO_ENABLED=0 GOOS=linux GOARCH=$(GOARCH) go build -ldflags "-s -w" -trimpath -o reload-api .
	@echo "Build project for Linux successfully"

.PHONY: build-win
build-win: # 构建 Windows 可执行文件 | Build project for Windows
	env CGO_ENABLED=0 GOOS=windows GOARCH=$(GOARCH) go build -ldflags "-s -w" -trimpath -o reload-api.exe .
	@echo "Build project for Windows successfully"

.PHONY: build-mac
build-mac: # 构建 MacOS 可执行文件 | Build project for MacOS
	env CGO_ENABLED=0 GOOS=darwin GOARCH=$(GOARCH) go build -ldflags "-s -w" -trimpath -o reload-api .
	@echo "Build project for MacOS successfully"

.PHONY: clean
clean: # 清理本地镜像和编译文件 | Clean up local images and build files
	docker rmi $(DOCKER_REPO)/$(SERVICE_DASH):${VERSION} 2>/dev/null || true
	rm -f reload-api reload-api.exe
	@echo "Clean up successfully"

.PHONY: fmt
fmt: # 格式化代码 | Format the codes
	go fmt ./...

.PHONY: test
test: # 运行测试 | Run tests
	go test ./... -v

.PHONY: lint
lint: # 检查代码风格 | Check code style
	golangci-lint run ./...
