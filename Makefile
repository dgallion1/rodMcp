# RodMCP Makefile
# Provides build automation and development tools for RodMCP

# Project configuration
PROJECT_NAME := rodmcp
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Go configuration
GO_VERSION_REQUIRED := 1.21.0
LOCAL_GO_DIR := $(HOME)/.local/go
LOCAL_GO_BIN := $(LOCAL_GO_DIR)/bin/go
SYSTEM_GO_BIN := $(shell which go 2>/dev/null)
GO_VERSION_LATEST := $(shell curl -s https://go.dev/VERSION?m=text 2>/dev/null | head -1 | sed 's/go//' || echo "1.24.5")

# Function to check if Go version meets minimum requirement
define check_go_version
$(shell \
  if [ -z "$(1)" ]; then \
    echo "false"; \
  else \
    current=$$(echo "$(1)" | sed 's/go//'); \
    required="$(GO_VERSION_REQUIRED)"; \
    printf '%s\n%s\n' "$$current" "$$required" | sort -V | head -1 | grep -q "$$required" && echo "true" || echo "false"; \
  fi)
endef

# Determine which Go to use
ifeq ($(SYSTEM_GO_BIN),)
    ifeq ($(shell test -x $(LOCAL_GO_BIN) && echo "yes"),yes)
        GO_BIN := $(LOCAL_GO_BIN)
        GO_VERSION := $(shell $(LOCAL_GO_BIN) version | cut -d' ' -f3)
        GO_VERSION_OK := $(call check_go_version,$(GO_VERSION))
    else
        GO_BIN := 
        GO_VERSION := not-installed
        GO_VERSION_OK := false
    endif
else
    GO_BIN := $(SYSTEM_GO_BIN)
    GO_VERSION := $(shell $(SYSTEM_GO_BIN) version | cut -d' ' -f3)
    GO_VERSION_OK := $(call check_go_version,$(GO_VERSION))
endif

# Build configuration
BUILD_DIR := ./bin
BINARY_NAME := rodmcp
MAIN_PACKAGE := ./cmd/server

# Installation paths
LOCAL_BIN := $(HOME)/.local/bin
SYSTEM_BIN := /usr/local/bin

# Go build flags
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildDate=$(BUILD_DATE) -X main.GoVersion=$(GO_VERSION)"
BUILD_FLAGS := -trimpath $(LDFLAGS)

# Go installation URLs (adjust for different OS/arch as needed)
UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)

ifeq ($(UNAME_S),Linux)
    ifeq ($(UNAME_M),x86_64)
        GO_ARCH := amd64
    else ifeq ($(UNAME_M),aarch64)
        GO_ARCH := arm64
    else ifeq ($(UNAME_M),armv7l)
        GO_ARCH := armv6l
    else
        GO_ARCH := $(UNAME_M)
    endif
    GO_OS := linux
else ifeq ($(UNAME_S),Darwin)
    ifeq ($(UNAME_M),x86_64)
        GO_ARCH := amd64
    else ifeq ($(UNAME_M),arm64)
        GO_ARCH := arm64
    else
        GO_ARCH := amd64
    endif
    GO_OS := darwin
else
    GO_OS := linux
    GO_ARCH := amd64
endif

