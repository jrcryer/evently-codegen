# AsyncAPI Go Code Generator - Build Configuration
# This Makefile provides build, test, and release automation

# Project information
PROJECT_NAME := evently-codegen
MODULE_NAME := github.com/jrcryer/evently-codegen
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")

# Build configuration
GO_VERSION := 1.19
BINARY_NAME := evently-codegen
BUILD_DIR := bin
DIST_DIR := dist
COVERAGE_DIR := coverage

# Go build flags
LDFLAGS := -ldflags "\
	-X '$(MODULE_NAME)/internal/version.Version=$(VERSION)' \
	-X '$(MODULE_NAME)/internal/version.BuildTime=$(BUILD_TIME)' \
	-X '$(MODULE_NAME)/internal/version.GitCommit=$(GIT_COMMIT)' \
	-X '$(MODULE_NAME)/internal/version.GitBranch=$(GIT_BRANCH)' \
	-s -w"

# Cross-compilation targets
PLATFORMS := \
	linux/amd64 \
	linux/arm64 \
	darwin/amd64 \
	darwin/arm64 \
	windows/amd64 \
	windows/arm64

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
PURPLE := \033[0;35m
CYAN := \033[0;36m
NC := \033[0m # No Color

.PHONY: help
help: ## Show this help message
	@echo "$(CYAN)AsyncAPI Go Code Generator - Build System$(NC)"
	@echo ""
	@echo "$(YELLOW)Available targets:$(NC)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "$(YELLOW)Build information:$(NC)"
	@echo "  Version:    $(VERSION)"
	@echo "  Git Commit: $(GIT_COMMIT)"
	@echo "  Git Branch: $(GIT_BRANCH)"
	@echo "  Build Time: $(BUILD_TIME)"

.PHONY: clean
clean: ## Clean build artifacts and temporary files
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	@rm -rf $(BUILD_DIR)
	@rm -rf $(DIST_DIR)
	@rm -rf $(COVERAGE_DIR)
	@rm -f coverage.out coverage.html
	@go clean -cache -testcache -modcache
	@echo "$(GREEN)✓ Clean completed$(NC)"

.PHONY: deps
deps: ## Download and verify dependencies
	@echo "$(YELLOW)Downloading dependencies...$(NC)"
	@go mod download
	@go mod verify
	@go mod tidy
	@echo "$(GREEN)✓ Dependencies updated$(NC)"

.PHONY: fmt
fmt: ## Format Go code
	@echo "$(YELLOW)Formatting Go code...$(NC)"
	@gofmt -s -w .
	@go mod tidy
	@echo "$(GREEN)✓ Code formatted$(NC)"

.PHONY: lint
lint: ## Run linters
	@echo "$(YELLOW)Running linters...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "$(RED)golangci-lint not found. Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)✓ Linting completed$(NC)"

.PHONY: vet
vet: ## Run go vet
	@echo "$(YELLOW)Running go vet...$(NC)"
	@go vet ./...
	@echo "$(GREEN)✓ go vet completed$(NC)"

.PHONY: security
security: ## Run security checks
	@echo "$(YELLOW)Running security checks...$(NC)"
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "$(YELLOW)gosec not found. Install it with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest$(NC)"; \
	fi
	@echo "$(GREEN)✓ Security checks completed$(NC)"

.PHONY: test
test: ## Run tests
	@echo "$(YELLOW)Running tests...$(NC)"
	@go test -v -race ./...
	@echo "$(GREEN)✓ Tests completed$(NC)"

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@echo "$(YELLOW)Running tests with coverage...$(NC)"
	@mkdir -p $(COVERAGE_DIR)
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -html=coverage.out -o coverage.html
	@go tool cover -func=coverage.out | tail -1
	@echo "$(GREEN)✓ Coverage report generated: coverage.html$(NC)"

.PHONY: test-integration
test-integration: build ## Run integration tests
	@echo "$(YELLOW)Running integration tests...$(NC)"
	@go test -v -tags=integration ./...
	@echo "$(GREEN)✓ Integration tests completed$(NC)"

.PHONY: test-performance
test-performance: ## Run performance tests
	@echo "$(YELLOW)Running performance tests...$(NC)"
	@go test -v -bench=. -benchmem ./...
	@echo "$(GREEN)✓ Performance tests completed$(NC)"

.PHONY: build
build: deps ## Build the binary
	@echo "$(YELLOW)Building $(BINARY_NAME)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/$(PROJECT_NAME)
	@echo "$(GREEN)✓ Build completed: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

.PHONY: build-debug
build-debug: deps ## Build the binary with debug information
	@echo "$(YELLOW)Building $(BINARY_NAME) with debug info...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 go build -gcflags="all=-N -l" -o $(BUILD_DIR)/$(BINARY_NAME)-debug ./cmd/$(PROJECT_NAME)
	@echo "$(GREEN)✓ Debug build completed: $(BUILD_DIR)/$(BINARY_NAME)-debug$(NC)"

.PHONY: install
install: build ## Install the binary to GOPATH/bin
	@echo "$(YELLOW)Installing $(BINARY_NAME)...$(NC)"
	@go install $(LDFLAGS) ./cmd/$(PROJECT_NAME)
	@echo "$(GREEN)✓ Installation completed$(NC)"

.PHONY: uninstall
uninstall: ## Uninstall the binary from GOPATH/bin
	@echo "$(YELLOW)Uninstalling $(BINARY_NAME)...$(NC)"
	@rm -f $(shell go env GOPATH)/bin/$(BINARY_NAME)
	@echo "$(GREEN)✓ Uninstallation completed$(NC)"

.PHONY: build-all
build-all: deps ## Build binaries for all platforms
	@echo "$(YELLOW)Building for all platforms...$(NC)"
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		os=$$(echo $$platform | cut -d'/' -f1); \
		arch=$$(echo $$platform | cut -d'/' -f2); \
		output_name=$(BINARY_NAME)-$$os-$$arch; \
		if [ "$$os" = "windows" ]; then \
			output_name=$$output_name.exe; \
		fi; \
		echo "Building for $$os/$$arch..."; \
		GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build $(LDFLAGS) \
			-o $(DIST_DIR)/$$output_name ./cmd/$(PROJECT_NAME); \
	done
	@echo "$(GREEN)✓ Cross-compilation completed$(NC)"

.PHONY: package
package: build-all ## Create release packages
	@echo "$(YELLOW)Creating release packages...$(NC)"
	@cd $(DIST_DIR) && for binary in $(BINARY_NAME)-*; do \
		if [[ "$$binary" == *".exe" ]]; then \
			zip "$$binary.zip" "$$binary"; \
		else \
			tar -czf "$$binary.tar.gz" "$$binary"; \
		fi; \
	done
	@echo "$(GREEN)✓ Release packages created in $(DIST_DIR)/$(NC)"

.PHONY: checksums
checksums: package ## Generate checksums for release packages
	@echo "$(YELLOW)Generating checksums...$(NC)"
	@cd $(DIST_DIR) && sha256sum *.tar.gz *.zip > checksums.txt
	@echo "$(GREEN)✓ Checksums generated: $(DIST_DIR)/checksums.txt$(NC)"

.PHONY: release
release: clean test lint vet security build-all package checksums ## Create a full release
	@echo "$(GREEN)✓ Release $(VERSION) created successfully!$(NC)"
	@echo "$(CYAN)Release artifacts:$(NC)"
	@ls -la $(DIST_DIR)/

.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "$(YELLOW)Building Docker image...$(NC)"
	@docker build -t $(PROJECT_NAME):$(VERSION) -t $(PROJECT_NAME):latest .
	@echo "$(GREEN)✓ Docker image built: $(PROJECT_NAME):$(VERSION)$(NC)"

.PHONY: docker-run
docker-run: docker-build ## Run Docker container
	@echo "$(YELLOW)Running Docker container...$(NC)"
	@docker run --rm -it $(PROJECT_NAME):latest --help

.PHONY: examples
examples: build ## Run all examples
	@echo "$(YELLOW)Running examples...$(NC)"
	@cd examples/basic_usage && go run main.go
	@cd examples/file_operations && go run main.go
	@cd examples/advanced_features && go run main.go
	@cd examples/integration && go run main.go
	@echo "$(GREEN)✓ All examples completed$(NC)"

.PHONY: benchmark
benchmark: ## Run benchmarks and save results
	@echo "$(YELLOW)Running benchmarks...$(NC)"
	@mkdir -p $(COVERAGE_DIR)
	@go test -bench=. -benchmem -cpuprofile=$(COVERAGE_DIR)/cpu.prof -memprofile=$(COVERAGE_DIR)/mem.prof ./...
	@echo "$(GREEN)✓ Benchmarks completed. Profiles saved in $(COVERAGE_DIR)/$(NC)"

.PHONY: profile-cpu
profile-cpu: benchmark ## Analyze CPU profile
	@echo "$(YELLOW)Analyzing CPU profile...$(NC)"
	@go tool pprof $(COVERAGE_DIR)/cpu.prof

.PHONY: profile-mem
profile-mem: benchmark ## Analyze memory profile
	@echo "$(YELLOW)Analyzing memory profile...$(NC)"
	@go tool pprof $(COVERAGE_DIR)/mem.prof

.PHONY: dev
dev: clean fmt vet test build ## Development workflow: clean, format, vet, test, build
	@echo "$(GREEN)✓ Development build completed$(NC)"

.PHONY: ci
ci: clean deps fmt vet lint security test test-coverage build ## CI workflow
	@echo "$(GREEN)✓ CI pipeline completed$(NC)"

.PHONY: pre-commit
pre-commit: fmt vet lint test ## Pre-commit checks
	@echo "$(GREEN)✓ Pre-commit checks passed$(NC)"

.PHONY: version
version: ## Show version information
	@echo "$(CYAN)Version Information:$(NC)"
	@echo "  Version:    $(VERSION)"
	@echo "  Git Commit: $(GIT_COMMIT)"
	@echo "  Git Branch: $(GIT_BRANCH)"
	@echo "  Build Time: $(BUILD_TIME)"
	@echo "  Go Version: $(shell go version)"

.PHONY: deps-update
deps-update: ## Update all dependencies
	@echo "$(YELLOW)Updating dependencies...$(NC)"
	@go get -u ./...
	@go mod tidy
	@echo "$(GREEN)✓ Dependencies updated$(NC)"

.PHONY: deps-check
deps-check: ## Check for dependency vulnerabilities
	@echo "$(YELLOW)Checking for dependency vulnerabilities...$(NC)"
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "$(YELLOW)govulncheck not found. Install it with: go install golang.org/x/vuln/cmd/govulncheck@latest$(NC)"; \
	fi
	@echo "$(GREEN)✓ Dependency vulnerability check completed$(NC)"

# Default target
.DEFAULT_GOAL := help