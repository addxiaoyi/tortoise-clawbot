# Tortoise Makefile

.PHONY: help build test lint clean install dev docker-build docker-run release docs

# 默认目标
help:
	@echo "Tortoise AI Agent Framework"
	@echo ""
	@echo "可用命令:"
	@echo "  make build         - 构建项目"
	@echo "  make test         - 运行测试"
	@echo "  make lint         - 代码检查"
	@echo "  make clean        - 清理构建文件"
	@echo "  make install      - 安装依赖"
	@echo "  make dev          - 开发模式"
	@echo "  make docker-build - 构建 Docker 镜像"
	@echo "  make docker-run    - 运行 Docker 容器"
	@echo "  make release      - 构建发布版本"
	@echo "  make docs         - 生成文档"

# === Rust Core ===
build-core:
	cd core && cargo build --release

test-core:
	cd core && cargo test --all

lint-core:
	cd core && cargo fmt --check && cargo clippy -- -D warnings

# === Go Plugins ===
build-plugins:
	cd plugins && go build -v ./...

test-plugins:
	cd plugins && go test -v ./...

lint-plugins:
	cd plugins && golangci-lint run

# === Flutter UI ===
build-ui:
	cd ui && flutter build apk --release

test-ui:
	cd ui && flutter test

lint-ui:
	cd ui && flutter analyze

# === 构建所有 ===
build: build-core build-plugins build-ui

# === 测试所有 ===
test: test-core test-plugins test-ui

# === 代码检查 ===
lint: lint-core lint-plugins lint-ui

# === 清理 ===
clean:
	cd core && cargo clean
	cd ui && flutter clean
	rm -rf target/

# === 安装依赖 ===
install:
	cargo install cargo-watch cargo-edit
	cd core && cargo fetch
	cd plugins && go mod download
	cd ui && flutter pub get

# === 开发模式 ===
dev:
	cd core && cargo watch -x build -x test

# === Docker ===
docker-build:
	docker build -t tortoise/tortoise:latest .

docker-run:
	docker-compose up -d

docker-stop:
	docker-compose down

# === 发布 ===
release:
	cd core && cargo build --release --bin tortoise
	cd ui && flutter build apk --release
	cd plugins && go build -o tortoise-plugins.so ./...

# === 文档 ===
docs:
	cd core && cargo doc --no-deps
	cd docs && ./generate.sh

# === 代码格式化 ===
format:
	cd core && cargo fmt
	cd ui && flutter format .
	cd plugins && go fmt ./...

# === 代码统计 ===
stats:
	@echo "=== Rust ==="
	@find core/src -name "*.rs" | xargs wc -l | tail -1
	@echo "=== Go ==="
	@find plugins -name "*.go" | xargs wc -l | tail -1
	@echo "=== Dart ==="
	@find ui/lib -name "*.dart" | xargs wc -l | tail -1

# === 安全检查 ===
security:
	cd core && cargo audit
	cd plugins && gosec ./...

# === 性能基准 ===
bench:
	cd core && cargo bench
