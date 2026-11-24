SHELL := /bin/bash
.DEFAULT_GOAL := help

PROJECT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
SCRIPT_DIR := $(PROJECT_DIR)/scripts

.PHONY: help pull push

help: ## Show available commands
	@printf '%s\n' \
	  'Available commands:' \
	  '  make help   Show this help' \
	  '  make pull   Copy server configs into this staging directory' \
	  '  make push   Back up and apply staged configs to server paths'

pull: ## Copy server configs into this staging directory
	@bash "$(SCRIPT_DIR)/pull.sh"

push: ## Back up and apply staged configs to server paths
	@bash "$(SCRIPT_DIR)/push.sh"