.PHONY: run down

# Docker Compose 目录
COMPOSE_DIR=./deploy/local

# 默认目标
all: help

# 启动服务
run:
	@echo "启动服务..."
	@cd $(COMPOSE_DIR) && docker-compose up -d
	@echo "服务已启动"

# 停止服务
down:
	@echo "停止服务..."
	@cd $(COMPOSE_DIR) && docker-compose down --volumes
	@echo "服务已停止"


# 帮助信息
help:
	@echo "任务管理系统 Makefile 命令:"
	@echo "  make up        - 启动所有服务 (docker-compose up -d)"
	@echo "  make down      - 停止并移除所有服务"