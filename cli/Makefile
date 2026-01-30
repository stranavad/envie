# Envie CLI Makefile

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')

BINARY_NAME = envie
BUILD_DIR = dist

# Go build flags
LDFLAGS = -s -w \
	-X 'github.com/stranavad/envie/cli/cmd.version=$(VERSION)' \
	-X 'github.com/stranavad/envie/cli/cmd.commit=$(COMMIT)' \
	-X 'github.com/stranavad/envie/cli/cmd.buildTime=$(BUILD_TIME)'

# Platforms to build for
PLATFORMS = \
	darwin/amd64 \
	darwin/arm64 \
	linux/amd64 \
	linux/arm64 \
	linux/arm \
	windows/amd64

.PHONY: all build clean test lint build-all release help

## help: Show this help message
help:
	@echo "Envie CLI Build System"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

## build: Build for current platform
build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) .

## build-all: Build for all platforms
build-all: clean
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		os=$$(echo $$platform | cut -d'/' -f1); \
		arch=$$(echo $$platform | cut -d'/' -f2); \
		ext=""; \
		if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
		output="$(BUILD_DIR)/$(BINARY_NAME)-$$os-$$arch$$ext"; \
		echo "Building $$output..."; \
		GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $$output . || exit 1; \
	done
	@echo "Build complete. Binaries in $(BUILD_DIR)/"

## test: Run tests
test:
	go test -v ./...

## test-cover: Run tests with coverage
test-cover:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

## lint: Run linters
lint:
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, running go vet"; \
		go vet ./...; \
	fi

## clean: Remove build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

## install: Install to /usr/local/bin
install: build
	sudo cp $(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)

## uninstall: Remove from /usr/local/bin
uninstall:
	sudo rm -f /usr/local/bin/$(BINARY_NAME)

## version: Show version info
version:
	@echo "Version:    $(VERSION)"
	@echo "Commit:     $(COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"

## release: Create a new release (usage: make release VERSION=v1.0.0)
release: build-all
	@echo "Creating release $(VERSION)"
	@echo "Binaries ready in $(BUILD_DIR)/"
	@echo ""
	@echo "To create GitHub release:"
	@echo "  gh release create $(VERSION) $(BUILD_DIR)/* --title 'Envie CLI $(VERSION)' --notes 'Release notes here'"
