# OpenClaw SaaS 平台 Makefile
# 常用命令集合

.PHONY: help build test run clean docker-build docker-push k8s-deploy \
        backend-dev frontend-dev db-migrate lint

# 默认目标：显示帮助
help:
	@echo "OpenClaw SaaS 平台 - 可用命令："
	@echo ""
	@echo "开发命令:"
	@echo "  make backend-dev      启动后端开发服务器（支持热重载）"
	@echo "  make frontend-dev     启动前端开发服务器"
	@echo "  make dev              同时启动前后端开发服务器"
	@echo ""
	@echo "构建命令:"
	@echo "  make build            构建后端二进制文件"
	@echo "  make frontend-build   构建前端静态文件"
	@echo "  make docker-build     构建 Docker 镜像"
	@echo "  make docker-push      推送 Docker 镜像到仓库"
	@echo ""
	@echo "数据库命令:"
	@echo "  make db-migrate       执行数据库迁移"
	@echo "  make db-reset         重置数据库（危险！）"
	@echo ""
	@echo "测试命令:"
	@echo "  make test             运行所有测试"
	@echo "  make test-backend     运行后端测试"
	@echo "  make lint             代码风格检查"
	@echo ""
	@echo "Docker 命令:"
	@echo "  make docker-up        启动所有服务（docker-compose）"
	@echo "  make docker-down      停止所有服务"
	@echo "  make docker-logs      查看服务日志"
	@echo ""
	@echo "K8s 部署命令:"
	@echo "  make k8s-deploy       部署到 K8s 集群"
	@echo "  make k8s-delete       从 K8s 集群删除"

# ============ 后端命令 ============

# 构建后端二进制
build:
	@echo "正在构建后端服务..."
	cd backend && go build -o bin/server ./cmd/server
	@echo "构建完成：backend/bin/server"

# 后端开发模式（需要安装 air: go install github.com/cosmtrek/air@latest）
backend-dev:
	@echo "启动后端开发服务器（热重载）..."
	cd backend && go run ./cmd/server

# 运行所有测试
test:
	@echo "运行所有测试..."
	cd backend && go test ./...

# 运行后端测试
test-backend:
	cd backend && go test -v ./...

# 代码检查（需要安装 golangci-lint）
lint:
	@echo "运行代码检查..."
	cd backend && golangci-lint run ./...

# ============ 前端命令 ============

# 安装前端依赖
frontend-install:
	cd frontend && npm install

# 启动前端开发服务器
frontend-dev:
	@echo "启动前端开发服务器..."
	cd frontend && npm run dev

# 构建前端
frontend-build:
	@echo "构建前端静态文件..."
	cd frontend && npm run build

# ============ 数据库命令 ============

# 执行数据库迁移
db-migrate:
	@echo "执行数据库迁移..."
	@for f in backend/migrations/*.sql; do \
		echo "执行迁移: $$f"; \
		PGPASSWORD=openclaw_dev_password psql -h localhost -U openclaw -d openclaw -f $$f; \
	done
	@echo "数据库迁移完成"

# 重置数据库（开发用，危险操作）
db-reset:
	@echo "警告：这将删除所有数据！"
	@read -p "确认重置数据库？(yes/no): " confirm && [ "$$confirm" = "yes" ]
	PGPASSWORD=openclaw_dev_password psql -h localhost -U openclaw -c "DROP DATABASE IF EXISTS openclaw;"
	PGPASSWORD=openclaw_dev_password psql -h localhost -U openclaw -c "CREATE DATABASE openclaw;"
	@$(MAKE) db-migrate

# ============ Docker 命令 ============

# 构建 Docker 镜像
docker-build:
	@echo "构建后端 Docker 镜像..."
	docker build -t openclaw-backend:latest ./backend
	@echo "构建前端 Docker 镜像..."
	docker build -t openclaw-frontend:latest ./frontend

# 启动所有服务
docker-up:
	docker-compose up -d
	@echo "服务已启动："
	@echo "  前端: http://localhost:3000"
	@echo "  后端: http://localhost:8080"

# 停止所有服务
docker-down:
	docker-compose down

# 查看服务日志
docker-logs:
	docker-compose logs -f

# ============ K8s 部署命令 ============

# 部署 SaaS 平台到 K8s
k8s-deploy:
	@echo "部署 OpenClaw SaaS 平台到 K8s..."
	kubectl apply -f k8s/base/
	kubectl apply -f k8s/saas-platform/
	@echo "部署完成"

# 从 K8s 删除
k8s-delete:
	kubectl delete -f k8s/saas-platform/
	kubectl delete -f k8s/base/

# 安装 OpenClaw Operator
install-operator:
	kubectl apply -f k8s/operator/openclaw-operator-install.yaml

# ============ 开发快捷命令 ============

# 同时启动前后端（需要 tmux 或 screen，或使用 & 后台运行）
dev:
	@echo "启动完整开发环境..."
	docker-compose up -d postgres redis
	@echo "PostgreSQL 和 Redis 已启动"
	@echo "请分别在两个终端运行:"
	@echo "  make backend-dev"
	@echo "  make frontend-dev"

# 清理构建产物
clean:
	rm -rf backend/bin/
	rm -rf frontend/dist/
	rm -rf frontend/node_modules/
