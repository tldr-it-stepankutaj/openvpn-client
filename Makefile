.PHONY: all build clean install test lint fmt vet

# Build variables
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

# Go variables
GO := go
GOFLAGS := -trimpath
LDFLAGS := -s -w \
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.buildTime=$(BUILD_TIME)

# Output directories
BUILD_DIR := build
INSTALL_DIR := /usr/local/bin

# Binary names
BINARIES := openvpn-login openvpn-connect openvpn-disconnect openvpn-firewall

# Default target
all: build

# Build all binaries
build: $(BINARIES)

openvpn-login:
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$@ ./cmd/login

openvpn-connect:
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$@ ./cmd/connect

openvpn-disconnect:
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$@ ./cmd/disconnect

openvpn-firewall:
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$@ ./cmd/firewall

# Build for Linux (for deployment)
build-linux:
	GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/linux-amd64/openvpn-login ./cmd/login
	GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/linux-amd64/openvpn-connect ./cmd/connect
	GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/linux-amd64/openvpn-disconnect ./cmd/disconnect
	GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/linux-amd64/openvpn-firewall ./cmd/firewall

build-linux-arm64:
	GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/linux-arm64/openvpn-login ./cmd/login
	GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/linux-arm64/openvpn-connect ./cmd/connect
	GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/linux-arm64/openvpn-disconnect ./cmd/disconnect
	GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/linux-arm64/openvpn-firewall ./cmd/firewall

# Install binaries
install: build
	install -m 755 $(BUILD_DIR)/openvpn-login $(INSTALL_DIR)/
	install -m 755 $(BUILD_DIR)/openvpn-connect $(INSTALL_DIR)/
	install -m 755 $(BUILD_DIR)/openvpn-disconnect $(INSTALL_DIR)/
	install -m 755 $(BUILD_DIR)/openvpn-firewall $(INSTALL_DIR)/

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)

# Run tests
test:
	$(GO) test -v -race ./...

# Run tests with coverage
test-coverage:
	$(GO) test -v -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

# Lint code
lint:
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed" && exit 1)
	golangci-lint run ./...

# Format code
fmt:
	$(GO) fmt ./...

# Vet code
vet:
	$(GO) vet ./...

# Tidy dependencies
tidy:
	$(GO) mod tidy

# Download dependencies
deps:
	$(GO) mod download

# Show help
help:
	@echo "Available targets:"
	@echo "  all             - Build all binaries (default)"
	@echo "  build           - Build all binaries"
	@echo "  build-linux     - Build for Linux amd64"
	@echo "  build-linux-arm64 - Build for Linux arm64"
	@echo "  install         - Install binaries to $(INSTALL_DIR)"
	@echo "  clean           - Remove build artifacts"
	@echo "  test            - Run tests"
	@echo "  test-coverage   - Run tests with coverage"
	@echo "  lint            - Run linter"
	@echo "  fmt             - Format code"
	@echo "  vet             - Vet code"
	@echo "  tidy            - Tidy dependencies"
	@echo "  deps            - Download dependencies"
	@echo ""
	@echo "Binaries: $(BINARIES)"
