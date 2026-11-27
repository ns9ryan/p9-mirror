GO ?= go

# RPC 配置
ADDR ?= 192.168.10.220:39005
OPERATOR_CODE ?= OP_034A80
TIMEOUT ?= 5s

# 钱包配置
CURRENCY_CODE ?= USDT
WALLETS ?= 128
WALLET_CONCURRENCY ?= 20
MEMBER_ID_START ?= 9000000001
PARTITIONS ?= 32

# 资金操作配置
MEMBER_GLOBAL_ID ?=
AMOUNT ?=
CONCURRENCY ?= 20

.PHONY: help ping prepare deposit concurrent multi-wallet fmt test build

# 显示命令帮助
help:
	@echo Available commands:
	@echo   make ping
	@echo   make prepare
	@echo   make deposit
	@echo   make concurrent
	@echo   make multi-wallet
	@echo   make fmt
	@echo   make test
	@echo   make build

# RPC 连通性测试
ping:
	$(GO) run . -mode ping -addr "$(ADDR)" -operator-code "$(OPERATOR_CODE)" -timeout "$(TIMEOUT)"

# 准备测试钱包
prepare:
	$(GO) run . -mode prepare -addr "$(ADDR)" -operator-code "$(OPERATOR_CODE)" -timeout "$(TIMEOUT)" -currency-code "$(CURRENCY_CODE)" -wallets "$(WALLETS)" -wallet-concurrency "$(WALLET_CONCURRENCY)" -member-id-start "$(MEMBER_ID_START)" -partitions "$(PARTITIONS)"

# 单笔存款测试
deposit:
	$(GO) run . -mode deposit -addr "$(ADDR)" -operator-code "$(OPERATOR_CODE)" -timeout "$(TIMEOUT)" -currency-code "$(CURRENCY_CODE)" -member-global-id "$(MEMBER_GLOBAL_ID)" -amount "$(AMOUNT)"

# 同钱包并发存款测试
concurrent:
	$(GO) run . -mode concurrent -addr "$(ADDR)" -operator-code "$(OPERATOR_CODE)" -timeout "$(TIMEOUT)" -currency-code "$(CURRENCY_CODE)" -member-global-id "$(MEMBER_GLOBAL_ID)" -amount "$(AMOUNT)" -concurrency "$(CONCURRENCY)"

# 多钱包并发存款测试
multi-wallet:
	$(GO) run . -mode multi-wallet -addr "$(ADDR)" -operator-code "$(OPERATOR_CODE)" -timeout "$(TIMEOUT)" -currency-code "$(CURRENCY_CODE)" -wallets "$(WALLETS)" -wallet-concurrency "$(WALLET_CONCURRENCY)" -member-id-start "$(MEMBER_ID_START)" -partitions "$(PARTITIONS)" -amount "$(AMOUNT)" -concurrency "$(CONCURRENCY)"

# 格式化代码
fmt:
	$(GO) fmt ./...

# 执行测试
test:
	$(GO) test ./...

# 编译检查
build:
	$(GO) build ./...