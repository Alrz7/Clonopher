GO ?= go
BINARY_NAME ?= clonopher
MAIN_PACKAGE ?= .
BUILD_DIR ?= bin
GOARCH ?= amd64

.PHONY: build build-all build-linux build-windows build-macos clean help

# Build for all supported operating systems.
build: build-all
build-all: build-linux build-windows build-macos

# Build for Linux.
build-linux:
	@mkdir -p "$(BUILD_DIR)"
	GOOS=linux GOARCH=$(GOARCH) $(GO) build -o "$(BUILD_DIR)/$(BINARY_NAME)-linux-$(GOARCH)" "$(MAIN_PACKAGE)"

# Build for Windows.
build-windows:
	@mkdir -p "$(BUILD_DIR)"
	GOOS=windows GOARCH=$(GOARCH) $(GO) build -o "$(BUILD_DIR)/$(BINARY_NAME)-windows-$(GOARCH).exe" "$(MAIN_PACKAGE)"

# Build for macOS.
build-macos:
	@mkdir -p "$(BUILD_DIR)"
	GOOS=darwin GOARCH=$(GOARCH) $(GO) build -o "$(BUILD_DIR)/$(BINARY_NAME)-macos-$(GOARCH)" "$(MAIN_PACKAGE)"

# Remove generated build artifacts.
clean:
	rm -rf "$(BUILD_DIR)"

# Show available targets.
help:
	@printf '%s\n' \
	  'build / build-all  Build for Linux, Windows, and macOS' \
	  'build-linux       Build for Linux' \
	  'build-windows     Build for Windows' \
	  'build-macos       Build for macOS' \
	  'clean             Remove the tmp directory' \
	  'help              Show this help' \
	  '' \
	  'Variables: GOARCH=amd64 (default), MAIN_PACKAGE=., BUILD_DIR=tmp'
