# Mihomo-sync Makefile

APP_NAME := mihomo-sync
BUILD_DIR := build

GO := go
LDFLAGS := -ldflags "-s -w"

PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

.PHONY: all build build-all clean

all: build

build:
	@echo "Building $(APP_NAME)..."
	$(GO) build $(LDFLAGS) -o $(APP_NAME) .
	@echo "Build complete: $(APP_NAME)"

build-all:
	@mkdir -p $(BUILD_DIR)
	@echo "Building for all platforms..."
	@for platform in $(PLATFORMS); do \
		os=$$(echo $$platform | cut -d'/' -f1); \
		arch=$$(echo $$platform | cut -d'/' -f2); \
		ext=""; \
		[ "$$os" = "windows" ] && ext=".exe"; \
		output=$(BUILD_DIR)/$(APP_NAME)-$$os-$$arch$$ext; \
		printf "Building %-40s" "$$output... "; \
		if GOOS=$$os GOARCH=$$arch $(GO) build $(LDFLAGS) -o $$output . 2>&1; then \
			echo "OK"; \
		else \
			echo "FAILED"; \
			exit 1; \
		fi; \
	done
	@echo "All builds complete!"

clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR) $(APP_NAME)
	@echo "Clean complete"