GO_TARBALL := go$(GO_VERSION_LATEST).$(GO_OS)-$(GO_ARCH).tar.gz
GO_URL := https://golang.org/dl/$(GO_TARBALL)

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
MAGENTA := \033[0;35m
CYAN := \033[0;36m
NC := \033[0m # No Color

# Process management functions
define stop_existing_processes
	@echo "$(YELLOW)Stopping existing rodmcp processes...$(NC)"; \
	PIDS=$$(pgrep -f "rodmcp" 2>/dev/null | grep -v $$$$ || true); \
	if [ -n "$$PIDS" ]; then \
		echo "  Found running processes: $$PIDS"; \
		for pid in $$PIDS; do \
			if kill -0 $$pid 2>/dev/null; then \
				CMDLINE=$$(ps -p $$pid -o args= 2>/dev/null || true); \
				if echo "$$CMDLINE" | grep -q "rodmcp" && ! echo "$$CMDLINE" | grep -q "make" && [ "$$pid" != "$$$$" ]; then \
					echo "  Stopping process $$pid: $$CMDLINE"; \
					kill -TERM $$pid 2>/dev/null || true; \
					sleep 1; \
					if kill -0 $$pid 2>/dev/null; then \
						echo "  Force killing process $$pid..."; \
						kill -KILL $$pid 2>/dev/null || true; \
					fi; \
				fi; \
			fi; \
		done; \
		echo "$(GREEN)✓ Existing processes stopped$(NC)"; \
	else \
		echo "  No existing processes found"; \
	fi; \
	rm -f /tmp/rodmcp-http-manager.* 2>/dev/null || true
endef

.PHONY: help build clean test install install-local uninstall demo config-visible config-headless dev check fmt lint vet deps update-deps install-go check-go stop-processes

# Default target
all: build

## Help - Show available targets
help:
	@echo "$(CYAN)RodMCP Build System$(NC)"
	@echo ""
	@echo "$(YELLOW)Build Targets:$(NC)"
	@echo "  $(GREEN)build$(NC)           - Build the binary"
	@echo "  $(GREEN)clean$(NC)           - Clean build artifacts"
	@echo "  $(GREEN)rebuild$(NC)         - Clean and build"
	@echo ""
	@echo "$(YELLOW)Installation:$(NC)"
	@echo "  $(GREEN)install$(NC)         - Install system-wide (requires sudo, stops existing processes)"
	@echo "  $(GREEN)install-local$(NC)   - Install to user bin (recommended, no sudo, stops existing processes)"
	@echo "  $(GREEN)uninstall$(NC)       - Uninstall from system"
	@echo "  $(GREEN)stop-processes$(NC)  - Stop all running rodmcp processes"
	@echo ""
	@echo "$(YELLOW)Configuration:$(NC)"
	@echo "  $(GREEN)config-visible$(NC)  - Configure visible browser mode"
	@echo "  $(GREEN)config-headless$(NC) - Configure headless browser mode"
	@echo ""
	@echo "$(YELLOW)Development:$(NC)"
	@echo "  $(GREEN)test$(NC)            - Run tests"
	@echo "  $(GREEN)demo$(NC)            - Run demo"
	@echo "  $(GREEN)dev$(NC)             - Start development mode"
	@echo "  $(GREEN)check$(NC)           - Run all code quality checks"
	@echo "  $(GREEN)fmt$(NC)             - Format code"
	@echo "  $(GREEN)lint$(NC)            - Run linter"
	@echo "  $(GREEN)vet$(NC)             - Run go vet"
	@echo ""
	@echo "$(YELLOW)Dependencies:$(NC)"
	@echo "  $(GREEN)deps$(NC)            - Download dependencies"
	@echo "  $(GREEN)update-deps$(NC)     - Update dependencies"
	@echo ""
	@echo "$(YELLOW)Go Installation:$(NC)"
	@echo "  $(GREEN)check-go$(NC)        - Check Go installation status"
	@echo "  $(GREEN)install-go$(NC)      - Install Go locally (no sudo)"
	@echo ""
	@echo "$(YELLOW)Version:$(NC) $(VERSION)"

## Check-go - Check Go installation status
check-go:
	@echo "$(CYAN)Go Installation Status:$(NC)"
	@if [ -z "$(GO_BIN)" ]; then \
		echo "  Status:     $(RED)Not installed$(NC)"; \
		echo "  Required:   Go $(GO_VERSION_REQUIRED)+"; \
		echo ""; \
		echo "$(YELLOW)To install Go locally, run: make install-go$(NC)"; \
	else \
		echo "  Status:     $(GREEN)Installed$(NC)"; \
		echo "  Location:   $(GO_BIN)"; \
		echo "  Version:    $(GO_VERSION)"; \
		echo "  Required:   Go $(GO_VERSION_REQUIRED)+"; \
		if [ "$(GO_VERSION_OK)" = "true" ]; then \
			echo "  Compatible: $(GREEN)Yes$(NC)"; \
		else \
			echo "  Compatible: $(RED)No (version too old)$(NC)"; \
		fi; \
	fi

## Install-go - Install Go locally (no sudo required)
install-go:
	@if [ -n "$(SYSTEM_GO_BIN)" ] && [ "$(GO_VERSION_OK)" = "true" ]; then \
		echo "$(GREEN)Go is already installed system-wide at $(SYSTEM_GO_BIN)$(NC)"; \
		echo "$(CYAN)Current version: $(GO_VERSION) (meets requirement)$(NC)"; \
		exit 0; \
	fi
	@if [ -x "$(LOCAL_GO_BIN)" ] && [ "$(GO_VERSION_OK)" = "true" ]; then \
		echo "$(GREEN)Go is already installed locally at $(LOCAL_GO_BIN)$(NC)"; \
		echo "$(CYAN)Current version: $(GO_VERSION) (meets requirement)$(NC)"; \
		exit 0; \
	fi
	@if [ -n "$(GO_BIN)" ] && [ "$(GO_VERSION_OK)" = "false" ]; then \
		echo "$(YELLOW)Existing Go version $(GO_VERSION) is older than required $(GO_VERSION_REQUIRED)$(NC)"; \
		echo "$(BLUE)Installing newer version locally...$(NC)"; \
	fi
	@echo "$(BLUE)Installing Go $(GO_VERSION_LATEST) locally...$(NC)"
	@echo "  Architecture: $(GO_OS)-$(GO_ARCH)"
	@echo "  Download URL: $(GO_URL)"
	@echo "  Install Path: $(LOCAL_GO_DIR)"
	@mkdir -p "$(HOME)/.local"
	@echo "$(CYAN)Downloading Go...$(NC)"
	@cd "$(HOME)/.local" && curl -sL "$(GO_URL)" -o "$(GO_TARBALL)"
	@echo "$(CYAN)Extracting Go...$(NC)"
	@cd "$(HOME)/.local" && tar -xf "$(GO_TARBALL)"
	@rm -f "$(HOME)/.local/$(GO_TARBALL)"
	@echo "$(GREEN)✓ Go installed successfully!$(NC)"
	@echo ""
	@echo "$(CYAN)Add this to your ~/.bashrc or ~/.zshrc:$(NC)"
	@echo "  export PATH=\"$(LOCAL_GO_DIR)/bin:\$$PATH\""
	@echo ""
	@echo "$(CYAN)Or run this now to update your current session:$(NC)"
	@echo "  export PATH=\"$(LOCAL_GO_DIR)/bin:\$$PATH\""

## Build - Build the binary
build: check-go-available
	@echo "$(BLUE)Building $(PROJECT_NAME)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	$(GO_BIN) build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "$(GREEN)✓ Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

# Internal target to check if Go is available and compatible
check-go-available:
	@if [ -z "$(GO_BIN)" ]; then \
		echo "$(RED)✗ Go is not installed$(NC)"; \
		echo ""; \
		echo "$(YELLOW)Install Go using one of these methods:$(NC)"; \
		echo "  1. $(CYAN)make install-go$(NC)     - Install Go locally (no sudo)"; \
		echo "  2. Install Go system-wide from https://golang.org/dl/"; \
		echo ""; \
		exit 1; \
	elif [ "$(GO_VERSION_OK)" = "false" ]; then \
		echo "$(RED)✗ Go version $(GO_VERSION) is too old$(NC)"; \
		echo "$(YELLOW)Minimum required: Go $(GO_VERSION_REQUIRED)$(NC)"; \
		echo ""; \
		echo "$(YELLOW)Update Go using one of these methods:$(NC)"; \
		echo "  1. $(CYAN)make install-go$(NC)     - Install newer Go locally (no sudo)"; \
		echo "  2. Update Go system-wide from https://golang.org/dl/"; \
		echo ""; \
		exit 1; \
	fi

## Clean - Remove build artifacts
clean:
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	rm -rf $(BUILD_DIR)
	@if [ -n "$(GO_BIN)" ]; then \
		$(GO_BIN) clean; \
	fi
	@echo "$(GREEN)✓ Clean complete$(NC)"

## Rebuild - Clean and build
rebuild: clean build

## Test - Run tests
test: check-go-available
	@echo "$(BLUE)Running tests...$(NC)"
	$(GO_BIN) test -v ./...
	@echo "$(GREEN)✓ Tests complete$(NC)"

## Test Comprehensive - Run comprehensive test suite for all MCP tools
test-comprehensive: check-go-available
	@echo "$(BLUE)Running comprehensive MCP test suite...$(NC)"
	@echo "Testing all 18 MCP tools across 5 categories"
	$(GO_BIN) run comprehensive_suite.go
	@echo "$(GREEN)✓ Comprehensive test suite complete$(NC)"

## Install - Install system-wide (requires sudo)
install: build
	@echo "$(BLUE)Installing $(PROJECT_NAME) system-wide...$(NC)"
	$(call stop_existing_processes)
	@if [ ! -w "$(SYSTEM_BIN)" ]; then \
		echo "$(YELLOW)Installing to $(SYSTEM_BIN) requires sudo$(NC)"; \
		sudo cp $(BUILD_DIR)/$(BINARY_NAME) $(SYSTEM_BIN)/$(BINARY_NAME); \
		sudo chmod +x $(SYSTEM_BIN)/$(BINARY_NAME); \
	else \
		cp $(BUILD_DIR)/$(BINARY_NAME) $(SYSTEM_BIN)/$(BINARY_NAME); \
		chmod +x $(SYSTEM_BIN)/$(BINARY_NAME); \
	fi
	@echo "$(GREEN)✓ Installed to $(SYSTEM_BIN)/$(BINARY_NAME)$(NC)"

## Install-local - Install to user bin (no sudo required)
install-local: build
	@echo "$(BLUE)Installing $(PROJECT_NAME) locally...$(NC)"
	$(call stop_existing_processes)
	@mkdir -p $(LOCAL_BIN)
	cp $(BUILD_DIR)/$(BINARY_NAME) $(LOCAL_BIN)/$(BINARY_NAME)
	chmod +x $(LOCAL_BIN)/$(BINARY_NAME)
	@echo "$(GREEN)✓ Installed to $(LOCAL_BIN)/$(BINARY_NAME)$(NC)"
	@echo "$(CYAN)Make sure $(LOCAL_BIN) is in your PATH$(NC)"

## Stop-processes - Stop all running rodmcp processes
stop-processes:
	@echo "$(BLUE)Stopping all rodmcp processes...$(NC)"
	$(call stop_existing_processes)

## Uninstall - Remove installed binary
uninstall:
	@echo "$(YELLOW)Uninstalling $(PROJECT_NAME)...$(NC)"
	@if [ -f "$(SYSTEM_BIN)/$(BINARY_NAME)" ]; then \
		if [ ! -w "$(SYSTEM_BIN)" ]; then \
			sudo rm -f $(SYSTEM_BIN)/$(BINARY_NAME); \
		else \
			rm -f $(SYSTEM_BIN)/$(BINARY_NAME); \
		fi; \
		echo "$(GREEN)✓ Removed from $(SYSTEM_BIN)$(NC)"; \
	fi
	@if [ -f "$(LOCAL_BIN)/$(BINARY_NAME)" ]; then \
		rm -f $(LOCAL_BIN)/$(BINARY_NAME); \
		echo "$(GREEN)✓ Removed from $(LOCAL_BIN)$(NC)"; \
	fi

## Config-visible - Configure visible browser mode
config-visible:
	@echo "$(BLUE)Configuring visible browser mode...$(NC)"
	@if [ -f "./configs/setup-visible-browser-local.sh" ]; then \
		echo "2" | ./configs/setup-visible-browser-local.sh; \
		echo "$(GREEN)✓ Visible browser mode configured$(NC)"; \
		echo "$(CYAN)Restart Claude Code for changes to take effect$(NC)"; \
	else \
		echo "$(RED)✗ Configuration script not found$(NC)"; \
		exit 1; \
	fi

## Config-headless - Configure headless browser mode
config-headless:
	@echo "$(BLUE)Configuring headless browser mode...$(NC)"
	@if [ -f "./configs/setup-headless-browser-local.sh" ]; then \
		echo "2" | ./configs/setup-headless-browser-local.sh; \
		echo "$(GREEN)✓ Headless browser mode configured$(NC)"; \
		echo "$(CYAN)Restart Claude Code for changes to take effect$(NC)"; \
	else \
		echo "$(RED)✗ Configuration script not found$(NC)"; \
		exit 1; \
	fi

## Demo - Run demonstration
demo:
	@echo "$(BLUE)Running demo...$(NC)"
	@if [ -f "./bin/demo" ]; then \
		./bin/demo; \
	else \
		echo "$(RED)✗ Demo script not found$(NC)"; \
		exit 1; \
	fi

## Dev - Start development mode
dev: build
	@echo "$(BLUE)Starting development mode...$(NC)"
	$(BUILD_DIR)/$(BINARY_NAME) --debug=true --log-level=debug

## Fmt - Format code
fmt: check-go-available
	@echo "$(BLUE)Formatting code...$(NC)"
	$(GO_BIN) fmt ./...
	@echo "$(GREEN)✓ Code formatted$(NC)"

## Vet - Run go vet
vet: check-go-available
	@echo "$(BLUE)Running go vet...$(NC)"
	$(GO_BIN) vet ./...
	@echo "$(GREEN)✓ Vet checks passed$(NC)"

## Lint - Run linter (if available)
lint:
	@echo "$(BLUE)Running linter...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
		echo "$(GREEN)✓ Linter checks passed$(NC)"; \
	else \
		echo "$(YELLOW)golangci-lint not found, skipping...$(NC)"; \
	fi

## Check - Run all code quality checks
check: fmt vet lint test
	@echo "$(GREEN)✓ All checks passed$(NC)"

## Deps - Download dependencies
deps: check-go-available
	@echo "$(BLUE)Downloading dependencies...$(NC)"
	$(GO_BIN) mod download
	$(GO_BIN) mod tidy
	@echo "$(GREEN)✓ Dependencies updated$(NC)"

## Update-deps - Update dependencies
update-deps: check-go-available
	@echo "$(BLUE)Updating dependencies...$(NC)"
	$(GO_BIN) get -u ./...
	$(GO_BIN) mod tidy
	@echo "$(GREEN)✓ Dependencies updated$(NC)"

# Development helpers
.PHONY: version status info

## Version - Show version information
version:
	@echo "$(CYAN)$(PROJECT_NAME) Build Information:$(NC)"
	@echo "  Version:    $(VERSION)"
	@echo "  Build Date: $(BUILD_DATE)"
	@echo "  Go Version: $(GO_VERSION)"

## Status - Show project status
status:
	@echo "$(CYAN)Project Status:$(NC)"
	@echo "  Project:    $(PROJECT_NAME)"
	@echo "  Version:    $(VERSION)"
	@echo "  Binary:     $(BUILD_DIR)/$(BINARY_NAME)"
	@if [ -f "$(BUILD_DIR)/$(BINARY_NAME)" ]; then \
		echo "  Built:      $(GREEN)Yes$(NC)"; \
	else \
		echo "  Built:      $(RED)No$(NC)"; \
	fi
	@if [ -f "$(LOCAL_BIN)/$(BINARY_NAME)" ]; then \
		echo "  Installed:  $(GREEN)Yes (local)$(NC)"; \
	elif [ -f "$(SYSTEM_BIN)/$(BINARY_NAME)" ]; then \
		echo "  Installed:  $(GREEN)Yes (system)$(NC)"; \
	else \
		echo "  Installed:  $(RED)No$(NC)"; \
	fi

## Info - Show detailed project information
info: version status
	@echo ""
	@echo "$(CYAN)Build Configuration:$(NC)"
	@echo "  Main Package:  $(MAIN_PACKAGE)"
	@echo "  Build Dir:     $(BUILD_DIR)"
	@echo "  Local Bin:     $(LOCAL_BIN)"
	@echo "  System Bin:    $(SYSTEM_BIN)"
	@echo "  Build Flags:   $(BUILD_FLAGS)"