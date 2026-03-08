# Mihomo-sync Makefile

APP_NAME := mihomo-sync
BUILD_DIR := build

# Go 参数
GO := go
GOFLAGS := -v
LDFLAGS := -ldflags "-s -w"

# 平台
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

.PHONY: all build build-all clean test help fmt tidy

# 默认目标
all: build

# 构建当前平台
build:
	@echo "Building $(APP_NAME) $(VERSION)..."
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(APP_NAME) .

# 构建所有平台
build-all:
	@mkdir -p $(BUILD_DIR)
	@echo "Building for all platforms..."
	@for platform in $(PLATFORMS); do \
		os=$$(echo $$platform | cut -d'/' -f1); \
		arch=$$(echo $$platform | cut -d'/' -f2); \
		ext=""; \
		[ "$$os" = "windows" ] && ext=".exe"; \
		output=$(BUILD_DIR)/$(APP_NAME)-$$os-$$arch$$ext; \
		echo "Building $$output"; \
		GOOS=$$os GOARCH=$$arch $(GO) build $(GOFLAGS) $(LDFLAGS) -o $$output .; \
	done
	@echo "Build complete. Artifacts in $(BUILD_DIR)/"

# 清理
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR) $(APP_NAME)

# 测试
test:
	@echo "Running tests..."
	$(GO) test $(GOFLAGS) ./...

# 格式化代码
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

# 整理依赖
tidy:
	@echo "Tidying dependencies..."
	$(GO) mod tidy

# 帮助
help:
	@echo "Mihomo-sync Build Tool"
	@echo ""
	@echo "Targets:"
	@echo "  build       - Build for current platform"
	@echo "  build-all   - Build for all platforms"
	@echo "  clean       - Remove build artifacts"
	@echo "  test        - Run tests"
	@echo "  fmt         - Format code"
	@echo "  tidy        - Tidy dependencies"
	@echo "  help        - Show this help"
