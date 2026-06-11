GOLANGCI_LINT_VERSION ?= v2.12.2
VERSION ?= dev
BUILD_DIR ?= build

.PHONY: lint test build build-multiarch clean

lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run

test:
	go test ./...

build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/undercov ./cmd/undercov

# Build binaries for multiple architectures
build-multiarch: clean
	@mkdir -p $(BUILD_DIR)
	@echo "Building undercov $(VERSION) for multiple architectures..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w -X main.version=$(VERSION)" -o $(BUILD_DIR)/undercov-$(VERSION)-linux-x86_64 ./cmd/undercov
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w -X main.version=$(VERSION)" -o $(BUILD_DIR)/undercov-$(VERSION)-linux-arm64 ./cmd/undercov
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -trimpath -ldflags="-s -w -X main.version=$(VERSION)" -o $(BUILD_DIR)/undercov-$(VERSION)-linux-armv7 ./cmd/undercov
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags="-s -w -X main.version=$(VERSION)" -o $(BUILD_DIR)/undercov-$(VERSION)-windows-x86_64.exe ./cmd/undercov
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -trimpath -ldflags="-s -w -X main.version=$(VERSION)" -o $(BUILD_DIR)/undercov-$(VERSION)-windows-arm64.exe ./cmd/undercov
	chmod +x $(BUILD_DIR)/undercov-*
	@printf 'Build complete:\n'
	@ls -1 $(BUILD_DIR)/undercov-*

clean:
	rm -rf $(BUILD_DIR)
