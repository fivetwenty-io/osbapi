# Colors for terminal output
GREEN  := \033[1;32m
YELLOW := \033[1;33m
BLUE   := \033[1;34m
CYAN   := \033[1;36m
WHITE  := \033[1;37m
RESET  := \033[0m

# Default target - show help
.DEFAULT_GOAL := help

# Variables
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
GIT_SHA := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
VERSION ?= "dev/$(GIT_BRANCH)/$(GIT_SHA)"
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GO_FILES := $(shell find . -name '*.go' -type f -not -path "./vendor/*")

##@ General

.PHONY: help
help: ## Display this help message
	@echo "$(BLUE)osbapi Go Library Makefile$(RESET)"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make $(CYAN)<target>$(RESET)\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  $(CYAN)%-20s$(RESET) %s\n", $$1, $$2 } /^##@/ { printf "\n$(YELLOW)%s$(RESET)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Testing & Quality

.PHONY: test
test: ## Run all tests
	@echo "$(GREEN)Running tests...$(RESET)"
	@go test -v $(shell go list ./... | grep -v vendor)
	@echo "$(GREEN)Tests complete$(RESET)"

.PHONY: test-short
test-short: ## Run tests in short mode
	@echo "$(GREEN)Running short tests...$(RESET)"
	@go test -short $(shell go list ./... | grep -v vendor)
	@echo "$(GREEN)Short tests complete$(RESET)"

.PHONY: test-race
test-race: ## Run tests with race condition detection
	@echo "$(GREEN)Running tests with race detection...$(RESET)"
	@go test -race $(shell go list ./... | grep -v vendor)
	@echo "$(GREEN)Race tests complete$(RESET)"

.PHONY: test-integration
test-integration: ## Run integration tests
	@echo "$(GREEN)Running integration tests...$(RESET)"
	@go test -tags integration -v ./test/integration/...
	@echo "$(GREEN)Integration tests complete$(RESET)"

.PHONY: coverage
coverage: ## Generate test coverage report
	@echo "$(GREEN)Generating coverage report...$(RESET)"
	@go test -coverprofile=coverage.out $(shell go list ./... | grep -v vendor)
	@go tool cover -func=coverage.out
	@echo "$(GREEN)Coverage report generated$(RESET)"

.PHONY: coverage-html
coverage-html: coverage ## Generate and open HTML coverage report
	@echo "$(GREEN)Opening HTML coverage report...$(RESET)"
	@go tool cover -html=coverage.out

.PHONY: test-all
test-all: test coverage ## Run all tests and generate coverage report
	@echo "$(GREEN)All tests and coverage complete$(RESET)"

.PHONY: benchmark
benchmark: ## Run benchmarks
	@echo "$(GREEN)Running benchmarks...$(RESET)"
	@go test -bench=. -benchmem $(shell go list ./... | grep -v vendor)
	@echo "$(GREEN)Benchmarks complete$(RESET)"

##@ Code Quality

.PHONY: fmt
fmt: ## Format all Go source files
	@echo "$(GREEN)Formatting code...$(RESET)"
	@go fmt $(shell go list ./... | grep -v vendor)
	@echo "$(GREEN)Code formatted$(RESET)"

.PHONY: vet
vet: ## Run go vet on all source files
	@echo "$(GREEN)Running go vet...$(RESET)"
	@go vet $(shell go list ./... | grep -v vendor)
	@echo "$(GREEN)Vet analysis complete$(RESET)"

.PHONY: lint
lint: fmt vet ## Run fmt and vet

.PHONY: govulncheck
govulncheck: ## Run vulnerability check on dependencies
	@echo "$(GREEN)Checking for vulnerabilities...$(RESET)"
	@command -v govulncheck >/dev/null 2>&1 || { \
		echo "$(YELLOW)Installing govulncheck...$(RESET)"; \
		go install golang.org/x/vuln/cmd/govulncheck@latest; \
	}
	@govulncheck $(shell go list ./... | grep -v vendor)
	@echo "$(GREEN)Vulnerability check complete$(RESET)"

.PHONY: gosec
gosec: ## Run security scanner on source code
	@echo "$(GREEN)Running security scan (gosec)...$(RESET)"
	@command -v gosec >/dev/null 2>&1 || { \
		echo "$(YELLOW)Installing gosec...$(RESET)"; \
		go install github.com/securego/gosec/v2/cmd/gosec@latest; \
	}
	@gosec -quiet -fmt text ./...
	@echo "$(GREEN)Security scan complete$(RESET)"

.PHONY: staticcheck
staticcheck: ## Run staticcheck static analysis
	@echo "$(GREEN)Running staticcheck...$(RESET)"
	@command -v staticcheck >/dev/null 2>&1 || { \
		echo "$(YELLOW)Installing staticcheck...$(RESET)"; \
		go install honnef.co/go/tools/cmd/staticcheck@latest; \
	}
	@staticcheck $(shell go list ./... | grep -v vendor)
	@echo "$(GREEN)Staticcheck analysis complete$(RESET)"

.PHONY: golangci
golangci: ## Run golangci-lint comprehensive linter
	@echo "$(GREEN)Running golangci-lint...$(RESET)"
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "$(YELLOW)Installing golangci-lint...$(RESET)"; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin; \
	}
	@golangci-lint run ./...
	@echo "$(GREEN)golangci-lint complete$(RESET)"

.PHONY: trivy
trivy: ## Run Trivy container and dependency scanner
	@echo "$(GREEN)Running Trivy scan...$(RESET)"
	@command -v trivy >/dev/null 2>&1 || { \
		echo "$(YELLOW)Trivy not found. Please install it:$(RESET)"; \
		echo "$(CYAN)  brew install trivy$(RESET) (macOS)"; \
		exit 1; \
	}
	@trivy fs --scanners vuln,misconfig,secret --severity HIGH,CRITICAL --skip-dirs vendor .
	@echo "$(GREEN)Trivy scan complete$(RESET)"

.PHONY: security
security: govulncheck gosec trivy ## Run all security scans

.PHONY: check
check: lint vet staticcheck test ## Run basic checks (lint, vet, staticcheck, test)
	@echo "$(GREEN)Basic checks passed$(RESET)"

##@ Documentation

.PHONY: docs
docs: ## Generate Go documentation
	@echo "$(GREEN)Generating documentation...$(RESET)"
	@go doc -all ./pkg/osbapi
	@echo "$(GREEN)Documentation generated$(RESET)"

##@ Examples

.PHONY: examples
examples: ## Build example programs
	@echo "$(GREEN)Building examples...$(RESET)"
	@if [ -d "examples" ]; then \
		for example in examples/*/; do \
			if [ -f "$$example/main.go" ]; then \
				echo "Building $$example..."; \
				go build -o "$$example/example" "$$example"; \
			fi \
		done; \
		echo "$(GREEN)Examples built$(RESET)"; \
	else \
		echo "$(YELLOW)No examples directory found$(RESET)"; \
	fi

##@ Cleanup

.PHONY: clean
clean: ## Clean build artifacts and test cache
	@echo "$(YELLOW)Cleaning up...$(RESET)"
	@rm -f coverage.out coverage.html test.cov
	@rm -rf artifacts/
	@go clean -testcache
	@if [ -d "examples" ]; then \
		find examples -name "example" -type f -delete; \
	fi
	@echo "$(GREEN)Cleanup complete$(RESET)"

##@ Release

.PHONY: tag
tag: ## Create a new version tag (use VERSION=vX.Y.Z)
	@test -n "$(VERSION)" || { echo "ERROR: VERSION not set. Use: make tag VERSION=vX.Y.Z"; exit 1; }
	@echo "$(GREEN)Creating tag $(VERSION)...$(RESET)"
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@echo "$(GREEN)Tag created. Push with: git push origin $(VERSION)$(RESET)"

.PHONY: version
version: ## Display the current version
	@echo "$(CYAN)Version: $(VERSION)$(RESET)"

##@ Dependencies

.PHONY: deps
deps: ## Download and verify dependencies
	@echo "$(GREEN)Downloading dependencies...$(RESET)"
	@go mod download
	@go mod verify
	@echo "$(GREEN)Dependencies ready$(RESET)"

.PHONY: deps-tidy
deps-tidy: ## Clean up go.mod and go.sum
	@echo "$(GREEN)Tidying dependencies...$(RESET)"
	@go mod tidy
	@echo "$(GREEN)Dependencies tidied$(RESET)"

##@ Development

.PHONY: setup
setup: deps ## Initial project setup
	@echo "$(GREEN)Setting up project...$(RESET)"
	@go mod download
	@echo "$(GREEN)Project setup complete$(RESET)"

.PHONY: ci
ci: lint vet test-race coverage security ## Run all CI checks
	@echo "$(GREEN)All CI checks passed$(RESET)"
